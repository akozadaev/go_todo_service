# syntax=docker/dockerfile:1
FROM golang:1.26.4-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/todo-api ./cmd/api

FROM alpine:3.23
RUN apk add --no-cache ca-certificates tzdata && addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/todo-api /usr/local/bin/todo-api
RUN mkdir -p /app/logs && chown -R app:app /app
USER app
EXPOSE 8080 50051
ENTRYPOINT ["todo-api"]
