# Orders Microservice

## Required tooling
- go 1.25.x +
- gnu make
- docker (or docker api compliant container orchestration engine like podman)

## Overview

This example service implements a minimal feature set microservice to create and get orders. It uses a schema first code generation paradigm where both the controller (API) and the data (repository) layer is generated from schema definitions. This enables the developer to keep documentation of the API in sync with the developed code. The sql codegen enables all time schema correctness and correct typing, which makes the application even less error prone.

API specification and generated scaffolding code can be found in `./api` where the OpenAPI documentation can be found under `./api/api.yaml`. The tool that generates API scaffolding from schema is [OAPI Codegen](https://github.com/oapi-codegen/oapi-codegen)

Migrations are defined under `./db/migrations`, the migration schemas are in the format of [dbmate](https://github.com/amacneil/dbmate) tool.

SQL queries can be found under `./internal/repo/query.sql` where queries are annotated for [sqlc](https://docs.sqlc.dev/en/latest/) tool, which is the tool of choise to generate type safe sql queries from migration + these query definitions. My postgres driver of choise is pgx and the connections are pooled.

For observability, I've used opentelemtry. There is a minimal tracing setup with otlp exporter pointing to the jaeger instance defined within docker compose.

For logging I've used `log/slog`, for the http server, the built in `net/http` capabilities were used, as for this small scale project I saw it as a minimal and good fit.

For configuration management I've used [kelseyhightower/envconfig](https://github.com/kelseyhightower/envconfig) which is a minimal env config tool.

## How to start

1. copy `.env.example` and create your own `.env` (or supply DATABASE_URL env var in any other preferred way)
2. `make install` (to install dependencies)
3. `docker compose up -d` to start application dependencies (postgres 18)
4. run database migrations via `make migrate-up`
5. `make start` (to start the application)

## Codegen

Both the repository and api layer is generated from schema (sql / openapi). If any of the schema files change (migrations, queries or API spec) please run `make generate` to apply those changes accordingly into generated code.

## OTEL

After running command `docker compose up` a local all-in-one jaeger instance is available at `http://localhost:16686/` from which all traces can be observed.

## TODO / What can be done to improve:

- outbox pattern for event handling, single transaction write to orders / outbox, ensuring event is emitted with entity persistence
- centralize error handling of common business errors, e.g. err no rows, validation errors, prefer it over in place error handling in handlers
- refactor to use a echo or similar for easier handlers, error handling, middlewares if project grows larger
- better input validation, e.g. right now the validation of uuid does happen but does not result in a descriptive error
- testing via more unit tests, integration tests
- lint via golangci lint
- add gh ci to repo to run test, fmt, codegen diff, govuln checks
- containerize the application
- add authorization to requests
- add feature flag to enable openapi browser of own schema
- create a producer interface so any queue / event solution can be easily connected later, or use something like watermill
