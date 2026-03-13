#!/bin/bash
# Hook: Stop — Rappel de vérification quand Claude finit de répondre
cd "$CLAUDE_PROJECT_DIR" 2>/dev/null || exit 0

UNSTAGED=$(git diff --name-only 2>/dev/null | wc -l)
UNTRACKED=$(git ls-files --others --exclude-standard 2>/dev/null | wc -l)

if [ "$UNSTAGED" -gt 0 ] || [ "$UNTRACKED" -gt 0 ]; then
  echo "Reminder: $UNSTAGED modified + $UNTRACKED untracked files pending"
fi

exit 0
