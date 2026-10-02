# Build context is the kit root: compose uses `context: .` and `dockerfile: backend/Dockerfile`.
FROM golang:1.27-alpine AS builder
WORKDIR /build/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -o /out/api ./cmd/api \
    && CGO_ENABLED=0 go build -trimpath -o /out/worker ./cmd/worker

FROM alpine:3.21
RUN apk add --no-cache ca-certificates \
    && adduser -S -H app

WORKDIR /app
COPY --from=builder /out/api /app/api
COPY --from=builder /out/worker /app/worker
COPY backend/migrations /app/migrations

ENV MIGRATIONS_DIR=/app/migrations
USER app
EXPOSE 8000
CMD ["/app/api"]
