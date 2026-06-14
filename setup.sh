#!/bin/bash
#
# ╔═══════════════════════════════════════════════════════════════╗
# ║                    🛡️  HEIMDALL SETUP                         ║
# ║         Installation automatique pour Linux/Kali/Parrot       ║
# ╚═══════════════════════════════════════════════════════════════╝
#
# Usage: sudo ./setup.sh
#

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'

# Versions minimales
GO_VERSION="1.22"
NODE_VERSION="20"

# Fonctions utilitaires
log()     { echo -e "${GREEN}[✓]${NC} $1"; }
info()    { echo -e "${BLUE}[i]${NC} $1"; }
warn()    { echo -e "${YELLOW}[!]${NC} $1"; }
error()   { echo -e "${RED}[✗]${NC} $1"; exit 1; }
step()    { echo -e "\n${PURPLE}${BOLD}▶ $1${NC}"; }

# Banner
show_banner() {
    echo -e "${CYAN}"
    cat << 'EOF'
    
    ██╗  ██╗███████╗██╗███╗   ███╗██████╗  █████╗ ██╗     ██╗     
    ██║  ██║██╔════╝██║████╗ ████║██╔══██╗██╔══██╗██║     ██║     
    ███████║█████╗  ██║██╔████╔██║██║  ██║███████║██║     ██║     
    ██╔══██║██╔══╝  ██║██║╚██╔╝██║██║  ██║██╔══██║██║     ██║     
    ██║  ██║███████╗██║██║ ╚═╝ ██║██████╔╝██║  ██║███████╗███████╗
    ╚═╝  ╚═╝╚══════╝╚═╝╚═╝     ╚═╝╚═════╝ ╚═╝  ╚═╝╚══════╝╚══════╝
                                                                   
              🛡️  WiFi Pentest Platform - Setup Script
    
EOF
    echo -e "${NC}"
}

# Vérifier si root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        error "Ce script doit être exécuté en root.\n   Relance avec: ${YELLOW}sudo $0${NC}"
    fi
}

# Détecter la distribution
detect_distro() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        DISTRO=$ID
        DISTRO_VERSION=$VERSION_ID
    else
        error "Distribution non reconnue"
    fi
    
    case $DISTRO in
        kali|parrot|debian|ubuntu|linuxmint)
            PKG_MANAGER="apt"
            PKG_INSTALL="apt-get install -y"
            PKG_UPDATE="apt-get update"
            ;;
        fedora|centos|rhel|rocky|alma)
            PKG_MANAGER="dnf"
            PKG_INSTALL="dnf install -y"
            PKG_UPDATE="dnf check-update || true"
            ;;
        arch|manjaro|endeavouros)
            PKG_MANAGER="pacman"
            PKG_INSTALL="pacman -S --noconfirm"
            PKG_UPDATE="pacman -Sy"
            ;;
        *)
            error "Distribution '$DISTRO' non supportée"
            ;;
    esac
    
    log "Distribution détectée: ${BOLD}$DISTRO${NC} ($DISTRO_VERSION)"
}

# Installer les paquets de base
install_base_packages() {
    step "Installation des paquets de base"
    
    $PKG_UPDATE > /dev/null 2>&1
    
    case $PKG_MANAGER in
        apt)
            $PKG_INSTALL curl wget git build-essential ca-certificates gnupg lsb-release
            ;;
        dnf)
            $PKG_INSTALL curl wget git gcc gcc-c++ make ca-certificates
            ;;
        pacman)
            $PKG_INSTALL curl wget git base-devel ca-certificates
            ;;
    esac
    
    log "Paquets de base installés"
}

# Installer Docker
install_docker() {
    step "Vérification de Docker"
    
    if command -v docker &> /dev/null; then
        DOCKER_VERSION=$(docker --version | grep -oP '\d+\.\d+\.\d+' | head -1)
        log "Docker déjà installé (v$DOCKER_VERSION)"
        return
    fi
    
    info "Installation de Docker..."
    
    case $PKG_MANAGER in
        apt)
            # Ajouter le repo Docker
            install -m 0755 -d /etc/apt/keyrings
            curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
            chmod a+r /etc/apt/keyrings/docker.gpg
            
            # Utiliser debian pour Kali/Parrot
            DOCKER_DISTRO="debian"
            CODENAME="bookworm"
            if [ "$DISTRO" = "ubuntu" ]; then
                DOCKER_DISTRO="ubuntu"
                CODENAME=$(lsb_release -cs)
            fi
            
            echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$DOCKER_DISTRO $CODENAME stable" | \
                tee /etc/apt/sources.list.d/docker.list > /dev/null
            
            apt-get update
            apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
            ;;
        dnf)
            dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo
            dnf install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
            ;;
        pacman)
            pacman -S --noconfirm docker docker-compose
            ;;
    esac
    
    # Démarrer Docker
    systemctl enable docker
    systemctl start docker
    
    # Ajouter l'utilisateur courant au groupe docker
    REAL_USER=${SUDO_USER:-$USER}
    if [ "$REAL_USER" != "root" ]; then
        usermod -aG docker "$REAL_USER"
        info "Utilisateur '$REAL_USER' ajouté au groupe docker"
    fi
    
    log "Docker installé avec succès"
}

