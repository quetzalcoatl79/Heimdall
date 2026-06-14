#!/bin/bash
#
# ╔═══════════════════════════════════════════════════════════════════════════╗
# ║                    🔧 KALI TOOLS INSTALLER FOR UBUNTU 24.04               ║
# ║                     Transform Ubuntu into a Pentest Machine               ║
# ╚═══════════════════════════════════════════════════════════════════════════╝
#
# Usage: sudo ./install_kali_tools.sh [--all|--minimal|--wifi|--web|--recon]
#
# Author: Heimdall Project
# Tested on: Ubuntu 24.04 LTS
#

set -e

# ═══════════════════════════════════════════════════════════════════════════
# CONFIGURATION
# ═══════════════════════════════════════════════════════════════════════════

SCRIPT_VERSION="1.0.0"
LOG_FILE="/var/log/kali_tools_install.log"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m'
BOLD='\033[1m'

# ═══════════════════════════════════════════════════════════════════════════
# FONCTIONS UTILITAIRES
# ═══════════════════════════════════════════════════════════════════════════

log() {
    echo -e "${GREEN}[✓]${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [SUCCESS] $1" >> "$LOG_FILE"
}

info() {
    echo -e "${BLUE}[i]${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [INFO] $1" >> "$LOG_FILE"
}

warn() {
    echo -e "${YELLOW}[!]${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [WARNING] $1" >> "$LOG_FILE"
}

error() {
    echo -e "${RED}[✗]${NC} $1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [ERROR] $1" >> "$LOG_FILE"
    exit 1
}

step() {
    echo -e "\n${PURPLE}${BOLD}══════════════════════════════════════════════════════════════${NC}"
    echo -e "${PURPLE}${BOLD}▶ $1${NC}"
    echo -e "${PURPLE}${BOLD}══════════════════════════════════════════════════════════════${NC}\n"
}

# Banner
show_banner() {
    clear
    echo -e "${CYAN}"
    cat << 'EOF'
    
    ██╗  ██╗ █████╗ ██╗     ██╗    ████████╗ ██████╗  ██████╗ ██╗     ███████╗
    ██║ ██╔╝██╔══██╗██║     ██║    ╚══██╔══╝██╔═══██╗██╔═══██╗██║     ██╔════╝
    █████╔╝ ███████║██║     ██║       ██║   ██║   ██║██║   ██║██║     ███████╗
    ██╔═██╗ ██╔══██║██║     ██║       ██║   ██║   ██║██║   ██║██║     ╚════██║
    ██║  ██╗██║  ██║███████╗██║       ██║   ╚██████╔╝╚██████╔╝███████╗███████║
    ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚═╝       ╚═╝    ╚═════╝  ╚═════╝ ╚══════╝╚══════╝
                                                                               
                    🔧 Kali Tools Installer for Ubuntu 24.04
                              Version: ${SCRIPT_VERSION}
    
EOF
    echo -e "${NC}"
}

# Vérification des prérequis
check_prerequisites() {
    step "Vérification des prérequis"
    
    # Vérifier si root
    if [[ $EUID -ne 0 ]]; then
        error "Ce script doit être exécuté en tant que root (sudo ./install_kali_tools.sh)"
    fi
    log "Exécution en tant que root"
    
    # Vérifier Ubuntu
    if ! grep -q "Ubuntu" /etc/os-release 2>/dev/null; then
        warn "Ce script est optimisé pour Ubuntu. Continuez à vos risques..."
    else
        UBUNTU_VERSION=$(grep VERSION_ID /etc/os-release | cut -d'"' -f2)
        log "Ubuntu $UBUNTU_VERSION détecté"
    fi
    
    # Vérifier connexion internet
    if ! ping -c 1 google.com &>/dev/null; then
        error "Connexion internet requise"
    fi
    log "Connexion internet OK"
    
    # Créer le fichier de log
    touch "$LOG_FILE" 2>/dev/null || true
}

# ═══════════════════════════════════════════════════════════════════════════
# INSTALLATION DES DÉPENDANCES DE BASE
# ═══════════════════════════════════════════════════════════════════════════

