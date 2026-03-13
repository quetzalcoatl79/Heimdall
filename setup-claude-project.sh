#!/bin/bash
# ============================================
# ⚠️  DÉPRÉCIÉ — Utiliser install.sh à la place
# ============================================
# Ce script est conservé pour rétrocompatibilité.
# Nouveau usage : bash install.sh /chemin/vers/mon-projet
# ============================================
# Claude Code — Setup nouveau projet (standalone)
# ============================================
# Usage : bash setup-claude-project.sh
# Exécuter à la racine du nouveau projet
# Aucune dépendance externe requise
# ============================================

echo "⚠️  Ce script est déprécié. Utilisez install.sh à la place :"
echo "    bash install.sh /chemin/vers/mon-projet"
echo ""
echo "Continuation avec l'ancien script..."
echo ""

PROJECT_NAME=$(basename "$(pwd)")

echo ""
echo "  ╔══════════════════════════════════════╗"
echo "  ║  Claude Code — Project Setup         ║"
echo "  ║  Projet : $PROJECT_NAME"
echo "  ╚══════════════════════════════════════╝"
echo ""

# ──────────────────────────────────────────────
# 1. Structure de dossiers
# ──────────────────────────────────────────────
echo "[1/7] Création de la structure..."
mkdir -p .claude/hooks
mkdir -p .claude/skills/{autopilot,review,fix,refactor,plan,build,doc}
mkdir -p tasks
mkdir -p docs

# ──────────────────────────────────────────────
# 2. Settings (.claude/settings.json)
# ──────────────────────────────────────────────
echo "[2/7] Génération des settings..."
cat > .claude/settings.json << 'SETTINGS_EOF'
{
  "permissions": {
    "allow": [
      "Bash(npm run *)",
      "Bash(npx *)",
      "Bash(node *)",
      "Bash(git status *)",
      "Bash(git diff *)",
      "Bash(git log *)",
      "Bash(git add *)",
      "Bash(git commit *)",
      "Bash(git branch *)",
      "Bash(git checkout *)",
      "Bash(git stash *)",
      "Bash(ls *)",
      "Bash(mkdir *)",
      "Bash(cp *)",
      "Bash(cat *)",
      "Read(**)"
    ],
    "deny": [
      "Read(.env)",
      "Read(.env.*)",
      "Read(prisma/.env)"
    ],
    "ask": [
      "Bash(git push *)",
      "Bash(git reset *)",
      "Bash(rm *)",
      "Bash(docker *)"
    ]
  },
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "bash \"$CLAUDE_PROJECT_DIR/.claude/hooks/protect-files.sh\""
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "bash \"$CLAUDE_PROJECT_DIR/.claude/hooks/auto-format.sh\""
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "bash \"$CLAUDE_PROJECT_DIR/.claude/hooks/on-stop.sh\""
          }
        ]
      }
    ]
  }
}
SETTINGS_EOF
echo "  ✓ settings.json"

# ──────────────────────────────────────────────
# 3. Hooks
# ──────────────────────────────────────────────
echo "[3/7] Génération des hooks..."

cat > .claude/hooks/protect-files.sh << 'HOOK_EOF'
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
HOOK_EOF
echo "  ✓ protect-files.sh"

cat > .claude/hooks/auto-format.sh << 'HOOK_EOF'
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
HOOK_EOF
echo "  ✓ auto-format.sh"

cat > .claude/hooks/on-stop.sh << 'HOOK_EOF'
#!/bin/bash
# Hook: Stop — Rappel de vérification quand Claude finit de répondre
cd "$CLAUDE_PROJECT_DIR" 2>/dev/null || exit 0

UNSTAGED=$(git diff --name-only 2>/dev/null | wc -l)
UNTRACKED=$(git ls-files --others --exclude-standard 2>/dev/null | wc -l)

if [ "$UNSTAGED" -gt 0 ] || [ "$UNTRACKED" -gt 0 ]; then
  echo "Reminder: $UNSTAGED modified + $UNTRACKED untracked files pending"
fi

