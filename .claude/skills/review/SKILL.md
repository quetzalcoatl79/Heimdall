---
name: review
description: Code review complet du code modifié ou d'un fichier spécifique
user-invocable: true
allowed-tools: Read, Glob, Grep, Bash(git diff *), Bash(git log *)
model: sonnet
argument-hint: [file-or-branch]
---

# Code Review

Effectue une revue de code rigoureuse sur les changements ou le fichier spécifié.

## Contexte
- Branche actuelle : !`git rev-parse --abbrev-ref HEAD`
- Fichiers modifiés : !`git diff --name-only HEAD~1 2>/dev/null || git diff --name-only`

## Instructions

1. **Analyser** les changements (diff si pas de fichier spécifié, sinon le fichier `$ARGUMENTS`)
2. **Vérifier** selon ces critères :
   - Bugs potentiels ou edge cases manqués
   - Vulnérabilités de sécurité (XSS, injection, secrets exposés)
   - Performance (N+1 queries, re-renders inutiles, imports lourds)
   - TypeScript : types corrects, pas de `any` injustifié
   - Convention du projet (cf CLAUDE.md)
3. **Résumer** avec un verdict :
   - APPROVE : Code propre, prêt à merge
   - REQUEST CHANGES : Lister les problèmes par priorité (critical > major > minor)
   - COMMENT : Suggestions d'amélioration non-bloquantes

## Format de sortie
```
## Review: [fichier ou scope]

### Verdict: [APPROVE | REQUEST CHANGES | COMMENT]

### Problèmes trouvés
- [critical] Description...
- [major] Description...
- [minor] Description...

### Points positifs
- ...
```
