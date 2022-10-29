FROM golang:alpine AS builder

WORKDIR /go/src/github.com/mushonnip/echo-server

COPY . .

RUN go build -o /go/bin/echo-server

RUN chmod +x /go/bin/echo-server
ENTRYPOINT ["/go/bin/echo-server"]