# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags "-s -w \
      -X github.com/filimonic/piwiw/internal/buildinfo.Version=${VERSION} \
      -X github.com/filimonic/piwiw/internal/buildinfo.CommitSHA=${COMMIT_SHA} \
      -X github.com/filimonic/piwiw/internal/buildinfo.BuildDate=${BUILD_DATE}" \
    -o /out/piwiw ./cmd/piwiw

FROM alpine:latest

# ca-certificates provides update-ca-certificates, run at container start (see
# docker-entrypoint.sh) so certs mounted into /usr/local/share/ca-certificates
# are picked up. su-exec drops from root back to the piwiw user afterwards.
RUN apk add --no-cache ca-certificates su-exec \
    && addgroup -S piwiw \
    && adduser -S -G piwiw -H -D piwiw

COPY --from=builder /out/piwiw /usr/local/bin/piwiw
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Defaults, mirroring the values documented in README.md.
# OPENAI_API_BASE_URL and OPENAI_API_KEY have no default and are required
# at runtime.
ENV SERVER_PORT=11434 \
    SKIP_TLS_VERIFY=false \
    OPENAI_API_CHAT_FORCED_PARAMS="" \
    OPENAI_API_CHAT_FORCED_PARAMS_B64="" \
    REQUEST_TIMEOUT=180 \
    MAX_RETRIES=3 \
    RETRY_DELAY=300 \
    EMPTY_CONTENT_TEXT="" \
    TRACE_FOLDER_PATH="" \
    TRACE_KEEP_HOURS=2160

EXPOSE 11434

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
