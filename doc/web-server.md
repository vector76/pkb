# Web Server

PKB is a Go HTTP server that runs inside a knowledge base directory. It is the human's interface to the knowledge base and the mechanism through which human actions trigger Raymond workflows.

## Startup

```
pkb                     # serve the repo in the current directory
pkb -C /path/to/repo    # serve a repo elsewhere (like git -C or make -C)
```

PKB expects to find (or be able to create) the standard knowledge base directory structure at the working directory. On first run against an empty directory it should initialize the structure.

## Rendering

All markdown rendering happens on demand — no pre-generation or caching beyond what is convenient. Each HTTP request for a wiki page or conversation reads the current file from disk and renders it to HTML.

Markdown links between pages (`[topic](other-page.md)`) are rewritten to their equivalent HTTP paths so they work as hyperlinks in the browser. Relative links within the wiki and between conversations and attachments should resolve correctly.

## Navigation

The entry point for wiki navigation is `wiki/index.md`. From there, navigation is purely hyperlink-driven: pages link to other pages, and the human follows links as they would on Wikipedia.

Additional navigation views:
- **Conversation list** — chronological list of all conversations in `conversations/`, with title and date.
- **Search** — full-text search across wiki pages and conversations. Implemented as a grep over the markdown files; no index required.
- **Ephemeral conversations** — a separate list of conversations in `ephemeral/`. Opening an ephemeral conversation shows a Promote button that moves it to `conversations/` and queues it for ingest.

## Conversation UI

Each conversation is rendered as a chat-like view with human and agent turns visually distinguished. The human turn and agent turn delimiters (`---` / `human:` / `agent:`) are not shown literally; they are used to structure the rendering.

At the bottom of an active conversation, the human can type a new message and submit it. PKB:
1. Appends the human turn to the conversation file with proper delimiters.
2. Creates `queue/reply/<conversation-id>`.
3. Shows a "waiting for agent" indicator.

When the SSE event arrives (Raymond has written the agent response), the conversation view updates.

New conversations can be started from the UI. They begin in `ephemeral/`. The human provides a title when starting the conversation.

## Attachments

The conversation UI allows file attachment. When the human attaches a file:
1. PKB writes the file to `attachments/` with a unique name.
2. PKB injects a markdown reference to the attachment into the human's current message draft.

The human sees the reference in their message before submitting. Attachments are served as static files by PKB.

## Operation Triggers

Three buttons appear in the UI:

- **Ingest** — available on any persistent conversation (or on an ephemeral conversation, which is promoted first). Creates `queue/ingest/<id>`.
- **Lint** — available globally. Creates `queue/lint`.
- **Commit** — available globally. Creates `queue/commit`.

These buttons are disabled when Raymond is not running (see liveness below).

## Real-Time Updates (SSE)

PKB maintains a file watcher on the knowledge base directory. When Raymond modifies a file, PKB pushes a server-sent event to connected browser clients. The browser uses these events to refresh the relevant view without a full page reload.

The SSE stream is a single endpoint (`/events`). Events carry a type (`conversation_updated` or `wiki_updated`) and the relative path of the changed file. The browser compares the event path against the currently-viewed page and acts only when they match: conversation views reload the turn list in place; wiki pages show a "page updated" banner prompting the human to reload. The SSE connection reconnects automatically after a short delay if it drops.

## Read-Only Mode

If Raymond does not appear to be running — heuristically, if signal files have been sitting in `queue/` without being consumed for longer than a threshold — PKB disables conversation submission and operation triggers. The human can still browse wiki pages, read conversations, and search. The threshold and detection logic can be refined over time; the initial implementation can be simple.

## Static Assets

The web UI is served as static assets embedded in the PKB binary. No separate deployment of frontend files is required.
