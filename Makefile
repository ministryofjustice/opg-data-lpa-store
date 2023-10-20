SHELL = '/bin/bash'
export AWS_ACCESS_KEY_ID ?= X
export AWS_SECRET_ACCESS_KEY ?= X

build:
	docker compose build --parallel lambda-create apigw

up:
	docker compose up -d apigw
	make create-tables

down:
	docker compose down

test-api: URL ?= http://localhost:9000/
test-api:
	go build -o ./signer/test-api ./signer && \
	chmod +x ./signer/test-api && \
	./signer/test-api PUT $(URL)/M-AL9A-7EY3-075D '{"uid":"M-AL9A-7EY3-075D","version":"1"}'

create-tables:
	docker compose run --rm aws dynamodb describe-table --table-name deeds || \
	docker compose run --rm aws dynamodb create-table \
		--table-name deeds \
		--attribute-definitions AttributeName=uid,AttributeType=S \
		--key-schema AttributeName=uid,KeyType=HASH \
		--billing-mode PAY_PER_REQUEST

	docker compose run --rm aws dynamodb describe-table --table-name events || \
	docker compose run --rm aws dynamodb create-table \
		--table-name events \
		--attribute-definitions AttributeName=uid,AttributeType=S AttributeName=created,AttributeType=S \
		--key-schema AttributeName=uid,KeyType=HASH AttributeName=created,KeyType=RANGE \
		--billing-mode PAY_PER_REQUEST

run-structurizr:
	docker pull structurizr/lite
	docker run -it --rm -p 4080:8080 -v $(PWD)/docs/architecture/dsl/local:/usr/local/structurizr structurizr/lite

run-structurizr-export:
	docker pull structurizr/cli:latest
	docker run --rm -v $(PWD)/docs/architecture/dsl/local:/usr/local/structurizr structurizr/cli \
	export -workspace /usr/local/structurizr/workspace.dsl -format mermaid
