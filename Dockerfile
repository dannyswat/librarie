# ---- build stage ----
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /librarie ./cmd/server

# ---- runtime stage ----
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /librarie .

EXPOSE 8080
ENTRYPOINT ["/app/librarie"]
