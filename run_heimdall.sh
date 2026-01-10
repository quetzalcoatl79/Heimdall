#!/bin/bash
#
# ╔═══════════════════════════════════════════════════════════════╗
# ║                    🛡️  HEIMDALL RUN                           ║
# ║              Démarrage des services Heimdall                   ║
# ╚═══════════════════════════════════════════════════════════════╝
#
# Usage: sudo ./run_heimdall.sh [--dev|--prod]
#

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="/opt/heimdall/logs"
PID_DIR="/var/run/heimdall"

# Fonctions utilitaires
log()     { echo -e "${GREEN}[✓]${NC} $1"; }
info()    { echo -e "${BLUE}[i]${NC} $1"; }
warn()    { echo -e "${YELLOW}[!]${NC} $1"; }
error()   { echo -e "${RED}[✗]${NC} $1"; exit 1; }

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
                                                                   
                    🚀  Starting Services...
    
EOF
    echo -e "${NC}"
}

# Vérifier les prérequis
check_prerequisites() {
    info "Vérification des prérequis..."
    
    # Docker
    if ! command -v docker &> /dev/null; then
        error "Docker n'est pas installé. Lancez d'abord: sudo ./setup.sh"
    fi
    
    # Docker daemon
    if ! docker info &> /dev/null; then
        error "Le daemon Docker n'est pas démarré. Lancez: sudo systemctl start docker"
    fi
    
    # Go
    if ! command -v go &> /dev/null; then
        export PATH=$PATH:/usr/local/go/bin
        if ! command -v go &> /dev/null; then
            error "Go n'est pas installé. Lancez d'abord: sudo ./setup.sh"
        fi
    fi
    
    # Node
    if ! command -v node &> /dev/null; then
        error "Node.js n'est pas installé. Lancez d'abord: sudo ./setup.sh"
    fi
    
    # .env
    if [ ! -f "$SCRIPT_DIR/.env" ]; then
        error "Fichier .env manquant. Lancez d'abord: sudo ./setup.sh"
    fi
    
    log "Prérequis OK"
}

# Créer les répertoires
setup_dirs() {
    mkdir -p "$LOG_DIR"
    mkdir -p "$PID_DIR"
    
    REAL_USER=${SUDO_USER:-$USER}
    if [ "$REAL_USER" != "root" ]; then
        chown -R "$REAL_USER:$REAL_USER" "$LOG_DIR"
    fi
}

# Charger les variables d'environnement
load_env() {
    set -a
    source "$SCRIPT_DIR/.env"
    set +a
}

# Démarrer les services Docker (Redis + PostgreSQL)
start_docker_services() {
    info "Démarrage de Redis et PostgreSQL..."
    
    cd "$SCRIPT_DIR"
    
    # Arrêter les anciens containers si existants
    docker compose down --remove-orphans 2>/dev/null || true
    
    # Démarrer uniquement Redis et PostgreSQL
    docker compose up -d postgres redis
    
    # Attendre que PostgreSQL soit prêt
    info "Attente de PostgreSQL..."
    for i in {1..30}; do
        if docker compose exec -T postgres pg_isready -U postgres &> /dev/null; then
            break
        fi
        sleep 1
    done
    
    # Vérifier que Redis est prêt
    info "Attente de Redis..."
    for i in {1..10}; do
        if docker compose exec -T redis redis-cli ping &> /dev/null; then
            break
        fi
        sleep 1
    done
    
    log "Services Docker démarrés"
}

