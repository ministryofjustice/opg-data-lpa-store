FROM node:lts-alpine AS builder

WORKDIR /app

COPY fixtures/package.json package.json
COPY fixtures/package-lock.json package-lock.json
RUN npm ci

FROM python:3-alpine3.17

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

EXPOSE 80

CMD [ "flask", "run", "--host", "0.0.0.0", "--port", "80"]
