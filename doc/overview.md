# PKB — Personal Knowledge Base

## Purpose

PKB is a tool for building and maintaining a personal knowledge base using an LLM as the synthesis engine. The human provides raw inputs — conversations and attachments. A Raymond workflow processes those inputs and maintains a structured, cross-linked wiki. The result is a compounding artifact: each new input enriches an existing body of knowledge rather than accumulating in an undifferentiated pile.

The idea is inspired by Andrej Karpathy's LLM Wiki pattern. The key insight is that synthesis should happen once and persist, rather than being reconstructed on every query. The wiki becomes the distilled, navigable representation of everything the human has fed into the system.

## Design Goals

- **Human-authored foundation.** Conversations and attachments are written by the human and are never modified by the LLM. They are the permanent record of what the human said and provided.
- **LLM-maintained wiki.** The wiki is generated and maintained by Raymond workflows. The human does not edit wiki pages directly — corrections happen through conversation.
- **Provenance via git.** The knowledge base is a git repository. Git history and blame provide a record of how wiki pages came to say what they say.
- **File-based, portable.** The knowledge base is a directory of markdown files. No database. No proprietary format. The files are readable and editable by humans outside the tool.
- **One PKB binary, many knowledge bases.** PKB is a standalone binary that operates on any knowledge base directory. Each person's knowledge base is their own git repository.

## Components

Three processes collaborate at runtime:

- **PKB server** — a Go HTTP server that runs inside the knowledge base directory. It renders markdown as HTML, serves the web UI, handles file uploads, manages the signal files that coordinate with Raymond, and pushes real-time updates to the browser.
- **Raymond workflow** — a separate process that runs the LLM-driven workflows: responding to conversation turns, ingesting conversations into the wiki, linting the wiki, and committing changes to git. Raymond has direct access to the knowledge base files.
- **Browser** — the user interface. The human reads wiki pages, conducts conversations, uploads attachments, and triggers operations like ingest, lint, and commit.

## What PKB Is Not

PKB is not a RAG system. There is no vector database, no embedding, no retrieval at query time. The wiki is the retrieval artifact — the LLM has already synthesized the raw inputs into navigable pages. When the human asks a question, the LLM reads the wiki.

PKB is not a multi-user or multi-client system. It is designed for a single human and a single Raymond workflow instance operating on a single knowledge base.
