import asyncio
from fractions import Fraction
import json
import logging
import shutil
from pathlib import Path
from uuid import UUID

from fastapi import HTTPException
from ffmpeg.asyncio import FFmpeg
from sqlalchemy.ext.asyncio import AsyncSession

from src.api import crud
from src.utils.video_processing import VideoProcessor
from src.utils.audio_processing import AudioProcessor
from src.aws.s3_storage import s3_client

logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO)

# Основная директория для загрузок
MAIN_DIR = Path("./uploads").resolve()
MAIN_DIR.mkdir(exist_ok=True, parents=True)

# Настройки качества видео
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


class VideoHandler:
    """
    Обработчик видео:
      - Сохранение загрузки и аудитных задач.
      - NSFW проверка.
      - Метаданные через ffprobe.
      - Загрузка оригинала и транскод.
      - Апскейл и финальные задачи.
    """

    def __init__(
        self,
        upload_id: UUID,
        session: AsyncSession,
        realistic_video: bool = True,
    ) -> None:
        self.upload_id = upload_id
        self.upload_dir = MAIN_DIR / self.upload_id.hex
        self.upload_dir.mkdir(exist_ok=True, parents=True)
        self.session = session

        self.video_processor = VideoProcessor(
            upload_id=self.upload_id,
            upload_dir=str(self.upload_dir),
            realistic_video=realistic_video,
        )
        self.audio_processor = AudioProcessor(
            upload_id=self.upload_id,
            session=self.session,
            upload_dir=str(self.upload_dir),
        )

    async def upload_video_handler(self, tmp_path: Path) -> None:
        orig = self.upload_dir / "original.mp4"

        try:
            await asyncio.to_thread(shutil.copy, tmp_path, orig)
            audio_path = await self.audio_processor.extract_audio(orig)
            await s3_client.upload_file(self.upload_id.hex, str(audio_path))

            meta = await self.get_metadata(orig)
            height = int(meta["height"])
            qualities = QUALITIES.get(height, [])
            await crud.update_quality(
                self.session, id=self.upload_id, qualities=qualities
            )

            await self.check_nsfw(orig, meta)

            sub_vtt = await self.audio_processor.transcribe_video(str(orig))
            await s3_client.upload_file(self.upload_id.hex, str(sub_vtt))

            max_q = qualities[-1] if qualities else height
            upscale_out = self.upload_dir / f"{max_q}.mp4"
            await self.video_processor.upscale_video(str(tmp_path), str(upscale_out))
            await s3_client.upload_file(self.upload_id.hex, str(upscale_out))

            for q in VIDEO_SETTINGS:
                if q["height"] < max_q:
                    await self.video_processor.transcode_video(
                        str(upscale_out), q["height"], q["crf"]
                    )

        except Exception as e:
            logger.error("Ошибка при обработке видео: %s", e)
            raise
        finally:
            # Удаляем только если файл существует и точно закрыт
            if tmp_path and tmp_path.exists():
                try:
                    tmp_path.unlink()
                except Exception as e:
                    logger.warning("Не удалось удалить временный файл: %s", e)

    async def get_metadata(self, file: Path) -> dict:
        try:
            probe = FFmpeg(executable="ffprobe").input(
                str(file), print_format="json", show_streams=None
            )
            out = await probe.execute()
            media = json.loads(out)
            stream = next(s for s in media["streams"] if s.get("codec_type") == "video")
            logger.info("Video resolution: %sp", stream["height"])
            return stream
        except Exception as e:
            logger.error("Metadata error: %s", e)
            raise HTTPException(400, f"Metadata error: {e}")

    async def check_nsfw(self, file: Path, meta: dict) -> None:
        ratio = meta.get("avg_frame_rate") or meta.get("r_frame_rate", "0/1")
        fps = float(Fraction(ratio))
        timestamps = await self.video_processor.process_video_nsfw(str(file), fps=fps)
        duration = float(meta.get("duration", 0))
        is_nsfw = len(timestamps) > duration * 0.1
        await crud.update_nsfw(self.session, id=self.upload_id, nsfw=is_nsfw)
        logger.info("NSFW: %s", is_nsfw)

    async def get_sub(self, lang: str) -> str:
        info = await crud.select_video(self.session, id=self.upload_id)
        sub_file = self.upload_dir / f"{lang}_subtitles.vtt"
        if not await s3_client.file_exists(f"{self.upload_id.hex}/{sub_file.name}"):
            translated = await self.audio_processor.translate_sub(info.language, lang)
            await s3_client.upload_file(self.upload_id.hex, translated)
        return f"{info.video_path}/{sub_file.name}"

    async def get_dub(self, to_code: str) -> str:
        info = await crud.select_video(self.session, id=self.upload_id)
        dub_file = self.upload_dir / f"{to_code}_audio.mp3"
        if not await s3_client.file_exists(f"{self.upload_id.hex}/{dub_file.name}"):
            generated = await self.audio_processor.dub_video(info.language, to_code)
            await s3_client.upload_file(self.upload_id.hex, generated)
        return f"{info.video_path}/{dub_file.name}"
