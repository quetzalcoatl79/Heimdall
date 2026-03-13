#!/bin/bash
# Hook: PreToolUse — Protège les fichiers sensibles contre l'édition
INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // .tool_input.filePath // empty')

# Fichiers protégés
PROTECTED=(".env" ".env.local" ".env.production" "package-lock.json" ".git/" ".claude/settings.json" "prisma/migrations/")

for pattern in "${PROTECTED[@]}"; do
  if [[ "$FILE_PATH" == *"$pattern"* ]]; then
    echo "BLOCKED: '$FILE_PATH' est un fichier protégé ($pattern)" >&2
    exit 2
  fi
done

exit 0
