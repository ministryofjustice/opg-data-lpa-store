SHELL = '/bin/bash'
LAMBDA_LIST=lambda-create lambda-get lambda-getlist lambda-getstatic lambda-getupdates lambda-update
export JWT_SECRET_KEY ?= mysupersecrettestkeythatis128bits

help:
	@grep --no-filename -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build containers
	docker compose build --parallel $(LAMBDA_LIST) apigw fixtures

up: ## Start application
	docker compose up -d --build apigw

watch: ## Start and build app and watch for changes
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build --watch $(LAMBDA_LIST) apigw

down: ## Stop application
	docker compose down

test-results:
	mkdir -p -m 0777 test-results .gocache .trivy-cache

setup-directories: test-results

scan-all:
	@make scan IMAGE_NAME=lpa-store/lambda/api-create
	@make scan IMAGE_NAME=lpa-store/lambda/api-get
	@make scan IMAGE_NAME=lpa-store/lambda/api-getstatic
	@make scan IMAGE_NAME=lpa-store/lambda/api-update
	@make scan IMAGE_NAME=lpa-store/lambda/api-getlist
	@make scan IMAGE_NAME=lpa-store/lambda/api-getupdates
	@make scan IMAGE_NAME=lpa-store/fixtures

scan: setup-directories
	docker compose run --rm trivy image --format table --exit-code 0 $(IMAGE_NAME)
	docker compose run --rm trivy image --format sarif --output /test-results/trivy.sarif --exit-code 1 $(IMAGE_NAME)

tf-sec:
	docker compose run --rm trivy filesystem --format table --exit-code 1 --scanners secret,misconfig --severity CRITICAL,HIGH,MEDIUM ./terraform/

docker-lint:
	docker compose run --rm trivy filesystem --format table --exit-code 1 --scanners secret,misconfig --severity CRITICAL,HIGH,MEDIUM ./lambda/Dockerfile
	docker compose run --rm trivy filesystem --format table --exit-code 1 --scanners secret,misconfig --severity CRITICAL,HIGH,MEDIUM ./fixtures/Dockerfile

test: ## Unit tests
	go test ./... -race -short -covermode=atomic -coverprofile=coverage.out

test-api:
	go test -count 1 .
.PHONY: test-api

test-pact:
	$(eval JWT := "$(shell go run scripts/make_jwt.go)")

	docker compose run --rm pact-verifier \
      --header="X-Jwt-Authorization=Bearer $(JWT)" \
      --consumer-version-selectors='{"mainBranch": true}'

run-structurizr:
	docker pull structurizr/lite
	docker run -it --rm -p 4080:8080 -v $(PWD)/docs/architecture/dsl/local:/usr/local/structurizr structurizr/lite

run-structurizr-export:
	docker pull structurizr/cli:latest
	docker run --rm -v $(PWD)/docs/architecture/dsl/local:/usr/local/structurizr structurizr/cli \
	export -workspace /usr/local/structurizr/workspace.dsl -format mermaid

go-lint: ## Lint Go code
	docker compose run --rm go-lint

check-code: go-lint test

up-fixtures: ## Bring up fixtures UI locally
	docker compose up -d --build fixtures

build-apigw-openapi-spec:
	yq -n 'load("./docs/openapi/openapi.yaml") * load("./docs/openapi/openapi-aws.override.yaml")' > ./docs/openapi/openapi-aws.compiled.yaml

tail-logs: ## Tails logs for lambdas and apigw
	docker compose --ansi=always logs $(LAMBDA_LIST) apigw -f
