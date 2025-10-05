import logging
from pathlib import Path
from typing import Optional
from uuid import UUID

import torch
from TTS.api import TTS
import webvtt
import argostranslate.package
import argostranslate.translate
import whisperx
from whisperx.SubtitlesProcessor import SubtitlesProcessor
import src.db as db
import asyncio
from src.s3_storage import s3_client


logger = logging.getLogger(__name__)


class AudioProcessor:
    """
    Asynchronous audio pipeline:
      1. transcribe_video() – Extract speech and generate subtitles
      2. translate_sub() – Translate subtitles to another language
      3. dub_video() – Generate TTS audio track
    """

    def __init__(self, upload_id: UUID):
        self.upload_id = upload_id
        self.device = "cuda" if torch.cuda.is_available() else "cpu"
        self._tts: Optional[TTS] = None
        self._whisper_model = None
        self._align_model = None
        self._align_meta = None
        self.language: Optional[str] = None

    async def transcribe_video(self, tmpdir: Path):
        """
        Transcribe audio using WhisperX and generate subtitles (VTT).
        """
        audio_path = Path(tmpdir) / "audio.mp3"
        await s3_client.get_file(f"{self.upload_id}/audio.mp3", audio_path)

        if self._whisper_model is None:
            self._whisper_model = whisperx.load_model(
                "medium", device=self.device, compute_type="float16"
            )

        audio = await asyncio.to_thread(whisperx.load_audio, str(audio_path))
        result = await asyncio.to_thread(
            self._whisper_model.transcribe, audio, batch_size=16
        )

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

        vtt = SubtitlesProcessor(
            aligned["segments"], self._align_meta["language"], is_vtt=True
        )
        sub_path = Path(tmpdir) / f"{self._align_meta['language']}_subtitles.vtt"
        vtt.save(str(sub_path), advanced_splitting=True)
        self.language = self._align_meta["language"]

        await db.update_lang(self.upload_id, self.language)
        await s3_client.upload_file(str(self.upload_id), str(sub_path))

    async def translate_sub(self, from_code: str, to_code: str) -> Path:
        """
        Translate subtitles from one language to another using ArgosTranslate.
        """
        input_vtt = Path(f"{from_code}_subtitles.vtt")
        output_vtt = Path(f"{to_code}_subtitles.vtt")

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
                translation = await asyncio.to_thread(
                    argostranslate.translate.translate, text, from_code, to_code
                )
            else:
                translation = ""
            cap.text = translation
            translated.captions.append(cap)

        translated.save(str(output_vtt))
        return output_vtt

    async def dub_video(self, to_code: str, subtitles_vtt: Path) -> Path:
        """
        Generate TTS audio for given subtitles in target language.
        """
        tts_path = Path(f"{to_code}_audio.mp3")

        if self._tts is None:
            self._tts = TTS(
                model_name="tts_models/multilingual/multi-dataset/xtts_v2",
                progress_bar=False,
                gpu=(self.device == "cuda"),
            )

        captions = webvtt.read(str(subtitles_vtt))
        text = "\n".join(cap.text for cap in captions if cap.text.strip())

        await asyncio.to_thread(
            self._tts.tts_to_file,
            text=text,
            speaker_wav="audio.mp3",
            language=to_code,
            file_path=str(tts_path),
        )
        return tts_path
