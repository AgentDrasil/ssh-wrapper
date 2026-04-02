help:
  just --list

test:
  go test -v ./...

format:
  uvx ruff check --select I --fix
  uvx ruff format

lint:
  golangci-lint run
  uvx ruff check
  uvx ruff format --check

lint-fix:
  uvx ruff check --fix

e2e-test:
  uv run pytest -v