# Installer Docker Compose (si nécessaire)
install_docker_compose() {
    step "Vérification de Docker Compose"
    
    if docker compose version &> /dev/null; then
        COMPOSE_VERSION=$(docker compose version --short)
        log "Docker Compose déjà installé (v$COMPOSE_VERSION)"
        return
    fi
    
    # Docker Compose v2 est normalement inclus avec docker-compose-plugin
    warn "Docker Compose non trouvé, installation du plugin..."
    
    case $PKG_MANAGER in
        apt)
            apt-get install -y docker-compose-plugin
            ;;
        dnf)
            dnf install -y docker-compose-plugin
            ;;
        pacman)
            pacman -S --noconfirm docker-compose
            ;;
    esac
    
    log "Docker Compose installé"
}

# Installer Go
install_go() {
    step "Vérification de Go"
    
    if command -v go &> /dev/null; then
        GO_INSTALLED=$(go version | grep -oP '\d+\.\d+' | head -1)
        if [ "$(printf '%s\n' "$GO_VERSION" "$GO_INSTALLED" | sort -V | head -n1)" = "$GO_VERSION" ]; then
            log "Go déjà installé (v$GO_INSTALLED)"
            return
        fi
        warn "Go version $GO_INSTALLED trouvée, mise à jour vers $GO_VERSION+"
    fi
    
    info "Installation de Go $GO_VERSION..."
    
    # Télécharger Go
    GO_ARCHIVE="go1.22.5.linux-amd64.tar.gz"
    wget -q "https://go.dev/dl/$GO_ARCHIVE" -O /tmp/$GO_ARCHIVE
    
    # Supprimer l'ancienne installation
    rm -rf /usr/local/go
    
    # Extraire
    tar -C /usr/local -xzf /tmp/$GO_ARCHIVE
    rm /tmp/$GO_ARCHIVE
    
    # Ajouter au PATH
    if ! grep -q '/usr/local/go/bin' /etc/profile.d/go.sh 2>/dev/null; then
        echo 'export PATH=$PATH:/usr/local/go/bin' > /etc/profile.d/go.sh
        echo 'export GOPATH=$HOME/go' >> /etc/profile.d/go.sh
        echo 'export PATH=$PATH:$GOPATH/bin' >> /etc/profile.d/go.sh
    fi
    
    export PATH=$PATH:/usr/local/go/bin
    
    log "Go $(go version | grep -oP '\d+\.\d+\.\d+') installé"
}

# Installer Node.js
install_node() {
    step "Vérification de Node.js"
    
    if command -v node &> /dev/null; then
        NODE_INSTALLED=$(node --version | grep -oP '\d+' | head -1)
        if [ "$NODE_INSTALLED" -ge "$NODE_VERSION" ]; then
            log "Node.js déjà installé (v$(node --version))"
            return
        fi
        warn "Node.js v$NODE_INSTALLED trouvé, mise à jour vers v$NODE_VERSION+"
    fi
    
    info "Installation de Node.js $NODE_VERSION..."
    
    case $PKG_MANAGER in
        apt)
            # NodeSource repo
            curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash -
            apt-get install -y nodejs
            ;;
        dnf)
            curl -fsSL https://rpm.nodesource.com/setup_${NODE_VERSION}.x | bash -
            dnf install -y nodejs
            ;;
        pacman)
            pacman -S --noconfirm nodejs npm
            ;;
    esac
    
    log "Node.js $(node --version) installé"
}

