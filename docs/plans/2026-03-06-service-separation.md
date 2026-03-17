# Service Separation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan
> task-by-task.

**Goal:** Split `services/backend` into `services/api` and `services/scheduler` with independent Go
modules, Docker images, and Helm charts.

**Architecture:** The API service keeps all business logic, database access, and HTTP handlers. The
scheduler is a lightweight Go binary with only the generated HTTP client — it calls the API's
internal endpoints on a schedule. Both services follow identical Go conventions (Cobra, Viper, slog,
cmd/internal layout).

**Tech Stack:** Go 1.25, Cobra, Viper, slog, oapi-codegen, Chi, Docker (distroless), Helm

---

### Task 1: Rename services/backend → services/api

**Files:**

- Rename: `services/backend/` → `services/api/`
- Modify: `services/api/go.mod` (update module path)
- Modify: `Taskfile.yml` (update include path)
- Modify: `CLAUDE.md` (update all references)
- Modify: `docs/development.md` (update all references)
- Modify: `docs/architecture.md` (update all references)
- Modify: `docs/contributing.md` (update all references)
- Modify: `.github/workflows/deploy.yaml` (update build context path)
- Modify: `services/api/Dockerfile` (no changes needed — relative paths)

**Step 1: Move the directory**

```bash
git mv services/backend services/api
```

**Step 2: Update Go module path**

In `services/api/go.mod`, change:

```
module github.com/retr0h/freebie
```

to:

```
module github.com/retr0h/freebie/services/api
```

**Step 3: Update all internal import paths**

Find and replace across all `.go` files in `services/api/`:

```
"github.com/retr0h/freebie/internal/ → "github.com/retr0h/freebie/services/api/internal/
```

**Step 4: Update Taskfile.yml**

Root `Taskfile.yml` — change:

```yaml
backend:
  taskfile: ./services/backend/Taskfile.yml
  dir: ./services/backend
```

to:

```yaml
api:
  taskfile: ./services/api/Taskfile.yml
  dir: ./services/api
```

Also update `clean` task:

```yaml
clean:
  desc: Wipe backend database (migrations run on next serve)
  cmds:
    - task: api:clean
```

**Step 5: Update CLAUDE.md**

Replace all `services/backend` references with `services/api`. Replace all `task backend:*` with
`task api:*`.

**Step 6: Update docs/development.md**

Replace all `services/backend` references with `services/api`. Replace all `task backend:*` with
`task api:*`. Replace all `backend` task names.

**Step 7: Update docs/architecture.md**

Replace `services/backend` references with `services/api`.

**Step 8: Update docs/contributing.md**

Replace `services/backend` references with `services/api`.

**Step 9: Update .github/workflows/deploy.yaml**

Change build context:

```yaml
docker build -t "$IMAGE_TAG" -t "${{ env.REGISTRY }}/${{ env.IMAGE }}:latest" services/api/
```

Change trigger paths:

```yaml
paths:
  - "services/api/**"
  - "services/scheduler/**"
  - "charts/**"
  - ".github/workflows/deploy.yaml"
```

**Step 10: Verify build**

```bash
cd services/api && go build ./...
```

**Step 11: Run tests**

```bash
cd services/api && go test ./...
```

**Step 12: Commit**

```bash
git add -A
git commit -m "refactor: Rename services/backend to services/api

Update Go module path, all internal imports, Taskfile,
CI/CD, and documentation references."
```

---

### Task 2: Remove local worker commands from API

The API no longer needs `worker`, `worker run`, `worker check-triggers`, `worker send-reminders`, or
`worker remote` CLI commands. The worker service logic stays (used by HTTP handlers). The
`internal/scheduler/` package is also removed — K8s CronJobs handle scheduling now.

**Files:**

- Delete: `services/api/cmd/worker.go`
- Delete: `services/api/cmd/worker_remote.go`
- Delete: `services/api/internal/scheduler/scheduler.go`
- Delete: `services/api/internal/client/` (generated HTTP client moves to scheduler)

**Step 1: Delete worker CLI commands**

```bash
rm services/api/cmd/worker.go
rm services/api/cmd/worker_remote.go
```

**Step 2: Delete the scheduler package**

```bash
rm -rf services/api/internal/scheduler/
```

**Step 3: Delete the client package**

```bash
rm -rf services/api/internal/client/
```

**Step 4: Clean up go.mod**

Remove dependencies only used by deleted code:

```bash
cd services/api && go mod tidy
```

**Step 5: Verify build**

```bash
cd services/api && go build ./...
```

**Step 6: Run tests**

```bash
cd services/api && go test ./...
```

**Step 7: Commit**

```bash
git add -A
git commit -m "refactor(api): Remove worker CLI and scheduler

Worker CLI commands and cron scheduler move to the
scheduler service. Worker service logic stays in the
API for internal HTTP endpoints."
```

