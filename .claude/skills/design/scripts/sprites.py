"""Source of truth for the food sprites seeded into the `sprites` table.

Rules (docs/ui.md): 16x16 canvas, '.' transparent, 1 px black outline (K),
light from the upper left, colors only from the NES master palette, at most two
3-color palettes per item (black outline shared). Generic food only: never a
restaurant's packaging, logo or wrapper.

Migrations are diffs. sprites.lock.json (next to this file) records what the last
shipped migration left in the database. `sql` compares SPRITES and EVENT_SPRITES to
the lock and writes an Up that adds or changes only what differs, and a Down that
undoes exactly that: new sprites are deleted, changed sprites and deal mappings are
restored to their locked values. `lock` records the current state after shipping.

Run it through the skill's wrappers (preview.sh, ship.sh), or directly:
  python3 -I sprites.py sql > <migrations>/NNN_<slug>_sprites.sql   # diff vs lock
  python3 -I sprites.py lock                                       # after shipping
  python3 -I sprites.py sheet preview.png                          # 8x contact sheet
"""
import json
import os
import struct
import sys
import zlib

# NES master palette entries used here (index -> hex), see docs/ui.md.
NES = {
    "0F": "#000000", "30": "#FCFCFC", "10": "#BCBCBC", "00": "#7C7C7C",
    "17": "#E45C10", "27": "#FCA044", "07": "#881400", "06": "#A81000",
    "16": "#F83800", "26": "#F87858", "28": "#F8B800", "38": "#F8D878",
    "18": "#AC7C00", "1A": "#00A800", "36": "#F0D0B0", "21": "#3CBCFC",
    "31": "#A4E4FC",
}

