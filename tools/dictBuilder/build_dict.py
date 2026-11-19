#!/usr/bin/env python3

from __future__ import annotations

import argparse
import json
import struct
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Iterable

MAGIC = b"TSIMDI01"
FORMAT_VERSION = 1


@dataclass(frozen=True)
class Entry:
    text: str
    syllables: tuple[str, ...]
    weight: int


def split_header_body(raw: str) -> tuple[str, str]:
    lines = raw.splitlines()
    for idx, line in enumerate(lines):
        if line.strip() == "...":
            return "\n".join(lines[:idx]), "\n".join(lines[idx + 1 :])
    raise ValueError("missing YAML body marker line '...'")


def parse_dict_yaml(path: Path) -> tuple[str, str, list[Entry]]:
    raw = path.read_text(encoding="utf-8")
    header, body = split_header_body(raw)
    name = ""
    version = ""
    for line in header.splitlines():
        line = line.strip()
        if line.startswith("name:"):
            name = line.split(":", 1)[1].strip().strip('"')
        elif line.startswith("version:"):
            version = line.split(":", 1)[1].strip().strip('"')
    entries: list[Entry] = []
    for line in body.splitlines():
        stripped = line.strip()
        if not stripped or stripped.startswith("#"):
            continue
        parts = stripped.split("\t")
        if len(parts) < 2:
            continue
        text = parts[0]
        syllables = tuple(s for s in parts[1].split() if s)
        if not syllables:
            continue
        weight = 0
        if len(parts) >= 3 and parts[2].lstrip("-").isdigit():
            weight = int(parts[2])
        if weight < 0:
            weight = 0
        entries.append(Entry(text=text, syllables=syllables, weight=weight))
    return name, version, entries


def merge_entries(groups: Iterable[list[Entry]]) -> list[Entry]:
    best: dict[tuple[str, tuple[str, ...]], Entry] = {}
    for group in groups:
        for entry in group:
            key = (entry.text, entry.syllables)
            prev = best.get(key)
            if prev is None or entry.weight > prev.weight:
                best[key] = entry
    return sorted(best.values(), key=lambda e: (e.syllables, -e.weight, e.text))


def write_index(path: Path, entries: list[Entry], meta: dict) -> None:
    syllables: dict[str, int] = {}
    for entry in entries:
        for syllable in entry.syllables:
            if syllable not in syllables:
                syllables[syllable] = len(syllables)
    ordered = sorted(syllables.items(), key=lambda item: item[1])
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("wb") as out:
        out.write(MAGIC)
        out.write(struct.pack("<II", FORMAT_VERSION, 0))
        out.write(struct.pack("<II", len(ordered), len(entries)))
        for syllable, _ in ordered:
            encoded = syllable.encode("utf-8")
            if len(encoded) > 255:
                raise ValueError(f"syllable too long: {syllable}")
            out.write(struct.pack("<B", len(encoded)))
            out.write(encoded)
        for entry in entries:
            text = entry.text.encode("utf-8")
            if len(text) > 65535:
                raise ValueError(f"text too long: {entry.text}")
            if len(entry.syllables) > 255:
                raise ValueError(f"too many syllables: {entry.text}")
            out.write(struct.pack("<HI", len(text), entry.weight))
            out.write(text)
            out.write(struct.pack("<B", len(entry.syllables)))
            for syllable in entry.syllables:
                out.write(struct.pack("<H", syllables[syllable]))
    meta_path = path.with_suffix(path.suffix + ".meta.json")
    meta_path.write_text(json.dumps(meta, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def build_license_record(source_root: Path, sources: list[dict]) -> dict:
    license_text = ""
    license_path = source_root / "LICENSE"
    if license_path.is_file():
        license_text = license_path.read_text(encoding="utf-8", errors="replace")[:4000]
    return {
        "source_project": "https://github.com/iDvel/rime-ice",
        "license": "GPL-3.0",
        "redistribution": (
            "Compiled dictionary data derived from rime-ice remains under GPL-3.0. "
            "Distribute corresponding source YAML tables and this license record with the index."
        ),
        "license_excerpt": license_text,
        "inputs": sources,
        "runtime_dependency": "none (no Rime Runtime)",
    }


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Build TSInputMethod compiled dictionary index from rime-ice YAML.")
    parser.add_argument("--source-dir", type=Path, required=True, help="Directory containing *.dict.yaml and LICENSE")
    parser.add_argument("--inputs", nargs="+", required=True, help="Dict basenames such as 8105.dict.yaml base.dict.yaml")
    parser.add_argument("--output", type=Path, required=True, help="Output .tsdict path")
    parser.add_argument("--license-out", type=Path, help="Optional license JSON path")
    args = parser.parse_args(argv)

    groups: list[list[Entry]] = []
    sources: list[dict] = []
    for name in args.inputs:
        path = args.source_dir / name
        if not path.is_file():
            print(f"missing input: {path}", file=sys.stderr)
            return 1
        dict_name, dict_version, entries = parse_dict_yaml(path)
        groups.append(entries)
        sources.append(
            {
                "file": name,
                "name": dict_name,
                "version": dict_version,
                "entries": len(entries),
            }
        )
        print(f"parsed {name}: name={dict_name} version={dict_version} entries={len(entries)}")

    merged = merge_entries(groups)
    license_record = build_license_record(args.source_dir, sources)
    meta = {
        "format": "TSIMDI01",
        "format_version": FORMAT_VERSION,
        "entry_count": len(merged),
        "syllable_count_estimate": len({s for e in merged for s in e.syllables}),
        "license": license_record,
    }
    write_index(args.output, merged, meta)
    license_out = args.license_out or args.output.with_name("LICENSE.rime-ice.json")
    license_out.parent.mkdir(parents=True, exist_ok=True)
    license_out.write_text(json.dumps(license_record, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"wrote {args.output} entries={len(merged)}")
    print(f"wrote {license_out}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
