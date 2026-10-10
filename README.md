# Name It

[![CI](https://img.shields.io/github/actions/workflow/status/Vinetwigs/name-it/ci.yml?branch=master&label=CI)](https://github.com/Vinetwigs/name-it/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Vinetwigs/name-it)](https://github.com/Vinetwigs/name-it/releases)
[![License](https://img.shields.io/github/license/Vinetwigs/name-it)](LICENSE)

A tiny, offline desktop nagger for creative file names.

Name It watches a single folder and violently disapproves the names you give
your files. It scores every file the moment it is created, escalates from polite
suggestions to all-caps despair, and takes over the screen to make sure you
notice. It never renames, moves, deletes or reads your files.

## Features

- Watches one folder per working session, including subfolders created later.
- Deterministic filename scoring with six severity levels: `ignore`, `gentle`,
  `normal`, `passive`, `aggressive` and `maximum`.
- Escalating full-window interruption plus a native OS notification.
- Naming pattern alerts: only complain when a file does not follow your chosen
  convention.
- Session statistics, a crime log and a 30-minute pause button.
- Optional update check against GitHub Releases, announced with an in-app modal.
- Completely offline. The only network request is the update check, and only on
  startup.

## How it works

When you start a session, Name It records a baseline of the folder and then
listens for file creation events. For files that do not already follow the
selected naming pattern, the name is tokenized and scored:

- suspicious tokens (`final`, `def`, `copy`, `test`, `temp`, …) add points;
- repeated or improvised numeric suffixes add points;
- a proper `_v1`-style version marker subtracts points;
- re-saving the same document over and over is treated as escalation.

The total score selects a severity tier, and the tier selects a message from one
of the built-in message pools. The more crimes you commit in a row, the louder
Name It gets.

## Download

Grab the latest compressed Windows build from the
[Releases](https://github.com/Vinetwigs/name-it/releases) page. The executable is
stripped and packed with [UPX](https://upx.github.io/) for a small download.

## Build from source

### Requirements

- [Go](https://go.dev/) 1.25 or newer
- [Wails v2](https://wails.io/) (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- [Node.js](https://nodejs.org/) 20 or newer
- [UPX](https://upx.github.io/) (optional, for compressed release builds)

### Development

```bash
wails dev
```

### Release build

The quickest path is the included PowerShell script, which strips symbols and
compresses the result with UPX:

```powershell
./build.ps1
```

Or drive Wails directly:

```bash
wails build -clean -trimpath -ldflags "-s -w"
upx --best --lzma build/bin/name-it.exe
```

Pushing a `vX.Y.Z` tag runs these same steps in CI and publishes the archive to
GitHub Releases. The tag must match the version in
`internal/version/version.go`.

### Tests

The Go entry point embeds the built frontend, so generate it once before
running the suite:

```bash
cd frontend && npm run build && cd ..
go test ./...
```

## Usage

1. Run `name-it.exe`.
2. Click **START WORKING** and pick the folder you want to monitor.
3. Keep working. Name It will let you know what it thinks of your file names.
4. Use **PAUSE 30 MIN** when you need a break.

### Aggression levels

`POLITE`, `NORMAL`, `ANNOYING` and `UNHINGED` choose how aggressively Name It
complains. The chosen level is a floor; a sustained run of bad names pushes the
tone higher.

### Naming patterns

The **ALERT MODE** selector controls which files are judged:

| Mode | Meaning |
| --- | --- |
| `ALERT EVERY FILE` | Judge every new file. |
| `filename_vX.extension` | Only complain when the name does not match this shape. |
| `YYYY-MM-DD_filename_vX.extension` | Same, with a date prefix. |
| `CUSTOM PATTERN` | Bring your own template. |

A custom pattern may use these placeholders:

| Token | Matches |
| --- | --- |
| `filename` | any file name |
| `extension` | the file extension |
| `X` | one or more digits |
| `YYYY` | four-digit year |
| `MM` | two-digit month |
| `DD` | two-digit day |

Everything else is treated as literal text. For example,
`project_filename_vX.extension` matches `project_report_v1.txt`.

## Project structure

```
.
├── app.go                  Wails bindings and the event pipeline
├── main.go                 application entry point
├── internal
│   ├── config              session settings
│   ├── history             rolling list of recent crimes
│   ├── pattern             naming-template matching
│   ├── reminder            the message pools
│   ├── scorer              filename scoring
│   ├── session             session lifecycle and file records
│   ├── updater             GitHub release check
│   ├── version             version constant
│   └── watcher             fsnotify wrapper
├── frontend                vanilla HTML/CSS/JS interface
├── build.ps1               local stripped + UPX release build
└── .github/workflows       CI and release automation
```

## Contributing

Bug reports and pull requests are welcome. Please run `go test ./...` and
`go vet ./...` before opening a pull request, and keep the desktop-only,
offline nature of the project intact.

## License

Name It is released under the terms of the **GNU General Public License,
version 3 or (at your option) any later version**. See [LICENSE](LICENSE) for the
full text.

Copyright (C) 2026 Vinetwigs.
