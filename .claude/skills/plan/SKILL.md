---
name: plan
description: Planifier une feature ou tâche complexe avant implémentation
user-invocable: true
allowed-tools: Read, Glob, Grep, Bash(git log *), Bash(git diff *), Bash(ls *)
disable-model-invocation: false
argument-hint: [description-de-la-tâche]
---

# Plan de Tâche

Planifie l'implémentation de : `$ARGUMENTS`

## Processus

1. **Explorer** — Scanner le codebase pour comprendre l'existant
2. **Identifier** — Lister les fichiers qui seront impactés
3. **Concevoir** — Proposer l'approche technique (avec alternatives si pertinent)
4. **Découper** — Créer des sous-tâches ordonnées et atomiques
5. **Risques** — Identifier les points de vigilance

## Format de sortie

```markdown
## Plan: [titre]

### Contexte
[Ce qui existe déjà et comment ça fonctionne]

### Approche retenue
[Description technique de la solution]

### Étapes
1. [ ] Étape 1 — description + fichiers impactés
2. [ ] Étape 2 — ...
3. [ ] Étape N — ...

### Fichiers impactés
- `path/file.ts` — ce qui change

### Risques et points d'attention
- ...

### Questions ouvertes (si applicable)
- ...
```

Écrire le plan dans `tasks/todos.md` une fois validé par l'utilisateur.
