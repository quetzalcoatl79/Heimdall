---
name: fix
description: Diagnostic et correction autonome d'un bug
user-invocable: true
allowed-tools: Read, Glob, Grep, Edit, Write, Bash(npm run *), Bash(npx *), Bash(git diff *)
argument-hint: [description-du-bug]
---

# Bug Fix Autonome

Diagnostique et corrige le bug décrit : `$ARGUMENTS`

## Processus

1. **Reproduire** — Comprendre le bug, localiser le code impacté
2. **Diagnostiquer** — Trouver la root cause (pas juste le symptôme)
3. **Corriger** — Appliquer le fix minimal et ciblé
4. **Vérifier** — S'assurer que le fix fonctionne et n'introduit pas de régression
5. **Documenter** — Résumer ce qui a été trouvé et corrigé

## Règles
- Fix minimal : ne toucher que ce qui est nécessaire
- Ne pas refactorer le code autour du bug
- Si le fix nécessite un changement d'architecture, s'arrêter et demander confirmation
- Mettre à jour `tasks/issues.md` avec le pattern si c'est un bug récurrent

## Format de sortie
```
## Bug Fix: [titre court]

### Root Cause
[Explication concise]

### Fix appliqué
[Description des changements]

### Fichiers modifiés
- path/to/file.ts : description du changement

### Vérification
[Comment vérifier que le fix fonctionne]
```
