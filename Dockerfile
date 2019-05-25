FROM jrottenberg/ffmpeg:4.0-ubuntu AS build-env

RUN apt-get update -y && apt-get install -y curl git

WORKDIR /
RUN curl -O https://storage.googleapis.com/golang/go1.12.4.linux-amd64.tar.gz 
RUN tar xvf go1.12.4.linux-amd64.tar.gz && rm -rf go1.12.4.linux-amd64.tar.gz

ENV PATH $PATH:/go/bin

RUN mkdir -p /gifer

ARG PORT=8080

EXPOSE 8080 8080

WORKDIR /gifer

COPY go.mod go.sum main.go ./
RUN go mod download

RUN ["go", "build"]

FROM jrottenberg/ffmpeg:3.4-alpine 
WORKDIR /app

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY --from=build-env /gifer /bin/
# https://stackoverflow.com/a/35613430/3105368
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

ENTRYPOINT ["gifer"]