---

### Task 3: Create services/scheduler

**Files:**

- Create: `services/scheduler/main.go`
- Create: `services/scheduler/cmd/root.go`
- Create: `services/scheduler/cmd/check_triggers.go`
- Create: `services/scheduler/cmd/send_reminders.go`
- Create: `services/scheduler/internal/config/config.go`
- Create: `services/scheduler/internal/client/gen/` (copy from old backend)
- Create: `services/scheduler/go.mod`
- Create: `services/scheduler/Taskfile.yml`

**Step 1: Initialize Go module**

```bash
mkdir -p services/scheduler
cd services/scheduler
go mod init github.com/retr0h/freebie/services/scheduler
```

**Step 2: Create main.go**

```go
package main

import "github.com/retr0h/freebie/services/scheduler/cmd"

func main() {
	cmd.Execute()
}
```

**Step 3: Create cmd/root.go**

Mirror the API's root.go structure — Cobra root command, Viper config with `SCHEDULER_` prefix, slog
logger setup.

```go
package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/retr0h/freebie/services/scheduler/internal/config"
)

var (
	cfg     config.Config
	cfgFile string
	debug   bool
	jsonLog bool
)

var rootCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "Freebie scheduled task runner",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&jsonLog, "json", false, "output logs as JSON")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	viper.SetEnvPrefix("SCHEDULER")
	viper.AutomaticEnv()
	viper.ReadInConfig()

	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing config:", err)
		os.Exit(1)
	}

	// Setup logger
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	var handler slog.Handler
	if jsonLog {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	} else {
		handler = tint.NewHandler(os.Stderr, &tint.Options{Level: level})
	}
	slog.SetDefault(slog.New(handler))
}
```

**Step 4: Create internal/config/config.go**

```go
package config

type Config struct {
	API    API    `mapstructure:"api"`
	Worker Worker `mapstructure:"worker"`
}

type API struct {
	URL string `mapstructure:"url"`
}

type Worker struct {
	Secret string `mapstructure:"secret"`
}
```

**Step 5: Copy generated client code**

Copy from the old backend's `internal/client/gen/` directory. The OpenAPI spec and generated client
code are identical.

```bash
mkdir -p services/scheduler/internal/client/gen
cp services/api/internal/api/worker/gen/api.yaml services/scheduler/internal/client/gen/
```

Create `services/scheduler/internal/client/gen/cfg.yaml`:

```yaml
package: gen
output: client.gen.go
generate:
  client: true
```

Create `services/scheduler/internal/client/gen/generate.go`:

```go
package gen

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=cfg.yaml api.yaml
```

Generate the client:

```bash
cd services/scheduler
go get github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
go generate ./internal/client/gen/
```

**Step 6: Create cmd/check_triggers.go**

```go
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/spf13/cobra"

	clientgen "github.com/retr0h/freebie/services/scheduler/internal/client/gen"
)

var checkDate string

var checkTriggersCmd = &cobra.Command{
	Use:   "check-triggers",
	Short: "Check game results and create triggered events",
	RunE:  runCheckTriggers,
}

func init() {
	rootCmd.AddCommand(checkTriggersCmd)
	checkTriggersCmd.Flags().StringVar(&checkDate, "date", "", "Check triggers for a specific date (YYYY-MM-DD)")
}

func runCheckTriggers(cmd *cobra.Command, args []string) error {
	logger := slog.Default()

	secret := cfg.Worker.Secret
	if secret == "" {
		return fmt.Errorf("worker secret is required (SCHEDULER_WORKER_SECRET)")
	}

	apiURL := cfg.API.URL
	if apiURL == "" {
		apiURL = "http://freebie-api:8080"
	}

	client, err := clientgen.NewClient(apiURL)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	logger.Info("calling check-triggers", "url", apiURL)

	var params *clientgen.CheckTriggersParams
	if checkDate != "" {
		params = &clientgen.CheckTriggersParams{Date: &checkDate}
	}

	resp, err := client.CheckTriggers(context.Background(), params, bearerAuth(secret))
	if err != nil {
		return fmt.Errorf("calling check-triggers: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("check-triggers failed (HTTP %d): %s", resp.StatusCode, body)
	}

	var result clientgen.CheckTriggersResult
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	logger.Info("check-triggers complete",
		"triggered", result.Triggered,
		"notified", result.Notified,
		"totalEvents", result.TotalEvents,
	)
	return nil
}

func bearerAuth(token string) clientgen.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}
}
```

**Step 7: Create cmd/send_reminders.go**