install_base_dependencies() {
    step "Installation des dépendances de base"
    
    info "Mise à jour du système..."
    apt update && apt upgrade -y
    
    info "Installation des paquets essentiels..."
    apt install -y \
        git \
        curl \
        wget \
        vim \
        nano \
        build-essential \
        gcc \
        g++ \
        make \
        cmake \
        automake \
        autoconf \
        pkg-config \
        libssl-dev \
        libffi-dev \
        libpcap-dev \
        libsqlite3-dev \
        libreadline-dev \
        libbz2-dev \
        zlib1g-dev \
        libncurses5-dev \
        libgdbm-dev \
        libnss3-dev \
        python3 \
        python3-pip \
        python3-venv \
        python3-dev \
        ruby \
        ruby-dev \
        perl \
        golang-go \
        default-jdk \
        net-tools \
        dnsutils \
        whois \
        traceroute \
        tcpdump \
        nmap \
        tree \
        jq \
        unzip \
        p7zip-full \
        htop \
        tmux \
        screen
    
    log "Dépendances de base installées"
}

# ═══════════════════════════════════════════════════════════════════════════
# AJOUT DU DÉPÔT KALI LINUX
# ═══════════════════════════════════════════════════════════════════════════

add_kali_repository() {
    step "Configuration du dépôt Kali Linux"
    
    # Ajouter la clé GPG Kali
    info "Ajout de la clé GPG Kali..."
    wget -q -O - https://archive.kali.org/archive-key.asc | gpg --dearmor -o /usr/share/keyrings/kali-archive-keyring.gpg
    
    # Ajouter le dépôt Kali
    info "Ajout du dépôt Kali..."
    cat > /etc/apt/sources.list.d/kali.list << 'EOF'
# Kali Linux Repository
deb [signed-by=/usr/share/keyrings/kali-archive-keyring.gpg] http://http.kali.org/kali kali-rolling main contrib non-free non-free-firmware
EOF
    
    # Configurer les priorités pour éviter les conflits
    info "Configuration des priorités APT..."
    cat > /etc/apt/preferences.d/kali.pref << 'EOF'
# Priorité basse pour Kali par défaut (évite de casser Ubuntu)
Package: *
Pin: release o=Kali
Pin-Priority: 50

# Priorité haute uniquement pour les outils de pentest spécifiques
Package: aircrack-ng airgeddon wifite* hashcat john* hydra* sqlmap burpsuite metasploit* nmap zenmap nikto dirb dirbuster gobuster feroxbuster ffuf nuclei subfinder amass masscan rustscan netcat-openbsd socat proxychains* tor wireshark* ettercap* bettercap* mitmproxy arpspoof dsniff macchanger responder impacket-scripts crackmapexec evil-winrm bloodhound* neo4j enum4linux smbclient smbmap rpcclient nbtscan onesixtyone snmp-mibs-downloader exploitdb searchsploit msfpc veil-* unicorn-magic shellter wpscan joomscan droopescan whatweb wafw00f arjun paramspider gau waybackurls hakrawler gospider katana httpx nuclei-templates dnsx subfinder findomain assetfinder massdns dnsrecon fierce theHarvester shodan recon-ng spiderfoot maltego osrframework exiftool binwalk foremost steghide stegseek zsteg openstego outguess pdfparser oletools yara volatility3 autopsy sleuthkit regripper chntpw samdump2 mimikatz-* powershell-empire starkiller covenant sliver havoc-* villain pwncat-cs ligolo-* chisel sshuttle proxychains-ng redsocks iodine dns2tcp ptunnel-ng udptunnel icmpsh
Pin: release o=Kali
Pin-Priority: 500
EOF
    
    apt update
    log "Dépôt Kali configuré"
}

# ═══════════════════════════════════════════════════════════════════════════
# OUTILS WIFI / WIRELESS
# ═══════════════════════════════════════════════════════════════════════════

install_wifi_tools() {
    step "Installation des outils WiFi / Wireless"
    
    info "Installation depuis les dépôts..."
    apt install -y \
        aircrack-ng \
        reaver \
        pixiewps \
        bully \
        cowpatty \
        mdk3 \
        mdk4 \
        wifite \
        hostapd \
        dnsmasq \
        macchanger \
        iw \
        wireless-tools \
        rfkill \
        wpasupplicant \
        horst \
        kismet \
        fern-wifi-cracker 2>/dev/null || warn "Certains paquets WiFi non disponibles"
    
    # Installer Airgeddon
    info "Installation d'Airgeddon..."
    if [[ ! -d /opt/airgeddon ]]; then
        git clone --depth 1 https://github.com/v1s1t0r1sh3r3/airgeddon.git /opt/airgeddon
        ln -sf /opt/airgeddon/airgeddon.sh /usr/local/bin/airgeddon
    fi
    
    # Installer Fluxion
    info "Installation de Fluxion..."
    if [[ ! -d /opt/fluxion ]]; then
        git clone --depth 1 https://github.com/FluxionNetwork/fluxion.git /opt/fluxion
        ln -sf /opt/fluxion/fluxion.sh /usr/local/bin/fluxion
    fi
    
    # Installer Wifiphisher
    info "Installation de Wifiphisher..."
    pip3 install wifiphisher 2>/dev/null || warn "Wifiphisher installation échouée"
    
    # Installer hcxtools & hcxdumptool
    info "Installation de hcxtools..."
    apt install -y hcxtools hcxdumptool 2>/dev/null || {
        git clone --depth 1 https://github.com/ZerBea/hcxtools.git /tmp/hcxtools
        cd /tmp/hcxtools && make && make install
        cd -
    }
    
    log "Outils WiFi installés"
}

