<p align="center">
  <img src="build/appicon.png" width="190" alt="Name It icon">
</p>

<h1 align="center">Name It</h1>

<p align="center">
  <strong>A tiny, offline, deliberately obnoxious desktop nagger for creative file names.</strong>
</p>

<p align="center">
  <a href="https://github.com/Vinetwigs/name-it/actions/workflows/ci.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/Vinetwigs/name-it/ci.yml?branch=master&label=CI" alt="CI">
  </a>
  <a href="https://github.com/Vinetwigs/name-it/releases">
    <img src="https://img.shields.io/github/v/release/Vinetwigs/name-it" alt="Latest release">
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/github/license/Vinetwigs/name-it" alt="License">
  </a>
</p>

---

Name It watches a folder and forms opinions. Very specific opinions about the
names you give your files. The moment a new file appears it reads the name,
scores it, and depending on how badly things are going makes sure you hear
about it. It starts civil. It rarely stays civil.

It never renames, moves, deletes or reads your files. It just refuses to stay
quiet about `finale_finale_2.psd`. The renaming is still your job. That is,
more or less, the entire point.

## The pitch

Because `progetto`, `prova`, `finale`, `definitivo` and `nuovo` are not
versioning strategies. Future-you would rather find `report_v3.txt` than spend
an afternoon on an archaeological dig through a folder full of `def_ok_vero`.

Name It is the nagging voice you would have if you wrote a to-do list and then
ignored it for six months. Offline, no account, no cloud, no telemetry. It
judges locally and it judges hard.

## Features

- Watches one folder per working session, including subfolders created later.
- Deterministic filename scoring with six escalating severity levels:
  `ignore`, `gentle`, `normal`, `passive`, `aggressive` and `maximum`.
- A full-window interruption and a native OS notification when your naming
  habits deserve attention.
- Naming pattern alerts, so it only complains when a file breaks your chosen
  convention.
- Session statistics, a crime log and a 30-minute pause button for when you
  need to make poor decisions in private.
- Optional startup update check against GitHub Releases, announced with an
  in-app modal.
- Vanilla Go and vanilla JavaScript. No database, no cloud, no AI.

## Download

Grab the compressed archive for your platform from the
[Releases](https://github.com/Vinetwigs/name-it/releases) page:

| Platform | Archive | Notes |
| --- | --- | --- |
| Windows | `name-it-*-windows-amd64.zip` | No installer. Unzip and run `name-it.exe`. |
| Linux | `name-it-*-linux-amd64.tar.gz` | Needs WebKit2GTK 4.1 at runtime. |
| macOS | `name-it-*-macos-universal.zip` | Universal build (Intel + Apple Silicon). |

Windows and Linux binaries are stripped and packed with
[UPX](https://upx.github.io/) using its strongest practical settings. The macOS
bundle is left unpadded so it stays friendly with Gatekeeper.

## How it works

When you start a session, Name It takes a baseline of the folder and then
listens for files that get created. Files that already match your naming
pattern are left alone. Everything else is tokenized and scored:

- suspicious tokens (`final`, `def`, `copy`, `test`, `temp`, …) add points;
- repeated words and improvised numeric suffixes add more;
- a proper `_v1` version marker subtracts points, because that is a sign of a
  functioning adult;
- re-saving the same document over and over is treated as escalation.

The score picks a severity, the severity picks a message, and the message is
delivered with the subtlety of a car alarm.

## Build from source

### Requirements

- [Go](https://go.dev/) 1.25 or newer
- [Wails v2](https://wails.io/) (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- [Node.js](https://nodejs.org/) 20 or newer
- [UPX](https://upx.github.io/) (optional, for compressed release builds)
- Linux only: `libgtk-3-dev` and `libwebkit2gtk-4.1-dev`

### Development

```bash
wails dev
```

### Release build

On Windows, the included script strips symbols and compresses the result:

```powershell
./build.ps1
```

Or drive Wails directly:

```bash
wails build -clean -trimpath -ldflags "-s -w"
upx --best --lzma build/bin/name-it.exe
```

Pushing a `vX.Y.Z` tag builds Windows, Linux and macOS in CI and publishes
every archive to GitHub Releases. The tag must match the version in
`internal/version/version.go`.

### Tests

The Go entry point embeds the built frontend, so generate it once before
running the suite:

```bash
cd frontend && node build.mjs && cd ..
go test ./...
```

## Usage

1. Run `name-it`.
2. Click **START WORKING** and pick the folder you want to monitor.
3. Work normally. Name It will let you know what it thinks of your file names.
4. Use **PAUSE 30 MIN** when you need a moment of peace.

### Aggression levels

`POLITE`, `NORMAL`, `ANNOYING` and `UNHINGED` choose the baseline tone. The
selected level is a floor, not a ceiling: a sustained run of bad names will
drag the tone upward until Name It is essentially shouting.

### Naming patterns

The **ALERT MODE** selector controls which files get judged:

| Mode | Meaning |
| --- | --- |
| `ALERT EVERY FILE` | Judge every new file. No survivors. |
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
└── .github/workflows       CI and multi-platform release automation
```

## Contributing

Bug reports and pull requests are welcome. Please run the tests and `go vet`
before opening a pull request, and keep the desktop-only, offline nature of the
project intact. Name It should never need your data to function.

## License

Name It is released under the terms of the **GNU General Public License,
version 3 or (at your option) any later version**. See [LICENSE](LICENSE) for the
full text.

Copyright (C) 2026 Vinetwigs.

If you still name your files `final_final_2`, that is between you and your
folder. Name It is just the messenger.
