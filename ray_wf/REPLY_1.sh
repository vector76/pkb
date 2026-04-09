#!/bin/sh
# REPLY_1.sh — copy conversation to tmp/ for safe editing, then hand off to REPLY_2

id="$RAYMOND_RESULT"
dst="tmp/${id}.md"
mkdir -p tmp

if [ -f "conversations/${id}.md" ]; then
  cp "conversations/${id}.md" "$dst"
  echo "<goto input=\"$id\">REPLY_2</goto>"
elif [ -f "ephemeral/${id}.md" ]; then
  cp "ephemeral/${id}.md" "$dst"
  echo "<goto input=\"$id\">REPLY_2e</goto>"
fi

