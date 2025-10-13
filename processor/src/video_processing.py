import asyncio
import logging
import os
import subprocess
import time
from pathlib import Path
from typing import List, Tuple, Optional
from uuid import UUID

import cv2
import numpy as np
import torch
import onnxruntime as ort
from src.s3_storage import s3_client
from config import settings

logger = logging.getLogger(__name__)

QUALITIES = [144, 240, 360, 480, 720, 1080, 1440, 2160, 4320]


class VideoProcessor:
    _ort_session: Optional[ort.InferenceSession] = None

    def __init__(self, upload_id: UUID, realistic: bool) -> None:
        self.upload_id = upload_id
        self.device = "cuda" if torch.cuda.is_available() else "cpu"
        self.upscale_factor = 2
        self.upscale_model = Path(
            settings.UPSCALE_3D_PATH if realistic else settings.UPSCALE_2D_PATH
        )

        if VideoProcessor._ort_session is None:
            options = ort.SessionOptions()
            options.intra_op_num_threads = os.cpu_count() or 1
            options.inter_op_num_threads = 1
            options.execution_mode = ort.ExecutionMode.ORT_PARALLEL
            options.graph_optimization_level = ort.GraphOptimizationLevel.ORT_ENABLE_ALL

            providers = (
                ["CUDAExecutionProvider", "CPUExecutionProvider"]
                if "CUDAExecutionProvider" in ort.get_available_providers()
                else ["CPUExecutionProvider"]
            )

            VideoProcessor._ort_session = ort.InferenceSession(
                str(self.upscale_model), sess_options=options, providers=providers
            )

        self.ort_session = VideoProcessor._ort_session
        self.input_name = self.ort_session.get_inputs()[0].name

    def preprocess_batch(
        self, frames: List[np.ndarray], target_size: Tuple[int, int]
    ) -> np.ndarray:
        h, w = target_size
        batch = (
            np.stack(
                [
                    cv2.resize(f, (w, h), interpolation=cv2.INTER_LANCZOS4)
                    for f in frames
                ],
                axis=0,
            ).astype(np.float32)
            / 255.0
        )
        return batch.transpose(0, 3, 1, 2)

    def postprocess_batch(self, outputs: np.ndarray) -> List[np.ndarray]:
        outs = outputs.transpose(0, 2, 3, 1)
        return list(np.rint(outs * 255.0).clip(0, 255).astype(np.uint8))

    async def _reader(
        self, cap: cv2.VideoCapture, queue: asyncio.Queue, batch_size: int
    ):
        batch = []
        while True:
            ret, frame = await asyncio.to_thread(cap.read)
            if not ret:
                break
            batch.append(frame)
            if len(batch) == batch_size:
                await queue.put(batch)
                batch = []
        if batch:
            await queue.put(batch)
        await queue.put(None)

    async def _worker(
        self, queue_in: asyncio.Queue, queue_out: asyncio.Queue, size: Tuple[int, int]
    ):
        while True:
            batch = await queue_in.get()
            if batch is None:
                await queue_out.put(None)
                queue_in.task_done()
                break
            inputs = await asyncio.to_thread(self.preprocess_batch, batch, size)
            preds = await asyncio.to_thread(
                self.ort_session.run, None, {self.input_name: inputs}
            )
            outs = await asyncio.to_thread(self.postprocess_batch, preds[0])
            await queue_out.put(outs)
            queue_in.task_done()

    async def _writer(self, writer: cv2.VideoWriter, queue: asyncio.Queue):
        while True:
            outs = await queue.get()
            if outs is None:
                queue.task_done()
                break
            for img in outs:
                await asyncio.to_thread(writer.write, img)
            queue.task_done()

    def _extract_audio(self, input_video: Path, tmp_audio: Path) -> Optional[Path]:
        probe_cmd = [
            "ffprobe",
            "-v",
            "error",
            "-select_streams",
            "a:0",
            "-show_entries",
            "stream=index",
            "-of",
            "csv=p=0",
            str(input_video),
        ]
        result = subprocess.run(
            probe_cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE
        )
        if not result.stdout.strip():
            return None

        cmd = [
            "ffmpeg",
            "-y",
            "-i",
            str(input_video),
            "-vn",
            "-acodec",
            "mp3",
            "-f",
            "mp3",
            "-loglevel",
            "error",
            str(tmp_audio),
        ]
        subprocess.run(cmd, check=True)
        return tmp_audio

    def _merge_streams(
        self, input_video: Path, tmp_audio: Optional[Path], output_path: Path, crf: int
    ):
        if tmp_audio is None:
            cmd = [
                "ffmpeg",
                "-y",
                "-i",
                str(input_video),
                "-c:v",
                "libx264",
                "-preset",
                "medium",
                "-crf",
                str(crf),
                "-an",
                "-loglevel",
                "error",
                str(output_path),
            ]
        else:
            cmd = [
                "ffmpeg",
                "-y",
                "-i",
                str(input_video),
                "-i",
                str(tmp_audio),
                "-c:v",
                "libx264",
                "-preset",
                "medium",
                "-crf",
                str(crf),
                "-c:a",
                "mp3",
                "-b:a",
                "192k",
                "-map",
                "0:v:0",
                "-map",
                "1:a:0?",
                "-loglevel",
                "error",
                str(output_path),
            ]
        subprocess.run(cmd, check=True)

    def _transcode_quality(self, src_path: Path, dst_path: Path, h: int, crf: int):
        cmd = [
            "ffmpeg",
            "-i",
            str(src_path),
            "-map",
            "0:v:0",
            "-c:v",
            "libx264",
            "-crf",
            str(crf),
            "-vf",
            f"scale=-2:{h}",
            "-map",
            "0:a?",
            "-c:a",
            "mp3",
            "-b:a",
            "128k",
            "-fflags",
            "+genpts",
            "-loglevel",
            "error",
            str(dst_path),
        ]
        subprocess.run(cmd, check=True)

    async def upscale_video(
        self, id: UUID, file_name: str, tmpdir: str, batch_size: int = 4
    ):
        start = time.time()
        input_video = Path(tmpdir) / "original.mp4"
        await s3_client.get_file(f"{id}/{file_name}.mp4", input_video)

        cap = cv2.VideoCapture(str(input_video))
        if not cap.isOpened():
            raise IOError(f"Cannot open video {input_video}")

        width = int(cap.get(cv2.CAP_PROP_FRAME_WIDTH))
        height = int(cap.get(cv2.CAP_PROP_FRAME_HEIGHT))
        fps = cap.get(cv2.CAP_PROP_FPS)
        if fps <= 0:
            fps = 30.0

        new_size = (width * self.upscale_factor, height * self.upscale_factor)

        if height in QUALITIES:
            idx = QUALITIES.index(height)
            next_h = QUALITIES[min(idx + 2, len(QUALITIES) - 1)]
        else:
            next_h = height * self.upscale_factor

        output_path = Path(tmpdir) / f"{next_h}.mp4"
        tmp_audio = Path(tmpdir) / "audio.mp3"
        tmp_audio = self._extract_audio(input_video, tmp_audio)

        fourcc = cv2.VideoWriter_fourcc(*"mp4v")
        writer = cv2.VideoWriter(str(output_path), fourcc, fps, new_size)

        queue_frames: asyncio.Queue = asyncio.Queue(maxsize=8)
        queue_results: asyncio.Queue = asyncio.Queue(maxsize=8)

        await asyncio.gather(
            self._reader(cap, queue_frames, batch_size),
            self._worker(queue_frames, queue_results, (height, width)),
            self._writer(writer, queue_results),
        )

        await asyncio.to_thread(writer.release)
        await asyncio.to_thread(cap.release)

        crf = 26 - 2 * QUALITIES.index(height) if height in QUALITIES else 23
        self._merge_streams(input_video, tmp_audio, output_path, crf)

        if height in QUALITIES:
            idx = QUALITIES.index(height)
            lower_h = QUALITIES[min(idx + 1, len(QUALITIES) - 1)]
        else:
            lower_h = next_h // 2

        transcode_path = Path(tmpdir) / f"{lower_h}.mp4"
        crf = 26 - 2 * QUALITIES.index(lower_h) if lower_h in QUALITIES else 23
        self._transcode_quality(output_path, transcode_path, lower_h, crf)

        logger.info(f"Upscaling completed in {time.time() - start:.2f} seconds")

        await s3_client.upload_file(str(id), output_path)
        await s3_client.upload_file(str(id), transcode_path)
