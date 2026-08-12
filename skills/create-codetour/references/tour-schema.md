# CodeTour Authoring Reference

This is a concise authoring reference for the [Microsoft CodeTour schema](https://github.com/microsoft/codetour#tour-schema).

## Tour fields

- `title` (string, required): Display name.
- `steps` (array, required): Ordered walkthrough steps.
- `description` (string): Tour summary shown in pickers and tooltips.
- `ref` (string): Git branch, tag, or commit the tour describes. Prefer the full commit SHA for an immutable snapshot; omit it for a tour that should follow changing checked-out code.
- `isPrimary` (boolean): Mark the recommended starting tour.
- `nextTour` (string): Exact title of a tour that should follow this one.
- `when` (string): JavaScript condition controlling visibility, such as `isLinux`, `isMac`, or `isWindows`.
- `stepMarker` (string): Prefix used by marker-based tours.
- `$schema` (string): Use `https://aka.ms/codetour-schema` for editor assistance.

Numbered titles such as `1 - Orientation` and `2 - Internals` form an implicit sequence. The first numbered tour is also treated as primary by CodeTour players.

## Step fields

Every step requires a non-empty Markdown `description`. A leading Markdown heading also provides a natural step label.

Location fields:

- `file` (string): Workspace-relative source path.
- `line` (number): 1-based target line in `file`.
- `pattern` (string): Regular expression searched in `file`; considered only when `line` is absent.
- `selection` (object): 1-based start/end positions with `{ "line": number, "character": number }`. Keep the start at or before the end.
- `directory` (string): Workspace-relative directory. It takes precedence over `file`, so do not combine them.
- `uri` (string): Absolute URI. Do not combine it with `file`.
- No location: Create a content-only introduction or transition.

Optional behavior and labeling:

- `title` (string): Explicit step label. Usually prefer a heading in `description` so the label and content stay together.
- `view` (string): VS Code view ID to focus, such as `explorer`, `terminal`, or an extension view ID.
- `commands` (string array): VS Code command URIs invoked when the step is visited. Treat these as executable behavior and omit them unless requested.

Hand-authored tours should point at workspace files. Embedded `contents` may appear in exported tours but is not needed for normal repository authoring.

## Versioning and anchors

Choose one strategy deliberately:

- **Living tour:** Omit `ref` and use unique `pattern` anchors. This is the default for documentation expected to evolve with the main branch.
- **Pinned snapshot:** Confirm the working tree has no relevant uncommitted changes, run `git rev-parse HEAD`, and copy the full commit SHA into the top-level `ref`. Use `line` and `selection` anchors against that exact revision.

Do not pin an uncommitted working tree to `HEAD`: the commit does not contain those edits, so its line numbers may describe different source. Commit the changes first or use the living-tour strategy. A branch name is not immutable; continue to prefer patterns when `ref` names a moving branch.

A pinned tour has this shape:

```json
{
  "$schema": "https://aka.ms/codetour-schema",
  "title": "Review commit internals",
  "ref": "<full output of git rev-parse HEAD>",
  "steps": [
    {
      "file": "internal/service.go",
      "line": 37,
      "description": "### Service boundary\n\nThis revision introduces the boundary used by the request path."
    }
  ]
}
```

Replace the angle-bracket placeholder with the actual SHA before validation.

## Anchoring rules

- Make `pattern` unique within its file. Escape backslashes for JSON: the regex `^func main\(\)` becomes `"^func main\\(\\)"` in the tour file.
- Use regular expression syntax portable across JavaScript and Go/RE2. Avoid lookaround and backreferences.
- Prefer stable declarations over formatting-sensitive full-line matches.
- Recalculate numeric lines and selections after code moves.
- Keep paths relative to the workspace root and use `/`, even on Windows.

## Example

```json
{
  "$schema": "https://aka.ms/codetour-schema",
  "title": "Request lifecycle",
  "description": "Follow an HTTP request from startup to its handler.",
  "steps": [
    {
      "description": "### Orientation\n\nThis tour follows one request through the service boundaries."
    },
    {
      "file": "cmd/server/main.go",
      "pattern": "^func main\\(\\)",
      "description": "### Process entry point\n\nStartup assembles the dependencies used by every request."
    },
    {
      "file": "internal/http/router.go",
      "line": 42,
      "selection": {
        "start": { "line": 42, "character": 1 },
        "end": { "line": 47, "character": 2 }
      },
      "description": "### Route registration\n\nThe router connects the public endpoint to the application handler."
    }
  ]
}
```
