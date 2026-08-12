---
name: create-codetour
description: Create and edit Microsoft CodeTour .tour walkthroughs by inspecting a repository, choosing stable code anchors, writing useful Markdown step descriptions, and validating the result. Use when asked to create, author, generate, revise, or repair a CodeTour or a guided codebase walkthrough.
---

# Create a CodeTour

Create a reader-first walkthrough that explains how a codebase works. Write a standard CodeTour JSON file that both the terminal player and the VS Code extension can follow.

## Workflow

1. Inspect the repository before writing.
   - Read the relevant README, contributor guidance, entry points, implementation, and tests.
   - Trace the actual control or data flow for the requested topic.
   - Infer the audience and learning goal from the request and repository. Ask only when the missing choice would materially change the tour.

2. Design a short narrative.
   - Start with a content-only orientation step when readers need context.
   - Follow the concept in a useful order, such as entry point → orchestration → core logic → boundary or test.
   - Prefer 4–10 purposeful steps for a focused tour. Do not annotate every function.
   - Make each step explain why the location matters and how it connects to adjacent steps; do not merely restate the code.

3. Choose a versioning and anchoring strategy.
   - For a living tour that should follow the changing default branch, omit `ref` and prefer `pattern` anchors.
   - For a snapshot tour, first require a clean working tree so the checked-out files match the commit. Run `git rev-parse HEAD`, set the top-level `ref` to that full commit SHA, and prefer `line` or `selection` anchors.
   - If the working tree has relevant uncommitted changes, do not pin the tour to `HEAD`; either commit them first or create an unpinned tour with patterns.
   - Use workspace-relative, forward-slash paths.
   - Prefer `pattern` for a distinctive named declaration or other unique, stable text. Confirm it matches exactly once.
   - Prefer `line` when the tour is pinned to an exact commit, or when the exact ordinal location is important and no robust pattern exists.
   - Never set both `line` and `pattern`; CodeTour ignores `pattern` when `line` exists.
   - Use one location kind per step: `file`, `directory`, `uri`, or none for a content-only step.
   - Use `selection` only to emphasize a meaningful range within a file. All positions are 1-based.

4. Write `.tours/<kebab-case-name>.tour` as formatted JSON.
   - Include `"$schema": "https://aka.ms/codetour-schema"`.
   - Include required top-level `title` and `steps` fields.
   - Begin each step description with a concise Markdown heading such as `### Request routing`.
   - Keep descriptions compact, specific, and understandable without prior tribal knowledge.
   - Link related workspace files with Markdown paths such as `[router](./internal/http/router.go)`.
   - Add `commands` only when the user explicitly wants an interactive tutorial. Never execute tour commands while authoring or validating.

5. Verify every anchor and validate the result.
   - Confirm every referenced file and directory exists inside the workspace.
   - Compile each pattern mentally and search the target file to confirm it is unique. Use portable regular expressions without lookaround or backreferences.
   - Re-read the steps in order and remove jumps, repetition, and implementation trivia.
   - Run `tour validate .tours/<name>.tour`.
   - Fix every validation or resolution error before finishing. If practical, open it with `tour --tour .tours/<name>.tour` for a final visual check.

6. Report the created path, step count, subject covered, and validation result.

## Editing an Existing Tour

Preserve its title, intent, and useful ordering unless the user asks for a redesign. Re-inspect changed code, repair stale anchors, improve unclear descriptions, and run the same validation workflow. Do not silently discard optional metadata such as `ref`, `isPrimary`, `nextTour`, or `when`.

## Schema Details

Read [references/tour-schema.md](references/tour-schema.md) before writing or substantially editing a tour. It contains the supported fields, location rules, and a complete example.