# ═══════════════════════════════════════════════════════════════════════════
# OUTILS WEB / APPLICATION
# ═══════════════════════════════════════════════════════════════════════════

install_web_tools() {
    step "Installation des outils Web / Application"
    
    info "Installation depuis les dépôts..."
    apt install -y \
        nikto \
        dirb \
        gobuster \
        whatweb \
        wpscan \
        sqlmap \
        commix \
        xsser \
        skipfish 2>/dev/null || warn "Certains paquets web non disponibles"
    
    # Feroxbuster
    info "Installation de Feroxbuster..."
    if ! command -v feroxbuster &>/dev/null; then
        curl -sL https://raw.githubusercontent.com/epi052/feroxbuster/main/install-nix.sh | bash -s /usr/local/bin
    fi
    
    # FFuF
    info "Installation de FFuF..."
    go install github.com/ffuf/ffuf/v2@latest 2>/dev/null || apt install -y ffuf 2>/dev/null
    [[ -f ~/go/bin/ffuf ]] && cp ~/go/bin/ffuf /usr/local/bin/
    
    # Nuclei
    info "Installation de Nuclei..."
    go install github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest 2>/dev/null
    [[ -f ~/go/bin/nuclei ]] && cp ~/go/bin/nuclei /usr/local/bin/
    
    # httpx
    info "Installation de httpx..."
    go install github.com/projectdiscovery/httpx/cmd/httpx@latest 2>/dev/null
    [[ -f ~/go/bin/httpx ]] && cp ~/go/bin/httpx /usr/local/bin/
    
    # Arjun (parameter discovery)
    info "Installation d'Arjun..."
    pip3 install arjun 2>/dev/null || warn "Arjun installation échouée"
    
    # Dalfox (XSS scanner)
    info "Installation de Dalfox..."
    go install github.com/hahwul/dalfox/v2@latest 2>/dev/null
    [[ -f ~/go/bin/dalfox ]] && cp ~/go/bin/dalfox /usr/local/bin/
    
    # Katana (crawler)
    info "Installation de Katana..."
    go install github.com/projectdiscovery/katana/cmd/katana@latest 2>/dev/null
    [[ -f ~/go/bin/katana ]] && cp ~/go/bin/katana /usr/local/bin/
    
    log "Outils Web installés"
}

# ═══════════════════════════════════════════════════════════════════════════
# OUTILS RECONNAISSANCE / OSINT
# ═══════════════════════════════════════════════════════════════════════════

install_recon_tools() {
    step "Installation des outils Reconnaissance / OSINT"
    
    info "Installation depuis les dépôts..."
    apt install -y \
        nmap \
        masscan \
        netdiscover \
        arp-scan \
        nbtscan \
        enum4linux \
        smbclient \
        smbmap \
        ldapsearch \
        onesixtyone \
        snmpwalk \
        snmp-mibs-downloader \
        dnsenum \
        dnsrecon \
        fierce \
        theharvester \
        recon-ng \
        spiderfoot \
        maltego \
        metagoofil \
        exiftool 2>/dev/null || warn "Certains paquets recon non disponibles"
    
    # Subfinder
    info "Installation de Subfinder..."
    go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest 2>/dev/null
    [[ -f ~/go/bin/subfinder ]] && cp ~/go/bin/subfinder /usr/local/bin/
    
    # Amass
    info "Installation d'Amass..."
    go install github.com/owasp-amass/amass/v4/...@master 2>/dev/null
    [[ -f ~/go/bin/amass ]] && cp ~/go/bin/amass /usr/local/bin/
    
    # Assetfinder
    info "Installation d'Assetfinder..."
    go install github.com/tomnomnom/assetfinder@latest 2>/dev/null
    [[ -f ~/go/bin/assetfinder ]] && cp ~/go/bin/assetfinder /usr/local/bin/
    
    # Rustscan
    info "Installation de Rustscan..."
    if ! command -v rustscan &>/dev/null; then
        wget -q https://github.com/RustScan/RustScan/releases/download/2.1.1/rustscan_2.1.1_amd64.deb -O /tmp/rustscan.deb
        dpkg -i /tmp/rustscan.deb 2>/dev/null || apt install -f -y
    fi
    
    # Shodan CLI
    info "Installation de Shodan CLI..."
    pip3 install shodan 2>/dev/null
    
    # theHarvester (version récente)
    info "Installation de theHarvester..."
    if [[ ! -d /opt/theHarvester ]]; then
        git clone --depth 1 https://github.com/laramies/theHarvester.git /opt/theHarvester
        cd /opt/theHarvester
        pip3 install -r requirements.txt 2>/dev/null
        ln -sf /opt/theHarvester/theHarvester.py /usr/local/bin/theHarvester
        cd -
    fi
    
    log "Outils Reconnaissance installés"
}

