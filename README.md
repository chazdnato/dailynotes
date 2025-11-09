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
  --dry-run - Show what would be created without writing files
  --force - Overwrite existing file without prompting (default: ask)
  --list - List all daily note files in the directory, with some task stats
  --template string - Path to template file (uses default if not specified)

```

### Template format

The default format of the template is as follows:

```markdown
# {{.Date}}

## Useful Links

## Read, Attend or Watch

## Doing

- [ ]

## Longer Term

- [ ]
```

The import piece is `{{.Date}}` as this is parsed by the program to generate the date stamp. The rest is entirely up to you!

## Development Setup

This project uses [mise](https://mise.jdx.dev/) for tool version management. Or just have the right `golang` installed.

```bash
# Install mise
curl https://mise.run | sh

# Install project tools
./scripts/prepare-dev-env.sh

# Build
go build -o dailynotes

# Run tests
go test ./...
```
