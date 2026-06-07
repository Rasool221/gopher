# List all available recipes (default when running `just` with no arguments).
default:
  @just --list

# Run gopher locally. Pass URL and flags after `run`, e.g.:
#   just run http://localhost:8080 -w 4 -v
run *ARGS:
  go run ./main.go {{ARGS}}

# Fast tests, no network or docker required.
unit-tests:
  go test -v ./...

# Run integration tests against the dockerized test sites.
# Runs `go test` inside the compose network (the `tests` service), which is what lets it
# resolve the primary.com / external.com aliases. depends_on waits for the sites'
# healthchecks first; sites are left running afterwards — use `just test-site-down`.
integration-tests:
  docker compose -f test/docker-compose.yml run --rm tests
  docker compose -f test/docker-compose.yml down

test-site-up:
  docker compose -f test/docker-compose.yml up -d

test-site-down:
  docker compose -f test/docker-compose.yml down

test-site-logs:
  docker compose -f test/docker-compose.yml logs -f