```go
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/spf13/cobra"

	clientgen "github.com/retr0h/freebie/services/scheduler/internal/client/gen"
)

var sendRemindersCmd = &cobra.Command{
	Use:   "send-reminders",
	Short: "Send reminder notifications for expiring deals",
	RunE:  runSendReminders,
}

func init() {
	rootCmd.AddCommand(sendRemindersCmd)
}

func runSendReminders(cmd *cobra.Command, args []string) error {
	logger := slog.Default()

	secret := cfg.Worker.Secret
	if secret == "" {
		return fmt.Errorf("worker secret is required (SCHEDULER_WORKER_SECRET)")
	}

	apiURL := cfg.API.URL
	if apiURL == "" {
		apiURL = "http://freebie-api:8080"
	}

	client, err := clientgen.NewClient(apiURL)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	logger.Info("calling send-reminders", "url", apiURL)
	resp, err := client.SendReminders(context.Background(), bearerAuth(secret))
	if err != nil {
		return fmt.Errorf("calling send-reminders: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send-reminders failed (HTTP %d): %s", resp.StatusCode, body)
	}

	var result clientgen.SendRemindersResult
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	logger.Info("send-reminders complete",
		"sent", result.Sent,
		"failed", result.Failed,
		"expiringDeals", result.ExpiringDeals,
	)
	return nil
}
```

**Step 8: Create Taskfile.yml**

```yaml
version: "3"

tasks:
  default:
    desc: Show available tasks
    cmds:
      - task --list

  build:
    desc: Build scheduler binary
    cmds:
      - go build -o scheduler .

  test:
    desc: Run tests
    cmds:
      - go test ./...

  generate:
    desc: Regenerate OpenAPI client
    cmds:
      - go generate ./internal/client/gen/
```

**Step 9: Add to root Taskfile.yml**

Add scheduler include:

```yaml
scheduler:
  taskfile: ./services/scheduler/Taskfile.yml
  dir: ./services/scheduler
```

**Step 10: Install dependencies and verify build**

```bash
cd services/scheduler
go mod tidy
go build ./...
```

**Step 11: Run tests**

```bash
cd services/scheduler && go test ./...
```

**Step 12: Commit**

```bash
git add -A
git commit -m "feat(scheduler): Add scheduler service

Lightweight Go binary that calls the API's internal worker
endpoints via HTTP. Commands: check-triggers, send-reminders.
Uses generated OpenAPI client with bearer token auth."
```

---

### Task 4: Create scheduler Dockerfile

**Files:**

- Create: `services/scheduler/Dockerfile`
- Create: `services/scheduler/.dockerignore`

**Step 1: Create .dockerignore**

```
*

!go.mod
!go.sum

!*.go
!cmd/
!cmd/**/*.go
!internal/
!internal/**/*.go
```

**Step 2: Create Dockerfile**

No CGO needed — pure Go binary, smaller and faster to build.

```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary (no CGO needed)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-s -w' -o scheduler .

# Runtime stage - distroless
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /app/scheduler .

# Run as nonroot user (uid 65532)
USER nonroot:nonroot

ENTRYPOINT ["./scheduler"]
```

**Step 3: Verify Docker build**

```bash
docker build --platform linux/amd64 -t freebie-scheduler:test services/scheduler/
```

**Step 4: Commit**

```bash
git add -A
git commit -m "feat(scheduler): Add Dockerfile

Pure Go build (no CGO), distroless runtime, nonroot user."
```

---

### Task 5: Rename charts/cronjobs → charts/scheduler

**Files:**

- Rename: `charts/cronjobs/` → `charts/scheduler/`
- Modify: `charts/scheduler/Chart.yaml` (rename chart)
- Modify: `charts/scheduler/values.yaml` (update image to scheduler)
- Modify: `charts/scheduler/templates/_helpers.tpl` (rename)
- Modify: `charts/scheduler/templates/cronjob-triggers.yaml` (update command)
- Modify: `charts/scheduler/templates/cronjob-reminders.yaml` (update command)
- Modify: `charts/scheduler/templates/secret.yaml` (update naming)
- Modify: `.github/workflows/deploy.yaml` (update chart name + add scheduler image build)

**Step 1: Move the directory**

```bash
git mv charts/cronjobs charts/scheduler
```

**Step 2: Update Chart.yaml**

Change `name: freebie-cronjobs` to `name: freebie-scheduler`.

**Step 3: Update values.yaml**

Change image repository to `freebie-scheduler`. Change `pullPolicy` default. Update env var name
from `FREEBIE_WORKER_SECRET` to `SCHEDULER_WORKER_SECRET`.

```yaml
image:
  repository: registry.digitalocean.com/freebie/freebie-scheduler
  tag: latest
  pullPolicy: IfNotPresent

api:
  serviceName: freebie-api
  port: 8080

workerSecret: ""

checkTriggers:
  schedule: "0 14 * * *" # 6am PT (UTC-8)

sendReminders:
  schedule: "0 2 * * *" # 6pm PT (UTC-8)
```

