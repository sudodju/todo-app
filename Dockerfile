FROM golang:1.23.9 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /my_app

FROM ubuntu:latest

WORKDIR /app

COPY --from=builder /my_app /app/
COPY --from=builder /app/web /app/web

ENV TODO_PORT=7540 \
    TODO_DBFILE=scheduler.db \
    TODO_PASSWORD=

EXPOSE 7540

CMD ["/app/my_app"]

