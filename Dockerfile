# syntax=docker/dockerfile:1.4
ARG PACKAGE_VER=develop
ARG PACKAGE_REV=

# --- Stage 1: Frontend ---
FROM node:23-slim AS fe-build
ENV NODE_ENV=production NODE_OPTIONS="--max-old-space-size=4096"
WORKDIR /frontend
COPY McAfee.pe[m] /usr/local/share/ca-certificates/McAfee.crt
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates git && \
    update-ca-certificates && rm -rf /var/lib/apt/lists/*
COPY ./backend/pkg/graph/schema.graphqls ../backend/pkg/graph/
COPY frontend/ .
RUN --mount=type=cache,target=/root/.npm npm ci --include=dev
RUN npm run build -- --mode production --minify esbuild --outDir dist \
    --emptyOutDir --sourcemap false --target es2020

# --- Stage 2: Backend ---
FROM golang:1.24-bookworm AS be-build
ARG PACKAGE_VER
ARG PACKAGE_REV
ENV CGO_ENABLED=0 GO111MODULE=on
COPY McAfee.pe[m] /usr/local/share/ca-certificates/McAfee.crt
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates git && \
    update-ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /backend
COPY backend/ .
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download || go mod download
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    for bin in pentagi ctester ftester etester; do \
      go build -trimpath \
        -ldflags "-s -w \
          -X pentagi/pkg/version.PackageName=${bin} \
          -X pentagi/pkg/version.PackageVer=${PACKAGE_VER} \
          -X pentagi/pkg/version.PackageRev=${PACKAGE_REV}" \
        -o /out/${bin} ./cmd/${bin}; \
    done

# --- Stage 3: Final ---
FROM alpine:3.23.3
COPY McAfee.pe[m] /etc/ssl/certs/McAfee.pem
RUN if [ -f /etc/ssl/certs/McAfee.pem ]; then \
      cat /etc/ssl/certs/McAfee.pem >> /etc/ssl/certs/ca-certificates.crt; \
    fi && \
    apk --no-cache add ca-certificates openssl openssh-keygen shadow && \
    if [ -f /etc/ssl/certs/McAfee.pem ]; then \
      cp /etc/ssl/certs/McAfee.pem /usr/local/share/ca-certificates/McAfee.crt && \
      update-ca-certificates; \
    fi && \
    addgroup -g 998 docker && \
    addgroup -S appuser && adduser -S appuser -G appuser && \
    addgroup appuser docker

WORKDIR /opt/mtsdevsecops
RUN mkdir -p bin ssl fe logs data conf

COPY scripts/entrypoint.sh bin/entrypoint.sh
RUN sed -i 's/\r//' bin/entrypoint.sh && chmod +x bin/entrypoint.sh

COPY --from=be-build /out/pentagi bin/mtsdevsecops
COPY --from=be-build /out/ctester /out/ftester /out/etester bin/
COPY --from=fe-build /frontend/dist fe/
COPY examples/configs/*.provider.yml conf/
COPY LICENSE NOTICE EULA.md ./
COPY EULA.md fe/EULA.md

RUN chown -R appuser:appuser /opt/mtsdevsecops
USER appuser

ENTRYPOINT ["/opt/mtsdevsecops/bin/entrypoint.sh", "/opt/mtsdevsecops/bin/mtsdevsecops"]

LABEL org.opencontainers.image.source="https://gitlab.ecsbr.net/pagueveloz/infra/serasa-cyber-shield" \
      org.opencontainers.image.description="Serasa Cyber Shield AI - Autonomous Penetration Testing" \
      org.opencontainers.image.authors="Serasa Experian Security Team" \
      org.opencontainers.image.licenses="MIT"
