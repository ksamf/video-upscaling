FROM ghcr.io/astral-sh/uv:python3.12-bookworm-slim AS builder

WORKDIR /processor

COPY pyproject.toml uv.lock ./

RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --locked --no-dev

COPY . .

FROM nvidia/cuda:12.8.1-base-ubuntu24.04 AS runtime

WORKDIR /processor

RUN apt-get update -q && \
    DEBIAN_FRONTEND=noninteractive apt-get install -qy --no-install-recommends \
        python3.12 \
        python3.12-venv \
        python3-pip \
        ffmpeg \
        ca-certificates \
        libsndfile1 \
        nvidia-cudnn \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=builder /processor /processor

ENV PATH="/processor/.venv/bin:/usr/local/bin:${PATH}" \
    PYTHONUNBUFFERED=1 \
    LD_LIBRARY_PATH="${LD_LIBRARY_PATH}:/usr/local/cuda/lib64"


EXPOSE 8080

CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8080"]
