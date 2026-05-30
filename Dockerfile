FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /migrate ./cmd/migrate

# ---

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /api /api
COPY --from=builder /migrate /migrate

EXPOSE 8080

CMD ["/api"]
