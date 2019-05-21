FROM jrottenberg/ffmpeg:4.0-ubuntu

LABEL version="1.0"
LABEL maintainer="shindu666@gmail.com"

RUN apt-get update -y && apt-get install -y curl git

WORKDIR /
RUN curl -O https://storage.googleapis.com/golang/go1.12.4.linux-amd64.tar.gz 
RUN tar xvf go1.12.4.linux-amd64.tar.gz && rm -rf go1.12.4.linux-amd64.tar.gz

ENV PATH $PATH:/go/bin

RUN mkdir -p /gifer

ARG PORT=8080

EXPOSE 8080 8080

COPY go.mod /gifer
COPY main.go /gifer
WORKDIR /gifer

ENTRYPOINT ["go", "run", "main.go"]
