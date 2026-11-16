set shell := ["bash", "-eu", "-o", "pipefail", "-c"]

image := "registry.digitalocean.com/freebie/freebie-api:latest"

default:
    @just --list

# Install the app's dependencies (first-time setup).
setup: mobile-install

# Wipe the local API database; the next serve migrates a fresh one.
clean: api-clean

# Run the Go tests for the API and the scheduler.
test: api-test scheduler-test mobile-test

# Run the local checks without starting the API or the app.
ready: just-fmt-check md-lint test mobile-typecheck

# Start the API server on localhost:8080; it migrates the database on start.
[working-directory('services/api')]
api-serve:
    go run . serve

# Run the API, rebuilding on Go changes (needs watchexec).
[working-directory('services/api')]
api-dev:
    watchexec -r -e go -- go run . serve

# Build the production API binary into services/api/freebie.
[working-directory('services/api')]
api-build:
    go build -o freebie .

# Run the API's Go tests.
[working-directory('services/api')]
api-test:
    go test ./...

# Regenerate sqlc code from services/api/db/queries.sql.
[working-directory('services/api/db')]
api-generate:
    go tool sqlc generate

# Delete services/api/freebie.db and its journal files.
[working-directory('services/api')]
api-clean:
    rm -f freebie.db freebie.db-journal freebie.db-wal freebie.db-shm

# Create three test deals (24h, 6h and 2h expiry), then list them.
[working-directory('services/api')]
api-deals:
    go run . deals create --hours 24
    go run . deals create --hours 6
    go run . deals create --hours 2
    go run . deals list

# List active deals.
[working-directory('services/api')]
api-deals-list:
    go run . deals list

# List all users.
[working-directory('services/api')]
api-users:
    go run . users list

# Send a test notification to all users.
[working-directory('services/api')]
api-notify:
    go run . notify test

# Send a test notification to one user.
[working-directory('services/api')]
api-notify-send user:
    go run . notify send {{ user }}

# Build the scheduler binary into services/scheduler/scheduler.
[working-directory('services/scheduler')]
scheduler-build:
    go build -o scheduler .

# Run the scheduler's Go tests.
[working-directory('services/scheduler')]
scheduler-test:
    go test ./...

# Regenerate the scheduler's OpenAPI client.
[working-directory('services/scheduler')]
scheduler-generate:
    go generate ./internal/client/gen/

# Install the app's dependencies with bun.
[working-directory('apps/mobile')]
mobile-install:
    bun install

# Start the Expo dev server for Expo Go; press i for the iOS simulator.
[working-directory('apps/mobile')]
mobile-serve:
    bunx expo start --go

# Start Expo and open the iOS simulator.
[working-directory('apps/mobile')]
mobile-ios:
    bunx expo start --ios

# Start Expo and open the Android emulator.
[working-directory('apps/mobile')]
mobile-android:
    bunx expo start --android

# Start the web version of the app.
[working-directory('apps/mobile')]
mobile-web:
    bunx expo start --web

# Type-check the app with tsc.
[working-directory('apps/mobile')]
mobile-typecheck:
    bunx tsc --noEmit

# Run the app's unit tests (deep link routes and pending links) with bun.
[working-directory('apps/mobile')]
mobile-test:
    bun test src

# Lint the app with ESLint.
[working-directory('apps/mobile')]
mobile-lint:
    bunx eslint . --ext .ts,.tsx

# Clear the app's build caches.
[working-directory('apps/mobile')]
mobile-clean:
    rm -rf node_modules/.cache .expo

# Remove node_modules and caches, then reinstall.
[working-directory('apps/mobile')]
mobile-reset:
    rm -rf node_modules .expo
    bun install

# Build an iOS development client with EAS.
[working-directory('apps/mobile')]
mobile-build-ios:
    bunx eas build --platform ios --profile development

# Build an Android development client with EAS.
[working-directory('apps/mobile')]
mobile-build-android:
    bunx eas build --platform android --profile development

# Install prettier for formatting Markdown.
[working-directory('docs')]
docs-install:
    bun install

# Format every Markdown file to 100 columns with prettier.
[working-directory('docs')]
docs-fmt:
    bunx prettier --write --prose-wrap always --print-width 100 "../**/*.md"

# Check Markdown formatting without writing.
[working-directory('docs')]
docs-fmt-check:
    bunx prettier --check --prose-wrap always --print-width 100 "../**/*.md"

# Check Markdown using the same markdownlint-cli2 version as CI.
md-lint:
    bunx --bun markdownlint-cli2@0.24.0 '**/*.md'

# Build the API and scheduler image for linux/amd64.
docker-build:
    docker build --platform linux/amd64 -t {{ image }} .

# Push the image to the DigitalOcean registry.
docker-push:
    docker push {{ image }}

# Build and push the image, then restart the API deployment.
deploy: docker-build docker-push
    kubectl rollout restart deployment freebie-api -n freebies

# Format the justfile.
just-fmt:
    just --unstable --fmt

# Check the justfile's formatting.
just-fmt-check:
    just --unstable --fmt --check

# Preview notes for commits since the last release; writes nothing.
release-notes:
    git-cliff --config cliff.toml --unreleased --strip header
