FROM golang:1.24-alpine AS build

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/feed-api ./services/feed-api/cmd

FROM gcr.io/distroless/static:nonroot

COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/feed-api /feed-api

EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/feed-api"]
