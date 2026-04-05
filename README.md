# PKB — Personal Knowledge Base

PKB is a web-based tool for building a personal knowledge base with an LLM as the synthesis engine. You have conversations, and a background agent (Raymond) distills them into a structured, cross-linked wiki. The result is a compounding body of knowledge rather than an undifferentiated pile of chat logs.

Everything is plain markdown files in a git repository. No database.

## Install

**Prerequisites:** Go 1.22+

### From anywhere (no checkout needed)

```sh
go install github.com/vector76/pkb@latest
```

This puts the `pkb` binary in `$GOPATH/bin`.

### From a local checkout

```sh
git clone https://github.com/vector76/pkb.git
cd pkb
go install .
```

### Pre-built binaries

Download from the [Releases](https://github.com/vector76/pkb/releases) page.

## Quick Start

```sh
# Run against the current directory (creates the KB structure if needed)
pkb

# Or specify a directory and listen address
pkb -C /path/to/my-kb -addr 127.0.0.1:4242
```

Then open http://127.0.0.1:4242 in your browser.

PKB creates the directory structure on first run:

```
my-kb/
  wiki/           # LLM-maintained wiki pages (start at wiki/index.md)
  conversations/  # Persistent conversation logs
  ephemeral/      # Temporary conversations (promote to persistent when ready)
  attachments/    # Uploaded files
  queue/          # Signal files for coordinating with Raymond
```

## Usage

- **Wiki** — Browse the knowledge base starting from the index page.
- **Conversations** — Start a new conversation, exchange messages with the agent, then promote and ingest it into the wiki.
- **Search** — Full-text search across wiki pages and conversations.
- **Lint / Commit** — Trigger Raymond to check wiki consistency or commit changes to git.

Raymond (the LLM agent) is a separate process. Without it running, PKB operates in read-only mode — you can browse the wiki and read conversations, but message submission and operations are disabled.

## Documentation

See [doc/](doc/README.md) for design documents covering architecture, file formats, and the Raymond integration protocol.
