SHELL = '/bin/bash'
export JWT_SECRET_KEY ?= secret

help:
	@grep --no-filename -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build containers
	docker compose build --parallel lambda-create lambda-update lambda-get lambda-getlist apigw

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

	# JWT required
	JWT_SECRET_KEY=bad ./api-test/tester -expectedStatus=401 REQUEST PUT $(URL)/lpas/$(LPA_UID) '{"version":"1"}'
	JWT_SECRET_KEY=bad ./api-test/tester -expectedStatus=401 REQUEST POST $(URL)/lpas/$(LPA_UID)/updates '{"type":"BUMP_VERSION","changes":[{"key":"/version","old":"1","new":"2"}]}'
	JWT_SECRET_KEY=bad ./api-test/tester -expectedStatus=401 REQUEST GET $(URL)/lpas/$(LPA_UID) ''

	# create
	cat ./docs/example-lpa.json | ./api-test/tester -expectedStatus=201 REQUEST PUT $(URL)/lpas/$(LPA_UID) "`xargs -0`"
	./api-test/tester -expectedStatus=200 -write REQUEST GET $(URL)/lpas/$(LPA_UID) '' > $(TMPFILE)

	# missing fields
	diff <(jq --sort-keys 'del(.status,.uid,.updatedAt)' < $(TMPFILE)) <(jq --sort-keys . < docs/example-lpa.json)
	./api-test/tester -expectedStatus=400 REQUEST PUT $(URL)/lpas/$(LPA_UID) '{"version":"2"}'

	# certificate provider sign
	cat ./docs/certificate-provider-sign.json | ./api-test/tester -expectedStatus=201 REQUEST POST $(URL)/lpas/$(LPA_UID)/updates "`xargs -0`"

	# attorney sign
	cat ./docs/attorney-sign.json | ./api-test/tester -expectedStatus=201 REQUEST POST $(URL)/lpas/$(LPA_UID)/updates "`xargs -0`"

	# donor id check complete
	cat ./docs/donor-confirm-identity.json | ./api-test/tester -expectedStatus=201 REQUEST POST $(URL)/lpas/$(LPA_UID)/updates "`xargs -0`"

	# certificate provider id check complete
	cat ./docs/certificate-provider-confirm-identity.json | ./api-test/tester -expectedStatus=201 REQUEST POST $(URL)/lpas/$(LPA_UID)/updates "`xargs -0`"

	# trust corporation sign
	cat ./docs/trust-corporation-sign.json | ./api-test/tester -expectedStatus=201 REQUEST POST $(URL)/lpas/$(LPA_UID)/updates "`xargs -0`"

	# lpa enters statutory waiting period
	cat ./docs/statutory-waiting-period.json | ./api-test/tester -expectedStatus=201 REQUEST POST $(URL)/lpas/$(LPA_UID)/updates "`xargs -0`"

	# get lpa
	./api-test/tester -expectedStatus=200 REQUEST GET $(URL)/lpas/$(LPA_UID) ''

	# get lpas
	./api-test/tester -expectedStatus=200 REQUEST POST $(URL)/lpas '{"uids": [$(LPA_UID)]}'

	# certificate provider opt out
	$(eval LPA_UID := "$(shell ./api-test/tester UID)")
	cat ./docs/example-lpa.json | ./api-test/tester -expectedStatus=201 REQUEST PUT $(URL)/lpas/$(LPA_UID) "`xargs -0`"
	cat ./docs/certificate-provider-opt-out.json | ./api-test/tester -expectedStatus=201 REQUEST POST $(URL)/lpas/$(LPA_UID)/updates "`xargs -0`"

	# donor withdraws lpa
	$(eval LPA_UID := "$(shell ./api-test/tester UID)")
	cat ./docs/example-lpa.json | ./api-test/tester -expectedStatus=201 REQUEST PUT $(URL)/lpas/$(LPA_UID) "`xargs -0`"
	cat ./docs/donor-withdraw-lpa.json | ./api-test/tester -expectedStatus=201 REQUEST POST $(URL)/lpas/$(LPA_UID)/updates "`xargs -0`"
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