**Step 4: Update \_helpers.tpl**

Replace all `freebie-cronjobs` with `freebie-scheduler`.

**Step 5: Update cronjob-triggers.yaml**

Change command from:

```yaml
command:
  - ./freebie
  - worker
  - remote
  - check-triggers
  - --api-url
  - "http://{{ .Values.api.serviceName }}:{{ .Values.api.port }}"
```

to:

```yaml
command:
  - ./scheduler
  - check-triggers
env:
  - name: SCHEDULER_API_URL
    value: "http://{{ .Values.api.serviceName }}:{{ .Values.api.port }}"
  - name: SCHEDULER_WORKER_SECRET
    valueFrom:
      secretKeyRef:
        name: { { include "freebie-scheduler.fullname" . } }
        key: worker-secret
```

Remove the separate `FREEBIE_WORKER_SECRET` env block since it's now inline.

**Step 6: Update cronjob-reminders.yaml**

Same pattern as triggers — update command and env vars.

**Step 7: Update secret.yaml**

Update naming from `freebie-cronjobs` helpers to `freebie-scheduler`.

**Step 8: Update .github/workflows/deploy.yaml**

Add second image build for scheduler. Update chart name from `freebie-cronjobs` to
`freebie-scheduler`. Pass `SCHEDULER_WORKER_SECRET` instead of `FREEBIE_WORKER_SECRET` for the
scheduler chart.

**Step 9: Commit**

```bash
git add -A
git commit -m "refactor(charts): Rename cronjobs chart to scheduler

Update Helm chart to use dedicated scheduler image with
SCHEDULER_* env vars. Update CI/CD to build and deploy
both images."
```

---

### Task 6: Update all documentation

**Files:**

- Modify: `CLAUDE.md`
- Modify: `docs/development.md`
- Modify: `docs/architecture.md`
- Modify: `docs/contributing.md`
- Modify: `services/api/docs/deployment.md`
- Modify: `services/api/docs/architecture.md`

**Step 1: Update CLAUDE.md**

- Replace `services/backend` → `services/api`
- Replace `task backend:*` → `task api:*`
- Add `services/scheduler/` to directory structure
- Add `task scheduler:build` to quick reference

**Step 2: Update docs/development.md**

- Replace `services/backend` → `services/api`
- Replace `task backend:*` → `task api:*`
- Add scheduler tasks to task table

**Step 3: Update docs/architecture.md**

- Replace `services/backend` references with `services/api`
- Update CronJob section to reference scheduler binary
- Update CI/CD section for two images

**Step 4: Update docs/contributing.md**

- Replace `services/backend` → `services/api`
- Replace `task backend:*` → `task api:*`

**Step 5: Update services/api/docs/deployment.md**

- Update Docker build commands for both images
- Update Helm deploy commands for `charts/scheduler/`
- Update env var table with `SCHEDULER_*` vars

**Step 6: Update services/api/docs/architecture.md**

- Reference scheduler as separate service
- Update component diagram

**Step 7: Verify no stale references**

```bash
grep -r "services/backend" --include="*.md" --include="*.yml" --include="*.yaml" .
grep -r "task backend:" --include="*.md" --include="*.yml" --include="*.yaml" .
grep -r "freebie-cronjobs" .
grep -r "charts/cronjobs" .
```

**Step 8: Commit**

```bash
git add -A
git commit -m "docs: Update all references for service separation

Rename services/backend → services/api and
charts/cronjobs → charts/scheduler across all docs,
Taskfiles, and CI/CD configuration."
```

---

### Task 7: Final verification

**Step 1: Build API**

```bash
cd services/api && go build ./... && go test ./...
```

**Step 2: Build scheduler**

```bash
cd services/scheduler && go build ./... && go test ./...
```

**Step 3: Docker build both images**

```bash
docker build --platform linux/amd64 -t freebie-api:test services/api/
docker build --platform linux/amd64 -t freebie-scheduler:test services/scheduler/
```

**Step 4: Helm template validation**

```bash
helm template freebie-api charts/api/ --set databasePath=test --set workerSecret=test
helm template freebie-scheduler charts/scheduler/ --set workerSecret=test
helm template freebie-cloudflare charts/cloudflare/ --set tunnelToken=test
```

**Step 5: Verify no stale references**

```bash
grep -rn "services/backend" . --include="*.go" --include="*.md" --include="*.yml" --include="*.yaml" | grep -v ".git/"
grep -rn "freebie-cronjobs" . | grep -v ".git/"
grep -rn "worker remote" . --include="*.go" --include="*.yaml" | grep -v ".git/"
```

**Step 6: Commit any fixes**

```bash
git add -A
git commit -m "chore: Final verification and cleanup"
```
