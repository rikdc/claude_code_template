---
name: changelog
description: Creates and maintains CHANGELOG.md in Keep a Changelog format, including adding entries and cutting releases. Use when the user asks to update the changelog, add a changelog entry, or record a release.
user-invocable: true
argument-hint: "[--create] [--add-entry \"description\" --type TYPE] [--release X.Y.Z]"
allowed-tools: Read, Write(CHANGELOG.md), Edit(CHANGELOG.md), Bash(date:*), Bash(git add:*), Bash(git status:*), Bash(git commit:*), Bash(git log:*)
---

# Changelog - Keep a Changelog Maintenance

Maintains CHANGELOG.md in [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
format with [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Usage

- `/git-workflow:changelog --create` - Create CHANGELOG.md
- `/git-workflow:changelog --add-entry "description" --type TYPE` - Add an entry
- `/git-workflow:changelog --release X.Y.Z` - Cut a release

With no flags, infer the operation from the request. Categories: `added`,
`changed`, `deprecated`, `removed`, `fixed`, `security`.

## Skeleton

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.3] - 2024-01-15

### Added

- User authentication
```

Category headings exist only where they have entries. Never write a heading
with a placeholder bullet under it.

## Operation: --create

1. Read CHANGELOG.md. If it exists, stop and report it — do not overwrite.
2. Write the skeleton above with an empty `## [Unreleased]` and no version
   sections.
3. Offer to seed it from `git log`, but do not do so unasked.

## Operation: --add-entry

1. Read CHANGELOG.md; run `--create` first if it is missing.
2. Reject a `--type` outside the six categories.
3. Append the entry as a bullet under `## [Unreleased]` and its `### <Type>`
   heading, adding that heading if absent.
4. Preserve the existing order of categories and the file's surrounding
   formatting.

## Operation: --release X.Y.Z

Check before writing, and stop with the reason if any fails:

- `## [Unreleased]` has at least one entry — nothing to release otherwise.
- `X.Y.Z` does not already appear in the file.
- `X.Y.Z` is higher than the most recent version present.

Then:

1. Get today's date with `date +%F`. Do not guess it.
2. Rename `## [Unreleased]` to `## [X.Y.Z] - YYYY-MM-DD`, keeping its entries.
3. Insert a fresh empty `## [Unreleased]` above it.
4. If the file uses link references at the bottom, add one for `X.Y.Z` and
   repoint `[Unreleased]` to compare against the new tag.

Do not tag, commit, or push unless asked.

## Entry Style

- One line per entry, describing a user-facing change.
- Lead with a verb: "Add session refresh endpoint", not "Added a new endpoint
  that lets users refresh sessions".
- No commit hashes, file paths, or internal refactors that users cannot observe.
