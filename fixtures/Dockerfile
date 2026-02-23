FROM node:lts-alpine3.23@sha256:4f696fbf39f383c1e486030ba6b289a5d9af541642fc78ab197e584a113b9c03 AS builder

WORKDIR /app

COPY fixtures/package.json package.json
COPY fixtures/package-lock.json package-lock.json
RUN npm ci --ignore-scripts

FROM python:3.14-alpine3.23@sha256:31da4cb527055e4e3d7e9e006dffe9329f84ebea79eaca0a1f1c27ce61e40ca5

WORKDIR /app

COPY fixtures/requirements.txt requirements.txt
RUN pip3 install --upgrade pip && \
    pip3 install -r requirements.txt && \
    pip3 cache purge

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

RUN addgroup -S app && adduser -S -g app app
USER app

CMD [ "flask", "run", "--host", "0.0.0.0", "--port", "8080"]
