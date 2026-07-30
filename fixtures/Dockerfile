FROM node:lts-alpine3.23@sha256:2bdb65ed1dab192432bc31c95f94155ca5ad7fc1392fb7eb7526ab682fa5bf14 AS builder

RUN apk upgrade zlib

WORKDIR /app

COPY fixtures/package.json package.json
COPY fixtures/package-lock.json package-lock.json
RUN npm ci --ignore-scripts

FROM python:3.14-alpine3.23@sha256:5a824eb82cc75361f98611f3cfc5091ea33f10a6ccea4d4ebdabbc523b9a1614

RUN apk upgrade libcrypto3 libuuid musl musl-utils zlib

COPY --from=ghcr.io/astral-sh/uv:0.11.2@sha256:c4f5de312ee66d46810635ffc5df34a1973ba753e7241ce3a08ef979ddd7bea5 /uv /uvx /bin/
ENV UV_MALWARE_CHECK=1
ENV UV_SYSTEM_PYTHON=1
ENV UV_NO_BUILD=1
ENV UV_FROZEN=1

WORKDIR /app

COPY fixtures/pyproject.toml pyproject.toml
COPY fixtures/uv.lock uv.lock
RUN uv sync --no-dev

COPY fixtures/app.py app.py
COPY fixtures/lib lib
COPY --from=builder /app/node_modules/govuk-frontend/dist/govuk/assets static
COPY --from=builder /app/node_modules/govuk-frontend/dist/govuk/govuk-frontend.min.css static
COPY --from=builder /app/node_modules/govuk-frontend/dist/govuk/govuk-frontend.min.js static
COPY --from=builder /app/node_modules/@ministryofjustice/frontend/moj/assets static
COPY --from=builder /app/node_modules/@ministryofjustice/frontend/moj/moj-frontend.min.css static
COPY --from=builder /app/node_modules/@ministryofjustice/frontend/moj/moj-frontend.min.js static
COPY fixtures/static static
COPY fixtures/templates templates
COPY docs/schemas static/schemas

EXPOSE 8080

RUN adduser \
  --disabled-password \
  --gecos "" \
  --home "/nonexistent" \
  --shell "/sbin/nologin" \
  --no-create-home \
  --uid 65532 \
  app
USER app

CMD ["uv", "run", "flask", "run", "--host", "0.0.0.0", "--port", "8080"]
