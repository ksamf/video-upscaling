from src.databases.base import Base
from sqlalchemy.orm import Mapped
from sqlalchemy.orm import mapped_column
from sqlalchemy import ARRAY, Integer


class Video(Base):
    __tablename__ = "videos"
    name: Mapped[str]
    video_path: Mapped[str]
    language: Mapped[str]
    nsfw: Mapped[bool]
    qualities: Mapped[list[int]] = mapped_column(ARRAY(Integer), nullable=True)
