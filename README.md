# 🛡️ Heimdall

<div align="center">

![Heimdall Logo](frontend/public/Heimdall.png)

**Plateforme de pentest WiFi extensible par plugins**

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Next.js](https://img.shields.io/badge/Next.js-14+-000000?style=for-the-badge&logo=next.js&logoColor=white)](https://nextjs.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

*Conçu pour Kali Linux, Parrot OS et autres distributions de pentest*

</div>

---

## 📋 Table des matières

- [Fonctionnalités](#-fonctionnalités)
- [Prérequis](#-prérequis)
- [Installation Linux](#-installation-linux)
- [Installation Windows](#-installation-windows-développement)
- [Utilisation](#-utilisation)
- [Plugins](#-plugins-disponibles)
- [Architecture](#-architecture)
- [Configuration](#-configuration)
- [Développement](#-développement)
- [Sécurité](#-sécurité)

---

## ✨ Fonctionnalités

### 🔌 Système de Plugins
- Architecture modulaire extensible
- Activation/désactivation à chaud
- Interface UI dynamique par plugin
- Permissions granulaires

### 📡 Plugin WiFi Pentest
- **Scan** des réseaux WiFi environnants
- **Capture** de handshakes WPA/WPA2
- **Bruteforce** avec wordlists personnalisées
- Détection automatique des interfaces monitor
- Support des chipsets Atheros, Realtek, Ralink

### 🌐 Interface Web Moderne
- Dashboard temps réel
- Thème clair/sombre
- Multilingue (FR/EN)
- Responsive design

### ⚡ Workers & Jobs
- Exécution asynchrone des tâches
- File d'attente Redis
- Monitoring des jobs en cours

---

## 📦 Prérequis

### Linux (Kali/Parrot OS/Debian/Ubuntu)

| Composant | Version | Installé par setup.sh |
|-----------|---------|----------------------|
| Docker | 24.0+ | ✅ |
| Docker Compose | 2.20+ | ✅ |
| Go | 1.22+ | ✅ |
| Node.js | 20+ | ✅ |
| Outils WiFi | aircrack-ng, hashcat... | ✅ |

**Chipsets WiFi recommandés :**
- Atheros AR9271 (ALFA AWUS036NHA)
- Realtek RTL8812AU (ALFA AWUS036ACH)
- Ralink RT3070 (ALFA AWUS036NH)

---

## 🐧 Installation

### Installation automatique (recommandée)

```bash
# Cloner le projet
git clone https://github.com/YOUR_REPO/heimdall.git
cd heimdall

# Lancer l'installation complète
sudo ./setup.sh
```

Le script `setup.sh` va :
1. ✅ Détecter votre distribution (Kali, Parrot, Debian, Ubuntu, Fedora, Arch...)
2. ✅ Installer Docker et Docker Compose si absents
3. ✅ Installer Go 1.22+ si absent
4. ✅ Installer Node.js 20+ si absent
5. ✅ Installer les outils WiFi pentest (aircrack-ng, reaver, hashcat...)
6. ✅ Installer les dépendances Go et npm du projet
7. ✅ Configurer l'environnement (.env)
8. ✅ Créer les répertoires de captures et wordlists

### Démarrer Heimdall

```bash
# Lancer tous les services
sudo ./run_heimdall.sh

# Ou en mode production
sudo ./run_heimdall.sh --prod
```

### Arrêter Heimdall

```bash
sudo ./stop_heimdall.sh
```

### Installation manuelle

```bash
# 1. Cloner le projet
git clone https://github.com/YOUR_REPO/heimdall.git
cd heimdall

# 2. Créer le fichier .env
cp .env.example .env
nano .env  # Modifier si nécessaire

# 3. Démarrer Redis et PostgreSQL
docker compose up -d postgres redis

# 4. Compiler et lancer le backend
cd backend
go build -o ./bin/api ./cmd/api
go build -o ./bin/worker ./cmd/worker
./bin/api &
./bin/worker &

# 5. Lancer le frontend
cd ../frontend
npm ci --legacy-peer-deps
npm run dev
```

### Vérifier l'installation

```bash
# Interfaces WiFi détectées
ip link show | grep -E "wlan|wlp"

# Logs
tail -f /opt/heimdall/logs/backend.log
tail -f /opt/heimdall/logs/frontend.log
```

### Accès à l'interface

| Service | URL |
|---------|-----|
| **Frontend** | http://localhost:3001 |
| **API Backend** | http://localhost:3000 |
| **Adminer (DB)** | http://localhost:8082 (avec profil debug) |

**Identifiants par défaut :**
```
Email:    admin@heimdall.local
Password: admin123
```

---

## 🎯 Utilisation

### 1. Connexion

1. Ouvrir http://localhost:3001
2. Se connecter avec `admin@heimdall.local` / `admin123`
3. Changer le mot de passe dans les paramètres

### 2. Activer le plugin WiFi

1. Aller dans **Plugins**
2. Cliquer sur **Activer** pour le plugin WiFi
3. Le plugin apparaît dans la sidebar

### 3. Pentest WiFi

```
┌─────────────────────────────────────────────────┐
│  1. Sélectionner l'interface WiFi               │
│     └─> wlan0, wlan1, etc.                      │
│                                                 │
│  2. Scanner les réseaux                         │
│     └─> Liste des AP avec SSID, BSSID, canal   │
│                                                 │
│  3. Sélectionner une cible                      │
│     └─> Cliquer sur le réseau à attaquer       │
│                                                 │
│  4. Capturer le handshake                       │
│     └─> Attente de la capture WPA              │
│                                                 │
│  5. Bruteforce                                  │
│     └─> Attaque dictionnaire avec wordlist     │
└─────────────────────────────────────────────────┘
```

### 4. Mode Monitor

```bash
# Activer le mode monitor manuellement (si nécessaire)
sudo airmon-ng start wlan0

# Vérifier
iwconfig wlan0mon
```

---

## 🔌 Plugins disponibles

| Plugin | Description | Permissions |
|--------|-------------|-------------|
| 📡 **WiFi** | Pentest WiFi complet | `wifi:scan`, `wifi:capture`, `wifi:crack` |
| 💚 **Healthcheck** | Monitoring applicatif | - |
| 🔧 **Sample** | Plugin de démonstration | - |

---

## 🏗️ Architecture Logicielle

### Vue Globale

```
┌────────────────────────────────────────────────────────────────────────────────┐
│                              HEIMDALL ARCHITECTURE                              │
└────────────────────────────────────────────────────────────────────────────────┘

                              ┌─────────────────────┐
                              │     NAVIGATEUR      │
                              │    (localhost:3001) │
                              └──────────┬──────────┘
                                         │
                                         ▼
┌────────────────────────────────────────────────────────────────────────────────┐
│                            FRONTEND (Next.js 14)                                │
│                                                                                 │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐  ┌──────────────┐  │
│  │    Pages       │  │  DynamicRender │  │   i18n (FR/EN) │  │  Auth Store  │  │
│  │  /admin/*      │  │  UI Components │  │   Context      │  │   Zustand    │  │
│  └────────────────┘  └────────────────┘  └────────────────┘  └──────────────┘  │
└────────────────────────────────────────┬───────────────────────────────────────┘
                                         │ REST API (JSON)
                                         ▼
┌────────────────────────────────────────────────────────────────────────────────┐
│                            BACKEND (Go Buffalo)                                 │
│                              (localhost:3000)                                   │
│                                                                                 │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐  ┌──────────────┐  │
│  │   Handlers     │  │   Middleware   │  │    Services    │  │   Workers    │  │
│  │  Auth, User    │  │  JWT, CORS     │  │  Auth, Plugin  │  │  Background  │  │
│  └────────────────┘  └────────────────┘  └────────────────┘  └──────────────┘  │
│                                                                                 │
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │                         PLUGIN SYSTEM                                     │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │  │
│  │  │  Healthcheck │  │    WiFi      │  │    Sample    │  │  Your Plugin │  │  │
│  │  │   Plugin     │  │   Plugin     │  │    Plugin    │  │    Here...   │  │  │
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘  │  │
│  └──────────────────────────────────────────────────────────────────────────┘  │
└────────┬───────────────────────┬───────────────────────┬───────────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌────────────────┐      ┌────────────────┐      ┌────────────────────────────────┐
│  PostgreSQL    │      │     Redis      │      │      SYSTÈME LINUX             │
│  (Docker)      │      │   (Docker)     │      │  ┌────────────────────────┐   │
│  Port 5433     │      │   Port 6379    │      │  │  Interfaces WiFi       │   │
│                │      │                │      │  │  aircrack-ng, reaver   │   │
│  - users       │      │  - sessions    │      │  │  hashcat               │   │
│  - plugins     │      │  - jobs queue  │      │  └────────────────────────┘   │
│  - jobs        │      │  - cache       │      │                                │
└────────────────┘      └────────────────┘      └────────────────────────────────┘
```

### Architecture Server-Driven UI (SDUI)

Heimdall utilise une architecture **SDUI** : le backend définit l'interface via un schéma JSON, et le frontend l'interprète dynamiquement. Cela permet d'ajouter des plugins **sans modifier le code frontend**.

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                          FLUX SERVER-DRIVEN UI                                   │
└─────────────────────────────────────────────────────────────────────────────────┘

   BACKEND                                              FRONTEND
   ───────                                              ────────

┌────────────────────┐                           ┌────────────────────┐
│  Plugin Go         │                           │  Page Plugin       │
│                    │   GET /plugins/{key}/view │  [key]/page.tsx    │
│  ui.NewView(...)   │ ◄─────────────────────────│                    │
│  .AddComponent()   │                           │  useQuery(...)     │
└─────────┬──────────┘                           └─────────┬──────────┘
          │                                                │
          │ Retourne ViewSchema JSON                       │
          ▼                                                ▼
┌────────────────────┐                           ┌────────────────────┐
│  {                 │                           │  DynamicRenderer   │
│    "title": "...", │ ──────────────────────────│                    │
│    "components": [ │                           │  ComponentRenderer │
│      {type:"card"} │                           │  ┌──────────────┐  │
│      {type:"stat"} │                           │  │ CardComponent│  │
│    ]               │                           │  │ StatComponent│  │
│  }                 │                           │  │ TableComponent│ │
└────────────────────┘                           │  │ ChartComponent│ │
                                                 │  │ FormComponent │ │
                                                 │  │ ... (+20)     │ │
                                                 │  └──────────────┘  │
                                                 └────────────────────┘
```

---

## 🔌 Système de Plugins

### Interface Plugin (Go)

Chaque plugin implémente cette interface :

```go
type Plugin interface {
    Key() string                                      // Identifiant unique (url-safe)
    Version() string                                  // Version sémantique
    Description() string                              // Description courte
    RegisterRoutes(group *buffalo.App, deps Deps)     // Endpoints API
    Manifest() map[string]any                         // Métadonnées (menu, permissions)
}
```

### Composants UI Disponibles

| Type | Description | Props Principales |
|------|-------------|-------------------|
| `card` | Carte avec titre/contenu | `title`, `subtitle`, `footer` |
| `stats` | Grille de statistiques | `children[]` |
| `stat` | Un indicateur | `label`, `value`, `icon`, `color`, `trend` |
| `table` | Tableau de données | `columns[]`, `data[]` |
| `chart` | Graphique | `chartType`, `data[]`, `xKey`, `yKey` |
| `alert` | Message d'alerte | `variant`, `message` |
| `form` | Formulaire dynamique | `fields[]`, `submitUrl` |
| `grid` | Layout en grille | `cols`, `gap`, `children[]` |
| `text` | Texte simple | `content`, `variant` |
| `heading` | Titre | `level`, `content` |
| `badge` | Badge/tag | `label`, `variant` |
| `progress` | Barre de progression | `value`, `max`, `label` |
| `list` | Liste | `children[]` |
| `codeBlock` | Code formaté | `code`, `language` |
| `json` | Affichage JSON | `data` |

### Processus d'Intégration d'un Plugin

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                     CRÉER UN NOUVEAU PLUGIN - 5 ÉTAPES                          │
└─────────────────────────────────────────────────────────────────────────────────┘

ÉTAPE 1: Créer le fichier plugin
────────────────────────────────
📁 backend/internal/plugins/builtin/myplugin/myplugin.go

    package myplugin

    import (
        "github.com/gobuffalo/buffalo"
        "github.com/nxo/engine/internal/plugins"
        "github.com/nxo/engine/internal/ui"
    )

    type MyPlugin struct{}

    func (p *MyPlugin) Key() string         { return "myplugin" }
    func (p *MyPlugin) Version() string     { return "1.0.0" }
    func (p *MyPlugin) Description() string { return "Mon super plugin" }

    func (p *MyPlugin) Manifest() map[string]any {
        return map[string]any{
            "name":        "Mon Plugin",
            "version":     p.Version(),
            "description": p.Description(),
            "menu_items": []map[string]any{{
                "label":    "Mon Plugin",
                "path":     "/admin/plugins/myplugin",
                "icon":     "star",
                "position": 50,
            }},
        }
    }

    func (p *MyPlugin) RegisterRoutes(g *buffalo.App, deps plugins.Deps) {
        g.GET("/view", p.viewHandler(deps))
    }

    func (p *MyPlugin) viewHandler(deps plugins.Deps) buffalo.Handler {
        return func(c buffalo.Context) error {
            view := ui.NewView("Mon Plugin").
                WithDescription("Interface de mon plugin").
                WithIcon("star").
                AddComponent(ui.Card("Bienvenue",
                    ui.Stat("Statut", "Actif", ui.WithIcon("check"), ui.WithColor("green")),
                    ui.Text("Contenu de mon plugin ici..."),
                )).
                AddComponent(ui.Table(
                    []ui.TableColumn{
                        {Key: "id", Label: "ID"},
                        {Key: "name", Label: "Nom"},
                        {Key: "status", Label: "Statut", Render: "badge"},
                    },
                    []map[string]any{
                        {"id": 1, "name": "Item 1", "status": "active"},
                        {"id": 2, "name": "Item 2", "status": "pending"},
                    },
                ))

            return c.Render(200, render.JSON(view))
        }
    }

    func init() {
        plugins.Register(&MyPlugin{})  // ← Auto-registration
    }


ÉTAPE 2: Importer dans builtin.go
─────────────────────────────────
📁 backend/internal/plugins/builtin/builtin.go

    import (
        _ "github.com/nxo/engine/internal/plugins/builtin/healthcheck"
        _ "github.com/nxo/engine/internal/plugins/builtin/sample"
        _ "github.com/nxo/engine/internal/plugins/builtin/wifi"
        _ "github.com/nxo/engine/internal/plugins/builtin/myplugin"  // ← Ajouter
    )


ÉTAPE 3: Recompiler le backend
──────────────────────────────
    cd backend
    go build -o ./bin/api ./cmd/api
    go build -o ./bin/worker ./cmd/worker


ÉTAPE 4: Redémarrer les services
────────────────────────────────
    sudo ./stop_heimdall.sh
    sudo ./run_heimdall.sh


ÉTAPE 5: Activer dans l'interface admin
───────────────────────────────────────
    1. Aller sur http://localhost:3001/admin/plugins
    2. Trouver "Mon Plugin" dans la liste
    3. Cliquer sur "Activer"
    4. Le plugin apparaît dans la sidebar !
```

### Exemple Concret : ViewSchema JSON

Voici ce que le plugin Healthcheck retourne au frontend :

```json
{
  "title": "État de santé",
  "description": "Surveillance en temps réel",
  "icon": "activity",
  "refresh": { "enabled": true, "interval": 5 },
  "components": [
    {
      "type": "alert",
      "props": { "variant": "success", "message": "Systèmes opérationnels" }
    },
    {
      "type": "grid",
      "props": { "cols": 4 },
      "children": [
        {
          "type": "card",
          "props": { "title": "Database" },
          "children": [{
            "type": "stat",
            "props": { "label": "PostgreSQL", "value": "healthy", "icon": "database", "color": "green" }
          }]
        },
        {
          "type": "card",
          "props": { "title": "Cache" },
          "children": [{
            "type": "stat",
            "props": { "label": "Redis", "value": "healthy", "icon": "server", "color": "green" }
          }]
        }
      ]
    }
  ]
}
```

### Arborescence des Fichiers Plugin

```
backend/internal/plugins/
├── registry.go          # Interface Plugin + Register()
├── mount.go             # Monte les routes /api/v1/plugins/{key}/*
├── sync.go              # Synchronise plugins Go ↔ DB
├── enabled_middleware.go # Middleware vérifie si plugin activé
└── builtin/
    ├── builtin.go       # Imports des plugins (auto-register)
    ├── healthcheck/
    │   └── healthcheck.go
    ├── sample/
    │   └── sample.go
    ├── wifi/
    │   └── wifi.go
    └── myplugin/        # ← Votre nouveau plugin
        └── myplugin.go

frontend/src/
├── components/ui/
│   └── DynamicRenderer.tsx  # Interprète ViewSchema JSON
├── app/admin/plugins/
│   └── [key]/
│       └── page.tsx         # Page générique pour tous les plugins
└── hooks/
    └── useActivePlugins.ts  # Hook pour les plugins actifs
```

---

## ⚙️ Configuration

### Variables d'environnement (.env)

```bash
# Application
APP_ENV=production
APP_DEBUG=false

# Database (localhost car pas en Docker)
DB_HOST=localhost
DB_PORT=5433
DB_NAME=heimdall_dev
DB_USER=postgres
DB_PASSWORD=postgres

# Redis (localhost car pas en Docker)
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT (générer avec: openssl rand -hex 32)
JWT_SECRET=CHANGE_ME_IN_PRODUCTION
JWT_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# Frontend
NEXT_PUBLIC_API_URL=http://localhost:3000/api/v1
```

---

## 🛠️ Développement

### Commandes utiles

```bash
# Démarrer Heimdall
sudo ./run_heimdall.sh

# Arrêter Heimdall
sudo ./stop_heimdall.sh

# Voir les logs
tail -f /opt/heimdall/logs/backend.log
tail -f /opt/heimdall/logs/worker.log
tail -f /opt/heimdall/logs/frontend.log

# Reconstruire le backend manuellement
cd backend
go build -o ./bin/api ./cmd/api
go build -o ./bin/worker ./cmd/worker

# Accéder à PostgreSQL
docker exec -it heimdall_postgres psql -U postgres -d heimdall_dev

# Reset la base de données
docker compose down -v
docker compose up -d postgres redis
# Puis relancer: sudo ./run_heimdall.sh
```

### Structure du code

```
backend/
├── cmd/
│   ├── api/main.go       # Point d'entrée API
│   ├── worker/main.go    # Point d'entrée Worker
│   └── migrate/main.go   # Outil migrations
├── internal/
│   ├── plugins/
│   │   ├── registry.go   # Registre des plugins
│   │   └── builtin/
│   │       ├── wifi/     # Plugin WiFi
│   │       └── ...
│   └── ...
└── migrations/           # Scripts SQL
```

---

## 🔒 Sécurité

### ⚠️ Avertissement légal

> **Heimdall est un outil de sécurité destiné aux professionnels autorisés.**
> 
> L'utilisation de cet outil pour attaquer des réseaux sans autorisation explicite est **illégale** et peut entraîner des poursuites judiciaires.
> 
> Utilisez uniquement sur des réseaux dont vous êtes propriétaire ou pour lesquels vous avez une autorisation écrite.

### Bonnes pratiques

1. **Changer les mots de passe par défaut**
2. **Utiliser HTTPS** en production
3. **Restreindre l'accès** au réseau local
4. **Mettre à jour** régulièrement les dépendances
5. **Ne pas exposer** les ports sur Internet

---

## 📄 Licence

MIT License - Voir [LICENSE](LICENSE)

---

<div align="center">

**Heimdall** - *Le gardien de vos audits WiFi* 🛡️

Made with for security researchers

</div>
