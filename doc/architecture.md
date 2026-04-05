# Architecture

## Runtime Overview

```
Browser  ──HTTP/SSE──►  PKB Server  ◄──files──►  Knowledge Base (git repo)
                                                         ▲
                                                         │ files
                                                   Raymond Workflow
```

PKB server and the Raymond workflow both operate directly on the files in the knowledge base directory. There is no API between them for data transfer — the filesystem is the shared store. Coordination (whose turn is it, what work is pending) happens through a small set of signal files described in [raymond-integration.md](raymond-integration.md).

## PKB Server

PKB is a Go binary invoked from within a knowledge base directory:

```
pkb                     # serve the repo in the current directory
pkb -C /path/to/repo    # serve a repo elsewhere (like git -C)
```

The server's responsibilities:

- **Render and serve** markdown files from the knowledge base as HTML on demand. No pre-generation; every page render is fresh from the file.
- **Navigation UI** — Wikipedia-style: pages link to other pages. The entry point is `wiki/index.md`. Additional views include a chronological conversation list and full-text search.
- **Conversation UI** — a chat-like interface for creating and continuing conversations. The server writes the human turn to the conversation file and creates a signal file to wake Raymond.
- **Attachment handling** — accepts file uploads, writes them to `attachments/`, and injects a reference into the current conversation turn.
- **Operation triggers** — buttons for Ingest, Lint, and Commit. Each creates a signal file that causes the corresponding Raymond workflow to proceed.
- **Real-time updates** — watches the knowledge base for file changes and pushes updates to connected browsers via SSE. When Raymond writes an agent response, the browser refreshes without a manual reload.
- **Read-only mode** — if Raymond is not running (detected by absence of recent activity on signal files), conversation submission and operation triggers are disabled. The human can still browse and search.

The server does not write to the wiki. It does not run git commands. It does not direct the LLM. All of that belongs to Raymond.

## Raymond Workflow

Raymond runs as a separate long-running process. It is launched independently of PKB. The workflow:

- Watches for signal files in the `queue/` directory of the knowledge base.
- When a signal appears, invokes the appropriate LLM-driven workflow (reply, ingest, lint, commit).
- Reads and writes knowledge base files directly — conversations (agent turns only), wiki pages, log, attachments.
- Deletes signal files when work is complete.

Raymond handles all git operations: staging, commit message generation, and committing. The workflow is responsible for producing coherent, attributed commits.

See [raymond-integration.md](raymond-integration.md) for the signal file protocol.

## Knowledge Base

The knowledge base is an ordinary git repository containing markdown files and attachments. PKB and Raymond both treat it as their working directory. Structure is described in [kb-structure.md](kb-structure.md).

## Data Flow: Conversation Turn

1. Human types a message and submits via the browser.
2. PKB server appends the human turn to the conversation file (with turn delimiter).
3. PKB server creates `queue/reply/<conversation-id>`.
4. Raymond workflow detects the signal file, reads the conversation, generates a response, appends the agent turn to the conversation file, deletes the signal file.
5. PKB server's file watcher detects the conversation file change, pushes an SSE event to the browser.
6. Browser refreshes the conversation view.

## Data Flow: Ingest

1. Human clicks Ingest on one or more conversations.
2. PKB server creates `queue/ingest/<conversation-id>` for each.
3. Raymond ingest workflow processes each conversation: reads the conversation and any referenced attachments, updates relevant wiki pages, appends to `log.md`, deletes signal files.
4. PKB server detects wiki file changes, updates browser if viewing affected pages.

## Data Flow: Commit

1. Human clicks Commit in the UI.
2. PKB server creates `queue/commit`.
3. Raymond commit workflow stages changes, generates a commit message by reviewing what changed, commits.
4. Signal file deleted on completion.

## Design Decisions

**No CLI for Raymond coordination.** Unlike the backlog manager pattern (which uses a CLI with blocking HTTP calls for multi-client work distribution), PKB uses plain files. Both processes share a filesystem, so there is no reason to introduce a network boundary for coordination. Signal file presence means work is pending; absence means it is done.

**No database.** The knowledge base is self-contained markdown. This makes it portable, human-readable, git-friendly, and independent of any runtime state. The server and Raymond reconstruct all necessary context from the files themselves.

**Raymond owns git.** The server never runs git commands. This keeps the commit history clean and attributed to LLM-driven operations, and allows the commit workflow to generate meaningful messages and handle edge cases (like unexpected remote changes).

**Ephemeral conversations.** Conversations begin as ephemeral and are not committed or indexed until the human explicitly queues them for ingest. This prevents noise from accumulating in the knowledge base.