# ═══════════════════════════════════════════════════════════════════════════
# OUTILS EXPLOITATION
# ═══════════════════════════════════════════════════════════════════════════

install_exploit_tools() {
    step "Installation des outils Exploitation"
    
    info "Installation depuis les dépôts..."
    apt install -y \
        exploitdb \
        metasploit-framework \
        armitage \
        beef-xss \
        set \
        netcat-openbsd \
        socat \
        pwncat 2>/dev/null || warn "Certains paquets exploit non disponibles"
    
    # Metasploit Framework (si pas déjà installé)
    if ! command -v msfconsole &>/dev/null; then
        info "Installation de Metasploit Framework..."
        curl https://raw.githubusercontent.com/rapid7/metasploit-omnibus/master/config/templates/metasploit-framework-wrappers/msfupdate.erb > /tmp/msfinstall
        chmod 755 /tmp/msfinstall
        /tmp/msfinstall
    fi
    
    # SearchSploit (exploitdb)
    info "Mise à jour de SearchSploit..."
    searchsploit -u 2>/dev/null || warn "SearchSploit update échoué"
    
    # Impacket
    info "Installation d'Impacket..."
    pip3 install impacket 2>/dev/null
    
    # CrackMapExec / NetExec
    info "Installation de NetExec (successeur de CrackMapExec)..."
    pip3 install netexec 2>/dev/null || pip3 install crackmapexec 2>/dev/null
    
    # Evil-WinRM
    info "Installation d'Evil-WinRM..."
    gem install evil-winrm 2>/dev/null
    
    # Chisel (tunneling)
    info "Installation de Chisel..."
    go install github.com/jpillora/chisel@latest 2>/dev/null
    [[ -f ~/go/bin/chisel ]] && cp ~/go/bin/chisel /usr/local/bin/
    
    # Ligolo-ng (tunneling)
    info "Installation de Ligolo-ng..."
    if [[ ! -d /opt/ligolo-ng ]]; then
        mkdir -p /opt/ligolo-ng
        wget -q https://github.com/nicocha30/ligolo-ng/releases/latest/download/ligolo-ng_proxy_0.6.1_linux_amd64.tar.gz -O /tmp/ligolo-proxy.tar.gz
        tar -xzf /tmp/ligolo-proxy.tar.gz -C /opt/ligolo-ng
        wget -q https://github.com/nicocha30/ligolo-ng/releases/latest/download/ligolo-ng_agent_0.6.1_linux_amd64.tar.gz -O /tmp/ligolo-agent.tar.gz
        tar -xzf /tmp/ligolo-agent.tar.gz -C /opt/ligolo-ng
    fi
    
    log "Outils Exploitation installés"
}

# ═══════════════════════════════════════════════════════════════════════════
# OUTILS PASSWORD CRACKING
# ═══════════════════════════════════════════════════════════════════════════

