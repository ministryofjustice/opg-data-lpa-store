FROM golang:1.21.5 AS build-env

WORKDIR /app
COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

RUN go mod download

COPY ./lambda/shared /app/lambda/shared

ARG DIR
COPY ./lambda/$DIR /app/lambda/$DIR

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/main ./lambda/$DIR

FROM alpine:3

RUN apk upgrade libssl3 libcrypto3

COPY --from=build-env /go/bin/main /var/task/main

ENTRYPOINT [ "/var/task/main" ]
