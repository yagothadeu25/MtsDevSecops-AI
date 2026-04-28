## Stage 1: Build Go proxy
FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY main.go go.mod ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o scraper .

## Stage 2: Final image based on browserless-chrome
FROM browserless/chrome:1.61.1-chrome-stable

USER root

RUN apt-get update && apt-get install -y --no-install-recommends \
    python3 python3-pip openssl xvfb \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY requirements.txt .
RUN pip3 install --no-cache-dir --trusted-host pypi.org --trusted-host pypi.python.org --trusted-host files.pythonhosted.org -r requirements.txt

COPY main.py .
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

COPY --from=builder /build/scraper /app/scraper

ENV USERNAME=someuser \
    PASSWORD=somepass \
    USE_SSL=true \
    SERVER_NAME=scraper.local \
    LISTEN=0.0.0.0:443 \
    BASE_URL=http://127.0.0.1:3000 \
    CONVERT_URL=http://127.0.0.1:8000 \
    SSL_CRT=ssl/service.crt \
    SSL_KEY=ssl/service.key \
    MAX_CONCURRENT_SESSIONS=10 \
    CONNECTION_TIMEOUT=60000 \
    CHROME_PATH=/usr/bin/google-chrome \
    HOST=0.0.0.0 \
    PORT=3000 \
    IS_DOCKER=true \
    NODE_ENV=production \
    PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=true

EXPOSE 443 3000

CMD ["/app/entrypoint.sh"]
