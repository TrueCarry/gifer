FROM golang:1.12-alpine AS build-env

RUN apk update && apk add --no-cache curl git

WORKDIR /gifer

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN go build -ldflags="-s -w" -o gifer

FROM jrottenberg/ffmpeg:3.4-alpine
WORKDIR /app

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY --from=build-env /gifer/gifer /bin/
# https://stackoverflow.com/a/35613430/3105368
# RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

ENTRYPOINT ["gifer"]