# Exécuter les migrations
run_migrations() {
    info "Exécution des migrations..."
    
    cd "$SCRIPT_DIR/backend"
    
    # Compiler l'outil de migration si nécessaire
    if [ ! -f "./bin/migrate" ]; then
        go build -o ./bin/migrate ./cmd/migrate
    fi
    
    # Exécuter les migrations
    ./bin/migrate up 2>/dev/null || {
        # Si ça échoue, essayer avec psql directement
        warn "Migration via Go échouée, tentative avec psql..."
        for f in migrations/*.up.sql; do
            docker compose exec -T postgres psql -U postgres -d heimdall_dev -f - < "$f" 2>/dev/null || true
        done
    }
    
    log "Migrations exécutées"
}

# Compiler le backend
build_backend() {
    info "Compilation du backend..."
    
    cd "$SCRIPT_DIR/backend"
    
    export CGO_ENABLED=0
    export GOOS=linux
    
    go build -o ./bin/api ./cmd/api
    go build -o ./bin/worker ./cmd/worker
    
    log "Backend compilé"
}

# Démarrer le backend Go
start_backend() {
    info "Démarrage du backend Go..."
    
    cd "$SCRIPT_DIR/backend"
    load_env
    
    # Tuer l'ancien processus si existant
    if [ -f "$PID_DIR/backend.pid" ]; then
        OLD_PID=$(cat "$PID_DIR/backend.pid")
        kill $OLD_PID 2>/dev/null || true
    fi
    
    # Démarrer en arrière-plan
    nohup ./bin/api > "$LOG_DIR/backend.log" 2>&1 &
    echo $! > "$PID_DIR/backend.pid"
    
    # Attendre que le backend soit prêt
    sleep 2
    for i in {1..10}; do
        if curl -s http://localhost:3000/api/v1/health &> /dev/null; then
            break
        fi
        sleep 1
    done
    
    log "Backend démarré (PID: $(cat $PID_DIR/backend.pid))"
}

# Démarrer le worker Go
start_worker() {
    info "Démarrage du worker..."
    
    cd "$SCRIPT_DIR/backend"
    load_env
    
    # Tuer l'ancien processus si existant
    if [ -f "$PID_DIR/worker.pid" ]; then
        OLD_PID=$(cat "$PID_DIR/worker.pid")
        kill $OLD_PID 2>/dev/null || true
    fi
    
    # Démarrer en arrière-plan
    nohup ./bin/worker > "$LOG_DIR/worker.log" 2>&1 &
    echo $! > "$PID_DIR/worker.pid"
    
    log "Worker démarré (PID: $(cat $PID_DIR/worker.pid))"
}

# Démarrer le frontend Next.js
start_frontend() {
    info "Démarrage du frontend Next.js..."
    
    cd "$SCRIPT_DIR/frontend"
    
    # Tuer l'ancien processus si existant
    if [ -f "$PID_DIR/frontend.pid" ]; then
        OLD_PID=$(cat "$PID_DIR/frontend.pid")
        kill $OLD_PID 2>/dev/null || true
    fi
    
    # Installer les dépendances si node_modules absent
    if [ ! -d "node_modules" ]; then
        npm ci --legacy-peer-deps
    fi
    
    # Mode dev ou prod
    if [ "$1" = "--prod" ]; then
        npm run build
        nohup npm run start > "$LOG_DIR/frontend.log" 2>&1 &
    else
        nohup npm run dev > "$LOG_DIR/frontend.log" 2>&1 &
    fi
    
    echo $! > "$PID_DIR/frontend.pid"
    
    # Attendre que le frontend soit prêt
    sleep 3
    
    log "Frontend démarré (PID: $(cat $PID_DIR/frontend.pid))"
}

# Afficher le status
show_status() {
    echo -e "\n${GREEN}${BOLD}"
    echo "╔═══════════════════════════════════════════════════════════════╗"
    echo "║              🚀  HEIMDALL EN COURS D'EXÉCUTION                ║"
    echo "╚═══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    echo -e "${CYAN}Services:${NC}"
    echo -e "  • PostgreSQL:   ${GREEN}●${NC} Running (Docker)"
    echo -e "  • Redis:        ${GREEN}●${NC} Running (Docker)"
    echo -e "  • Backend:      ${GREEN}●${NC} Running (PID: $(cat $PID_DIR/backend.pid 2>/dev/null || echo 'N/A'))"
    echo -e "  • Worker:       ${GREEN}●${NC} Running (PID: $(cat $PID_DIR/worker.pid 2>/dev/null || echo 'N/A'))"
    echo -e "  • Frontend:     ${GREEN}●${NC} Running (PID: $(cat $PID_DIR/frontend.pid 2>/dev/null || echo 'N/A'))"
    
    echo -e "\n${CYAN}Accès:${NC}"
    echo -e "  • Frontend:     ${YELLOW}http://localhost:3001${NC}"
    echo -e "  • API Backend:  ${YELLOW}http://localhost:3000${NC}"
    
    echo -e "\n${CYAN}Logs:${NC}"
    echo -e "  • Backend:      ${YELLOW}tail -f $LOG_DIR/backend.log${NC}"
    echo -e "  • Worker:       ${YELLOW}tail -f $LOG_DIR/worker.log${NC}"
    echo -e "  • Frontend:     ${YELLOW}tail -f $LOG_DIR/frontend.log${NC}"
    
    echo -e "\n${CYAN}Identifiants:${NC}"
    echo -e "  • Email:        ${YELLOW}admin@heimdall.local${NC}"
    echo -e "  • Password:     ${YELLOW}admin123${NC}"
    
    echo -e "\n${CYAN}Arrêter:${NC}"
    echo -e "  ${YELLOW}sudo ./stop_heimdall.sh${NC}"
    
    # Lister les interfaces WiFi
    echo -e "\n${CYAN}Interfaces WiFi détectées:${NC}"
    ip link show 2>/dev/null | grep -E "wlan|wlp|ath|wifi" | awk '{print "  • " $2}' | tr -d ':' || echo "  Aucune interface détectée"
    
    echo ""
}

# Main
main() {
    show_banner
    
    # Mode
    MODE="${1:---dev}"
    
    check_prerequisites
    setup_dirs
    start_docker_services
    run_migrations
    build_backend
    start_backend
    start_worker
    start_frontend "$MODE"
    show_status
}

main "$@"
