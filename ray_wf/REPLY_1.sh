#!/bin/sh
# REPLY_1.sh — copy conversation to tmp/ for safe editing, then hand off to REPLY_2

id="$RAYMOND_RESULT"
src="conversations/${id}.md"
dst="tmp/${id}.md"

mkdir -p tmp
cp "$src" "$dst"

echo "<goto input=\"$id\">REPLY_2</goto>"

