FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o app cmd/app/main.go

FROM golang:1.23-alpine as release

WORKDIR /app

COPY --from=builder /app/app .
COPY --from=builder /app/config.yml .
COPY --from=builder /app/migrations /app/migrations

CMD ["./app"]