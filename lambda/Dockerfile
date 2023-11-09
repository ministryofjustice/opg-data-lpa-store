FROM golang:1.21.4 AS build-env

WORKDIR /app

COPY go.work .
COPY go.work.sum .

COPY . .

ARG DIR
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/main ./lambda/$DIR

FROM alpine:3

RUN apk upgrade libssl3 libcrypto3

COPY --from=build-env /go/bin/main /var/task/main

ENTRYPOINT [ "/var/task/main" ]
