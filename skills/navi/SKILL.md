---
name: navi
description: Send a notification to the user's phone. Use when the user asks to be notified or pinged, or when something urgent needs them.
---

# navi: notify the user on their phone

```sh
navi send "<message>" [-p high] [--tag <emoji>] [--click <url>]
```

The title defaults to `<repo> on <machine>`. Replace it with `-t "<title>"` only if that would mislead.

## When to send

- When the user asks. "Tell me when it's done" means last, after the work and its checks are finished.
- Unprompted only when something is urgent: you are blocked on the user, or something broke that they need to act on. Use `-p high`. Never for progress.

## What to write

- **NEVER send secrets** (API keys, tokens, passwords, credentials, .env contents) or other sensitive information (personal data, private code, customer details). The ntfy server can read every message, and a lock screen shows it to anyone nearby. Point to where the details are instead.
- A sentence or two that says how it went: "Build passed, PR #42 is open" beats "Done".
- `--click <url>` opens a link when tapped, e.g. the PR.

## If it fails

`command not found`, `no topic configured`, an HTTP error, or output about cheatsheets (a different `navi` is first on `PATH`): tell the user. Do not install navi, run `navi setup`, or read its config: the topic in it works like a password.
