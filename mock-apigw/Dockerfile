FROM golang:1.21.3 AS build-env

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/main .

FROM alpine:3

COPY --from=build-env /go/bin/main /var/task/main

ENTRYPOINT [ "/var/task/main" ]
