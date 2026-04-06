---
allowed_transitions:
  - { tag: reset, target: WAIT }
---

Process the incoming reply and integrate it into the knowledge base.

Steps:
1. Read the reply source (queue entry or injected turn) and identify the conversation it belongs to.
2. Append the reply as a new attributed turn to the appropriate conversation file.
3. Run any post-ingest steps (e.g. updating wiki pages, tagging, linking) that the reply content warrants.
4. If the reply introduces new topics or substantially updates existing ones, update the relevant wiki pages.

## Escalation

Issue creation is reserved for **genuinely unresolvable** situations — not every minor problem. Examples that warrant escalation:
- The conversation file targeted by the reply is so malformed that a new turn cannot be appended safely.
- The ingest step fails in a way that cannot be retried (e.g. the file is locked, the path does not exist, or the data is irrecoverably contradictory).
- The reply references context that is completely absent from the KB and cannot be reconstructed.

If you encounter such a situation, write a single new markdown file to `issues/` at the KB root using this format:

```
# <short one-line title>

**Workflow:** REPLY
**Time:** <ISO 8601 timestamp>
**Related:** <path to file or conversation ID>

<description of what was found, what was attempted, what the user should do>
```

Use a descriptive filename such as `reply-failure-{id}.md`. Check that the filename does not already exist in `issues/` before writing; if it does, append a short disambiguating suffix. Do not touch any existing files in `issues/`.

After writing the issue file, stop processing — do not attempt to ingest or modify the conversation further.
