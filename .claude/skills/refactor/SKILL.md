---
name: refactor
description: Refactoring ciblé d'un fichier ou module
user-invocable: true
allowed-tools: Read, Glob, Grep, Edit, Write, Bash(npm run *), Bash(git diff *)
argument-hint: [fichier-ou-module]
---

# Refactoring

Refactore le code ciblé : `$ARGUMENTS`

## Processus

1. **Analyser** — Lire et comprendre le code actuel à 100%
2. **Planifier** — Identifier les améliorations (présenter le plan AVANT d'agir)
3. **Exécuter** — Appliquer les changements de façon incrémentale
4. **Vérifier** — Le comportement doit être IDENTIQUE (pas de changement fonctionnel)

## Principes
- Refactoring pur : pas de changement de comportement
- Commits atomiques par transformation
- Garder la rétro-compatibilité des exports/API publiques
- "Earned elegance" : commencer simple, puis améliorer
- Si le refactoring touche trop de fichiers, découper en étapes

## Transformations courantes
- Extraction de composants/fonctions
- Simplification de conditions complexes
- Suppression de code mort
- Amélioration des types TypeScript
- Réduction de la duplication (DRY)
