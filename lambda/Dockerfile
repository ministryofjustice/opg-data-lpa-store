FROM golang:1.21.4 AS build-env

ARG DIR
WORKDIR /app/lambda/$DIR
COPY ./lambda/shared ../shared
COPY ./lambda/$DIR/go.mod ./go.mod
COPY ./lambda/$DIR/go.sum ./go.sum

RUN go mod download

COPY ./lambda/$DIR .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/main .

FROM alpine:3

RUN apk upgrade libssl3 libcrypto3

COPY --from=build-env /go/bin/main /var/task/main

ENTRYPOINT [ "/var/task/main" ]
