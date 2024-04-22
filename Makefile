SHELL = '/bin/bash'
export JWT_SECRET_KEY ?= secret

help:
	@grep --no-filename -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build containers
	docker compose build --parallel lambda-create lambda-update lambda-get apigw

up: ## Start application
	docker compose up -d apigw

down: ## Stop application
	docker compose down

test: ## Unit tests
	go test ./... -race -covermode=atomic -coverprofile=coverage.out

test-api: URL ?= http://localhost:9000
# test-api: export JWT_SECRET_KEY ?= secret
test-api:
	$(shell go build -o ./api-test/tester ./api-test && chmod +x ./api-test/tester)
	$(eval LPA_UID := "$(shell ./api-test/tester UID)")
	$(eval TMPFILE := "$(shell mktemp)")

	JWT_SECRET_KEY=bad ./api-test/tester -expectedStatus=401 REQUEST PUT $(URL)/lpas/$(LPA_UID) '{"version":"1"}'
	JWT_SECRET_KEY=bad ./api-test/tester -expectedStatus=401 REQUEST POST $(URL)/lpas/$(LPA_UID)/updates '{"type":"BUMP_VERSION","changes":[{"key":"/version","old":"1","new":"2"}]}'
	JWT_SECRET_KEY=bad ./api-test/tester -expectedStatus=401 REQUEST GET $(URL)/lpas/$(LPA_UID) ''

	cat ./docs/example-lpa.json | ./api-test/tester -expectedStatus=201 REQUEST PUT $(URL)/lpas/$(LPA_UID) "`xargs -0`"
	./api-test/tester -expectedStatus=200 -write REQUEST GET $(URL)/lpas/$(LPA_UID) '' > $(TMPFILE)

	diff <(jq --sort-keys 'del(.status,.uid,.updatedAt)' < $(TMPFILE)) <(jq --sort-keys . < docs/example-lpa.json)
	./api-test/tester -expectedStatus=400 REQUEST PUT $(URL)/lpas/$(LPA_UID) '{"version":"2"}'

	cat ./docs/certificate-provider-change.json | ./api-test/tester -expectedStatus=201 REQUEST POST $(URL)/lpas/$(LPA_UID)/updates "`xargs -0`"
	./api-test/tester -expectedStatus=200 REQUEST GET $(URL)/lpas/$(LPA_UID) ''
	./api-test/tester -expectedStatus=200 REQUEST POST $(URL)/lpas '{"uids": [$(LPA_UID)]}'
.PHONY: test-api

test-pact:
	$(eval JWT := "$(shell JWT_SECRET_KEY=secret ./api-test/tester JWT)")

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

gosec: ## Scan Go code for security flaws
	docker compose run --rm gosec

check-code: go-lint gosec test

up-fixtures: ## Bring up fixtures UI locally
	docker compose up -d --build fixtures

build-apigw-openapi-spec:
	yq -n 'load("./docs/openapi/openapi.yaml") * load("./docs/openapi/openapi-aws.override.yaml")' > ./docs/openapi/openapi-aws.compiled.yaml
