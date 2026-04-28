# syntax=docker/dockerfile:1.4

# STEP 1: Build the frontend
FROM node:23-slim as fe-build

ENV NODE_ENV=production
ENV VITE_BUILD_MEMORY_LIMIT=4096
ENV NODE_OPTIONS="--max-old-space-size=4096"

WORKDIR /frontend

# Install build essentials
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    gcc \
    g++ \
    make \
    git

COPY ./backend/pkg/graph/schema.graphqls ../backend/pkg/graph/
COPY frontend/ .

# Install dependencies with package manager detection for SBOM
RUN --mount=type=cache,target=/root/.npm \
    npm ci --include=dev

# Build frontend with optimizations and parallel processing
RUN npm run build -- \
    --mode production \
    --minify esbuild \
    --outDir dist \
    --emptyOutDir \
    --sourcemap false \
    --target es2020

# STEP 2: Build the backend
FROM golang:1.24-bookworm as be-build

# Build arguments for version information
ARG PACKAGE_VER=develop
ARG PACKAGE_REV=

ENV CGO_ENABLED=0
ENV GO111MODULE=on

# Install build essentials
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    gcc \
    g++ \
    make \
    git \
    musl-dev

WORKDIR /backend

COPY backend/ .

# Download dependencies with module detection for SBOM
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Build backend with version information
RUN go build -trimpath \
    -ldflags "\
        -X pentagi/pkg/version.PackageName=monkeyg \
        -X pentagi/pkg/version.PackageVer=${PACKAGE_VER} \
        -X pentagi/pkg/version.PackageRev=${PACKAGE_REV}" \
    -o /pentagi ./cmd/pentagi

# Build ctester utility
RUN go build -trimpath \
    -ldflags "\
        -X pentagi/pkg/version.PackageName=ctester \
        -X pentagi/pkg/version.PackageVer=${PACKAGE_VER} \
        -X pentagi/pkg/version.PackageRev=${PACKAGE_REV}" \
    -o /ctester ./cmd/ctester

# Build ftester utility
RUN go build -trimpath \
    -ldflags "\
        -X pentagi/pkg/version.PackageName=ftester \
        -X pentagi/pkg/version.PackageVer=${PACKAGE_VER} \
        -X pentagi/pkg/version.PackageRev=${PACKAGE_REV}" \
    -o /ftester ./cmd/ftester

# Build etester utility
RUN go build -trimpath \
    -ldflags "\
        -X pentagi/pkg/version.PackageName=etester \
        -X pentagi/pkg/version.PackageVer=${PACKAGE_VER} \
        -X pentagi/pkg/version.PackageRev=${PACKAGE_REV}" \
    -o /etester ./cmd/etester

# STEP 3: Build the final image
FROM alpine:3.23.3

# Create non-root user and docker group with specific GID
RUN addgroup -g 998 docker && \
    addgroup -S monkeyg && \
    adduser -S monkeyg -G monkeyg && \
    addgroup monkeyg docker

# Install required packages
RUN apk --no-cache add ca-certificates openssl openssh-keygen shadow

ADD scripts/entrypoint.sh /opt/monkeyg/bin/

RUN sed -i 's/\r//' /opt/monkeyg/bin/entrypoint.sh && \
    chmod +x /opt/monkeyg/bin/entrypoint.sh

RUN mkdir -p \
    /root/.ollama \
    /opt/monkeyg/bin \
    /opt/monkeyg/ssl \
    /opt/monkeyg/fe \
    /opt/monkeyg/logs \
    /opt/monkeyg/data \
    /opt/monkeyg/conf && \
    chmod 777 /root/.ollama

COPY --from=be-build /pentagi /opt/monkeyg/bin/monkeyg
COPY --from=be-build /ctester /opt/monkeyg/bin/ctester
COPY --from=be-build /ftester /opt/monkeyg/bin/ftester
COPY --from=be-build /etester /opt/monkeyg/bin/etester
COPY --from=fe-build /frontend/dist /opt/monkeyg/fe

# Copy provider configuration files
COPY examples/configs/custom-openai.provider.yml /opt/monkeyg/conf/
COPY examples/configs/deepinfra.provider.yml /opt/monkeyg/conf/
COPY examples/configs/deepseek.provider.yml /opt/monkeyg/conf/
COPY examples/configs/moonshot.provider.yml /opt/monkeyg/conf/
COPY examples/configs/ollama-llama318b-instruct.provider.yml /opt/monkeyg/conf/
COPY examples/configs/ollama-llama318b.provider.yml /opt/monkeyg/conf/
COPY examples/configs/ollama-qwen332b-fp16-tc.provider.yml /opt/monkeyg/conf/
COPY examples/configs/ollama-qwq32b-fp16-tc.provider.yml /opt/monkeyg/conf/
COPY examples/configs/openrouter.provider.yml /opt/monkeyg/conf/
COPY examples/configs/novita.provider.yml /opt/monkeyg/conf/
COPY examples/configs/vllm-qwen3.5-27b-fp8.provider.yml /opt/monkeyg/conf/
COPY examples/configs/vllm-qwen3.5-27b-fp8-no-think.provider.yml /opt/monkeyg/conf/
COPY examples/configs/vllm-qwen332b-fp16.provider.yml /opt/monkeyg/conf/

COPY LICENSE /opt/monkeyg/LICENSE
COPY NOTICE /opt/monkeyg/NOTICE
COPY EULA.md /opt/monkeyg/EULA
COPY EULA.md /opt/monkeyg/fe/EULA.md

RUN chown -R monkeyg:monkeyg /opt/monkeyg

WORKDIR /opt/monkeyg

USER monkeyg

ENTRYPOINT ["/opt/monkeyg/bin/entrypoint.sh", "/opt/monkeyg/bin/monkeyg"]

