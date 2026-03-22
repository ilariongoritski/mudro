FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY services/casino ./services/casino

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/casino ./services/casino/cmd/casino

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/casino /usr/local/bin/casino

EXPOSE 8081

CMD ["/usr/local/bin/casino"]
