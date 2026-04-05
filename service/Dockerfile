# syntax=docker/dockerfile:1

# --- build stage ---
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=arm64
# For RPi 3 running a 32-bit OS, build with: --build-arg TARGETARCH=arm --build-arg GOARM=7
ARG GOARM=

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=$GOARM \
    go build -trimpath -ldflags="-s -w" -o /app/server ./cmd/server

# --- final stage ---
FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/server /server

EXPOSE 8080

ENTRYPOINT ["/server"]
