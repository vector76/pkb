---
allowed_transitions:
  - { tag: reset, target: WAIT }
---

Look at all the new or changed files in wiki, conversations, and attachments.  
These need to be committed and pushed to the remote repository.  The log.md 
file in the repo root should also be committed.

Use 'git add' to stage all the appropriate files.  Commit with a descriptive 
message.  Do not mention Claude as coauthor or contributor.  Push with 
'git push'.

If push fails, then create a new markdown file in issues/ and describe the 
problem.  Be sure not to overwrite an existing file.  Write one new file in 
that folder and do not touch any existing files.
