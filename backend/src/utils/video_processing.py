import asyncio
import hashlib
import logging
import os
import time
from pathlib import Path
from typing import List, Tuple, Optional
from uuid import UUID

import cv2
import numpy as np
import onnxruntime as ort
import torch
from ffmpeg import FFmpegInvalidCommand
from ffmpeg.asyncio import FFmpeg
from PIL import Image
from src.utils.video_pool import task_queue
from transformers import AutoModelForImageClassification, ViTImageProcessor
from src.aws.s3_storage import s3_client

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class VideoProcessor:
    """
    Класс для обработки видео, полностью неблокирующий main event loop FastAPI.
      • Транскодирование
      • Апскейл через ONNX в to_thread
      • NSFW проверка в thread pool
    """

    def __init__(self, upload_id: UUID, upload_dir: str, realistic_video: bool) -> None:
        self.upload_id = upload_id
        self.upload_dir = Path(upload_dir)
        self.device = "cuda" if torch.cuda.is_available() else "cpu"
        self.upscale_factor = 2
        self.upscale_model = Path(
            "backend/src/models/RealESRGAN_x2.onnx"
            if realistic_video
            else "backend/src/models/2xHigurashi_v1_compact_270k.onnx"
        )
        logger.info("Используемая модель: %s", self.upscale_model)

        self.nsfw_model = AutoModelForImageClassification.from_pretrained(
            "Falconsai/nsfw_image_detection"
        ).to(self.device)
        self.nsfw_processor = ViTImageProcessor.from_pretrained(
            "Falconsai/nsfw_image_detection"
        )

        self.frames_dir = self.upload_dir / "frames"
        self.processed_dir = self.upload_dir / "processed"
        self.frames_dir.mkdir(exist_ok=True, parents=True)
        self.processed_dir.mkdir(exist_ok=True, parents=True)

        options = ort.SessionOptions()
        options.intra_op_num_threads = os.cpu_count()
        options.execution_mode = ort.ExecutionMode.ORT_PARALLEL
        options.inter_op_num_threads = 1
        options.graph_optimization_level = ort.GraphOptimizationLevel.ORT_ENABLE_ALL
        options.enable_mem_pattern = True
        options.enable_cpu_mem_arena = True
        options.log_severity_level = 3
        providers = [
            "CUDAExecutionProvider",
            "CPUExecutionProvider",
        ]
        self.ort_session = ort.InferenceSession(
            str(self.upscale_model), sess_options=options, providers=providers
        )
        self.input_name = self.ort_session.get_inputs()[0].name

    async def transcode_video(
        self, input_file: str, target_height: int, crf: int = 23
    ) -> None:
        await task_queue.add_task(self.upload_id, "transcode")
        output_file = self.upload_dir / f"{target_height}.mp4"
        try:
            ffmpeg = (
                FFmpeg()
                .option("y")
                .input(str(input_file))
                .output(
                    str(output_file),
                    vcodec="libx264",
                    vf=f"scale=-2:{target_height}",
                    preset="slow",
                    crf=str(crf),
                    threads=str(os.cpu_count()),
                )
            )

            logger.info("Начало транскодирования видео в %s", output_file)
            await ffmpeg.execute()
            logger.info("Видео перекодировано: %s", output_file)
            await s3_client.upload_file(
                folder_name=self.upload_id.hex, file_path=str(output_file)
            )
        except FFmpegInvalidCommand as e:
            error_msg = e.stderr.decode() if e.stderr else str(e)
            logger.error("Ошибка при транскодировании: %s", error_msg)
            raise

    @staticmethod
    def _compute_hash(frame: np.ndarray, factor: int) -> str:
        small = frame[::factor, ::factor]
        small = np.ascontiguousarray(small)
        return hashlib.md5(small.tobytes()).hexdigest()

    @staticmethod
    def _run_inference(
        session: ort.InferenceSession, name: str, inp: np.ndarray
    ) -> np.ndarray:
        return session.run(None, {name: inp})[0]

    def preprocess_batch(
        self, frames: List[np.ndarray], target_size: Tuple[int, int]
    ) -> np.ndarray:
        """Resize and normalize batch to CHW float32."""
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
        """Convert CHW float output to list of HWC uint8 frames."""
        outs = outputs.transpose(0, 2, 3, 1)
        outs = np.rint(outs * 255.0).clip(0, 255).astype(np.uint8)
        return list(outs)

    async def upscale_video(
        self, input_video: str, output_path: str, batch_size: int = 4
    ) -> None:
        start = time.time()
        cap = cv2.VideoCapture(str(input_video), cv2.CAP_FFMPEG)
        cap.set(cv2.CAP_PROP_HW_ACCELERATION, cv2.VIDEO_ACCELERATION_ANY)
        cap.set(cv2.CAP_PROP_HW_DEVICE, 0)
        if not await asyncio.to_thread(cap.isOpened):
            raise IOError(f"Не удалось открыть видео {input_video}")

        width = int(await asyncio.to_thread(cap.get, cv2.CAP_PROP_FRAME_WIDTH))
        height = int(await asyncio.to_thread(cap.get, cv2.CAP_PROP_FRAME_HEIGHT))
        fps = await asyncio.to_thread(cap.get, cv2.CAP_PROP_FPS)
        new_size = (width * self.upscale_factor, height * self.upscale_factor)

        tmp_out = Path(output_path).with_suffix(".tmp.mp4")
        fourcc = cv2.VideoWriter_fourcc(*"mp4v")
        writer = await asyncio.to_thread(
            cv2.VideoWriter, str(tmp_out), fourcc, fps, new_size
        )

        prev_hash: Optional[str] = None
        prev_frame: Optional[np.ndarray] = None
        batch: List[np.ndarray] = []
        hashes: List[str] = []

        while True:
            ret, frame = await asyncio.to_thread(cap.read)
            if not ret:
                break
            h = await asyncio.to_thread(
                self._compute_hash, frame, self.upscale_factor * 10
            )
            batch.append(frame)
            hashes.append(h)

            if len(batch) == batch_size:
                unique_idx = [i for i, hx in enumerate(hashes) if hx != prev_hash]
                outs: List[np.ndarray] = []
                if unique_idx:
                    inputs = await asyncio.to_thread(
                        self.preprocess_batch,
                        [batch[i] for i in unique_idx],
                        (height, width),
                    )
                    preds = await asyncio.to_thread(
                        self._run_inference, self.ort_session, self.input_name, inputs
                    )
                    outs = await asyncio.to_thread(self.postprocess_batch, preds)
                    prev_hash = hashes[unique_idx[-1]]
                    prev_frame = outs[-1]

                outputs = [
                    prev_frame if idx not in unique_idx else None
                    for idx in range(len(batch))
                ]
                for idx_u, img in zip(unique_idx, outs):
                    outputs[idx_u] = img

                for img in outputs:
                    await asyncio.to_thread(writer.write, img)
                batch.clear()
                hashes.clear()

        if batch:
            inputs = await asyncio.to_thread(
                self.preprocess_batch, batch, (height, width)
            )
            preds = await asyncio.to_thread(
                self._run_inference, self.ort_session, self.input_name, inputs
            )
            results = await asyncio.to_thread(self.postprocess_batch, preds)
            for img in results:
                await asyncio.to_thread(writer.write, img)

        await asyncio.to_thread(writer.release)
        await asyncio.to_thread(cap.release)

        merge = (
            FFmpeg()
            .option("y")
            .input(str(tmp_out))
            .input(str(self.upload_dir / "audio.mp3"))
            .output(
                output_path,
                vcodec="copy",
                acodec="copy",
            )
        )
        await merge.execute()
        os.remove(tmp_out)
        logger.info(f"Upscaled saved: {output_path} in {time.time() - start:.2f}s")

    async def extract_frames(self, input_video: str, fps: float = 30.0) -> None:
        self.frames_dir.mkdir(exist_ok=True)
        pattern = str(self.frames_dir / "frame_%05d.png")
        await FFmpeg().input(input_video).output(pattern, vf=f"fps={fps}").execute()

    def process_frame_nsfw(self, path: str, out_path: str) -> Tuple[int, str]:
        img = Image.open(path)
        label = self.classify_frame(img)
        img.save(out_path)
        return (int(Path(path).stem.split("_")[-1]), label)

    def classify_frame(self, image: Image.Image) -> str:
        image = image.convert("RGB")
        inputs = self.nsfw_processor(images=image, return_tensors="pt")
        if torch.cuda.is_available():
            inputs = {k: v.to("cuda") for k, v in inputs.items()}
        with torch.no_grad():
            logits = self.nsfw_model(**inputs).logits
        return self.nsfw_model.config.id2label[int(logits.argmax(-1))]

    async def process_video_nsfw(
        self, input_video: str, fps: float = 30.0
    ) -> List[float]:
        start = time.time()
        await self.extract_frames(input_video, fps)
        files = sorted(self.frames_dir.glob("*.png"))
        sample = [f for i, f in enumerate(files) if i % int(fps) == 0]
        tasks = [
            asyncio.to_thread(
                self.process_frame_nsfw, str(f), str(self.processed_dir / f.name)
            )
            for f in sample
        ]
        results = await asyncio.gather(*tasks)
        times = [idx / int(fps) for idx, label in results if label == "nsfw"]
        logger.info(f"NSFW done in {time.time() - start:.2f}s")
        return times