SPRITES = {
    "burger": ("Burger", {"K": "0F", "b": "17", "h": "27", "g": "1A", "p": "07"}, 1, [
        "....KKKKKKKK....",
        "..KKbbhbbbhbKK..",
        ".KbhhbbbhbbbbbK.",
        ".KbhbbbbbbbhbbK.",
        "KbbbbbbbbbbbbbbK",
        "KKKKKKKKKKKKKKKK",
        "KggggggggggggggK",
        "KppppppppppppppK",
        "KppppppppppppppK",
        "KKKKKKKKKKKKKKKK",
        "KbbbbbbbbbbbbbbK",
        ".KbbbbbbbbbbbbK.",
        "..KKKKKKKKKKKK..",
    ]),
    "taco": ("Taco", {"K": "0F", "s": "28", "h": "38", "g": "1A", "t": "16"}, 2, [
        "....g.t..g.t....",
        "..ggtggggtgggg..",
        ".KggggggggggggK.",
        "KhhhsssssssssssK",
        "KhsssssssssssssK",
        "KssssssssssssssK",
        ".KssssssssssssK.",
        "..KssssssssssK..",
        "...KKssssssKK...",
        ".....KKKKKK.....",
    ]),
    "street-tacos": ("Street tacos", {"K": "0F", "w": "36", "d": "27", "m": "07", "g": "1A", "o": "30", "l": "10"}, 3, [
        ".gmgm.gmgm.gmgm.",
        "KwmmwKwmmwKwmmwK",
        "KwwwwKwwwwKwwwwK",
        "KwwwwKwwwwKwwwwK",
        "KwwwdKwwwdKwwwdK",
        ".KwdK.KwdK.KwdK.",
        "KKKKKKKKKKKKKKKK",
        "KooooooooooooooK",
        ".KllllllllllllK.",
        "..KKKKKKKKKKKK..",
    ]),
    "burrito": ("Burrito", {"K": "0F", "w": "36", "l": "30", "f": "10", "F": "30"}, 4, [
        "...KKKKKKKKKK...",
        "..KfFFfKlwwwwwK.",
        ".KfFfffKwwwwwwwK",
        "KfffffffKwwwwwwK",
        "KffffffKwwwwwwwK",
        "KffffffKwwwwwwK.",
        ".KffffKwwwwwwK..",
        "..KKKKKKKKKKK...",
    ]),
    "griddle-sandwich": ("Griddle cake sandwich", {"K": "0F", "c": "28", "d": "18", "e": "30", "s": "07"}, 2, [
        "....KKKKKKKK....",
        "..KKccdccdccKK..",
        ".KccdccccccdccK.",
        "KccccccdcccccccK",
        "KKKKKKKKKKKKKKKK",
        "KeeeeeeeeeeeeeeK",
        "KssssssssssssssK",
        "KssssssssssssssK",
        "KKKKKKKKKKKKKKKK",
        "KccccdcccccdcccK",
        ".KccccccccccccK.",
        "..KKKKKKKKKKKK..",
    ]),
    "chicken-sandwich": ("Chicken sandwich", {"K": "0F", "b": "17", "h": "27", "c": "28", "p": "1A"}, 1, [
        "....KKKKKKKK....",
        "..KKbhhbbbbbKK..",
        ".KbhbbbbbbbbbbK.",
        "KbbbbbbbbbbbbbbK",
        "KKKKKKKKKKKKKKKK",
        "KppKppKppKppKppK",
        "KccccccccccccccK",
        "KccKcccKcccKcccK",
        "KccccccccccccccK",
        "KKKKKKKKKKKKKKKK",
        "KbbbbbbbbbbbbbbK",
        ".KbbbbbbbbbbbbK.",
        "..KKKKKKKKKKKK..",
    ]),
    "plate-lunch": ("Plate lunch", {"K": "0F", "w": "30", "g": "10", "m": "07", "v": "1A"}, 4, [
        "..KKKK...KKKKK..",
        ".KwwwwK.KmmmmmK.",
        ".KwwwwwKmmmvmmK.",
        "KwwwwwwwwwwwwwwK",
        "KgwwwwwwwwwwwwgK",
        ".KggggggggggggK.",
        "..KKKKKKKKKKKK..",
    ]),
    "fries": ("Fries", {"K": "0F", "f": "28", "h": "38", "r": "16", "R": "06"}, 2, [
        "....KK.KK.KK....",
        "...KhfKhfKhfK...",
        "..KhfKhfKhfKhK..",
        "..KffKffKffKfK..",
        ".KKKKKKKKKKKKKK.",
        ".KrrrrrrrrrrrrK.",
        "..KrrrrrrrrrrK..",
        "..KrrRRRRRRrrK..",
        "..KrrrrrrrrrrK..",
        "...KrrrrrrrrK...",
        "...KRRRRRRRRK...",
        "....KKKKKKKK....",
    ]),
    "nuggets": ("Nuggets", {"K": "0F", "n": "27", "h": "38"}, 3, [
        "..KKKK..........",
        ".KhnnnK...KKKK..",
        ".KnnnnK..KhnnnK.",
        "..KKKK...KnnnnK.",
        "....KKKKK.KKKK..",
        "...KhnnnnK......",
        "...KnnnnnK......",
        "....KKKKK.......",
    ]),
    "pizza": ("Pizza slice", {"K": "0F", "y": "28", "c": "17", "p": "16", "h": "38"}, 1, [
        "KKKKKKKKKKKKKKKK",
        "KccccccccccccccK",
        "KccccccccccccccK",
        ".KyhyyyypyyyyyK.",
        ".KyppyyyyyyppyK.",
        "..KyyyyypyyyyK..",
        "..KyyyyppyyyyK..",
        "...KyyyyyyyyK...",
        "...KyypyyyyyK...",
        "....KyyyyyyK....",
        ".....KyyyyK.....",
        "......KyyK......",
        ".......KK.......",
    ]),
    "hot-dog": ("Hot dog", {"K": "0F", "b": "17", "h": "27", "s": "06", "m": "28"}, 4, [
        "..KKKKKKKKKKKK..",
        ".KssmsssmsssmsK.",
        "KbsmsssmsssmssbK",
        "KbhbbbbbbbbbbbbK",
        "KbbbbbbbbbbbbbbK",
        ".KbbbbbbbbbbbbK.",
        "..KKKKKKKKKKKK..",
    ]),
    "drink": ("Drink", {"K": "0F", "r": "16", "w": "30", "l": "10"}, 2, [
        "..........KK....",
        ".........KwK....",
        "........KwK.....",
        "..KKKKKKKwKKKK..",
        "..KllllllllllK..",
        "...KrwrrrrrrK...",
        "...KrwrrrrrrK...",
        "...KrrwwwwrrK...",
        "....KrrrrrrK....",
        "....KrrrrrrK....",
        "....KKKKKKKK....",
    ]),
    # Sport badges (16x16) for team tiles; served as team.badge.
    "sport-basketball": ("Basketball badge", {"K": "0F", "f": "27", "h": "27", "s": "0F"}, 0, [
        "................",
        "....KKKKKKKK....",
        "...KKffsfffKK...",
        "..KsfffsffffsK..",
        ".KKfshfsfffsfKK.",
        ".KffhhhsffssffK.",
        ".KfffhfsffsfffK.",
        ".KssssssssssssK.",
        ".KfffsfsffsfffK.",
        ".KfffsfsffsfffK.",
        ".KffssfsffssffK.",
        ".KKfsffsfffsfKK.",
        "..KsfffsffffsK..",
        "...KKffsfffKK...",
        "....KKKKKKKK....",
        "................",
    ]),
    "sport-baseball": ("Baseball badge", {"K": "0F", "f": "30", "h": "30", "r": "16"}, 0, [
        "................",
        "....KKKKKKKK....",
        "...KKffffffKK...",
        "..KrffffffffrK..",
        ".KKrrhfffffrrKK.",
        ".KffhhhffffrffK.",
        ".KffrhfffffrffK.",
        ".KffrrffffrrffK.",
        ".KffrrffffrrffK.",
        ".KffrffffffrffK.",
        ".KffrffffffrffK.",
        ".KKrrffffffrrKK.",
        "..KrffffffffrK..",
        "...KKffffffKK...",
        "....KKKKKKKK....",
        "................",
    ]),
    "sport-football": ("Football badge", {"K": "0F", "f": "06", "h": "27", "w": "30"}, 0, [
        "................",
        "................",
        "................",
        "....KKKKKKKK....",
        "..KKKhfffffKKK..",
        ".KKhhhhhfffffKK.",
        ".KfffhwfwfwfffK.",
        "KKfffwwwwwwfffKK",
        "KKffffwfwfwfffKK",
        ".KffffffffffffK.",
        ".KKffffffffffKK.",
        "..KKKffffffKKK..",
        "....KKKKKKKK....",
        "................",
        "................",
        "................",
    ]),
    "sport-soccer": ("Soccer badge", {"K": "0F", "f": "30", "h": "30", "s": "0F"}, 0, [
        "................",
        "....KKKKKKKK....",
        "...KKffffffKK...",
        "..KsffffffffsK..",
        ".KKsshfffffssKK.",
        ".KsshhhfffffssK.",
        ".KfffhfssfffffK.",
        ".KffffssssffffK.",
        ".KffffssssffffK.",
        ".KfffffssfffffK.",
        ".KssffffffffssK.",
        ".KKssffffffssKK.",
        "..KssffffffssK..",
        "...KKffssffKK...",
        "....KKKKKKKK....",
        "................",
    ]),
    "sport-hockey": ("Hockey badge", {"K": "0F", "b": "28", "t": "30", "g": "00"}, 0, [
        "............KK..",
        "...........KbK..",
        "...........KbK..",
        "..........KbK...",
        "..........KbK...",
        ".........KbK....",
        ".........KbK....",
        "........KbK.....",
        "........KtK.....",
        ".......KtK......",
        "..KKKKKKtK......",
        ".KbbbbbbbK......",
        ".KKKKKKKK.......",
        "..........KKKK..",
        ".........KggggK.",
        "..........KKKK..",
    ]),
    # Fallback: redrawn from 024 onto NES palette entries.
    "gift": ("Gift", {"K": "0F", "x": "21", "X": "31", "y": "16", "Y": "26"}, 1, [
        "...KK.....KK....",
        "..KyYK...KYyK...",
        "...KyYK.KYyK....",
        "....KKyKyKK.....",
        "KKKKKKKyKKKKKKKK",
        "KXxxxxxKyxxxxxxK",
        "KxxxxxxKyxxxxxxK",
        "KKKKKKKKyKKKKKKK",
        ".KXxxxxKyxxxxxK.",
        ".KxxxxxKyxxxxxK.",
        ".KxxxxxKyxxxxxK.",
        ".KxxxxxKyxxxxxK.",
        ".KxxxxxKyxxxxxK.",
        ".KKKKKKKKKKKKKK.",
    ]),
}

