FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /movie-catalog ./services/movie-catalog/cmd

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /movie-catalog /usr/local/bin/movie-catalog

EXPOSE 8091

ENTRYPOINT ["/usr/local/bin/movie-catalog"]
