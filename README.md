# Tourminal

[![CI](https://github.com/jillesme/tourminal/actions/workflows/ci.yml/badge.svg)](https://github.com/jillesme/tourminal/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Follow and create [Microsoft CodeTour](https://github.com/microsoft/codetour)
walkthroughs from the terminal. Tourminal combines syntax-highlighted source,
rendered Markdown explanations, and keyboard navigation in a single TUI.

## Install

Install the latest source with Go:

```sh
go install github.com/jillesme/tourminal/cmd/tourminal@latest
```

Tagged releases provide macOS and Linux binaries for Intel and Apple/ARM
machines. A Homebrew tap is planned for the first release; its install command
will be:

```sh
brew install jillesme/tap/tourminal
```

To build the current checkout:

```sh
go build -o tourminal ./cmd/tourminal
```

## Follow a tour

Run Tourminal from a repository that contains CodeTours, or pass its path:

```sh
tourminal
tourminal /path/to/repository
```

Open a specific tour or begin at a particular step:

```sh
tourminal --tour .tours/intro.tour
tourminal --tour .tours/intro.tour --step 2
```

Inside the player, use `n`/`p` for next/previous, `g` to choose a step,
arrow keys or `j`/`k` to scroll, `?` for help, and `q` to quit. For file
steps, the Markdown description appears immediately above the target line.

Tourminal discovers tours in the standard locations:

- `.tour`, `main.tour`, and `.vscode/main.tour`
- `.tours/**`, `.vscode/tours/**`, and `.github/tours/**`

## Create a tour with an agent

Tourminal ships its authoring instructions inside the binary. Ask a coding
agent to load them before creating or editing a tour:

```text
Create a CodeTour for the request lifecycle. Run `tourminal skill` first and
follow the instructions it prints. Validate the finished tour with Tourminal.
```

The agent can load the skill with:

```sh
tourminal skill
```

`tourminal --skill` is a convenience alias. The skill tells the agent how to
inspect the codebase, write useful descriptions, choose unique regex anchors
for living tours, or pin a snapshot tour to `git rev-parse HEAD` and use exact
line/selection anchors.

## Validate tours

Validation parses the schema and resolves every location against the current
workspace. It checks file and directory existence, path containment, line and
selection bounds, regex validity and uniqueness, absolute URIs, `when`
expressions, pinned Git refs, and `nextTour` links.

```sh
tourminal validate .
tourminal validate .tours/intro.tour
```

This command is suitable for CI and exits non-zero when a tour is invalid.

## Commands

| Command | Purpose |
| --- | --- |
| `tourminal [workspace]` | Discover and follow tours |
| `tourminal --tour FILE` | Follow one `.tour` file |
| `tourminal validate [PATH]` | Strictly validate a workspace or tour |
| `tourminal skill` | Print the bundled AI authoring skill |
| `tourminal version` | Print build version information |

Use `tourminal --help` for all flags.

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
go build ./cmd/tourminal
```

Releases are built by GoReleaser when a `v*` tag is pushed. CI tests macOS and
Linux and cross-builds all supported OS/architecture pairs.

## Project status

Tourminal is an independent, early-stage project. It reads the open CodeTour
format but is not affiliated with, endorsed by, or maintained by Microsoft.
“CodeTour” and “Visual Studio Code” are names of their respective Microsoft
projects.

Tourminal is available under the [MIT License](LICENSE).
