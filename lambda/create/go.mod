module github.com/ministryofjustice/opg-data-lpa-store/lambda/create

go 1.20

replace github.com/ministryofjustice/opg-data-lpa-store/lambda/shared => ../shared

require (
	github.com/aws/aws-lambda-go v1.41.0
	github.com/ministryofjustice/opg-data-lpa-store/lambda/shared v0.0.0-00010101000000-000000000000
	github.com/ministryofjustice/opg-go-common v0.0.0-20231106092059-b3dcf8bd1eeb
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/aws/aws-sdk-go v1.47.9 // indirect
	github.com/aws/aws-xray-sdk-go v1.8.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.1.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.34.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	google.golang.org/grpc v1.56.3 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