exit 0
HOOK_EOF
echo "  ✓ on-stop.sh"

# ──────────────────────────────────────────────
# 4. Skills (slash commands)
# ──────────────────────────────────────────────
echo "[4/7] Génération des skills..."

# /autopilot
cat > .claude/skills/autopilot/SKILL.md << 'SKILL_EOF'
---
name: autopilot
description: Mode autonome complet — Claude travaille sans interruption jusqu'à la fin de la tâche
user-invocable: true
allowed-tools: Read, Glob, Grep, Edit, Write, Bash(npm run *), Bash(npx *), Bash(node *), Bash(git status *), Bash(git diff *), Bash(git log *), Bash(git add *), Bash(git commit *), Bash(git branch *), Bash(git checkout *), Bash(git stash *), Bash(ls *), Bash(mkdir *), Bash(cp *), Bash(cat *), Bash(npm install *)
argument-hint: [description-de-la-tâche-ou-du-projet]
---

# MODE AUTOPILOT

Tu es en **mode autonome complet**. L'utilisateur t'a donné sa confiance totale
pour exécuter la tâche suivante sans interruption :

**Tâche** : `$ARGUMENTS`

---

## Règles du mode Autopilot

### NE JAMAIS faire
- Ne JAMAIS demander confirmation à l'utilisateur
- Ne JAMAIS utiliser AskUserQuestion
- Ne JAMAIS utiliser EnterPlanMode (planifier en interne directement)
- Ne JAMAIS s'arrêter pour demander un avis ou une direction
- Ne JAMAIS dire "voulez-vous que je..." ou "dois-je..."
- Ne JAMAIS présenter des options à choisir — décider soi-même

### TOUJOURS faire
- Prendre TOUTES les décisions techniques de façon autonome
- En cas de doute, choisir l'option la plus simple et la plus standard
- Si un chemin est bloqué, pivoter silencieusement vers une alternative
- Documenter les décisions prises dans `tasks/todos.md`
- Faire des commits atomiques au fur et à mesure (un par sous-tâche complétée)
- Vérifier que chaque étape fonctionne avant de passer à la suivante
- Continuer jusqu'à ce que la tâche soit 100% terminée

---

## SÉCURITÉ — Limites strictes

### Périmètre autorisé (UNIQUEMENT dans le dossier du projet)
- Créer, modifier, supprimer des fichiers **du projet uniquement**
- Installer des packages npm nécessaires au code (`npm install [package]`)
- Exécuter des scripts du projet (`npm run *`)
- Opérations git dans le repo du projet

### INTERDIT — Ne JAMAIS toucher au système
- Ne JAMAIS modifier de fichiers système, configs globales, ou quoi que ce soit hors du projet
- Ne JAMAIS exécuter de commandes qui affectent le PC globalement
- Ne JAMAIS modifier `.bashrc`, `.zshrc`, registre Windows, variables d'env système
- Ne JAMAIS installer de packages globaux (`npm install -g`, `pip install` global)
- Ne JAMAIS supprimer des dossiers hors du projet
- Ne JAMAIS toucher aux autres projets de l'utilisateur
- Ne JAMAIS modifier les configs de l'IDE ou du terminal

