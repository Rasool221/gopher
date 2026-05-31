# Run gopher locally. Pass URL and flags after `run`, e.g.:
#   just run http://localhost:8080 -w 4 -v
run *ARGS:
  go run ./main.go {{ARGS}}

# Fast tests, no network or docker required.
unit-tests:
  go test -v ./...

# Run integration tests against the dockerized test site.
# Brings the site up if it isn't already, waits for it, then runs the tests.
# Site is left running afterwards — use `just test-site-down` when done.
integration-tests: test-site-up
  @echo "Waiting for test site at http://localhost:8080..."
  @until curl -sf http://localhost:8080/ > /dev/null; do sleep 0.2; done
  go test -tags=integration -v -count=1 ./internal/...

test-site-up:
  docker compose -f test/docker-compose.yml up -d

test-site-down:
  docker compose -f test/docker-compose.yml down

test-site-logs:
  docker compose -f test/docker-compose.yml logs -f
