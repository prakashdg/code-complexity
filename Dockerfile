FROM golang:1.18 AS golang

ADD . /src

ENV HOME=/src \
    LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    PATH="$PATH:/src/bin" \
    PYTHONPATH=/src/src

VOLUME /src

WORKDIR /src

RUN go mod init complexity

RUN go mod tidy

RUN go build *.go
