module github.com/ministryofjustice/opg-data-lpa-store/lambda/get

go 1.20

replace github.com/ministryofjustice/opg-data-lpa-store/lambda/shared => ../shared

require (
	github.com/aws/aws-lambda-go v1.41.0
	github.com/ministryofjustice/opg-data-lpa-store/lambda/shared v0.0.0-00010101000000-000000000000
	github.com/ministryofjustice/opg-go-common v0.0.0-20231128145056-24628fba649c
)

require (
	github.com/andybalholm/brotli v1.0.6 // indirect
	github.com/aws/aws-sdk-go v1.48.11 // indirect
	github.com/aws/aws-xray-sdk-go v1.8.3 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.50.0 // indirect
	golang.org/x/net v0.18.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/grpc v1.59.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
