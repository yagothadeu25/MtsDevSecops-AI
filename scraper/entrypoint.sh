#!/bin/bash
set -e

SSL_DIR=/app/ssl
SSL_KEY=$SSL_DIR/service.key
SSL_CRT=$SSL_DIR/service.crt
SSL_CSR=$SSL_DIR/service.csr
SSL_CA_KEY=$SSL_DIR/service_ca.key
SSL_CA_CRT=$SSL_DIR/service_ca.crt
SERVER_NAME=${SERVER_NAME:-scraper.local}

mkdir -p $SSL_DIR

if [[ -f "$SSL_KEY" && -f "$SSL_CRT" ]]; then
    echo "SSL certs already exist"
elif [ "${USE_SSL:-true}" = "true" ]; then
    echo "Generating SSL certs..."
    openssl genrsa -out $SSL_CA_KEY 4096 2>/dev/null
    openssl req -new -x509 -days 3650 -key $SSL_CA_KEY \
        -subj "/C=US/ST=NY/L=NY/O=Scraper/CN=ScraperCA" -out $SSL_CA_CRT 2>/dev/null
    openssl req -newkey rsa:4096 -sha256 -nodes -keyout $SSL_KEY \
        -subj "/C=US/ST=NY/L=NY/O=Scraper/CN=$SERVER_NAME" -out $SSL_CSR 2>/dev/null
    openssl x509 -req -days 730 \
        -extfile <(printf "subjectAltName=DNS:$SERVER_NAME\nkeyUsage=critical,digitalSignature,keyAgreement") \
        -in $SSL_CSR -CA $SSL_CA_CRT -CAkey $SSL_CA_KEY -CAcreateserial -out $SSL_CRT 2>/dev/null
    cat $SSL_CA_CRT >> $SSL_CRT
    echo "SSL certs generated"
fi

export SSL_KEY SSL_CRT

_kill_procs() {
    kill -TERM $pid_proxy $pid_chrome $pid_uvicorn 2>/dev/null
    wait
}
trap _kill_procs SIGTERM SIGINT

# 1. Start Xvfb + browserless-chrome
[ -f /tmp/.X99-lock ] && rm -f /tmp/.X99-lock
if [ -z "$DISPLAY" ]; then
    Xvfb :99 -screen 0 1024x768x16 -nolisten tcp -nolisten unix &
    export DISPLAY=:99
fi
cd /usr/src/app && dumb-init -- node ./build/index.js &
pid_chrome=$!

# 2. Start FastAPI converter
cd /app && uvicorn main:app --host 127.0.0.1 --port 8000 --workers 2 &
pid_uvicorn=$!

# 3. Wait for browserless to be ready
sleep 3

# 4. Start Go proxy
/app/scraper &
pid_proxy=$!

echo "Scraper ready: chrome=$pid_chrome uvicorn=$pid_uvicorn proxy=$pid_proxy"
wait
