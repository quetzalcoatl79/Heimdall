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
