# NAME IT

A tiny, offline, deliberately obnoxious desktop nagger for creative file names.

## Stack

- Wails v2
- Go
- Vanilla JavaScript
- Vanilla CSS
- fsnotify
- beeep
- No database, account, cloud, AI, or network service.

## Run

```bash
wails dev
```

Build:

```bash
wails build
```

## Behavior

Name It watches only the directory chosen for the current work session. It reacts to file creation and rename events, scores naming patterns deterministically, and throws an increasingly aggressive full-window interruption plus an OS notification. It never renames, moves, deletes, or reads file contents.
