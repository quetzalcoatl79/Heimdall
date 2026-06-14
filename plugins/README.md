# Heimdall Plugins

Ce dossier contient les plugins modulaires de Heimdall.

## Structure d'un Plugin

Chaque plugin doit suivre cette structure :

```
/plugins/
  /mon-plugin/
    plugin.yaml           # Manifest obligatoire
    backend/
      plugin.go           # Implémente plugins.Plugin
      handlers.go         # Handlers HTTP optionnels
      models.go           # Modèles GORM optionnels
      *.go                # Autres fichiers Go
    frontend/
      components/         # Composants React réutilisables
      MonPluginView.tsx   # Vue principale du plugin
      *.tsx               # Autres composants
    migrations/
      NNNNNN_description.up.sql
      NNNNNN_description.down.sql
```

## Manifest (plugin.yaml)

```yaml
name: mon-plugin
version: 1.0.0
description: Description du plugin
author: Nom de l'auteur

# Dépendances système (apt/brew packages)
system_dependencies:
  - aircrack-ng
  - nmap

# Dépendances Go (modules)
go_dependencies:
  - github.com/google/uuid

# Dépendances npm frontend
npm_dependencies:
  - recharts

# Configuration du menu
menu:
  label: "Mon Plugin"
  icon: "plug"
  position: 100

# Permissions requises
permissions:
  - mon-plugin:read
  - mon-plugin:write

# Routes API exposées (pour documentation)
routes:
  - method: GET
    path: /view
    description: Retourne le schéma UI
  - method: POST
    path: /action
    description: Exécute une action
```

## Installation d'un Plugin

```bash
# Cloner le plugin dans /plugins
cd plugins
git clone https://github.com/user/heimdall-plugin-wifi.git wifi

# Régénérer le registre et recompiler
make plugins
make build

# Redémarrer Heimdall
./run_heimdall.sh
```

## Création d'un Plugin

1. Créer le dossier avec la structure ci-dessus
2. Implémenter l'interface `plugins.Plugin` dans `backend/plugin.go`
3. Créer la vue frontend dans `frontend/MonPluginView.tsx`
4. Ajouter les migrations si nécessaire
5. Créer le `plugin.yaml`

## Interface Plugin (Go)

```go
package monplugin

import (
    "github.com/gobuffalo/buffalo"
    "github.com/nxo/engine/internal/plugins"
)

type MonPlugin struct{}

func (p *MonPlugin) Key() string         { return "mon-plugin" }
func (p *MonPlugin) Version() string     { return "1.0.0" }
func (p *MonPlugin) Description() string { return "Mon super plugin" }

func (p *MonPlugin) Manifest() map[string]any {
    return map[string]any{
        "name":        "Mon Plugin",
        "version":     p.Version(),
        "description": p.Description(),
        "menu_items": []map[string]any{
            {"label": "Mon Plugin", "path": "/admin/plugins/mon-plugin", "icon": "plug"},
        },
    }
}

func (p *MonPlugin) RegisterRoutes(group *buffalo.App, deps plugins.Deps) {
    group.GET("/view", p.handleView)
    group.POST("/action", p.handleAction)
}

func init() {
    plugins.Register(&MonPlugin{})
}
```

## Composant Frontend

Le composant frontend doit être exporté par défaut et sera chargé automatiquement
basé sur le `Key()` du plugin :

```tsx
// frontend/MonPluginView.tsx
'use client';

export default function MonPluginView() {
    return (
        <div>
            <h1>Mon Plugin</h1>
        </div>
    );
}
```
