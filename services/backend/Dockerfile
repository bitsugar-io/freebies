# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary with CGO
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o freebie .

# Runtime stage - distroless
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy binary (migrations are embedded)
COPY --from=builder /app/freebie .

EXPOSE 8080

# Run as nonroot user (uid 65532)
USER nonroot:nonroot

# Fly.io processes override CMD via fly.toml [processes]
ENTRYPOINT ["./freebie"]
CMD ["serve", "--host", "0.0.0.0"]
