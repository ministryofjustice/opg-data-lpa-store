SHELL = '/bin/bash'
export JWT_SECRET_KEY ?= mysupersecrettestkeythatis128bits

help:
	@grep --no-filename -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build containers
	docker compose build --parallel lambda-create lambda-update lambda-get lambda-getlist lambda-getstatic apigw

up: ## Start application
	docker compose up -d --build apigw

down: ## Stop application
	docker compose down

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

tail-logs: ## tails logs for lambda-create lambda-update lambda-get lambda-getlist apigw
	docker compose --ansi=always logs lambda-create lambda-update lambda-get lambda-getlist apigw -f
