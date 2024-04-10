#base go image
FROM golang:1.21-alpine as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o backend ./cmd/api

RUN chmod +x /app/backend

#tiny docker image

FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/backend /app

CMD ["/app/backend"]