# dailynotes

This will create a markdown file in the format `YYYY-MM-DD.md` in the directory where the utility is run. This will look in that
directory for the most recent daily note file, and extract any incomplete tasks and merge them into the new file.

The concept was inspired by @skoretzGL's [`brain2`](https://gitlab.com/gitlab-com/gl-infra/gitlab-dedicated/sandbox/skoretzGL/brain2) project.

This was unabashedly vibe coded; any bugs are definitely not my own, but additions and improvements are welcome!

## Installation

Feel free to snag something from the releases that works for you. Or just run (assuming you have the go path in your `PATH`):

```bash
go install
```

## Usage

Basic usage is just to run the script. However, there are several ways to modify the behaviour of the script.

```bash

Usage of ./dailynotes:
  --debug - Enable debug output
  --dir string - Directory containing daily note files (default ".")
  --dry-run - Show what would be created or archived without writing/moving files
  --force - Overwrite existing file without prompting (default: ask)
  --list - List all daily note files in the directory, with some task stats
  --no-archive - Skip the automatic archive step on the main invocation
  --template string - Path to template file (uses default if not specified)

Subcommands:
  archive - Archive prior-month notes into YYYY/MM/ subdirectories.
            The subcommand must be the first argument, e.g.:
              dailynotes archive
              dailynotes archive -dry-run
              dailynotes archive -dir ~/notes

```

### Archive behavior

As months roll over, the utility automatically moves prior-month `YYYY-MM-DD.md`
files into `YYYY/MM/` subdirectories on each invocation. Current-month files stay
at the root, and the operation is idempotent.

Pass `--no-archive` to skip the automatic step, or run `dailynotes archive` as
a standalone subcommand. Collisions (a file already at the destination) are
reported and the source is left in place.

### Template format

The default format of the template is as follows:

```markdown
# {{.Date}}

## Useful Links

## Read, Attend, or Watch

## Doing

- [ ]

## Longer Term

- [ ]
```

The import piece is `{{.Date}}` as this is parsed by the program to generate the date stamp. The rest is entirely up to you!

## Development Setup

Prerequisites (install via Homebrew):

```bash
brew install go golangci-lint
```

```bash
# Build
go build -o dailynotes

# Run tests
go test ./...

# Lint
golangci-lint run
```

## TODO / Aspirational

- [ ] Releases: goreleaser cross-platform binaries on tag → GitHub Releases
- [ ] Docker: multi-arch image push to ghcr.io
- [ ] Signing: cosign keyless signing of archives + image
