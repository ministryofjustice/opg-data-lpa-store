FROM node:lts-alpine3.23@sha256:7fddd9ddeae8196abf4a3ef2de34e11f7b1a722119f91f28ddf1e99dcafdf114 AS builder

WORKDIR /app

COPY fixtures/package.json package.json
COPY fixtures/package-lock.json package-lock.json
RUN npm ci --ignore-scripts

FROM python:3.14-alpine3.23@sha256:faee120f7885a06fcc9677922331391fa690d911c020abb9e8025ff3d908e510

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
