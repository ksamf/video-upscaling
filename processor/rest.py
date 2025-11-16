import tempfile
from uuid import UUID

from fastapi import FastAPI, Response
from fastapi.middleware.cors import CORSMiddleware
import uvicorn
from src.video_processing import VideoProcessor
from src.audio_processing import AudioProcessor
from config import settings

app = FastAPI(title="Video Upscaling and Subtitle Service", debug=settings.APP_DEBUG)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/health")
async def health_check():
    return {"status": "ok"}


@app.post("/upscale/{id}")
async def upscale(
    id: UUID,
    file: str,
    real: bool,
):
    """
    Upscales video to x2 resolution and uploads to S3.
    Args:
        id (UUID): Unique identifier folder in S3.
        file (str): file name in S3 without extension.
        real (bool): Whether to use realistic upscaling model.
    """
    with tempfile.TemporaryDirectory() as tmpdir:
        video_processor = VideoProcessor(
            upload_id=id,
            realistic=real,
        )
        await video_processor.upscale_video(id, file, tmpdir)


@app.post("/subtitles/{id}")
async def create_sub(id: UUID) -> Response:
    """
    Creates subtitles original and english languages for the video and uploads to S3.
    Args:
        id (UUID): Unique identifier folder in S3.
    Returns:
        orig_lang (str): language video
    """
    with tempfile.TemporaryDirectory() as tmpdir:
        audio_processor = AudioProcessor(upload_id=id)
        orig_lang = await audio_processor.transcribe_video(tmpdir)
    return Response(orig_lang, media_type="text/plain")


@app.post("/translate/{id}")
async def translate_sub(id: UUID, lang: str):
    """
    Translates english subtitles to the specified language and uploads to S3.
    Args:
        id (UUID): Unique identifier folder in S3.
        lang (str): Language code to translate subtitles to.
    """
    with tempfile.TemporaryDirectory() as tmpdir:
        audio_processor = AudioProcessor(upload_id=id)
        await audio_processor.translate_sub(tmpdir, lang)


if __name__ == "__main__":
    import torch
    import onnxruntime as ort

    print(ort.get_available_providers())
    print(torch.cuda.is_available())
    uvicorn.run(
        "rest:app",
        host=settings.APP_HOST,
        port=settings.APP_PORT,
        reload=settings.APP_DEBUG,
    )
