#!/bin/bash
# Hook: PostToolUse — Auto-format les fichiers modifiés (si prettier est disponible)
INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // .tool_input.filePath // empty')

# Ne formater que les fichiers pertinents
if [[ "$FILE_PATH" == *.ts || "$FILE_PATH" == *.tsx || "$FILE_PATH" == *.js || "$FILE_PATH" == *.jsx || "$FILE_PATH" == *.css || "$FILE_PATH" == *.json ]]; then
  if command -v npx &> /dev/null && [ -f "$CLAUDE_PROJECT_DIR/node_modules/.bin/prettier" ]; then
    npx prettier --write "$FILE_PATH" 2>/dev/null
  fi
fi

exit 0