# Installer les outils WiFi pentest
install_wifi_tools() {
    step "Installation des outils WiFi pentest"
    
    case $PKG_MANAGER in
        apt)
            # Essayer d'abord l'installation standard
            $PKG_INSTALL aircrack-ng reaver pixiewps bully hashcat hcxtools hcxdumptool wifite \
                wireless-tools iw net-tools macchanger 2>/dev/null || {
                
                # Si ça échoue, essayer une version minimale
                warn "Installation complète échouée, tentative minimale..."
                $PKG_INSTALL aircrack-ng reaver hashcat wireless-tools iw net-tools 2>/dev/null || {
                    
                    # Dernier recours : ajouter le repo Kali si on est sur Ubuntu/Debian
                    if [[ "$DISTRO" == "ubuntu" || "$DISTRO" == "debian" || "$DISTRO" == "linuxmint" ]]; then
                        warn "Tentative avec le dépôt Kali..."
                        
                        # Ajouter le repo Kali si pas déjà présent
                        if ! grep -q "kali" /etc/apt/sources.list.d/* 2>/dev/null; then
                            echo "deb http://http.kali.org/kali kali-rolling main non-free contrib" > /etc/apt/sources.list.d/kali.list
                            wget -q -O - https://archive.kali.org/archive-key.asc | apt-key add - 2>/dev/null || true
                            $PKG_UPDATE > /dev/null 2>&1 || true
                        fi
                        
                        # Installer depuis Kali
                        apt-get install -y -t kali-rolling aircrack-ng 2>/dev/null || \
                        apt-get install -y aircrack-ng 2>/dev/null || \
                        error "Impossible d'installer aircrack-ng. Installez manuellement : sudo apt install aircrack-ng"
                    fi
                }
            }
            ;;
        dnf)
            $PKG_INSTALL aircrack-ng hashcat wireless-tools iw net-tools
            ;;
        pacman)
            $PKG_INSTALL aircrack-ng reaver hashcat iw net-tools macchanger
            ;;
    esac
    
    # Vérifier que aircrack-ng est bien installé
    if ! command -v airodump-ng &> /dev/null; then
        error "aircrack-ng n'a pas pu être installé. Le plugin WiFi ne fonctionnera pas.\n   Installez manuellement : sudo apt install aircrack-ng"
    fi
    
    log "Outils WiFi installés (airodump-ng: $(which airodump-ng))"
}

# Installer les dépendances du projet
install_project_deps() {
    step "Installation des dépendances du projet"
    
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    cd "$SCRIPT_DIR"
    
    # Dépendances Go (backend)
    if [ -d "backend" ]; then
        info "Installation des dépendances Go..."
        cd backend
        export PATH=$PATH:/usr/local/go/bin
        go mod download
        cd ..
        log "Dépendances Go installées"
    fi
    
    # Dépendances Node (frontend)
    if [ -d "frontend" ]; then
        info "Installation des dépendances Node.js..."
        cd frontend
        npm ci --legacy-peer-deps
        cd ..
        log "Dépendances Node.js installées"
    fi
}

# Créer le fichier .env
setup_env() {
    step "Configuration de l'environnement"
    
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    cd "$SCRIPT_DIR"
    
    if [ ! -f .env ]; then
        if [ -f .env.example ]; then
            cp .env.example .env
        else
            cat > .env << 'EOF'
# Heimdall Configuration
APP_ENV=development
APP_DEBUG=true
APP_PORT=3000

# Database
DB_HOST=localhost
DB_PORT=5433
DB_NAME=heimdall_dev
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT (change in production!)
JWT_SECRET=CHANGE_ME_IN_PRODUCTION_USE_OPENSSL_RAND
JWT_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# Frontend
NEXT_PUBLIC_API_URL=http://localhost:3000/api/v1

# Workers
WORKER_CONCURRENCY=5
WORKER_QUEUE_NAME=heimdall_jobs
EOF
        fi
        
        # Générer un secret JWT aléatoire
        JWT_SECRET=$(openssl rand -hex 32)
        sed -i "s/CHANGE_ME_IN_PRODUCTION_USE_OPENSSL_RAND/$JWT_SECRET/" .env
        
        log "Fichier .env créé"
    else
        log "Fichier .env existant conservé"
    fi
}

# Créer les répertoires
setup_directories() {
    step "Création des répertoires"

    WORDLISTS_DIR_DEFAULT="/opt/heimdall/wordlists"
    WORDLISTS_DIR_EFFECTIVE="$WORDLISTS_DIR_DEFAULT"
    if [ -n "${WORDLISTS_DIR}" ]; then
        WORDLISTS_DIR_EFFECTIVE="$WORDLISTS_DIR"
    elif [ -n "${HEIMDALL_ROOT}" ]; then
        WORDLISTS_DIR_EFFECTIVE="$HEIMDALL_ROOT/wordlists"
    fi
    
    mkdir -p /opt/heimdall/captures
    mkdir -p "$WORDLISTS_DIR_EFFECTIVE"
    mkdir -p /opt/heimdall/logs
    
    # Télécharger rockyou si absent
    if [ ! -f "$WORDLISTS_DIR_EFFECTIVE/rockyou.txt" ]; then
        if [ -f /usr/share/wordlists/rockyou.txt.gz ]; then
            info "Extraction de rockyou.txt..."
            gunzip -c /usr/share/wordlists/rockyou.txt.gz > "$WORDLISTS_DIR_EFFECTIVE/rockyou.txt"
        elif [ -f /usr/share/wordlists/rockyou.txt ]; then
            cp /usr/share/wordlists/rockyou.txt "$WORDLISTS_DIR_EFFECTIVE/"
        fi
    fi

    # Télécharger / mettre à jour des wordlists (FR/EN/ES) depuis Git
    if [ "${WORDLISTS_AUTO_DOWNLOAD:-1}" = "1" ]; then
        info "Téléchargement des wordlists (FR/EN/ES)"

        download_wordlist() {
            local name="$1"
            local url="$2"
            local dest="$WORDLISTS_DIR_EFFECTIVE/$name"

            if command -v curl &> /dev/null; then
                if [ -f "$dest" ]; then
                    curl -L -z "$dest" -o "$dest" "$url" || true
                else
                    curl -L -o "$dest" "$url" || true
                fi
            elif command -v wget &> /dev/null; then
                wget -q -N -O "$dest" "$url" || true
            else
                warn "curl/wget introuvable, skip téléchargement wordlists"
            fi
        }

        # FR
        download_wordlist "french-common-password-list-top-5000.txt" \
            "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Passwords/Common-Credentials/Language-Specific/French-common-password-list-top-5000.txt"

        # EN (général)
        download_wordlist "top1575-probable-v2.txt" \
            "https://github.com/berzerk0/Probable-Wordlists/raw/master/Real-Passwords/Top1575-probable-v2.txt"

        # ES
        download_wordlist "Spanish_1000-common-usernames-and-passwords.txt" \
            "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Passwords/Common-Credentials/Language-Specific/Spanish_1000-common-usernames-and-passwords.txt"

        # WiFi common (petit, utile)
        download_wordlist "wifi-common.txt" \
            "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Passwords/WiFi-WPA/wifi-common.txt"
    else
        info "Téléchargement wordlists désactivé (WORDLISTS_AUTO_DOWNLOAD=0)"
    fi
    
    # Permissions
    REAL_USER=${SUDO_USER:-$USER}
    if [ "$REAL_USER" != "root" ]; then
        chown -R "$REAL_USER:$REAL_USER" /opt/heimdall
        chown -R "$REAL_USER:$REAL_USER" "$WORDLISTS_DIR_EFFECTIVE"
    fi
    
    log "Répertoires créés"
}

# Rendre les scripts exécutables
setup_scripts() {
    step "Configuration des scripts"
    
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    cd "$SCRIPT_DIR"
    
    chmod +x run_heimdall.sh 2>/dev/null || true
    chmod +x stop_heimdall.sh 2>/dev/null || true
    
    log "Scripts configurés"
}

# Résumé final
show_summary() {
    echo -e "\n${GREEN}${BOLD}"
    echo "╔═══════════════════════════════════════════════════════════════╗"
    echo "║              ✅  INSTALLATION TERMINÉE                        ║"
    echo "╚═══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    echo -e "${CYAN}Versions installées:${NC}"
    echo -e "  • Docker:       $(docker --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' | head -1 || echo 'non installé')"
    echo -e "  • Compose:      $(docker compose version --short 2>/dev/null || echo 'non installé')"
    echo -e "  • Go:           $(go version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' || echo 'non installé')"
    echo -e "  • Node.js:      $(node --version 2>/dev/null || echo 'non installé')"
    echo -e "  • npm:          $(npm --version 2>/dev/null || echo 'non installé')"
    
    echo -e "\n${CYAN}Pour démarrer Heimdall:${NC}"
    echo -e "  ${YELLOW}sudo ./run_heimdall.sh${NC}"
    
    echo -e "\n${CYAN}Accès:${NC}"
    echo -e "  • Frontend:     ${YELLOW}http://localhost:3001${NC}"
    echo -e "  • API Backend:  ${YELLOW}http://localhost:3000${NC}"
    
    echo -e "\n${CYAN}Identifiants:${NC}"
    echo -e "  • Email:        ${YELLOW}admin@heimdall.local${NC}"
    echo -e "  • Password:     ${YELLOW}admin123${NC}"
    
    echo -e "\n${RED}⚠️  N'oubliez pas de changer le mot de passe par défaut!${NC}"
    echo ""
}

# Main
main() {
    show_banner
    check_root
    detect_distro
    install_base_packages
    install_docker
    install_docker_compose
    install_go
    install_node
    install_wifi_tools
    install_project_deps
    setup_env
    setup_directories
    setup_scripts
    show_summary
}

main "$@"
