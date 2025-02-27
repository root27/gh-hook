FROM golang:1.23.0 as builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o hook .

FROM gcr.io/distroless/static

COPY --from=builder /app/hook /hook
COPY --from=builder /app/.env /.env

ENTRYPOINT ["/hook"]
