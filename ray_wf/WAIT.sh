#!/bin/sh
# WAIT.sh — poll for signal files and print the matching state transition
# Intended to run from the knowledge base root.

outer=0
while [ "$outer" -lt 50 ]; do
  touch heartbeat.md

  inner=0
  while [ "$inner" -lt 10 ]; do
    # Reply signals: queue/reply/<conversation-id>
    for f in queue/reply/*; do
      [ -e "$f" ] || continue
      id=$(basename "$f")
      rm "$f"
      echo "<fork input=\"$id\" next=\"WAIT\">REPLY_1</fork>"
      exit 0
    done

    # Ingest signals: queue/ingest/<conversation-id>
    for f in queue/ingest/*; do
      [ -e "$f" ] || continue
      id=$(basename "$f")
      rm "$f"
      rm heartbeat.md
      echo "<goto input=\"$id\">INGEST</goto>"
      exit 0
    done

    # Lint signal
    if [ -f queue/lint ]; then
      rm queue/lint
      rm heartbeat.md
      echo "<goto>LINT</goto>"
      exit 0
    fi

    # Commit signal
    if [ -f queue/commit ]; then
      rm queue/commit
      rm heartbeat.md
      echo "<goto>COMMIT</goto>"
      exit 0
    fi

    sleep 1
    inner=$((inner + 1))
  done

  outer=$((outer + 1))
done

echo "<reset>WAIT</reset>"
