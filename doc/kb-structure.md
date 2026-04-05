# Knowledge Base Structure

The knowledge base is a git repository containing markdown files and attachments. PKB and Raymond operate directly on these files.

## Directory Layout

```
/                        <- knowledge base root (PKB's working directory)
├── wiki/
│   ├── index.md         <- entry point for wiki navigation
│   └── *.md             <- knowledge articles, cross-linked
├── conversations/
│   └── *.md             <- persistent conversations (human + agent turns)
├── ephemeral/
│   └── *.md             <- transient conversations, not committed or indexed
├── attachments/
│   └── *                <- files uploaded by the human (PDFs, images, etc.)
├── queue/
│   ├── reply/           <- signal files: pending agent replies
│   ├── ingest/          <- signal files: conversations queued for ingest
│   ├── lint             <- signal file: lint requested
│   └── commit           <- signal file: commit requested
├── heartbeat.md         <- touched by Raymond periodically to signal liveness
└── log.md               <- append-only record of ingest and maintenance operations
```

The `queue/` directory and `heartbeat.md` are coordination artifacts used by PKB and Raymond. They are not part of the knowledge base content and should be gitignored.

## Conversation Format

Each conversation is a single markdown file. Turns are separated by a horizontal rule (`---` on a line by itself). The line immediately following the `---` identifies the author: `human:` or `agent:`. Content follows on subsequent lines.

Example conversation file:

    # Conversation Title

    ---
    human: 2026-04-04T14:32:00Z
    I've been reading about sleep and want to understand the relationship
    between REM sleep and memory consolidation.

    ---
    agent: 2026-04-04T14:32:10Z
    REM sleep plays a central role in memory consolidation, particularly for
    procedural and emotional memories...

    ---
    human: 2026-04-04T14:35:00Z
    What about declarative memory? I thought slow-wave sleep was more important
    for that.

    ---
    agent: 2026-04-04T14:35:12Z
    You're right. Slow-wave sleep (SWS) is the primary stage for consolidating
    declarative memories...

Rules:
- `---` on a line by itself is the only turn delimiter. It must not appear within human or agent content.
- The line immediately after `---` must be `human:` or `agent:`, optionally followed by an RFC3339 timestamp.
- Content begins on the line after the author tag and continues until the next `---` or end of file.
- PKB always writes a timestamp on human turns. Raymond should write a timestamp on agent turns.
- The parser is permissive: a first turn without a preceding `---` is accepted, and timestamps are optional. But all PKB-managed files will have `---` before every turn.

## Wiki Format

Wiki pages are ordinary markdown files in `wiki/`. There are no structural requirements beyond being valid markdown. Pages link to each other with standard markdown links. The `wiki/index.md` page is the entry point for navigation and should provide an overview and links to major topic areas.

Raymond is responsible for creating, updating, and cross-linking wiki pages. The server renders them as HTML on demand.

## Attachments

Attachments are files uploaded by the human — documents, images, data files, or any other binary or text content. They live in `attachments/` and are referenced from conversation turns.

When the human uploads an attachment, PKB writes it to `attachments/` and injects a markdown link into the current conversation turn. The human's message provides the context (e.g. "see this chart"). The attachment and the conversation turn are processed together during ingest.

## Ephemeral Conversations

Conversations in `ephemeral/` behave identically to those in `conversations/` during a session — the human can chat, the agent responds, attachments can be referenced — but they are not committed to git and are not queued for ingest unless the human explicitly promotes them.

Promotion moves the file from `ephemeral/` to `conversations/` and queues it for ingest. The intent is to let the human have exploratory or low-value exchanges without polluting the knowledge base.

## Log

`log.md` is an append-only file maintained by Raymond. Each ingest, lint run, and commit appends a timestamped entry describing what was processed and what changed. It provides a human-readable audit trail alongside git history.