from src.databases.db_helper import db_helper
from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession

SessionDep: AsyncSession = Depends(db_helper.scoped_session_dependency)
