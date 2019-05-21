# Gifer

## Description

`gifer` converts a GIF image to a mp4/webm format.

## Run

0. Clone the repository.

1. Build and run container:

```bash
$> sudo docker build -t gifer .
$> sudo docker run -it --rm -p 8080:8080 -e PORT=8080 gifer
# You will see:
... Start gifer server on 8080
```

