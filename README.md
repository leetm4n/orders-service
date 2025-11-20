# Orders Microservice

## Required tooling
- go 1.25.x +
- gnu make
- docker (or docker api compliant container orchestration engine like podman)

## Overview

This example service implements a minimal feature set microservice to create and get orders. It uses a schema first codegeneration paradigm where both the controller (API) and the data (repository) layer is generated from schema definitions. This enables the developer to keep documentation of the API in sync with the developed code. The sql codegen enables all time schema correctness and correct typing, which makes the application even less error prone.

## How to start

1. copy `.env.example` and create your own `.env` (or supply DATABASE_URL env var in any other preferred way)
2. `make install` (to install dependencies)
3. `docker compose up -d` to start application dependencies (postgres 18)
4. run database migrations via `make migrate-up`
5. `make start` (to start the application)

## Codegen

Both the repository and api layer is generated from schema (sql / openapi). If any of the schema files change (migrations, queries or API spec) please run `make generate` to apply those changes accordingly into generated code.
