from uuid import UUID
from sqlalchemy import delete, insert, select, update
from sqlalchemy.ext.asyncio import AsyncSession
from src.databases.video import Video
from src.config import settings


async def insert_video(session: AsyncSession, name: str, id: UUID) -> None:
    stmt = insert(Video).values(
        name=name,
        video_path=f"{settings.ENDPOINT_URL}/{settings.BUCKET_NAME}/{id.hex}",
        language="",
        nsfw=False,
        id=id,
    )
    await session.execute(stmt)
    await session.commit()


async def select_all_videos(session: AsyncSession):
    stmt = select(Video).order_by(Video.video_path.desc())
    res = await session.execute(stmt)
    return res.scalars().all()


async def select_video(session: AsyncSession, id: UUID, fields: list | None):
    stmt = select(*fields or [Video]).where(Video.id == id)
    res = await session.execute(stmt)
    return res.first()


async def update_nsfw(session: AsyncSession, id: UUID, nsfw: bool) -> None:
    stmt = update(Video).where(Video.id == id).values(nsfw=nsfw)
    await session.execute(stmt)
    await session.commit()


async def update_lang(session: AsyncSession, id: UUID, lang: str) -> None:
    stmt = update(Video).where(Video.id == id).values(language=lang)
    await session.execute(stmt)
    await session.commit()


async def delete_video(session: AsyncSession, id: UUID) -> None:
    await session.execute(delete(Video).where(Video.id == id))
    await session.commit()


async def update_quality(session: AsyncSession, id: UUID, qualities: list) -> None:
    stmt = update(Video).where(Video.id == id).values(qualities=qualities)
    await session.execute(stmt)
    await session.commit()


async def select_videos_filtered(
    session: AsyncSession, nsfw: bool, language: str
) -> list:
    conditions = []
    if nsfw is not None:
        conditions.append(Video.nsfw == nsfw)
    if language:
        conditions.append(Video.language == language)
    stmt = select(Video).where(*conditions)
    res = await session.execute(stmt)
    return res.scalars().all()
