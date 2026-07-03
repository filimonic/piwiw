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

# Required for TLS verification when piwiw talks to the OpenAI API over HTTPS.
RUN apk add --no-cache ca-certificates \
    && addgroup -S piwiw \
    && adduser -S -G piwiw -H -D piwiw

COPY --from=builder /out/piwiw /usr/local/bin/piwiw

# Defaults, mirroring the values documented in README.md.
# OPENAI_API_BASE_URL and OPENAI_API_KEY have no default and are required
# at runtime.
ENV SERVER_PORT=11434 \
    SKIP_TLS_VERIFY=false \
    OPENAI_API_CHAT_FORCED_PARAMS="" \
    REQUEST_TIMEOUT=180 \
    MAX_RETRIES=3 \
    RETRY_DELAY=300 \
    EMPTY_CONTENT_TEXT="" \
    TRACE_FOLDER_PATH="" \
    TRACE_KEEP_HOURS=2160

EXPOSE 11434

USER piwiw

ENTRYPOINT ["/usr/local/bin/piwiw"]
