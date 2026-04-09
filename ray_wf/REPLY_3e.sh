#!/bin/sh
# REPLY_3e.sh — strip "---" lines from LLM response and append to ephemeral conversation

id="$RAYMOND_RESULT"
response="tmp/response_${id}.md"
conversation="ephemeral/${id}.md"

timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

printf '\n\n---\nagent: %s\n' "$timestamp" >> "$conversation"
grep -v '^[[:space:]]*---[[:space:]]*$' "$response" >> "$conversation"

rm -f "$response" "tmp/${id}.md"

# Since it was launched with a fork we need to terminate, not loop
echo "<result>REPLY DONE</result>"