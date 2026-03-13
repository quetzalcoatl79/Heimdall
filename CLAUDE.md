# CLAUDE.md — Heimdall

> Ce fichier est lu automatiquement par Claude Code au démarrage de chaque session.
> Il sert de mémoire persistante et de guide comportemental pour l'agent.

---

## Identité du projet

- **Nom** : Heimdall
- **Stack** : [À compléter — ex: Next.js 15, TypeScript, Tailwind CSS, Prisma]
- **Langue de communication** : Français (code et commits en anglais)
- **Repo** : Heimdall

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
Heimdall/
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
