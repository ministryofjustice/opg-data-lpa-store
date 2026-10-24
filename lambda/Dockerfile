FROM golang:1.27.2@sha256:820153096fda4de6ac1b7194d12e416b7dd8f4e9a33d17e0345b111b945162d8 AS build-env

WORKDIR /app
COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

RUN go mod download

COPY ./internal /app/internal

ARG DIR
COPY ./lambda/$DIR /app/lambda/$DIR

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/main "./lambda/$DIR"

FROM scratch

COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /go/bin/main /var/task/main

USER 65532

ENTRYPOINT [ "/var/task/main" ]
