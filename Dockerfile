FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o relay .

FROM scratch

COPY --from=builder /app/relay /relay

EXPOSE 8080

CMD ["/relay"]