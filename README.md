<img src="assets/icon.svg" width="80" alt="navi icon">

# navi

Hey, listen! Notifications on your phone from coding agents, through [ntfy](https://ntfy.sh).

Tell an agent to let you know when it is done, walk away, and your phone buzzes when it is. Each notification says which repo and machine it came from.

```sh
$ navi send "Tests pass, PR #42 is open" --click https://github.com/dittofleet/lichen/pull/42
Sent
```

```
lichen on devbox
Tests pass, PR #42 is open
```

`-p high` makes it louder, `--tag warning` adds an emoji, and `-t` replaces the title. Run `navi help` for the rest.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/dittofleet/navi/main/install.sh | sh
navi setup
```

The installer puts the latest release in `~/.local/bin/navi` (override with `NAVI_INSTALL_DIR`). Supported platforms: macOS and Linux, arm64 and x64.

`navi setup` generates a random topic and shows how to subscribe to it in the ntfy app (iOS or Android). The public server at ntfy.sh needs no account. Anyone who knows the topic can read it, so keep it to yourself. Run `navi setup` again to see it, or `navi setup --rotate` to replace it if it leaks.

To set up another machine with the same topic in one line (the variable goes after the pipe, so it reaches the shell running the script):

```sh
curl -fsSL https://raw.githubusercontent.com/dittofleet/navi/main/install.sh | NAVI_TOPIC=<your topic> sh
```

Or sync `~/.config/navi/config.json` with [lichen](https://github.com/dittofleet/lichen).

Not to be confused with [navi the cheatsheet tool](https://github.com/denisidoro/navi). Both can be installed at once without touching each other's files, but only one can be `navi` on your `PATH`.

## Your own server

```sh
navi setup --server https://ntfy.example.com --token tk_...
```

For instant delivery to iOS, the server needs `upstream-base-url: "https://ntfy.sh"` ([ntfy docs](https://docs.ntfy.sh/config/#ios-instant-notifications)).

## Configuration

`navi setup` writes `~/.config/navi/config.json` (respects `$XDG_CONFIG_HOME`), readable only by you:

```json
{
  "schemaVersion": 1,
  "server": "https://ntfy.sh",
  "topic": "navi-3f9a1c6e0b7d24e85a1f9c2d"
}
```

`NAVI_SERVER`, `NAVI_TOPIC` and `NAVI_TOKEN` override the file, and are enough without it.

## Agent skill

`skills/navi/SKILL.md` tells coding agents when to notify you and what to write, so "ping me when the build is done" just works. Unprompted, they only notify when something urgent needs you. Install with [skills.sh](https://github.com/vercel-labs/skills):

```sh
bunx skills add https://github.com/dittofleet/navi
```

## Updating and uninstalling

`navi update` installs the latest release. Once a day, navi prints a hint when one is out. It skips the check when `CI` or `NAVI_NO_UPDATE_CHECK` is set or stderr is not a terminal.

`navi uninstall` removes the binary, the config and the update cache, after asking (`--yes` skips the prompt). Your topic and the phone's subscription are untouched.
