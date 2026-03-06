# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary (no CGO needed)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-s -w' -o scheduler .

# Runtime stage - distroless
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /app/scheduler .

# Run as nonroot user (uid 65532)
USER nonroot:nonroot

ENTRYPOINT ["./scheduler"]
