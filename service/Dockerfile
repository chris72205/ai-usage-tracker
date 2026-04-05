# syntax=docker/dockerfile:1

# --- build stage ---
# Build on the native runner platform; cross-compile to target via Go toolchain.
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# TARGETOS, TARGETARCH, and TARGETVARIANT are set automatically by Docker buildx.
# TARGETVARIANT is "v7" for linux/arm/v7 — strip the leading "v" to get GOARM=7.
ARG TARGETOS TARGETARCH TARGETVARIANT
RUN CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    GOARM=${TARGETVARIANT#v} \
    go build -trimpath -ldflags="-s -w" -o /app/server ./cmd/server

# --- final stage ---
FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/server /server

EXPOSE 8080

ENTRYPOINT ["/server"]