# Existing deals -> sprite (ids from earlier migrations). Unlisted deals keep
# the gift fallback.
EVENT_SPRITES = {
    "event-mlb-lad-jitb": "burger",
    "event-mlb-sd-jitb": "burger",
    "event-mlb-ari-jitb": "burger",
    "event-mlb-lad-habit": "burger",
    "event-mlb-lad-panda": "plate-lunch",
    "event-mlb-lad-mcdonalds": "nuggets",
    "event-mlb-lad-ampm": "hot-dog",
    "event-mlb-laa-mcdonalds": "fries",
    "event-mlb-laa-deltaco": "taco",
    "event-mlb-col-tacobell": "taco",
    "event-mlb-sd-mcdonalds": "fries",
    "event-test-smoke": "pizza",
    # 034_deal_verification.sql
    "event-nfl-lar-carlsjr": "burger",
    "event-nhl-ana-chickfila": "chicken-sandwich",
    "event-nba-lac-chickfila": "chicken-sandwich",
    "event-mls-lafc-chickfila": "chicken-sandwich",
}

W = H = 16


def build(sid):
    name, keys, top, rows = SPRITES[sid]
    for i, r in enumerate(rows):
        if len(r) != W:
            raise SystemExit(f"{sid}: row {i} is {len(r)} wide: {r!r}")
        for ch in r:
            if ch != "." and ch not in keys:
                raise SystemExit(f"{sid}: row {i} uses unknown key {ch!r}")
    if top + len(rows) > H:
        raise SystemExit(f"{sid}: {len(rows)} rows at offset {top} overflow {H}")
    full = ["." * W] * top + rows + ["." * W] * (H - top - len(rows))
    used = {ch for r in full for ch in r if ch != "."}
    palette = {k: NES[v] for k, v in keys.items() if k in used}
    non_outline = {k for k in used if keys[k] != "0F"}
    if len(non_outline) > 6:
        raise SystemExit(f"{sid}: {len(non_outline)} fill colors, max 6 (two NES palettes)")
    return name, full, palette


