import tempfile
from uuid import UUID

import uvicorn
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from src.video_processing import VideoProcessor
from src.audio_processing import AudioProcessor

app = FastAPI(root_path="/api")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/upscale")
async def upscale(
    id: UUID,
    file: str,
    real: bool = True,
):
    with tempfile.TemporaryDirectory() as tmpdir:
        video_processor = VideoProcessor(
            upload_id=id,
            realistic_video=real,
        )
        await video_processor.upscale_video(id, file, tmpdir)


@app.get("/subtitles")
async def create_sub(id: UUID):
    with tempfile.TemporaryDirectory() as tmpdir:
        audio_processor = AudioProcessor(upload_id=id)
        await audio_processor.transcribe_video(tmpdir)


# if __name__ == "__main__":
#     uvicorn.run("main:app", host="0.0.0.0", port=8080)
