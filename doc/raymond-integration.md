# Raymond Integration

PKB and the Raymond workflow coordinate through signal files in the `queue/` directory of the knowledge base. Both processes have direct access to the knowledge base filesystem; the signal files exist purely to communicate intent and turn ownership, not to transfer data.

## Signal File Protocol

A signal file's presence means work is pending. Raymond detects it, does the work, and deletes the file. PKB creates signal files in response to human actions and watches for file changes to know when Raymond has completed work.

### Reply signals

Path: `queue/reply/<conversation-id>`

Created by PKB when the human submits a conversation turn. The filename is the conversation's identifier (the filename without extension from `conversations/` or `ephemeral/`).

Raymond's reply workflow:
1. Detects a file in `queue/reply/`.
2. Reads the full conversation from `conversations/<id>.md` (or `ephemeral/<id>.md`).
3. Generates an agent response.
4. Appends the agent turn to the conversation file (with `---` delimiter and `agent:` tag).
5. Deletes `queue/reply/<id>`.

PKB's file watcher detects the conversation file change and pushes an SSE event to the browser.

### Ingest signals

Path: `queue/ingest/<conversation-id>`

Created by PKB when the human queues a conversation for ingest (from either `conversations/` or `ephemeral/`). If the conversation is in `ephemeral/`, PKB moves it to `conversations/` before creating the signal.

Raymond's ingest workflow:
1. Detects one or more files in `queue/ingest/`.
2. For each: reads the conversation and any attachments it references.
3. Updates relevant wiki pages: creating new pages, extending existing ones, flagging contradictions.
4. Appends an entry to `log.md`.
5. Deletes the signal file.

### Lint signal

Path: `queue/lint`

Created by PKB when the human triggers a lint run. Content is ignored; presence is the signal.

Raymond's lint workflow:
1. Detects `queue/lint`.
2. Audits the wiki: orphaned pages, missing cross-links, contradictions between pages, gaps in coverage.
3. Produces corrections or flags issues (implementation of how findings are presented is up to the workflow design).
4. Appends to `log.md`.
5. Deletes `queue/lint`.

### Commit signal

Path: `queue/commit`

Created by PKB when the human triggers a commit. Content is ignored.

Raymond's commit workflow:
1. Detects `queue/commit`.
2. Reviews what has changed since the last commit.
3. Generates a commit message summarizing the changes.
4. Stages and commits all knowledge base changes.
5. Deletes `queue/commit`.

## Polling

Raymond workflow shell scripts detect signal files by polling — checking for file existence in a loop with a short sleep. This approach works on all platforms without requiring OS-specific file watching utilities. One second polling intervals are sufficient given the latency involved in human interaction.

Example shell pattern used in a Raymond workflow state:

    while [ ! -f "queue/lint" ]; do
      sleep 1
    done

The Raymond workflow is designed to handle multiple signal types within a single workflow by checking each queue directory in turn, or by forking agents for each type. Parallel agents (e.g. a reply agent and a lint agent) do not conflict because they operate on different parts of the knowledge base.

## PKB File Watching

PKB uses the OS file watching facilities available in Go to detect changes made by Raymond. When a conversation file is modified (agent turn appended), PKB pushes an SSE event to any browser clients viewing that conversation. When wiki files change, PKB can invalidate cached renders.

File watching is an optimization for responsiveness. The browser can also poll if SSE is unavailable.

## Raymond Liveness

PKB does not actively monitor whether Raymond is running. If the human submits a turn and no agent response appears after a reasonable time, the lack of response is itself the signal. The UI may indicate "waiting for agent" based on the presence of a signal file with no corresponding response yet. Formal liveness detection can be added later if needed.

## What Raymond Owns

Raymond is the sole writer of:
- Agent turns in conversation files
- All files in `wiki/`
- `log.md`
- Git commits

PKB is the sole writer of:
- Human turns in conversation files
- Files in `attachments/`
- Signal files in `queue/`

This division ensures clear provenance: every line in the wiki came from a Raymond workflow invocation, and every human turn came from the PKB server relaying what the human typed.