def sql_str(s):
    return "'" + s.replace("'", "''") + "'"


LOCK = os.path.join(os.path.dirname(os.path.abspath(__file__)), "sprites.lock.json")


def current_state():
    sprites = {}
    for sid in SPRITES:
        name, full, palette = build(sid)
        sprites[sid] = {"name": name, "pixels": full, "palette": palette}
    return {"sprites": sprites, "events": dict(EVENT_SPRITES)}


def load_lock():
    if not os.path.exists(LOCK):
        return {"sprites": {}, "events": {}}
    with open(LOCK) as f:
        return json.load(f)


def upsert(sid, spr):
    px = json.dumps(spr["pixels"], separators=(",", ":"))
    pal = json.dumps(spr["palette"], separators=(",", ":"))
    return (f"INSERT INTO sprites (id, name, width, height, pixels, palette) VALUES ({sql_str(sid)}, {sql_str(spr['name'])}, {W}, {H}, {sql_str(px)}, {sql_str(pal)})\n"
            f"    ON CONFLICT(id) DO UPDATE SET name = excluded.name, pixels = excluded.pixels, palette = excluded.palette, updated_at = CURRENT_TIMESTAMP;")


def set_event(eid, sid):
    value = sql_str(sid) if sid else "NULL"
    return f"UPDATE events SET sprite_id = {value} WHERE id = {sql_str(eid)};"


