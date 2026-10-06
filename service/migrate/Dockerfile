FROM golang:1.25.7-bookworm AS builder

WORKDIR /app

# Copy workspace and dependency manifests
COPY go.work go.work.sum* ./
COPY pkg/go.mod pkg/go.sum* ./pkg/
COPY shared/go.mod shared/go.sum* ./shared/
COPY tests/go.mod tests/go.sum* ./tests/

# Copy all service go.mod files
COPY service/ ./service/
RUN find service -type f ! -name "go.mod" ! -name "go.sum" -delete

WORKDIR /app/service/migrate

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /app

# Copy source code
COPY pkg/ ./pkg/
COPY shared/ ./shared/
COPY service/migrate/ ./service/migrate/

WORKDIR /app/service/migrate

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/migrate ./main.go


FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/migrate ./migrate

CMD ["./migrate", "up"]
