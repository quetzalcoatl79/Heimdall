#!/bin/bash
#
# ╔═══════════════════════════════════════════════════════════════╗
# ║                    🛡️  HEIMDALL STOP                          ║
# ║                 Arrêt des services Heimdall                    ║
# ╚═══════════════════════════════════════════════════════════════╝
#
# Usage: sudo ./stop_heimdall.sh
#

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_DIR="/var/run/heimdall"

log()     { echo -e "${GREEN}[✓]${NC} $1"; }
info()    { echo -e "${BLUE}[i]${NC} $1"; }

echo -e "${YELLOW}"
echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║              🛑  ARRÊT DE HEIMDALL                            ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Arrêter le frontend
if [ -f "$PID_DIR/frontend.pid" ]; then
    PID=$(cat "$PID_DIR/frontend.pid")
    info "Arrêt du frontend (PID: $PID)..."
    kill $PID 2>/dev/null || true
    # Tuer aussi les processus Node orphelins
    pkill -f "next dev" 2>/dev/null || true
    pkill -f "next start" 2>/dev/null || true
    rm -f "$PID_DIR/frontend.pid"
    log "Frontend arrêté"
fi

# Arrêter le worker
if [ -f "$PID_DIR/worker.pid" ]; then
    PID=$(cat "$PID_DIR/worker.pid")
    info "Arrêt du worker (PID: $PID)..."
    kill $PID 2>/dev/null || true
    rm -f "$PID_DIR/worker.pid"
    log "Worker arrêté"
fi

# Arrêter le backend
if [ -f "$PID_DIR/backend.pid" ]; then
    PID=$(cat "$PID_DIR/backend.pid")
    info "Arrêt du backend (PID: $PID)..."
    kill $PID 2>/dev/null || true
    rm -f "$PID_DIR/backend.pid"
    log "Backend arrêté"
fi

# Arrêter les services Docker
info "Arrêt des services Docker..."
cd "$SCRIPT_DIR"
docker compose down 2>/dev/null || true
log "Services Docker arrêtés"

echo -e "\n${GREEN}✅ Tous les services Heimdall sont arrêtés${NC}\n"
