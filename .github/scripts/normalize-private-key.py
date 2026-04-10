#!/usr/bin/env python3
import base64
import os
import sys


def strip_outer_quotes(value: str) -> str:
    if len(value) >= 2 and value[0] == value[-1] and value[0] in {"'", '"'}:
        return value[1:-1].strip()
    return value


def raw_payload() -> str:
    return strip_outer_quotes(os.environ["SINGLE_NODE_VPS_SSH_PRIVATE_KEY"].replace("\r", "").strip())


def normalize() -> str:
    raw = raw_payload()
    candidates = [raw]

    if "\\n" in raw:
        candidates.append(strip_outer_quotes(raw.replace("\\n", "\n").strip()))

    try:
        decoded = base64.b64decode(raw, validate=True).decode("utf-8").replace("\r", "").strip()
    except Exception:
        decoded = ""

    if decoded:
        candidates.append(strip_outer_quotes(decoded))

    for candidate in candidates:
        if "BEGIN " in candidate and "PRIVATE KEY" in candidate:
            return candidate.rstrip("\n") + "\n"

    return raw.rstrip("\n") + "\n"


def main() -> None:
    if len(sys.argv) != 2 or sys.argv[1] != "normalize":
        raise SystemExit("usage: normalize-private-key.py normalize")

    sys.stdout.write(normalize())


if __name__ == "__main__":
    main()
