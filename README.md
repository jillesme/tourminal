# Tourminal

[![CI](https://github.com/jillesme/tourminal/actions/workflows/ci.yml/badge.svg)](https://github.com/jillesme/tourminal/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Follow and create [Microsoft CodeTour](https://github.com/microsoft/codetour)
walkthroughs from the terminal. Tourminal combines syntax-highlighted source,
rendered Markdown explanations, and keyboard navigation in a single TUI.

## Install

### Homebrew

```sh
brew install jillesme/tap/tourminal
```

This installs the preferred `tour` command. The original `tourminal` command
remains available as a compatibility alias.

### Release binaries

Download a ready-to-run archive from the
[latest release](https://github.com/jillesme/tourminal/releases/latest). Builds
are provided for macOS and Linux on Intel and ARM, with both executable names
and SHA-256 checksums.

## Follow a tour

Run Tourminal from a repository that contains CodeTours, or pass its path:

```sh
tour
tour /path/to/repository
```

Open a specific tour or begin at a particular step:

```sh
tour --tour .tours/intro.tour
tour --tour .tours/intro.tour --step 2
```

Inside the player, use `n`/`p` for next/previous, `g` to choose a step,
arrow keys or `j`/`k` to scroll, `?` for help, and `q` to quit. For file
steps, the Markdown description appears immediately above the target line.

Tourminal detects light and dark terminal backgrounds automatically. If a
terminal cannot report its background reliably, select one explicitly with
`tour --theme light`, `tour --theme dark`, or the
`TOURMINAL_THEME` environment variable.

Tourminal discovers tours in the standard locations:

- `.tour`, `main.tour`, and `.vscode/main.tour`
- `.tours/**`, `.vscode/tours/**`, and `.github/tours/**`

## Create a tour with an agent

Tourminal ships its authoring instructions inside the binary. Ask a coding
agent to load them before creating or editing a tour:

```text
Create a CodeTour for the request lifecycle. Run `tour skill` first and
follow the instructions it prints. Validate the finished tour with Tourminal.
```

The agent can load the skill with:

```sh
tour skill
```

`tour --skill` is a convenience alias. The skill tells the agent how to
inspect the codebase, write useful descriptions, choose unique regex anchors
for living tours, or pin a snapshot tour to `git rev-parse HEAD` and use exact
line/selection anchors.

## Validate tours

Validation parses the schema and resolves every location against the current
workspace. It checks file and directory existence, path containment, line and
selection bounds, regex validity and uniqueness, absolute URIs, `when`
expressions, pinned Git refs, and `nextTour` links.

```sh
tour validate .
tour validate .tours/intro.tour
```

This command is suitable for CI and exits non-zero when a tour is invalid.

## Commands

| Command | Purpose |
| --- | --- |
| `tour [workspace]` | Discover and follow tours |
| `tour --tour FILE` | Follow one `.tour` file |
| `tour validate [PATH]` | Strictly validate a workspace or tour |
| `tour skill` | Print the bundled AI authoring skill |
| `tour version` | Print build version information |

Use `tour --help` for all flags.

## Compatibility and safety

The MVP supports content, file/line, selection, pattern, directory,
embedded-content, and URI steps; Markdown descriptions; syntax highlighting;
platform `when` expressions; marker-based and linked tours; and standard tour
locations. Images use their Markdown fallback, and VS Code-specific views are
informational.

Tour-provided commands are displayed but never executed. Source paths cannot
escape the selected workspace, source/tour sizes are bounded, binary sources
are rejected, and control characters from repository content are sanitized
before terminal rendering. Remote tour fetching and interactive tour editing
are not part of this release.

## Development

Tourminal requires the Go version declared in `go.mod`.

```sh
go test -race ./...
go vet ./...
go build -o tour ./cmd/tourminal
```

Releases are built by GoReleaser when a `v*` tag is pushed. CI tests macOS and
Linux and cross-builds all supported OS/architecture pairs.

## Project status

Tourminal is an independent, early-stage project. It reads the open CodeTour
format but is not affiliated with, endorsed by, or maintained by Microsoft.
“CodeTour” and “Visual Studio Code” are names of their respective Microsoft
projects.

Tourminal is available under the [MIT License](LICENSE).
