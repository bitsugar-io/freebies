# CLI Reference

## Server

```bash
./bin/freebie serve           # Start API server
./bin/freebie serve --debug   # With debug logging
```

## Worker

```bash
./bin/freebie worker check-triggers             # Check yesterday's games
./bin/freebie worker check-triggers --date 2024-10-30  # Check specific date
./bin/freebie worker send-reminders             # Notify expiring deals
```

## Flags

| Flag       | Description                               |
| ---------- | ----------------------------------------- |
| `--config` | Config file path (default: ./config.yaml) |
| `--db`     | Database file path                        |
| `--debug`  | Enable debug logging                      |
| `--json`   | JSON log output                           |
