FROM --platform=linux/amd64 golang:1.21-bookworm AS builder

RUN useradd -ms /bin/bash main

USER main

WORKDIR /home/main

COPY *.go ./

RUN mkdir -p ./bin && \
    CGO_ENABLED=0 GOOS=linux go build -o bin/server

ENTRYPOINT ["./bin/server"]