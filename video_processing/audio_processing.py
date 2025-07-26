import logging
from pathlib import Path
from typing import Optional
from uuid import UUID

import torch
from ffmpeg.asyncio import FFmpeg
from TTS.api import TTS
import webvtt
import argostranslate.package
import argostranslate.translate
import whisperx
from whisperx.SubtitlesProcessor import SubtitlesProcessor
from src.api import crud
from src.utils.video_pool import task_queue
from sqlalchemy.ext.asyncio import AsyncSession
import asyncio

logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO)

# Отключаем TF32 на GPU для максимально точных вычислений
torch.backends.cuda.matmul.allow_tf32 = False
torch.backends.cudnn.allow_tf32 = False


class AudioProcessor:
    """
    Асинхронный аудио‑пайплайн:
      1. extract_audio() – ffmpeg → .flac
      2. transcribe_video() – WhisperX
      3. translate_sub() – ArgosTranslate (пакетно)
      4. dub_video() – TTS
    """

    def __init__(self, upload_id: UUID, upload_dir: str, session: AsyncSession):
        self.upload_id = upload_id
        self.upload_dir = Path(upload_dir)
        self.session = session

        self.device = "cuda" if torch.cuda.is_available() else "cpu"
        self._tts: Optional[TTS] = None
        self._whisper_model = None
        self._align_model = None
        self._align_meta = None
        self.language: Optional[str] = None

    async def extract_audio(self, video_file: Path) -> Path:
        await task_queue.add_task(self.upload_id, "extract_audio")
        out = self.upload_dir / "audio.mp3"
        logger.info("Extracting audio to %s", out)
        cmd = FFmpeg().input(str(video_file)).output(str(out))
        try:
            await cmd.execute()
            logger.info("Audio extracted")
            return out
        except Exception as e:
            logger.error("Failed extract_audio: %s", e)
            raise

    async def transcribe_video(self, audio_path: Path) -> Path:
        await task_queue.add_task(self.upload_id, "transcribe")

        # Ленивая загрузка моделей
        if self._whisper_model is None:
            logger.info("Loading WhisperX transcription model...")
            self._whisper_model = whisperx.load_model(
                "large-v2", device=self.device, compute_type="float16"
            )

        audio = await asyncio.to_thread(whisperx.load_audio, str(audio_path))
        result = await asyncio.to_thread(
            self._whisper_model.transcribe, audio, batch_size=16
        )
        logger.info("Transcription done")

        # Ленивая загрузка alignment
        if self._align_model is None:
            self._align_model, self._align_meta = whisperx.load_align_model(
                language_code=result["language"], device=self.device
            )

        aligned = whisperx.align(
            result["segments"],
            self._align_model,
            self._align_meta,
            audio,
            self.device,
            return_char_alignments=False,
        )

        # Генерация VTT
        vtt = SubtitlesProcessor(
            aligned["segments"], self._align_meta["language"], is_vtt=True
        )
        sub_path = self.upload_dir / f"{self._align_meta['language']}_subtitles.vtt"
        vtt.save(str(sub_path), advanced_splitting=True)
        self.language = self._align_meta["language"]

        # Обновляем язык в БД
        await crud.update_lang(self.session, self.upload_id, self.language)

        logger.info("Subtitles saved: %s", sub_path)
        return sub_path

    async def translate_sub(self, from_code: str, to_code: str) -> Path:
        await task_queue.add_task("translate_sub", self.upload_id)

        input_vtt = self.upload_dir / f"{from_code}_subtitles.vtt"
        output_vtt = self.upload_dir / f"{to_code}_subtitles.vtt"
        logger.info("Translating %s → %s", from_code, to_code)

        # Обновляем пакеты один раз
        argostranslate.package.update_package_index()
        pkgs = argostranslate.package.get_available_packages()
        pkg = next(
            (p for p in pkgs if p.from_code == from_code and p.to_code == to_code), None
        )
        if not pkg:
            raise RuntimeError(f"No ArgosTranslate package for {from_code}->{to_code}")
        pkg_path = pkg.download()
        argostranslate.package.install_from_path(pkg_path)

        vtt_reader = webvtt.read(str(input_vtt))
        translated = webvtt.WebVTT()
        for cap in vtt_reader:
            text = cap.text.strip()
            if text:
                # CPU-bound перевод оборачиваем в to_thread
                translation = await asyncio.to_thread(
                    argostranslate.translate.translate, text, from_code, to_code
                )
            else:
                translation = ""
            cap.text = translation
            translated.captions.append(cap)

        translated.save(str(output_vtt))
        logger.info("Translation done: %s", output_vtt)
        return output_vtt

    async def dub_video(self, to_code: str, subtitles_vtt: Path) -> Path:
        await task_queue.add_task("dub_video", self.upload_id)
        tts_path = self.upload_dir / f"{to_code}_audio.mp3"
        logger.info("Generating TTS dub: %s", tts_path)

        # Ленивая инициализация TTS
        if self._tts is None:
            self._tts = TTS(
                model_name="tts_models/multilingual/multi-dataset/xtts_v2",
                progress_bar=False,
                gpu=(self.device == "cuda"),
            )

        # Собираем текст из VTT
        captions = webvtt.read(str(subtitles_vtt))
        text = "\n".join(cap.text for cap in captions if cap.text.strip())

        try:
            # Запуск TTS в потоке, чтобы не блокировать loop
            await asyncio.to_thread(
                self._tts.tts_to_file,
                text=text,
                speaker_wav=str(self.upload_dir / "audio.mp3"),
                language=to_code,
                file_path=str(tts_path),
            )
            logger.info("Dub saved: %s", tts_path)
            return tts_path
        except Exception as e:
            logger.error("TTS error: %s", e)
            raise
