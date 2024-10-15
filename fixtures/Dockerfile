FROM node:lts-alpine AS builder

WORKDIR /app

COPY fixtures/package.json package.json
COPY fixtures/package-lock.json package-lock.json
RUN npm ci --ignore-scripts

FROM python:3.13-alpine3.19

WORKDIR /app

COPY fixtures/requirements.txt requirements.txt
RUN pip3 install -r requirements.txt

COPY fixtures/app.py app.py
COPY fixtures/lib lib
COPY --from=builder /app/node_modules/govuk-frontend/dist/govuk/assets static
COPY --from=builder /app/node_modules/govuk-frontend/dist/govuk/govuk-frontend.min.css static
COPY --from=builder /app/node_modules/govuk-frontend/dist/govuk/govuk-frontend.min.js static
COPY fixtures/static static
COPY fixtures/templates templates
COPY docs/schemas static/schemas

EXPOSE 8080

RUN addgroup -S app && adduser -S -g app app
USER app

CMD [ "flask", "run", "--host", "0.0.0.0", "--port", "8080"]
