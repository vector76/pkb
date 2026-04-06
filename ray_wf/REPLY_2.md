---
allowed_transitions:
  - { tag: goto, target: REPLY_3, input: "{{result}}" }
---

There is a conversation history here: `tmp/{{result}}.md`

You are to continue the conversation, but *DO NOT WRITE TO THE FILE*.  Any
changes you make will be lost.

For context, you may refer to wiki/, which is the knowledge base and may
contain background for this conversation.  You should generally avoid referring
to conversations/ with the exception of cases where wiki/ documents explicitly
refer to conversations and those conversations are necessary.  This should be
very rare.

Instead, your response is to be written to `tmp/response_{{result}}.md`.  You
should include only your response turn and not the entire extended document.
(This will be post-processed into the conversation after applying appropriate
constraints).

Respond with markdown in the response file.  Avoid using hr ("---") in your
response because any occurrences will be stripped out from your response when
adding to the conversation.