install_password_tools() {
    step "Installation des outils Password Cracking"
    
    info "Installation depuis les dépôts..."
    apt install -y \
        hashcat \
        john \
        hydra \
        medusa \
        ncrack \
        cewl \
        crunch \
        rsmangler \
        hash-identifier \
        hashid \
        ophcrack \
        fcrackzip \
        pdfcrack \
        rarcrack \
        truecrack 2>/dev/null || warn "Certains paquets password non disponibles"
    
    # Wordlists
    info "Installation des wordlists..."
    mkdir -p /usr/share/wordlists
    
    # RockYou
    if [[ ! -f /usr/share/wordlists/rockyou.txt ]]; then
        info "Téléchargement de RockYou..."
        wget -q https://github.com/brannondorsey/naive-hashcat/releases/download/data/rockyou.txt -O /usr/share/wordlists/rockyou.txt 2>/dev/null || \
        curl -sL https://gitlab.com/kalilinux/packages/wordlists/-/raw/kali/master/rockyou.txt.gz | gunzip > /usr/share/wordlists/rockyou.txt
    fi
    
    # SecLists
    if [[ ! -d /usr/share/wordlists/SecLists ]]; then
        info "Téléchargement de SecLists (peut prendre du temps)..."
        git clone --depth 1 https://github.com/danielmiessler/SecLists.git /usr/share/wordlists/SecLists
    fi
    
    log "Outils Password installés"
}

# ═══════════════════════════════════════════════════════════════════════════
# OUTILS SNIFFING / MITM
# ═══════════════════════════════════════════════════════════════════════════

install_sniffing_tools() {
    step "Installation des outils Sniffing / MITM"
    
    info "Installation depuis les dépôts..."
    apt install -y \
        wireshark \
        tshark \
        tcpdump \
        ettercap-graphical \
        ettercap-text-only \
        bettercap \
        arpwatch \
        dsniff \
        mitmproxy \
        sslstrip \
        sslsplit \
        responder \
        tcpflow \
        tcpreplay \
        ngrep 2>/dev/null || warn "Certains paquets sniffing non disponibles"
    
    # Autoriser Wireshark pour les utilisateurs non-root
    info "Configuration de Wireshark..."
    dpkg-reconfigure -f noninteractive wireshark-common 2>/dev/null || true
    
    # Bettercap (version récente)
    info "Installation de Bettercap..."
    go install github.com/bettercap/bettercap@latest 2>/dev/null
    [[ -f ~/go/bin/bettercap ]] && cp ~/go/bin/bettercap /usr/local/bin/
    
    log "Outils Sniffing installés"
}

# ═══════════════════════════════════════════════════════════════════════════
# OUTILS FORENSICS / REVERSE
# ═══════════════════════════════════════════════════════════════════════════

install_forensics_tools() {
    step "Installation des outils Forensics / Reverse Engineering"
    
    info "Installation depuis les dépôts..."
    apt install -y \
        binwalk \
        foremost \
        scalpel \
        bulk-extractor \
        autopsy \
        sleuthkit \
        volatility3 \
        yara \
        radare2 \
        gdb \
        gdb-multiarch \
        ltrace \
        strace \
        hexedit \
        ghex \
        xxd \
        strings \
        file \
        pev \
        upx-ucl \
        apktool \
        dex2jar \
        jd-gui \
        jadx 2>/dev/null || warn "Certains paquets forensics non disponibles"
    
    # Ghidra
    info "Installation de Ghidra..."
    if [[ ! -d /opt/ghidra ]]; then
        GHIDRA_VERSION="11.0.1"
        wget -q "https://github.com/NationalSecurityAgency/ghidra/releases/download/Ghidra_${GHIDRA_VERSION}_build/ghidra_${GHIDRA_VERSION}_PUBLIC_20240130.zip" -O /tmp/ghidra.zip 2>/dev/null
        unzip -q /tmp/ghidra.zip -d /opt/ 2>/dev/null
        mv /opt/ghidra_* /opt/ghidra 2>/dev/null || true
        ln -sf /opt/ghidra/ghidraRun /usr/local/bin/ghidra 2>/dev/null || true
    fi
    
    # Pwntools
    info "Installation de Pwntools..."
    pip3 install pwntools 2>/dev/null
    
    # ROPgadget
    info "Installation de ROPgadget..."
    pip3 install ROPgadget 2>/dev/null
    
    log "Outils Forensics installés"
}

# ═══════════════════════════════════════════════════════════════════════════
# OUTILS STEGANOGRAPHY
# ═══════════════════════════════════════════════════════════════════════════

