FROM golang:1.12-alpine AS build-env

RUN apk update && apk add --no-cache curl git

WORKDIR /gifer

COPY go.mod go.sum ./
RUN go mod download

COPY main.go from_url.go from_file.go ./
RUN go build -ldflags="-s -w" -o gifer

FROM jrottenberg/ffmpeg:3.4-alpine
WORKDIR /app

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY --from=build-env /gifer/gifer /bin/

ENTRYPOINT ["gifer"]
