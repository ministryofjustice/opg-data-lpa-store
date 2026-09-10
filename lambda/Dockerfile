FROM golang:1.27.1@sha256:512690a5660563b57d37ecc31129e7f136e831db2aed24a1dbeb8ad7380dc0fa AS build-env

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
