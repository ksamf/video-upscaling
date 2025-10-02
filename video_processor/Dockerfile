FROM ghcr.io/astral-sh/uv:python3.12-bookworm-slim AS builder

WORKDIR /video_processor
ENV UV_COMPILE_BYTECODE=1 UV_LINK_MODE=copy UV_TOOL_BIN_DIR=/usr/local/bin

RUN --mount=type=cache,target=/root/.cache/uv \
    --mount=type=bind,source=uv.lock,target=uv.lock \
    --mount=type=bind,source=pyproject.toml,target=pyproject.toml \
    uv sync --locked --no-install-project --no-dev

COPY . /video_processor

RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --locked --no-dev

FROM nvidia/cuda:12.8.1-cudnn-runtime-ubuntu24.04

ENV PATH="/video_processor/.venv/bin:/root/.local/bin:${PATH}" \
    LD_LIBRARY_PATH="${LD_LIBRARY_PATH}:/usr/local/cuda/lib64" \
    UV_COMPILE_BYTECODE=1

WORKDIR /video_processor

RUN apt-get update -qy && \
    apt-get install -qy --no-install-recommends \
        python3.12 python3.12-venv \
        ffmpeg \
        ca-certificates \
        curl && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /video_processor/.venv /video_processor/.venv
COPY --from=builder /video_processor /video_processor

EXPOSE 8080

CMD ["/video_processor/.venv/bin/uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8080"]
