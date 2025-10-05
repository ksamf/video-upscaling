FROM ghcr.io/astral-sh/uv:python3.12-bookworm-slim AS builder

WORKDIR /video_processor
ENV UV_COMPILE_BYTECODE=1 UV_LINK_MODE=copy UV_TOOL_BIN_DIR=/usr/local/bin

RUN --mount=type=cache,target=/root/.cache/uv \
    --mount=type=bind,source=uv.lock,target=uv.lock \
    --mount=type=bind,source=pyproject.toml,target=pyproject.toml \
    uv sync --locked --no-install-project --no-dev

COPY . .

RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --locked --no-dev


FROM nvidia/cuda:12.8.1-base-ubuntu24.04

WORKDIR /video_processor

ENV PATH="/video_processor/.venv/bin:/usr/local/bin:${PATH}" \
    LD_LIBRARY_PATH="${LD_LIBRARY_PATH}:/usr/local/cuda/lib64" \
    UV_COMPILE_BYTECODE=1

RUN --mount=type=cache,target=/var/cache/apt \
    --mount=type=cache,target=/var/lib/apt \
    apt-get update -qy && \
    apt-get install -qy --no-install-recommends \
        nvidia-cudnn \
        ffmpeg \
        ca-certificates \
        curl && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/local/bin/uv /usr/local/bin/uv
COPY --from=builder /video_processor/.venv /video_processor/.venv
COPY --from=builder /video_processor /video_processor

EXPOSE 8080

CMD ["uv", "run", "main.py"]
