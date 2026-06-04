# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/api /usr/local/bin/api

USER 65534:65534
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
