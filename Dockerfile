FROM golang:1.23.0 as builder

WORKDIR /app

COPY . .

RUN go build -o hook .

FROM gcr.io/distroless/static

COPY --from=builder /app/hook /hook
COPY --from=builder /app/.env /.env

ENTRYPOINT ["/hook"]
