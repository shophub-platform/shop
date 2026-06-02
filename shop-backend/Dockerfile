# ─── Stage 1: Build ───────────────────────────────────────────────────────────
# Analogno Maven/Gradle build fazi u multi-stage Docker build-u za Spring Boot
FROM golang:1.26-alpine AS builder

# Instaliraj build dependency-je
RUN apk add --no-cache git

WORKDIR /app

# Kopiraj go.mod i go.sum prvo (layer caching - ako se dependencies ne menjaju,
# ovaj layer se kešira. Isti princip kao COPY pom.xml . && mvn dependency:go-offline)
COPY go.mod go.sum ./
RUN go mod download

# Kopiraj ostatak source koda
COPY . .

# Build binarnog fajla
# CGO_ENABLED=0 = statički linked binary (bez C runtime - može u scratch image)
# -ldflags="-s -w" = uklanja debug simbole (manji binary)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /app/server \
    ./cmd/server

# ─── Stage 2: Runtime ─────────────────────────────────────────────────────────
# Minimalni image - samo binary, bez Go toolchain-a
# Analogno distroless image-u za Spring Boot
FROM alpine:3.19

# Dodaj CA sertifikate (potrebni za HTTPS pozive, npr. ka Sepolia RPC)
RUN apk add --no-cache ca-certificates tzdata

# Kreiranje non-root korisnika (security best practice)
RUN addgroup -S shop && adduser -S shop -G shop
USER shop

WORKDIR /app

# Kopiraj samo binary iz builder stage-a
COPY --from=builder /app/server .

EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

ENTRYPOINT ["./server"]
