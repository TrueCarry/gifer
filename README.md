# Gifer

## Description

`gifer` resize GIF image.

This service implement the same API as [thumbor](https://github.com/thumbor/thumbor).

## Features

- [x] [Resize](#resize)
- [ ] [Crop](#crop)

## Resize

If we'd like to keep the aspect ratio,
we need to specify only one component, either width or height, 
and set the other component to 0.

For example: 0x400, 100x0, 200x200

```
curl -I 'http://localhost:8080/unsafe/0x400/filters:gifv(webm)/https://66.media.tumblr.com/bb202134de4a12f482e7d1637c0da733/tumblr_nnbt6tLsou1s7jx17o1_400.gif'
```

## Formats

* `filters:gifv(webm) -> gif`
* `filters:format(jpeg) -> jpg`
* `filters:format(png) -> jpg`

## Run

Clone the repository.

Build and run container:

```bash
$> sudo docker build -t gifer .
$> sudo docker run -it --rm -p 8080:8080 gifer
# You will see:
... Start gifer server on 8080
```