def emit_sql():
    old, new = load_lock(), current_state()
    removed = [sid for sid in old["sprites"] if sid not in new["sprites"]]
    if removed:
        raise SystemExit(f"sprites removed from SPRITES: {removed}. Deleting a shipped sprite needs a hand-written migration; restore them.")
    added = [sid for sid in new["sprites"] if sid not in old["sprites"]]
    changed = [sid for sid in new["sprites"] if sid in old["sprites"] and new["sprites"][sid] != old["sprites"][sid]]
    mapped = [eid for eid in new["events"] if new["events"][eid] != old["events"].get(eid)]
    if not (added or changed or mapped):
        raise SystemExit("nothing to ship: SPRITES and EVENT_SPRITES match sprites.lock.json")
    summary = ", ".join(filter(None, [
        added and "adds " + " ".join(added),
        changed and "changes " + " ".join(changed),
        mapped and f"maps {len(mapped)} deal(s)",
    ]))
    up = ["-- +goose Up",
          "-- Generated by the design skill (.claude/skills/design/scripts/sprites.py); edit that, not this.",
          f"-- Sprite diff vs sprites.lock.json: {summary}."]
    up += [upsert(sid, new["sprites"][sid]) for sid in added + changed]
    up += [set_event(eid, new["events"][eid]) for eid in mapped]
    down = ["", "-- +goose Down"]
    down += [set_event(eid, old["events"].get(eid)) for eid in mapped]
    down += [upsert(sid, old["sprites"][sid]) for sid in changed]
    if added:
        down.append("DELETE FROM sprites WHERE id IN (" + ", ".join(sql_str(s) for s in added) + ");")
    print("\n".join(up + down))


def write_lock():
    with open(LOCK, "w") as f:
        json.dump(current_state(), f, indent=1, sort_keys=True)
        f.write("\n")
    print(f"locked {len(SPRITES)} sprites, {len(EVENT_SPRITES)} deal mappings -> {LOCK}", file=sys.stderr)


def emit_sheet(path, scale=8, gap=2, cols=6):
    ids = list(SPRITES)
    rows_n = (len(ids) + cols - 1) // cols
    cw = (W + gap) * scale
    wpx, hpx = cols * cw + gap * scale, rows_n * cw + gap * scale
    bg = (0x00, 0x40, 0x58)  # NES $0C panel, so black outlines stay visible
    canvas = [[bg] * wpx for _ in range(hpx)]
    for n, sid in enumerate(ids):
        _, full, pal = build(sid)
        ox, oy = gap * scale + (n % cols) * cw, gap * scale + (n // cols) * cw
        for y, r in enumerate(full):
            for x, ch in enumerate(r):
                if ch == ".":
                    continue
                h = pal[ch]
                c = (int(h[1:3], 16), int(h[3:5], 16), int(h[5:7], 16))
                for dy in range(scale):
                    for dx in range(scale):
                        canvas[oy + y * scale + dy][ox + x * scale + dx] = c
    raw = b"".join(b"\x00" + bytes(v for px in row for v in px) for row in canvas)

    def chunk(t, d):
        return struct.pack(">I", len(d)) + t + d + struct.pack(">I", zlib.crc32(t + d) & 0xFFFFFFFF)

    with open(path, "wb") as f:
        f.write(b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", struct.pack(">IIBBBBB", wpx, hpx, 8, 2, 0, 0, 0))
                + chunk(b"IDAT", zlib.compress(raw, 9)) + chunk(b"IEND", b""))
    print("order:", ", ".join(ids), file=sys.stderr)


if __name__ == "__main__":
    cmd = sys.argv[1] if len(sys.argv) > 1 else ""
    if cmd == "sql":
        emit_sql()
    elif cmd == "lock":
        write_lock()
    elif cmd == "sheet" and len(sys.argv) == 3:
        emit_sheet(sys.argv[2])
    else:
        raise SystemExit(__doc__)
