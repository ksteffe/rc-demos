FROM mcr.microsoft.com/devcontainers/go:1-1.25-bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends curl python3 \
    && rm -rf /var/lib/apt/lists/*
