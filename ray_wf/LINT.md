---
allowed_transitions:
  - { tag: reset, target: WAIT }
---

Scan all files in wiki, conversations, and attachments for structural or content anomalies.

Checks to perform:
- Conversations: valid frontmatter, well-formed turn structure, no truncated or duplicate entries.
- Wiki: internally consistent content, no broken internal links, no contradictory statements across pages.
- Attachments: referenced files exist; orphaned attachments are noted but not deleted.

For minor, self-correctable anomalies (e.g. trailing whitespace, missing blank lines), fix them in place.

## Escalation

Issue creation is reserved for **genuinely unresolvable** situations — not every minor anomaly. Examples that warrant escalation:
- A conversation file is so malformed that the turn structure cannot be reconstructed.
- Two wiki pages make contradictory factual claims with no clear authoritative source to defer to.
- A file is corrupted or binary where markdown is expected.

If you encounter such a situation, write a single new markdown file to `issues/` at the KB root using this format:

```
# <short one-line title>

**Workflow:** LINT
**Time:** <ISO 8601 timestamp>
**Related:** <path to file, e.g. conversations/foo.md>

<description of what was found, what was attempted, what the user should do>
```

Use a descriptive filename such as `lint-conflict-{topic}.md`. Check that the filename does not already exist in `issues/` before writing; if it does, append a short disambiguating suffix. Do not touch any existing files in `issues/`.

After writing the issue file, stop processing — do not attempt further repairs to the problematic file.
