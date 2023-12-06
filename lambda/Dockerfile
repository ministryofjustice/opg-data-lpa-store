FROM golang:1.21.4 AS build-env

ARG DIR
WORKDIR /app/lambda
COPY ./lambda/go.mod /app/lambda/go.mod
COPY ./lambda/go.sum /app/lambda/go.sum

RUN go mod download

COPY ./lambda/shared /app/lambda/shared
COPY ./lambda/$DIR /app/lambda/$DIR

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/main ./$DIR

FROM alpine:3

RUN apk upgrade libssl3 libcrypto3

COPY --from=build-env /go/bin/main /var/task/main

ENTRYPOINT [ "/var/task/main" ]
