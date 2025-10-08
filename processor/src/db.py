from uuid import UUID
import asyncpg
from config import settings


async def select_video(id: UUID):
    conn = await asyncpg.connect(settings.DB_URL)
    try:
        query = "SELECT * FROM videos WHERE video_id=$1"
        return await conn.fetchrow(query, id)
    finally:
        await conn.close()


async def update_lang(id: UUID, lang: str) -> None:
    conn = await asyncpg.connect(settings.DB_URL)
    row = await conn.fetchrow(
        "SELECT language_id FROM languages WHERE language=$1", lang
    )

    if row:
        language_id = row["language_id"]
    else:
        row = await conn.fetchrow(
            "INSERT INTO languages (language) VALUES ($1) RETURNING language_id",
            lang,
        )
        language_id = row["language_id"]

    await conn.execute(
        "UPDATE videos SET language_id=$1 WHERE video_id=$2", language_id, id
    )
