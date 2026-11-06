# Build stage
FROM golang:1.27.3-alpine@sha256:fa3d596898d9f126af7c01d14ee7597bab3f39cccf392ef36b69fe77cd681f18 AS builder

# Install build dependencies (CGO required for SQLite in API)
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Build API
WORKDIR /app/api
COPY services/api/go.mod services/api/go.sum ./
RUN go mod download
COPY services/api/ .
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o /bin/freebie .

# Build scheduler (no CGO needed)
WORKDIR /app/scheduler
COPY services/scheduler/go.mod services/scheduler/go.sum ./
RUN go mod download
COPY services/scheduler/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-s -w' -o /bin/scheduler .

# Runtime stage - distroless
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy both binaries
COPY --from=builder /bin/freebie .
COPY --from=builder /bin/scheduler .

EXPOSE 8080

# Run as nonroot user (uid 65532)
USER nonroot:nonroot

# Default: run API server
ENTRYPOINT ["./freebie"]
CMD ["serve", "--host", "0.0.0.0"]