install_stego_tools() {
    step "Installation des outils Steganography"
    
    info "Installation depuis les dépôts..."
    apt install -y \
        steghide \
        stegsnow \
        outguess \
        pngcheck \
        zbarimg \
        imagemagick \
        gimp \
        audacity \
        sonic-visualiser 2>/dev/null || warn "Certains paquets stego non disponibles"
    
    # Stegseek (fast steghide cracker)
    info "Installation de Stegseek..."
    if ! command -v stegseek &>/dev/null; then
        wget -q https://github.com/RickdeJager/stegseek/releases/download/v0.6/stegseek_0.6-1.deb -O /tmp/stegseek.deb
        dpkg -i /tmp/stegseek.deb 2>/dev/null || apt install -f -y
    fi
    
    # zsteg (PNG/BMP stego)
    info "Installation de zsteg..."
    gem install zsteg 2>/dev/null
    
    log "Outils Steganography installés"
}

# ═══════════════════════════════════════════════════════════════════════════
# OUTILS ANONYMAT / PRIVACY
# ═══════════════════════════════════════════════════════════════════════════

install_anonymity_tools() {
    step "Installation des outils Anonymat / Privacy"
    
    info "Installation depuis les dépôts..."
    apt install -y \
        tor \
        torbrowser-launcher \
        proxychains4 \
        privoxy \
        openvpn \
        wireguard \
        wireguard-tools 2>/dev/null || warn "Certains paquets anonymat non disponibles"
    
    # Configuration ProxyChains
    info "Configuration de ProxyChains..."
    if [[ -f /etc/proxychains4.conf ]]; then
        sed -i 's/^strict_chain/#strict_chain/' /etc/proxychains4.conf
        sed -i 's/^#dynamic_chain/dynamic_chain/' /etc/proxychains4.conf
    fi
    
    log "Outils Anonymat installés"
}

# ═══════════════════════════════════════════════════════════════════════════
# OUTILS C2 / POST-EXPLOITATION
# ═══════════════════════════════════════════════════════════════════════════

