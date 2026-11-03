FROM node:lts-alpine3.23@sha256:e6c03d42b46af3b1e98207e59f06a43cfe5efd3ec4b4d3d9802bddf04f833374 AS builder

RUN apk upgrade zlib

WORKDIR /app

COPY fixtures/package.json package.json
COPY fixtures/package-lock.json package-lock.json
RUN npm ci --ignore-scripts

FROM python:3.14-alpine3.23@sha256:2fc6f06143be3b5ea0f7d8977773a22b398dc9a3641bb7b75940f98e9767631e

RUN apk upgrade libcrypto3 libuuid musl musl-utils zlib \
  && apk del .python-rundeps

COPY --from=ghcr.io/astral-sh/uv:0.13.5@sha256:32a7bb39b496babbd997b0a39a4d66c126bdaf5661b302bf6f06abda8deca6ff /uv /uvx /bin/
ENV UV_MALWARE_CHECK=1
ENV UV_SYSTEM_PYTHON=1
ENV UV_NO_BUILD=1

WORKDIR /app

RUN adduser \
  --disabled-password \
  --gecos "" \
  --home "/nonexistent" \
  --shell "/sbin/nologin" \
  --no-create-home \
  --uid 65532 \
  app

COPY fixtures/pyproject.toml pyproject.toml
COPY fixtures/uv.lock uv.lock
RUN uv sync --no-dev --locked

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

USER app

CMD ["/app/.venv/bin/flask", "run", "--host", "0.0.0.0", "--port", "8080"]
