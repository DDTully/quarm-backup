FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /app/db-backup .

FROM alpine:latest

RUN apk update && apk add --no-cache mariadb-client tzdata

WORKDIR /app

COPY --from=builder /app/db-backup .

RUN mkdir /app/backups

ENTRYPOINT ["/app/db-backup"]
