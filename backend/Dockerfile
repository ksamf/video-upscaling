# Базовый образ с поддержкой CUDA и cuDNN
FROM nvidia/cuda:12.8.1-cudnn-devel-ubuntu24.04

# Установим переменные окружения для uv
ENV UV_COMPILE_BYTECODE=1 \
    UV_LINK_MODE=copy

# Объединяем установку зависимостей в одну RUN-команду
RUN apt-get update -qy && \
    apt-get install -qy --no-install-recommends \
    ffmpeg \
    python3 \
    python3-pip \
    python3-dev \
    build-essential \
    git \
    curl \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Скачиваем и устанавливаем uv
ADD https://astral.sh/uv/install.sh /uv-installer.sh
RUN sh /uv-installer.sh && rm /uv-installer.sh

# Обновляем PATH, чтобы включить uv, CUDA и виртуальное окружение (если нужно)
ENV PATH="/root/.local/bin:/usr/local/cuda/bin:/app/.venv/bin:${PATH}"

# Синхронизируем зависимости uv
RUN --mount=type=cache,target=/root/.cache/uv \
    --mount=type=bind,source=uv.lock,target=uv.lock \
    --mount=type=bind,source=pyproject.toml,target=pyproject.toml \
    uv sync --frozen --no-install-project --no-dev

# Устанавливаем рабочую директорию
WORKDIR /backend

# Копируем весь проект
COPY . /backend/

# Повторная синхронизация после копирования исходного кода (при необходимости)
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-dev

# Даем права на выполнение entrypoint-скрипта
RUN chmod +x /backend/entrypoint.sh

# Настройка переменных для CUDA (если требуются)
ENV LD_LIBRARY_PATH=${LD_LIBRARY_PATH}:/usr/local/cuda/lib64

EXPOSE 8000

# Запускаем приложение
CMD ["./entrypoint.sh"]