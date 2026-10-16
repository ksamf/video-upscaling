from pydantic import BaseModel, ConfigDict
from typing import List, Optional
from uuid import UUID


class VideoBase(BaseModel):
    name: str
    video_path: str


class VideoCreate(VideoBase):
    pass


class Video(VideoBase):
    id: UUID
    language: Optional[str] = ""
    nsfw: bool
    qualities: Optional[List[int]] = None

    model_config = ConfigDict(from_attributes=True)


class VideoInfo(BaseModel):
    video_path: str
    nsfw: bool
    language: str
    qualities: List[int]


class AllowedQualitiesResponse(BaseModel):
    qualities: List[int]


class NSFWResponse(BaseModel):
    nsfw: bool


class DubbingResponse(BaseModel):
    dubbing: str


class SubtitlesResponse(BaseModel):
    subtitles: str


class VideoUrlResponse(BaseModel):
    video_path: str


class DeleteVideoResponse(BaseModel):
    status: str
    message: str


class Status(BaseModel):
    video_id: UUID
    status: str
