# Gifer

## Description

`gifer` - Docker image with REST API for image manipulation by ffmpeg.

This service implement the same API as [thumbor](https://github.com/thumbor/thumbor).

## Features

- [x] Resize
- [ ] Crop

## Run

Clone the repository.

Build and run container:

```bash
$> sudo docker build -t gifer .
$> sudo docker run -it --rm -p 8080:8080 gifer
# You will see:
... Start gifer server on 8080
```


