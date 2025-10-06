FROM golang:1.25.1-alpine3.22 AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o backend ./cmd
FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache bash ca-certificates ffmpeg

COPY --from=builder /app/backend .
COPY ./migrations ./migrations
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

EXPOSE 8000


CMD ["sh", "-c", "migrate -path migrations -database postgres://$DB_USER:$DB_PASS@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable up && ./backend"]

