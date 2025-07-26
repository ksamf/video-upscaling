from pathlib import Path
import tempfile
from uuid import UUID, uuid4
from fastapi import (
    APIRouter,
    BackgroundTasks,
    File,
    HTTPException,
    UploadFile,
    status,
    Query,
)
from fastapi_cache.decorator import cache
from sqlalchemy.ext.asyncio import AsyncSession

import src.api.crud as crud
from src.api.dependencies import SessionDep
from src.api.schemas import (
    AllowedQualitiesResponse,
    DeleteVideoResponse,
    DubbingResponse,
    Status,
    SubtitlesResponse,
    VideoInfo,
    VideoUrlResponse,
)
from src.databases.video import Video
from src.aws.s3_storage import s3_client
from src.utils.handler import VideoHandler
from src.utils.video_pool import task_queue

router = APIRouter(prefix="/api/videos", tags=["videos"])


@router.get("/", response_model=list[VideoInfo])
@cache(expire=60)
async def list_videos(
    nsfw: bool | None = Query(None),
    lang: str | None = Query(None),
    session: AsyncSession = SessionDep,
) -> list[VideoInfo]:
    """
    Получить список видео. Опционально фильтровать по NSFW и/или языку.
    """
    if nsfw is not None or lang is not None:
        return await crud.select_videos_filtered(
            session=session, nsfw=nsfw, language=lang
        )
    return await crud.select_all_videos(session=session)


@router.post("/", status_code=status.HTTP_201_CREATED, response_model=Status)
async def upload_video(
    background_tasks: BackgroundTasks,
    realistic_video: bool = True,
    video: UploadFile = File(...),
    session: AsyncSession = SessionDep,
) -> Status:
    """
    Загружает видео и инициирует асинхронную обработку.
    """
    if not video.content_type.startswith("video/"):
        raise HTTPException(400, "Неверный тип файла")
    video_id = uuid4()
    await crud.insert_video(
        session,
        id=video_id,
        name=video.filename,
    )
    handler = VideoHandler(
        upload_id=video_id, session=session, realistic_video=realistic_video
    )
    tmp = tempfile.NamedTemporaryFile(delete=False, suffix=".mp4")
    await video.seek(0)
    while chunk := await video.read(1024 * 1024):
        tmp.write(chunk)
    tmp.close()
    background_tasks.add_task(handler.upload_video_handler, Path(tmp.name))
    return Status(video_id=video_id, status="queued")


@router.get("/{video_id}", response_model=VideoUrlResponse)
@cache(expire=60)
async def get_video_url(
    id: UUID,
    session: AsyncSession = SessionDep,
) -> VideoUrlResponse:
    """
    Возвращает URL папки с  видео.
    """
    video = await crud.select_video(session=session, id=id, fields=[Video.video_path])
    if video is None:
        raise HTTPException(status_code=404, detail="Video not found")
    return VideoUrlResponse(video_path=video.video_path)


@router.delete("/{video_id}", response_model=DeleteVideoResponse)
async def delete_video(
    id: UUID,
    session: AsyncSession = SessionDep,
) -> DeleteVideoResponse:
    """
    Удаляет видео и связанные ресурсы.
    """
    video = await crud.select_video(session=session, id=id)
    if video is None:
        raise HTTPException(status_code=404, detail="Video not found")
    await crud.delete_video(session=session, id=id)
    await s3_client.delete_file(file_path=video.id)
    return DeleteVideoResponse(
        status="success",
        message=f"Video {id} deleted successfully",
    )


@router.get("/{video_id}/status", response_model=dict)
@cache(expire=5)
async def get_status(
    id: UUID,
) -> dict:
    """
    Статус фоновой обработки видео.
    """
    status = task_queue.get_status(id)
    if status is None:
        raise HTTPException(status_code=404, detail="Job not found")
    return status


@router.get("/{video_id}/info", response_model=VideoInfo)
@cache(expire=60)
async def get_video_info(
    id: UUID,
    session: AsyncSession = SessionDep,
) -> VideoInfo:
    """
    Общая информация о видео: путь, NSFW-флаг, язык, доступные качества.
    """
    video = await crud.select_video(
        session=session,
        id=id,
        fields=[
            Video.video_path,
            Video.nsfw,
            Video.language,
            Video.qualities,
        ],
    )
    if video is None:
        raise HTTPException(status_code=404, detail="Video not found")
    return VideoInfo(
        video_path=video.video_path,
        nsfw=video.nsfw,
        language=video.language,
        qualities=video.qualities or [],
    )


@router.get("/{video_id}/subtitles", response_model=SubtitlesResponse)
@cache(expire=60)
async def get_subtitles(
    id: UUID,
    lang: str,
    session: AsyncSession = SessionDep,
) -> SubtitlesResponse:
    """
    Возвращает путь к субтитрам или генерирует их при отсутствии.
    """
    sub_path = await VideoHandler(
        upload_id=id,
        session=session,
    ).get_sub(lang=lang)
    return SubtitlesResponse(subtitles=sub_path)


@router.get("/{video_id}/dubbing", response_model=DubbingResponse)
@cache(expire=60)
async def get_dubbing(
    id: UUID,
    lang: str,
    session: AsyncSession = SessionDep,
) -> DubbingResponse:
    """
    Возвращает путь к дубляжу или генерирует аудио, если отсутствует.
    """
    dub_path = await VideoHandler(
        upload_id=id,
        session=session,
    ).get_dub(to_code=lang)
    return DubbingResponse(dubbing=dub_path)


@router.get("/{video_id}/qualities", response_model=AllowedQualitiesResponse)
@cache(expire=60)
async def get_allowed_qualities(
    id: UUID,
    session: AsyncSession = SessionDep,
) -> AllowedQualitiesResponse:
    """
    Список доступных разрешений видео.
    """
    video = await crud.select_video(session=session, id=id, fields=[Video.qualities])
    if not video:
        raise HTTPException(status_code=404, detail="No qualities found")
    return AllowedQualitiesResponse(qualities=video.qualities)
