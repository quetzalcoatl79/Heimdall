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
