##################################################
# Multi-stage Dockerfile for all Go services
# Usage: docker build --build-arg SERVICE=cmd/api -t mudro-api .
##################################################
FROM golang:1.24-alpine AS build

ARG SERVICE=cmd/api

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/service ./${SERVICE}

##################################################
FROM gcr.io/distroless/static:nonroot

COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /bin/service /service

USER nonroot:nonroot
ENTRYPOINT ["/service"]
