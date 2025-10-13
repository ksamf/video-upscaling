import asyncio
import logging
import os
from pathlib import Path
from typing import Optional, List
from uuid import UUID

import torch
import webvtt
import whisperx
from whisperx.SubtitlesProcessor import SubtitlesProcessor
from transformers import MarianMTModel, MarianTokenizer

from src.s3_storage import s3_client
from config import settings

logger = logging.getLogger(__name__)


class AudioProcessor:
    """
    Asynchronous audio processing pipeline.

    Steps:
        1. transcribe_video() - Transcribe video to subtitles (VTT) using WhisperX.
        2. translate_sub() - Translate subtitles to target language.
        3. dub_video() - Generate TTS audio track for translated subtitles.
    """

    def __init__(self, upload_id: UUID):
        if not os.path.exists(settings.TRANSCRIBE_PATH):
            os.mkdir(settings.TRANSCRIBE_PATH)

        if not os.path.exists(settings.TRANSLATE_PATH):
            os.mkdir(settings.TRANSLATE_PATH)

        self.upload_id = upload_id
        self.device = "cuda" if torch.cuda.is_available() else "cpu"

        # self._tts: Optional[TTS] = None
        self._whisper_model = None
        self._align_model = None
        self._align_meta = None

        self.language: Optional[str] = None

    async def transcribe_video(self, tmpdir: Path) -> str:
        """
        Transcribes video audio to VTT subtitles using WhisperX and uploads them to S3.
        """
        logger.info(f"[{self.upload_id}] Starting transcription...")
        audio_path = Path(tmpdir) / "audio.mp3"

        await s3_client.get_file(f"{self.upload_id}/audio.mp3", audio_path)

        if self._whisper_model is None:
            self._whisper_model = whisperx.load_model(
                "medium",
                device=self.device,
                compute_type="float16",
                download_root=settings.TRANSCRIBE_PATH,
            )
            logger.info("Loaded WhisperX model")

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

        lang = self._align_meta["language"]
        sub_path = Path(tmpdir) / f"{lang}_sub.vtt"
        vtt = SubtitlesProcessor(aligned["segments"], lang, is_vtt=True)
        vtt.save(str(sub_path), advanced_splitting=True)

        self.language = lang
        if self.language != "en":
            await self.translate_sub(tmpdir, "en")
        await s3_client.upload_file(str(self.upload_id), str(sub_path))
        logger.info(f"[{self.upload_id}] Transcription complete: {lang}")

        return lang

    def _load_mt_model(self, src_lang: str, tgt_lang: str):
        model_name = f"Helsinki-NLP/opus-mt-{src_lang}-{tgt_lang}"
        tokenizer = MarianTokenizer.from_pretrained(
            model_name, cache_dir=settings.TRANSLATE_PATH
        )
        model = MarianMTModel.from_pretrained(
            model_name, cache_dir=settings.TRANSLATE_PATH
        )
        model = model.to(self.device)
        return tokenizer, model

    @staticmethod
    async def _translate_batch(
        lines: List[str], tokenizer, model, batch_size: int = 8
    ) -> List[str]:
        """Translate a batch of subtitle lines asynchronously."""
        device = next(model.parameters()).device
        results = []
        for i in range(0, len(lines), batch_size):
            batch = lines[i : i + batch_size]
            inputs = tokenizer(
                batch, return_tensors="pt", padding=True, truncation=True
            )
            inputs = {k: v.to(device) for k, v in inputs.items()}
            outputs = model.generate(**inputs)
            batch_translations = [
                tokenizer.decode(t, skip_special_tokens=True) for t in outputs
            ]
            results.extend(batch_translations)
        return results

    async def translate_sub(self, tmpdir: Path, tgt_lang: str) -> None:
        """
        Translate subtitles (VTT) from en to target language and upload to S3.
        """
        logger.info(f"[{self.upload_id}] Translating subtitles en → {tgt_lang}")

        if os.path.exists(Path(tmpdir) / f"{self.language}_sub.vtt"):
            tokenizer, model = self._load_mt_model(self.language, tgt_lang)
            input_vtt = Path(tmpdir) / f"{self.language}_sub.vtt"
        else:
            tokenizer, model = self._load_mt_model("en", tgt_lang)
            input_vtt = Path(tmpdir) / "en_sub.vtt"
            await s3_client.get_file(f"{self.upload_id}/en_sub.vtt", input_vtt)
        output_vtt = Path(tmpdir) / f"{tgt_lang}_sub.vtt"
        captions = list(webvtt.read(input_vtt))
        texts = [cap.text for cap in captions]

        translated_texts = await self._translate_batch(texts, tokenizer, model)

        for cap, new_text in zip(captions, translated_texts):
            cap.text = new_text.strip().replace("\n", " ")

        webvtt.WebVTT(captions=captions).save(output_vtt)
        await s3_client.upload_file(str(self.upload_id), str(output_vtt))

        logger.info(f"[{self.upload_id}] Translation complete → {tgt_lang}")

    # async def dub_video(self, lang_code: str, subtitles_vtt: Path) -> Path:
    #     """
    #     Generate a TTS audio track for the given translated subtitles.
    #     """
    #     logger.info(f"[{self.upload_id}] Generating TTS dub → {lang_code}")
    #     tts_path = Path(f"{lang_code}_audio.mp3")

    #     if self._tts is None:
    #         self._tts = TTS(
    #             model_name="tts_models/multilingual/multi-dataset/xtts_v2",
    #             progress_bar=False,
    #             gpu=(self.device == "cuda"),
    #         )
    #         logger.info("Loaded XTTSv2 model")

    #     captions = list(webvtt.read(str(subtitles_vtt)))
    #     text = "\n".join(cap.text for cap in captions if cap.text.strip())

    #     await asyncio.to_thread(
    #         self._tts.tts_to_file,
    #         text=text,
    #         speaker_wav="audio.mp3",
    #         language=lang_code,
    #         file_path=str(tts_path),
    #     )

    #     logger.info(f"[{self.upload_id}] TTS generation complete: {tts_path}")
    #     return tts_path
