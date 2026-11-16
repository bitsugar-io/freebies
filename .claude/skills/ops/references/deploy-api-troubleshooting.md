# Deploying: checking and troubleshooting

Read-only unless marked. Deploy steps are in [deploy-api.md](deploy-api.md).

## Checking and troubleshooting

All read-only unless marked.

### Look

```bash
.claude/skills/ops/scripts/deploy-status.sh             # pods, CronJobs, Jobs, releases, /healthz
kubectl get pods -n freebies -w                         # watch a rollout
kubectl logs -n freebies -l app.kubernetes.io/name=freebie-api -f
kubectl get events -n freebies --sort-by=.lastTimestamp
kubectl describe pod -n freebies <pod>
helm list -n freebies
```

Reach the API without the tunnel:

```bash
kubectl port-forward -n freebies svc/freebie-api 8080:8080
curl localhost:8080/healthz
```

Through the tunnel: `curl https://freebie-api.bitsugar.io/healthz`.

### Scheduler runs

```bash
kubectl get cronjobs,jobs -n freebies
kubectl logs -n freebies job/<job-name>
```

Each job logs a summary of what the API did (events checked, deals triggered, pushes sent).

Running a check by hand changes production and can send real pushes; ask the owner first:

```bash
kubectl create job -n freebies --from=cronjob/freebie-scheduler-check-triggers manual-check-$(date +%s)
```

### The image is distroless

There's no shell in the container. Attach a debug container instead:

```bash
kubectl debug -n freebies <pod> --image=busybox -it
```

### Common failures

**`ErrImagePull`, `no match for platform in manifest`.** The image was built on ARM. Rebuild with
`--platform linux/amd64` and push.

**`ErrImagePull`, `repository not found` or unauthorized.** The registry isn't linked to the
cluster: `doctl kubernetes cluster registry add <cluster-name>`.

**`CrashLoopBackOff`.** Read `kubectl logs -n freebies <pod> --previous`. Usually a bad Turso URL or
token, a missing secret value (an upgrade without `--set`), or a failing migration.

**Tunnel returns 502 or 1033.** The `cloudflared` pod is down, or the public hostname doesn't point
at `freebie-api.freebies.svc.cluster.local:8080`. Check its logs and the Zero Trust dashboard.

**Scheduler jobs fail with 401.** `workerSecret` differs between the api and scheduler releases.
Upgrade both with the same `FREEBIE_WORKER_SECRET`.

**No pushes after a game.** Check the `check-triggers` job log first, then the deal's `trigger_rule`
([deals.md](deals.md)).

**Push fails, or preflight says the registry is nearly full.** The Starter registry holds 500 MB and
every push of `latest` leaves the old layers behind. With the owner's go-ahead, run garbage
collection (the registry is read-only while it runs, so don't push meanwhile):

```bash
doctl registry garbage-collection start freebie --include-untagged-manifests --force
doctl registry garbage-collection get-active freebie   # wait until it's gone
```

The alternative is the Basic tier ($5/mo, 5 GB).
