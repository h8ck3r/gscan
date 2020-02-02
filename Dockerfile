FROM golang:1.13.7-alpine3.11

RUN apk add --no-cache gcc g++ libc-dev

RUN addgroup -g 1003 h8ck3r
RUN adduser -h /home/h8ck3r -s /bin/ash -u 1003 -D -G h8ck3r h8ck3r

RUN mkdir -p /home/h8ck3r/go/src/github.com/h8ck3r/gcrawl
RUN chown -R 1003:1003 /home/h8ck3r

ENV GOPATH /home/h8ck3r/go
ENV GOBIN /home/h8ck3r/go/bin
ENV GO111MODULE on

USER 1003:1003

WORKDIR /home/h8ck3r/go/src/github.com/h8ck3r/gcrawl