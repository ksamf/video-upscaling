#!/bin/sh
# Выполняем миграцию базы данных
/backend/.venv/bin/alembic upgrade head
# Запускаем сервер приложения
exec /backend/.venv/bin/uvicorn src.main:app --host 0.0.0.0 --port 8000
