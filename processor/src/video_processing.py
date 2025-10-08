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

VIDEO_SETTINGS = [
    {"height": h, "crf": crf}
    for h, crf in [
        (17280, 8),
        (8640, 12),
        (4320, 14),
        (2160, 16),
        (1440, 17),
        (1080, 18),
        (720, 20),
        (480, 22),
        (360, 24),
        (240, 26),
    ]
]

QUALITIES = {
    240: [240, 360, 480],
    360: [240, 360, 480, 720],
    480: [240, 360, 480, 720, 1080],
    720: [240, 360, 480, 720, 1080, 1440],
    1080: [240, 360, 480, 720, 1080, 1440, 2160],
    1440: [240, 360, 480, 720, 1080, 1440, 2160, 4320],
    2160: [240, 360, 480, 720, 1080, 1440, 2160, 4320, 8640],
    4320: [240, 360, 480, 720, 1080, 1440, 2160, 4320, 8640, 17280],
}


class VideoProcessor:
    _ort_session: Optional[ort.InferenceSession] = None

    def __init__(self, upload_id: UUID, realistic: bool) -> None:
        """Initialize ONNX session and model path"""
        self.upload_id = upload_id
        self.device = "cuda" if torch.cuda.is_available() else "cpu"
        self.upscale_factor = 2
        self.upscale_model = Path(
            settings.UPSCALE_3D_PATH if realistic else settings.UPSCALE_2D_PATH
        )
        if VideoProcessor._ort_session is None:
            options = ort.SessionOptions()
            options.intra_op_num_threads = os.cpu_count() or 1
            options.execution_mode = ort.ExecutionMode.ORT_PARALLEL
            options.inter_op_num_threads = 1
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
        """Resize frames, normalize and transpose to NCHW"""
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
        """Convert model output tensor back to uint8 BGR frames"""
        outs = outputs.transpose(0, 2, 3, 1)
        return list(np.rint(outs * 255.0).clip(0, 255).astype(np.uint8))

    async def _reader(
        self, cap: cv2.VideoCapture, queue: asyncio.Queue, batch_size: int
    ):
        """Read frames from video and put into queue as batches"""
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
        """Run inference on batches from input queue and put results to output queue"""
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
        """Write processed frames from queue into video file"""
        while True:
            outs = await queue.get()
            if outs is None:
                queue.task_done()
                break
            for img in outs:
                await asyncio.to_thread(writer.write, img)
            queue.task_done()

    def _extract_audio(self, input_video: Path, tmp_audio: Path) -> Optional[Path]:
        """Extract audio track if exists, otherwise return None"""
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
            "-loglevel",
            "error",
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
            "copy",
            "-loglevel",
            "error",
            str(tmp_audio),
        ]
        subprocess.run(cmd, check=True)
        return tmp_audio

    def _merge_streams(
        self, input_video: Path, tmp_audio: Optional[Path], output_path: Path, crf: int
    ):
        """Merge processed video with audio if available"""
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
                "aac",
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
        """Transcode final video to lower resolution variant"""
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
            "aac",
            "-b:a",
            "192k",
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
        """Download, upscale, merge audio and upload result"""
        start = time.time()
        input_video = Path(tmpdir) / "original.mp4"
        await s3_client.get_file(str(id) + "/" + file_name + ".mp4", input_video)

        cap = cv2.VideoCapture(input_video, cv2.CAP_FFMPEG)
        if not cap.isOpened():
            raise IOError(f"Cannot open video {input_video}")

        width = int(cap.get(cv2.CAP_PROP_FRAME_WIDTH))
        height = int(cap.get(cv2.CAP_PROP_FRAME_HEIGHT))
        fps = cap.get(cv2.CAP_PROP_FPS) or 30.0
        new_size = (width * self.upscale_factor, height * self.upscale_factor)

        if height in QUALITIES:
            height_up = str(QUALITIES.get(height)[-1]) + ".mp4"
        else:
            height_up = str(height * self.upscale_factor) + ".mp4"
        output_path = Path(tmpdir) / height_up

        tmp_audio = Path(tmpdir) / "audio.aac"
        self._extract_audio(input_video, tmp_audio)

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

        crf = next(
            (vs["crf"] for vs in VIDEO_SETTINGS if new_size[1] == vs["height"]), 23
        )
        self._merge_streams(input_video, tmp_audio, output_path, crf)

        if height in QUALITIES:
            h = str(QUALITIES.get(height)[-2])
        else:
            h = str(height * self.upscale_factor // 2)
        trascode_path = Path(tmpdir) / (h + ".mp4")
        crf = next((vs["crf"] for vs in VIDEO_SETTINGS if int(h) == vs["height"]), 23)
        self._transcode_quality(output_path, trascode_path, int(h), crf)
        logger.info(f"Upscaling completed in {time.time() - start:.2f} seconds")
        await s3_client.upload_file(str(id), output_path)
        await s3_client.upload_file(str(id), trascode_path)
