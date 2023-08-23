SHELL = '/bin/bash'
export AWS_ACCESS_KEY_ID ?= X
export AWS_SECRET_ACCESS_KEY ?= X

build:
  # Nothing to build yet

up:
	docker-compose up -d

down:
	docker-compose down

create-tables:
	aws --endpoint-url http://localhost:8030 dynamodb create-table \
		--no-cli-pager \
		--table-name deeds \
		--attribute-definitions AttributeName=uid,AttributeType=S \
		--key-schema AttributeName=uid,KeyType=HASH \
		--billing-mode PAY_PER_REQUEST

	aws --endpoint-url http://localhost:8030 dynamodb create-table \
		--no-cli-pager \
		--table-name events \
		--attribute-definitions AttributeName=uid,AttributeType=S AttributeName=created,AttributeType=S \
		--key-schema AttributeName=uid,KeyType=HASH AttributeName=created,KeyType=RANGE \
		--billing-mode PAY_PER_REQUEST