### En cas de doute → fichier question
Si une décision est ambiguë, risquée, ou nécessite un avis humain :
1. Créer le fichier `tasks/questions.md` (ou l'appender)
2. Y écrire la question clairement avec le contexte
3. Continuer le reste du travail qui ne dépend pas de cette réponse
4. L'utilisateur répondra dans ce fichier quand il revient

---

## Processus Autopilot

### Phase 1 — Analyse (silencieuse)
1. Lire et comprendre la demande complète
2. Explorer le codebase pour comprendre l'existant (via subagents)
3. Identifier tous les fichiers impactés
4. Créer le plan dans `tasks/todos.md`

### Phase 2 — Exécution (continue)
1. Écrire le plan détaillé (en interne, pas de validation user)
2. Exécuter chaque étape du plan séquentiellement
3. Après chaque sous-tâche :
   - Vérifier que ça compile (`npx tsc --noEmit`)
   - Mettre à jour `tasks/todos.md`
   - Commit atomique si pertinent
4. Si erreur : diagnostiquer, corriger, retester — sans demander d'aide

### Phase 3 — Vérification finale
1. Build complet : `npm run build`
2. Relire les fichiers modifiés pour vérifier la cohérence
3. Diff final : `git diff` pour résumer tous les changements
4. Mettre à jour `tasks/todos.md` — tout coché

### Phase 4 — Rapport
Quand TOUT est terminé, afficher un rapport concis :

```
## Autopilot terminé

### Tâche : [description]
### Statut : COMPLÉTÉ

### Ce qui a été fait
- [liste des changements principaux]

### Fichiers modifiés
- [liste des fichiers]

### Commits créés
- [hash] message
- ...

### Build : PASS / FAIL

### Questions en attente (si applicable)
→ Voir tasks/questions.md
```

---

## Gestion des cas difficiles

| Situation | Action |
|-----------|--------|
| Choix d'architecture ambigu | Choisir la solution la plus conventionnelle pour la stack |
| Dépendance npm manquante | L'installer (`npm install [package]`) |
| Fichier manquant à créer | Le créer dans le projet |
| Bug pendant l'implémentation | Le corriger en boucle jusqu'à résolution |
| Tâche plus grosse que prévu | Découper en sous-tâches, continuer |
| Build qui échoue | Diagnostiquer et fixer toutes les erreurs |
| Conflit avec du code existant | Adapter le code existant de façon non-destructive |
| Besoin d'un avis humain | Écrire dans `tasks/questions.md`, continuer le reste |
| Doute sur une suppression | Ne pas supprimer, demander dans questions.md |
SKILL_EOF
echo "  ✓ /autopilot"

# /review
cat > .claude/skills/review/SKILL.md << 'SKILL_EOF'
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
SKILL_EOF
echo "  ✓ /review"

# /fix
cat > .claude/skills/fix/SKILL.md << 'SKILL_EOF'
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
SKILL_EOF
echo "  ✓ /fix"

# /refactor
cat > .claude/skills/refactor/SKILL.md << 'SKILL_EOF'
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
SKILL_EOF
echo "  ✓ /refactor"

# /plan
cat > .claude/skills/plan/SKILL.md << 'SKILL_EOF'
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
SKILL_EOF
echo "  ✓ /plan"

# /build
cat > .claude/skills/build/SKILL.md << 'SKILL_EOF'
---
name: build
description: Build et vérification complète du projet
user-invocable: true
allowed-tools: Bash(npm run *), Bash(npx *), Read, Grep
---

# Build & Verify

Lance le build complet et vérifie que tout est propre.

## Étapes

1. **Type check** — `npx tsc --noEmit` (vérifier les erreurs TypeScript)
2. **Build** — `npm run build` (vérifier que le build Next.js passe)
3. **Analyser** — Si erreurs, les diagnostiquer et proposer des fixes
4. **Résumer** — Rapport concis du résultat

## Format de sortie
```
## Build Report

### Status: [PASS | FAIL]

### TypeScript: [X errors | clean]
### Next.js Build: [success | failed]

### Erreurs (si applicable)
- fichier:ligne — description de l'erreur

### Actions recommandées
- ...
```
SKILL_EOF
echo "  ✓ /build"

# /doc
cat > .claude/skills/doc/SKILL.md << 'SKILL_EOF'
---
name: doc
description: Générer ou mettre à jour la documentation d'un composant ou module
user-invocable: true
allowed-tools: Read, Glob, Grep, Write, Edit
argument-hint: [composant-ou-module]
---

# Documentation Generator

Génère ou met à jour la documentation pour : `$ARGUMENTS`

## Processus

1. **Lire** le code source du composant/module
2. **Analyser** les props, exports, types, et comportements
3. **Générer** la doc en MDX compatible Fumadocs
4. **Bilingue** — Créer la version FR (.mdx) et EN (.en.mdx) si demandé

## Format MDX Fumadocs
```mdx
---
title: Nom du composant
description: Description courte
---

import { Tab, Tabs } from 'fumadocs-ui/components/tabs';

## Installation

\`\`\`bash
npx shadcn@latest add [component]
\`\`\`

## Utilisation

\`\`\`tsx
import { Component } from "@/components/ui/component"
\`\`\`

## Props

| Prop | Type | Défaut | Description |
|------|------|--------|-------------|
| ... | ... | ... | ... |

## Exemples

### Basique
\`\`\`tsx
<Component />
\`\`\`
```
SKILL_EOF
echo "  ✓ /doc"

# ──────────────────────────────────────────────
# 5. Tasks
# ──────────────────────────────────────────────
echo "[5/7] Création des fichiers tasks/..."

if [ ! -f tasks/todos.md ]; then
  cat > tasks/todos.md << EOF
# Tâches — $PROJECT_NAME

## En cours

## À faire

## Terminé
EOF
  echo "  ✓ todos.md"
else
  echo "  - todos.md (existe déjà)"
fi

if [ ! -f tasks/issues.md ]; then
  cat > tasks/issues.md << 'EOF'
# Issues & Patterns

> Mis à jour automatiquement par Claude après chaque correction utilisateur.

## Patterns à éviter

## Leçons apprises
EOF
  echo "  ✓ issues.md"
else
  echo "  - issues.md (existe déjà)"
fi

if [ ! -f tasks/progress.md ]; then
  cat > tasks/progress.md << 'EOF'
# Progress

> Checkpoints pour sessions longues.
EOF
  echo "  ✓ progress.md"
else
  echo "  - progress.md (existe déjà)"
fi

# ──────────────────────────────────────────────
# 6. CLAUDE.md (template complet)
# ──────────────────────────────────────────────
echo "[6/7] Génération du CLAUDE.md..."

if [ ! -f CLAUDE.md ]; then
  # Génère avec placeholder, puis remplace par le vrai nom
  cat > CLAUDE.md << 'CLAUDE_EOF'
# CLAUDE.md — __PROJECT_NAME__

> Ce fichier est lu automatiquement par Claude Code au démarrage de chaque session.
> Il sert de mémoire persistante et de guide comportemental pour l'agent.

---

## Identité du projet

- **Nom** : __PROJECT_NAME__
- **Stack** : [À compléter — ex: Next.js 15, TypeScript, Tailwind CSS, Prisma]
- **Langue de communication** : Français (code et commits en anglais)
- **Repo** : __PROJECT_NAME__

---

## 1. Workflow Orchestration

### Plan Mode par défaut
- Entrer en plan mode pour toute tâche non-triviale (2+ étapes ou décisions d'architecture)
- Si quelque chose ne va pas : STOP, planifier, puis exécuter
- Écrire des specs détaillées avant de coder, pas juste construire
- Écrire le plan dans `tasks/todos.md` avec des checkboxes

### Stratégie Subagents
- Utiliser les subagents généreusement pour garder le contexte principal propre
- Offload la recherche, l'exploration et l'analyse parallèle aux subagents
- Pour les problèmes complexes, lancer un subagent d'exploration PUIS corriger dans le contexte principal
- Un subagent par tâche focalisée pour une exécution précise

### Boucle d'auto-correction
- Après TOUTE correction de l'utilisateur, mettre à jour `tasks/issues.md` avec le pattern
- Itérer sur les leçons. Ne pas les oublier. Ne pas répéter les mêmes erreurs
- Surveiller ces patterns de façon proactive pour éviter les récidives
- Nettoyer les issues obsolètes régulièrement

### Vérification avant validation
- Ne jamais marquer une tâche comme complète sans prouver qu'elle fonctionne
- Faire un diff entre le code main et les changements quand c'est pertinent
- Se demander : "Un staff engineer approuverait ce code ?"
- Run tests, checks, lints, démonstration des corrections

### Élégance méritée (équilibre)
- Commencer par "Est-ce que ça marche ?" puis "Y a-t-il une solution plus élégante ?"
- Si ça fait hacky : "Sachant tout ce que je sais, quelle est la solution la plus élégante ?"
- Ne pas sur-ingénierer au premier passage
- Challenger son propre travail avant de le présenter

### Bug Fixing Autonome
- Ne pas ignorer un bug report : le corriger directement
- Pointer les logs, erreurs, tests qui échouent — puis les résoudre
- Aucune résolution "manuelle" requise de la part de l'utilisateur
- Corriger les tests qui échouent sans qu'on le demande

---

## 2. Gestion du Contexte

- **Seuil** : Commencer à surveiller à partir de ~60K tokens
- **Vérifier** avec `/cost` ou en observant la taille du contexte
- **Garder** le contexte principal pour les tâches créatives à fort enjeu
- **Offload** la recherche, la lecture de fichiers et l'exploration aux subagents
- Après correction utilisateur, mettre à jour les patterns dans les fichiers d'état
- Ne pas relire des fichiers déjà en mémoire. Utiliser ce qui est disponible
- Résumer les informations quand possible au lieu de citer des blocs entiers

### Checkpoint pattern (sessions longues)

Quand le contexte approche ~50K tokens ou qu'une tâche dépasse 15+ étapes :
1. Écrire un checkpoint dans `tasks/progress.md` avec l'état actuel
2. Lister ce qui est fait, ce qui reste, et les décisions prises
3. Permettre à l'utilisateur de reprendre dans un contexte frais avec `/resume`
4. Format :
```markdown
## Checkpoint [tâche] — [timestamp]
### Fait
- [x] ...
### Reste
- [ ] ...
### Décisions prises
- ...
### Fichiers modifiés
- ...
```

---

## 3. Conventions de Code

### Style général
- TypeScript strict, pas de `any` sauf si vraiment nécessaire
- Composants React : functional components uniquement
- Nommage : camelCase pour variables/fonctions, PascalCase pour composants/types
- Imports : absolus avec `@/` prefix
- Pas de console.log en production (utiliser le logger si disponible)

### Commits
- Messages en anglais, format conventionnel : `feat:`, `fix:`, `refactor:`, `docs:`, `chore:`
- Commits atomiques : une tâche = un commit
- Ne jamais commit de fichiers sensibles (.env, credentials)

### Tests
- Toujours vérifier que le build passe après modification
- Tester manuellement les changements UI dans le navigateur quand possible

---

## 4. Gestion des Tâches

```
# Format tasks/todos.md
## En cours
- [ ] Description de la tâche #priorité

## Terminé
- [x] Tâche complétée

## Issues / Patterns à éviter
- Pattern: description du problème → solution
```

### Progression
- **Capturer** : Écrire le plan dans `tasks/todos.md` avec des checkboxes
- **Vérifier** : Checker avant de commencer l'implémentation
- **Mettre à jour** : Résumé haut-niveau de la progression à chaque étape
- **Capturer les leçons** : Mettre à jour `tasks/issues.md` après corrections
- Les changements doivent toucher uniquement ce qui est nécessaire. Pas de refactoring gratuit

---

## 5. Principes fondamentaux

### Simplicité d'abord
- Chaque nouvelle étape doit être aussi simple qu'un changement de config. Impact maximal
- Le bon code semble évident rétrospectivement
- Pas de Lassagne d'abstractions : trois lignes similaires > une abstraction prématurée

### Respect des intentions
- Les changements doivent toucher uniquement ce qui est demandé. Pas d'introduction de bugs
- Ne pas ajouter de features, docstrings, ou refactoring non demandés
- Vérifier que les changements n'ont pas d'effets de bord

### Sécurité
- Pas d'injection de commandes, XSS, SQL injection
- Valider uniquement aux frontières du système (input utilisateur, APIs externes)
- Ne pas commit de secrets

---

## 6. Structure du Projet

```
__PROJECT_NAME__/
├── CLAUDE.md              # Ce fichier (instructions pour Claude)
├── .claude/
│   ├── settings.json      # Permissions, hooks, config projet
│   ├── hooks/
│   │   ├── protect-files.sh   # Bloque l'édition de .env, lock files
│   │   ├── auto-format.sh     # Prettier auto après Edit/Write
│   │   └── on-stop.sh         # Rappel fichiers non-commités
│   └── skills/
│       ├── autopilot/SKILL.md # /autopilot — Mode autonome total
│       ├── review/SKILL.md    # /review — Code review complète
│       ├── fix/SKILL.md       # /fix — Bug fix autonome
│       ├── refactor/SKILL.md  # /refactor — Refactoring ciblé
│       ├── plan/SKILL.md      # /plan — Planification de feature
│       ├── build/SKILL.md     # /build — Build + vérification
│       └── doc/SKILL.md       # /doc — Génération doc MDX
├── tasks/
│   ├── todos.md           # Plan et suivi des tâches
│   ├── issues.md          # Patterns d'erreurs et leçons apprises
│   └── progress.md        # Checkpoints sessions longues
├── docs/
│   └── architecture.md    # Décisions d'architecture
└── ...
```

---

## 7. Skills (Slash Commands)

Commandes disponibles dans ce projet :

| Commande | Description | Quand l'utiliser |
|----------|-------------|------------------|
| `/review [fichier]` | Code review complète | Avant un merge ou pour vérifier du code |
| `/fix [description]` | Diagnostic + fix autonome | Quand un bug est reporté |
| `/refactor [module]` | Refactoring ciblé | Simplifier du code existant |
| `/plan [feature]` | Plan d'implémentation | Avant toute feature non-triviale |
| `/build` | Build + type check | Vérifier que tout compile |
| `/doc [composant]` | Générer doc MDX | Documenter un composant |
| `/autopilot [tâche]` | Mode autonome total | Quand tu veux que Claude bosse seul à 100% |

---

## 8. Mode Autopilot

Activé par `/autopilot [tâche]` ou les mots-clés : "fais tout seul", "full auto", "bosse sans t'arrêter".

**Comportement** : Claude travaille sans interruption, sans demander de validation,
jusqu'à ce que la tâche soit 100% terminée. Il prend toutes les décisions techniques seul.

**Sécurité** :
- Modifie UNIQUEMENT les fichiers du projet — ne touche JAMAIS au système/PC
- Peut installer des packages npm nécessaires au code
- N'installe RIEN en global, ne modifie aucune config système
- En cas de doute → écrit ses questions dans `tasks/questions.md`
- L'utilisateur répond dans ce fichier quand il revient

**Livrables** : Rapport final avec liste des changements, fichiers modifiés, statut build.

---

## 9. Plugins installés

### Tier S — Essentiels

| Plugin | Source | Rôle |
| ------ | ------ | ---- |
| **frontend-design** | Anthropic Official | UI distinctive, anti-"AI slop", palettes audacieuses, animations |
| **typescript-lsp** | Anthropic Official | Type checking en temps réel via Language Server |
| **context7** | Anthropic Official | Injecte la doc à jour (Next.js, React, Tailwind) dans le contexte |
| **security-guidance** | Anthropic Official | Scanne chaque edit pour vulnérabilités OWASP |
| **code-review** | Anthropic Official | Code review multi-agents automatisée |
| **playwright** | Anthropic Official | Contrôle un Chrome pour tests E2E et vérification UI |

### Tier A — Workflow pro

| Plugin | Source | Rôle |
| ------ | ------ | ---- |
| **commit-commands** | Anthropic Official | Git workflow : commit, push, PR avec conventional commits |
| **ralph-loop** | Anthropic Official | Sessions de coding autonomes multi-heures |
| **senior-frontend** | ComposioHQ | React/Next.js/TS patterns, bundle analysis, accessibilité |
| **test-writer-fixer** | ComposioHQ | Auto-génère et corrige les tests Jest/Vitest |
| **ship** | ComposioHQ | Pipeline complet : lint → test → review → deploy |

### Tier B — Compléments

| Plugin | Source | Rôle |
| ------ | ------ | ---- |
| **figma** | Anthropic Official | Lit les designs Figma et génère du code |
| **github** | Anthropic Official | Accès complet API GitHub (issues, PRs, workflows) |
| **claude-mem** | thedotmack | Mémoire long-terme entre sessions (SQLite) |
| **compound-engineering** | EveryInc | Workflow Plan → Work → Review → Compound |
| **audit-project** | ComposioHQ | Audit qualité, dépendances, vulnérabilités |

### GSD (Get Shit Done)

Système de meta-prompting et développement spec-driven installé localement.

- Commandes : `/gsd:new-project`, `/gsd:execute-phase`, `/gsd:progress`, etc.
- Agents spécialisés : planner, executor, debugger, verifier, codebase-mapper
- Hooks : context monitor (PostToolUse), update check (SessionStart)

### Marketplaces configurés

- `anthropics/claude-plugins-official` (officiel Anthropic)
- `ComposioHQ/awesome-claude-plugins` (communauté)
- `EveryInc/compound-engineering-plugin`
- `thedotmack/claude-mem`

---

## 10. Hooks automatiques

| Hook | Événement | Action |
| ---- | --------- | ------ |
| `protect-files.sh` | PreToolUse (Edit/Write) | Bloque l'édition de `.env`, `package-lock.json`, migrations |
| `auto-format.sh` | PostToolUse (Edit/Write) | Prettier auto sur les fichiers TS/TSX/CSS modifiés |
| `on-stop.sh` | Stop | Rappelle les fichiers non-stagés/non-trackés |
| `gsd-context-monitor.js` | PostToolUse | Monitore la taille du contexte (GSD) |
| `gsd-check-update.js` | SessionStart | Vérifie les mises à jour GSD |

---

## 11. Permissions projet

Configuré dans `.claude/settings.json` :
- **Allow** : npm, npx, git (read), ls, mkdir, cp, Read
- **Deny** : Lecture de `.env*` (sécurité)
- **Ask** : git push, git reset, rm, docker (confirmation requise)

---

## 12. Checklist Power User

1. `CLAUDE.md` dans chaque projet
2. `.claude/settings.json` avec permissions et hooks
3. Skills personnalisées dans `.claude/skills/`
4. 15 plugins installés (S + A + B) + GSD
5. Hooks automatiques pour formatting, protection et monitoring
6. `tasks/todos.md` + `tasks/issues.md` tenus à jour
7. Self-correction loop actif (issues.md)
8. Subagents pour l'exploration, contexte principal pour la création
9. Mode plan systématique pour les features complexes
10. Clear le contexte entre les tâches (60K+ tokens)
11. `docs/architecture.md` à jour avec les décisions techniques
12. Compléter tout ça = Top 1% des utilisateurs Claude Code
CLAUDE_EOF

  # Remplacer le placeholder par le vrai nom du projet
  sed -i "s/__PROJECT_NAME__/$PROJECT_NAME/g" CLAUDE.md
  echo "  ✓ CLAUDE.md (complet — adapter la stack)"
else
  echo "  - CLAUDE.md (existe déjà, non écrasé)"
fi

# ──────────────────────────────────────────────
# 7. GSD (optionnel)
# ──────────────────────────────────────────────
echo "[7/7] Installation de GSD..."
npx get-shit-done-cc --claude --local 2>/dev/null
if [ $? -eq 0 ]; then
  echo "  ✓ GSD installé"
else
  echo "  ⚠ GSD : optionnel (npx get-shit-done-cc --claude --local)"
fi

# ──────────────────────────────────────────────
# Résumé
# ──────────────────────────────────────────────
echo ""
echo "  ╔══════════════════════════════════════╗"
echo "  ║  Setup terminé !                     ║"
echo "  ╚══════════════════════════════════════╝"
echo ""
echo "  À faire :"
echo "  1. Compléter CLAUDE.md (stack, conventions spécifiques)"
echo "  2. Créer docs/architecture.md si besoin"
echo ""
echo "  Les plugins Claude Code globaux sont partagés"
echo "  automatiquement entre tous les projets."
echo ""