install_c2_tools() {
    step "Installation des outils C2 / Post-Exploitation"
    
    # Sliver C2
    info "Installation de Sliver C2..."
    if ! command -v sliver-server &>/dev/null; then
        curl https://sliver.sh/install | bash 2>/dev/null || warn "Sliver installation échouée"
    fi
    
    # Villain
    info "Installation de Villain..."
    if [[ ! -d /opt/Villain ]]; then
        git clone --depth 1 https://github.com/t3l3machus/Villain.git /opt/Villain
        cd /opt/Villain
        pip3 install -r requirements.txt 2>/dev/null
        cd -
    fi
    
    # Havoc C2
    info "Installation de Havoc C2..."
    if [[ ! -d /opt/Havoc ]]; then
        git clone --depth 1 https://github.com/HavocFramework/Havoc.git /opt/Havoc
        # Note: Havoc requires additional build steps
    fi
    
    # pwncat-cs
    info "Installation de pwncat-cs..."
    pip3 install pwncat-cs 2>/dev/null
    
    # LinPEAS / WinPEAS
    info "Téléchargement de PEAS..."
    mkdir -p /opt/PEAS
    wget -q https://github.com/carlospolop/PEASS-ng/releases/latest/download/linpeas.sh -O /opt/PEAS/linpeas.sh 2>/dev/null
    wget -q https://github.com/carlospolop/PEASS-ng/releases/latest/download/winPEASany.exe -O /opt/PEAS/winpeas.exe 2>/dev/null
    chmod +x /opt/PEAS/*.sh 2>/dev/null
    
    log "Outils C2 installés"
}

# ═══════════════════════════════════════════════════════════════════════════
# CONFIGURATION FINALE
# ═══════════════════════════════════════════════════════════════════════════

final_configuration() {
    step "Configuration finale"
    
    # Configurer Go path
    info "Configuration de Go..."
    echo 'export GOPATH=$HOME/go' >> /etc/profile.d/go-path.sh
    echo 'export PATH=$PATH:$GOPATH/bin:/usr/local/go/bin' >> /etc/profile.d/go-path.sh
    chmod +x /etc/profile.d/go-path.sh
    
    # Créer des alias utiles
    info "Création des alias..."
    cat >> /etc/bash.bashrc << 'EOF'

# Kali Tools Aliases
alias ll='ls -la'
alias ports='netstat -tulanp'
alias myip='curl -s ifconfig.me'
alias localip='hostname -I | awk "{print \$1}"'
alias scan='nmap -sV -sC'
alias quickscan='nmap -T4 -F'
alias serve='python3 -m http.server 8080'
alias listen='nc -lvnp'
alias msfstart='sudo systemctl start postgresql && msfdb init && msfconsole'
alias updatekali='sudo apt update && sudo apt upgrade -y'
alias wordlists='ls /usr/share/wordlists/'
EOF
    
    # Créer un fichier de référence rapide
    info "Création du guide de référence..."
    cat > /opt/kali-tools-reference.txt << 'EOF'
╔═══════════════════════════════════════════════════════════════════════════╗
║                    🔧 KALI TOOLS QUICK REFERENCE                          ║
╚═══════════════════════════════════════════════════════════════════════════╝

📡 WIFI TOOLS:
   airmon-ng start wlan0          # Mode monitor
   airodump-ng wlan0mon           # Scan réseaux
   aircrack-ng capture.cap        # Crack WPA
   wifite -i wlan0                # Auto WiFi attack
   airgeddon                      # GUI WiFi toolkit

🌐 WEB TOOLS:
   nikto -h http://target         # Web scanner
   gobuster dir -u URL -w WORDLIST
   ffuf -u URL/FUZZ -w WORDLIST
   sqlmap -u "URL?id=1" --dbs
   nuclei -u URL                  # Vuln scanner

🔍 RECON TOOLS:
   nmap -sV -sC target            # Service scan
   rustscan -a target             # Fast port scan
   subfinder -d domain.com        # Subdomain enum
   theHarvester -d domain -b all
   amass enum -d domain.com

💥 EXPLOITATION:
   msfconsole                     # Metasploit
   searchsploit <keyword>         # Search exploits
   netexec smb target -u user -p pass

🔐 PASSWORD CRACKING:
   hashcat -m 0 hash.txt wordlist # MD5
   john --wordlist=rockyou hash
   hydra -l user -P pass.txt ssh://target

📁 WORDLISTS:
   /usr/share/wordlists/rockyou.txt
   /usr/share/wordlists/SecLists/

🛠️ USEFUL DIRECTORIES:
   /opt/                          # Custom tools
   /usr/share/wordlists/          # Wordlists
   /var/log/kali_tools_install.log # Install log

EOF
    
    log "Configuration finale terminée"
}

# ═══════════════════════════════════════════════════════════════════════════
# MENU PRINCIPAL
# ═══════════════════════════════════════════════════════════════════════════

show_menu() {
    echo -e "\n${WHITE}${BOLD}Sélectionnez le type d'installation:${NC}\n"
    echo -e "  ${GREEN}1)${NC} Installation complète (tous les outils)"
    echo -e "  ${GREEN}2)${NC} Installation minimale (outils essentiels)"
    echo -e "  ${GREEN}3)${NC} Outils WiFi uniquement"
    echo -e "  ${GREEN}4)${NC} Outils Web uniquement"
    echo -e "  ${GREEN}5)${NC} Outils Reconnaissance uniquement"
    echo -e "  ${GREEN}6)${NC} Outils Password Cracking uniquement"
    echo -e "  ${GREEN}7)${NC} Outils Exploitation uniquement"
    echo -e "  ${GREEN}8)${NC} Installation personnalisée"
    echo -e "  ${RED}9)${NC} Quitter"
    echo ""
}

custom_install() {
    echo -e "\n${WHITE}${BOLD}Sélectionnez les catégories à installer:${NC}\n"
    
    local install_wifi=false
    local install_web=false
    local install_recon=false
    local install_exploit=false
    local install_password=false
    local install_sniffing=false
    local install_forensics=false
    local install_stego=false
    local install_anonymity=false
    local install_c2=false
    
    read -p "Installer les outils WiFi? (y/n): " choice
    [[ "$choice" =~ ^[Yy]$ ]] && install_wifi=true
    
    read -p "Installer les outils Web? (y/n): " choice
    [[ "$choice" =~ ^[Yy]$ ]] && install_web=true
    
    read -p "Installer les outils Reconnaissance? (y/n): " choice
    [[ "$choice" =~ ^[Yy]$ ]] && install_recon=true
    
    read -p "Installer les outils Exploitation? (y/n): " choice
    [[ "$choice" =~ ^[Yy]$ ]] && install_exploit=true
    
    read -p "Installer les outils Password Cracking? (y/n): " choice
    [[ "$choice" =~ ^[Yy]$ ]] && install_password=true
    
    read -p "Installer les outils Sniffing/MITM? (y/n): " choice
    [[ "$choice" =~ ^[Yy]$ ]] && install_sniffing=true
    
    read -p "Installer les outils Forensics? (y/n): " choice
    [[ "$choice" =~ ^[Yy]$ ]] && install_forensics=true
    
    read -p "Installer les outils Steganography? (y/n): " choice
    [[ "$choice" =~ ^[Yy]$ ]] && install_stego=true
    
    read -p "Installer les outils Anonymat? (y/n): " choice
    [[ "$choice" =~ ^[Yy]$ ]] && install_anonymity=true
    
    read -p "Installer les outils C2? (y/n): " choice
    [[ "$choice" =~ ^[Yy]$ ]] && install_c2=true
    
    install_base_dependencies
    add_kali_repository
    
    $install_wifi && install_wifi_tools
    $install_web && install_web_tools
    $install_recon && install_recon_tools
    $install_exploit && install_exploit_tools
    $install_password && install_password_tools
    $install_sniffing && install_sniffing_tools
    $install_forensics && install_forensics_tools
    $install_stego && install_stego_tools
    $install_anonymity && install_anonymity_tools
    $install_c2 && install_c2_tools
    
    final_configuration
}

install_all() {
    install_base_dependencies
    add_kali_repository
    install_wifi_tools
    install_web_tools
    install_recon_tools
    install_exploit_tools
    install_password_tools
    install_sniffing_tools
    install_forensics_tools
    install_stego_tools
    install_anonymity_tools
    install_c2_tools
    final_configuration
}

install_minimal() {
    install_base_dependencies
    add_kali_repository
    
    step "Installation minimale"
    apt install -y \
        nmap \
        masscan \
        nikto \
        gobuster \
        sqlmap \
        hydra \
        john \
        hashcat \
        netcat-openbsd \
        wireshark \
        tcpdump \
        aircrack-ng \
        metasploit-framework \
        exploitdb \
        proxychains4 \
        tor
    
    # Wordlists essentielles
    mkdir -p /usr/share/wordlists
    wget -q https://github.com/brannondorsey/naive-hashcat/releases/download/data/rockyou.txt -O /usr/share/wordlists/rockyou.txt 2>/dev/null || true
    
    final_configuration
}

# ═══════════════════════════════════════════════════════════════════════════
# POINT D'ENTRÉE
# ═══════════════════════════════════════════════════════════════════════════

main() {
    show_banner
    check_prerequisites
    
    # Traiter les arguments CLI
    case "${1:-}" in
        --all)
            install_all
            ;;
        --minimal)
            install_minimal
            ;;
        --wifi)
            install_base_dependencies
            add_kali_repository
            install_wifi_tools
            final_configuration
            ;;
        --web)
            install_base_dependencies
            add_kali_repository
            install_web_tools
            final_configuration
            ;;
        --recon)
            install_base_dependencies
            add_kali_repository
            install_recon_tools
            final_configuration
            ;;
        *)
            # Mode interactif
            while true; do
                show_menu
                read -p "Choix [1-9]: " choice
                
                case $choice in
                    1) install_all; break ;;
                    2) install_minimal; break ;;
                    3) install_base_dependencies; add_kali_repository; install_wifi_tools; final_configuration; break ;;
                    4) install_base_dependencies; add_kali_repository; install_web_tools; final_configuration; break ;;
                    5) install_base_dependencies; add_kali_repository; install_recon_tools; final_configuration; break ;;
                    6) install_base_dependencies; add_kali_repository; install_password_tools; final_configuration; break ;;
                    7) install_base_dependencies; add_kali_repository; install_exploit_tools; final_configuration; break ;;
                    8) custom_install; break ;;
                    9) info "Au revoir!"; exit 0 ;;
                    *) warn "Choix invalide" ;;
                esac
            done
            ;;
    esac
    
    # Résumé final
    echo ""
    step "Installation terminée! 🎉"
    echo -e "${GREEN}"
    cat << 'EOF'
    ╔═══════════════════════════════════════════════════════════════════════╗
    ║                    ✅ INSTALLATION RÉUSSIE!                           ║
    ╠═══════════════════════════════════════════════════════════════════════╣
    ║                                                                       ║
    ║  📋 Guide de référence: /opt/kali-tools-reference.txt                ║
    ║  📁 Wordlists: /usr/share/wordlists/                                 ║
    ║  📝 Log: /var/log/kali_tools_install.log                             ║
    ║                                                                       ║
    ║  🔄 Redémarrez votre terminal pour charger les alias                 ║
    ║                                                                       ║
    ║  ⚠️  Utilisez ces outils de manière éthique et légale uniquement!    ║
    ║                                                                       ║
    ╚═══════════════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
    
    read -p "Voulez-vous redémarrer le système maintenant? (y/n): " reboot_choice
    [[ "$reboot_choice" =~ ^[Yy]$ ]] && reboot
}

# Exécution
main "$@"
