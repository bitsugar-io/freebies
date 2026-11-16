# Sprite workflow

All paths are relative to the repo root. Scripts live in `.claude/skills/design/scripts/`.

## 1. Draw

Add an entry to `SPRITES` in `scripts/sprites.py`:

```python
"churro": ("Churro", {"K": "0F", "c": "17", "h": "27", "s": "38"}, 3, [
    "................",   # exactly 16 characters per row
    # ...
]),
```

- The tuple is `(name, {key: NES index}, top_offset, rows)`.
- `top_offset` pads blank rows above the art to center it vertically. Rows plus offset must be 16 or
  fewer.
- Every non-`.` character in the rows must be a key in the palette map.

## 2. Preview and look

```bash
.claude/skills/design/scripts/preview.sh            # writes $TMPDIR/sprite-sheet.png
```

Open or Read the PNG. Check that the art reads at small size, the outline is closed, the highlight
sits upper left, and the density matches its neighbors. Iterate until all hold.

## 3. Map deals

Add `"event-id": "sprite-id"` pairs to `EVENT_SPRITES` in `sprites.py`. A new event can also set
`sprite_id` in the migration that creates it. Deals with no sprite show `gift`.

## 4. Ship

```bash
.claude/skills/design/scripts/ship.sh churro         # e.g. writes 032_churro_sprites.sql
```

- Migrations are **diffs** against `scripts/sprites.lock.json` (what the last shipped migration left
  in the database). Up adds or changes only what differs; Down deletes new sprites and restores
  changed sprites and deal mappings, so each migration rolls back on its own.
- The script writes the next free number, runs the sprite tests, then updates the lock. If tests
  fail it removes the migration and leaves the lock alone. Never edit a shipped migration or the
  lock by hand.
- Removing a shipped sprite is refused; that needs a deliberate hand-written migration.
- `TestSpriteMigrationsRollBackIndependently` proves each sprite migration rolls back cleanly.

## 5. Verify end to end

```bash
just api-serve                                              # auto-migrates; skip if already running
.claude/skills/design/scripts/fetch.sh churro 12      # writes $TMPDIR/churro.png
```

Look at the PNG, then reload the app. Every deal mapped to the sprite now shows it.

## API surface

- `GET /api/v1/sprites`: all sprites.
- `GET /api/v1/sprites/{id}`: one sprite as JSON.
- `GET /api/v1/sprites/{id}.png?scale=N`: PNG, where N is 1–32 (default 8).
- Any future admin UI or CLI that writes the `sprites` table must apply the same checks as
  `internal/sprites.Sprite.Validate` plus the palette rule.
