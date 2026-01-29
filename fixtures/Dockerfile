FROM node:lts-alpine3.23@sha256:cd6fb7efa6490f039f3471a189214d5f548c11df1ff9e5b181aa49e22c14383e AS builder

WORKDIR /app

COPY fixtures/package.json package.json
COPY fixtures/package-lock.json package-lock.json
RUN npm ci --ignore-scripts

FROM python:3.13-alpine3.23@sha256:9298134acf0820ed89d8259944248b24b41e4de911a083f33bd70b583acc368a

WORKDIR /app

COPY fixtures/requirements.txt requirements.txt
RUN pip3 install -r requirements.txt

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
