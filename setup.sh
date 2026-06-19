#!/usr/bin/env bash
# ──────────────────────────────────────────────────────────────────────────────
# OneforDrive — One-click dev environment setup
#
# Installs Docker, Python 3, Node.js 20, Go 1.22, generates .env, and
# starts all services via Docker Compose.
#
# Usage:
#   chmod +x setup.sh && ./setup.sh
#
# Supports: Ubuntu/Debian, macOS (Intel + Apple Silicon)
# ──────────────────────────────────────────────────────────────────────────────
set -euo pipefail

# ── Colours & helpers ────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[0;36m'; BOLD='\033[1m'; NC='\033[0m'

info()  { echo -e "${CYAN}▸${NC} $*"; }
ok()    { echo -e "${GREEN}✔${NC} $*"; }
warn()  { echo -e "${YELLOW}⚠${NC} $*"; }
fail()  { echo -e "${RED}✖${NC} $*"; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# ── OS detection ─────────────────────────────────────────────────────────────
OS="$(uname -s)"
if [ "$OS" = "Linux" ]; then
    if command -v apt-get &>/dev/null; then
        PKG_MGR="apt"
    elif command -v dnf &>/dev/null; then
        PKG_MGR="dnf"
    elif command -v yum &>/dev/null; then
        PKG_MGR="yum"
    else
        PKG_MGR="none"
    fi
elif [ "$OS" = "Darwin" ]; then
    PKG_MGR="brew"
else
    fail "Unsupported OS: $OS — this script supports Linux and macOS only."
fi

echo -e "\n${BOLD}╔══════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║      OneforDrive — Dev Setup             ║${NC}"
echo -e "${BOLD}╚══════════════════════════════════════════╝${NC}\n"

# ── 0. Fix broken apt (common on fresh Ubuntu) ──────────────────────────────
if [ "$OS" = "Linux" ] && [ "$PKG_MGR" = "apt" ]; then
    if ! sudo apt-get update -qq &>/dev/null; then
        warn "apt-get update failed — fixing broken cnf-update-db hook..."
        sudo rm -f /etc/apt/apt.conf.d/50command-not-found 2>/dev/null || true
        sudo dpkg --configure -a 2>/dev/null || true
        sudo apt-get update -qq 2>/dev/null || true
    fi
fi

# ── 1. Docker ────────────────────────────────────────────────────────────────
info "Checking Docker..."
if command -v docker &>/dev/null; then
    ok "Docker already installed: $(docker --version)"
else
    info "Installing Docker..."
    if [ "$OS" = "Darwin" ]; then
        if ! command -v brew &>/dev/null; then
            fail "Homebrew not found. Install it from https://brew.sh"
        fi
        brew install --cask docker
        warn "Docker Desktop installed. Please open Docker Desktop once, then re-run this script."
        exit 0
    else
        curl -fsSL https://get.docker.com | sudo sh
        sudo usermod -aG docker "$USER" 2>/dev/null || true
        ok "Docker installed."
    fi
fi

# Docker Compose (plugin)
if docker compose version &>/dev/null; then
    ok "Docker Compose already installed: $(docker compose version --short)"
else
    info "Installing Docker Compose plugin..."
    sudo apt-get update -qq && sudo apt-get install -y docker-compose-plugin
    ok "Docker Compose installed."
fi

# Wrapper: if docker needs sudo (user just added to group, session not refreshed),
# use sg docker to run the remaining docker compose commands.
if ! docker info &>/dev/null 2>&1; then
    if groups "$USER" 2>/dev/null | grep -qw docker; then
        warn "Docker group not active in this session — using sg docker wrapper."
        docker() { command sg docker -c "docker $*"; }
    fi
fi

# ── 2. Python 3 ─────────────────────────────────────────────────────────────
info "Checking Python 3..."
if command -v python3 &>/dev/null; then
    ok "Python 3 already installed: $(python3 --version)"
else
    info "Installing Python 3..."
    if [ "$OS" = "Darwin" ]; then
        brew install python3
    elif [ "$PKG_MGR" = "apt" ]; then
        sudo apt-get update && sudo apt-get install -y python3 python3-pip
    elif [ "$PKG_MGR" = "dnf" ] || [ "$PKG_MGR" = "yum" ]; then
        sudo $PKG_MGR install -y python3 python3-pip
    fi
    ok "Python 3 installed: $(python3 --version)"
fi

info "Checking rich package..."
if python3 -c "import rich" 2>/dev/null; then
    ok "rich already installed"
else
    info "Installing rich..."
    python3 -m pip install --user rich 2>/dev/null || pip3 install rich
    ok "rich installed."
fi

# ── 3. Node.js 20 ───────────────────────────────────────────────────────────
info "Checking Node.js 20..."
NEED_NODE=false
if command -v node &>/dev/null; then
    NODE_VER="$(node -v | sed 's/v//' | cut -d. -f1)"
    if [ "$NODE_VER" -eq 20 ]; then
        ok "Node.js already installed: $(node -v)"
    else
        warn "Node.js v$(node -v) detected but v20 is required (v22/v24 break npm install)."
        NEED_NODE=true
    fi
else
    NEED_NODE=true
fi

if [ "$NEED_NODE" = true ]; then
    info "Installing Node.js 20..."
    if [ "$OS" = "Darwin" ]; then
        brew install node@20
        brew link --overwrite node@20
    elif [ "$PKG_MGR" = "apt" ]; then
        curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
        sudo apt-get install -y nodejs
    elif [ "$PKG_MGR" = "dnf" ] || [ "$PKG_MGR" = "yum" ]; then
        curl -fsSL https://rpm.nodesource.com/setup_20.x | sudo bash -
        sudo $PKG_MGR install -y nodejs
    fi
    ok "Node.js installed: $(node -v)"
fi

# ── 4. Go 1.22 ──────────────────────────────────────────────────────────────
info "Checking Go 1.22..."
NEED_GO=false
if command -v go &>/dev/null; then
    GO_VER="$(go version | awk '{print $3}' | sed 's/go//')"
    GO_MAJOR="$(echo "$GO_VER" | cut -d. -f1)"
    GO_MINOR="$(echo "$GO_VER" | cut -d. -f2)"
    if [ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -eq 22 ]; then
        ok "Go already installed: $(go version)"
    else
        warn "Go v$GO_VER detected but v1.22 is required."
        NEED_GO=true
    fi
else
    NEED_GO=true
fi

if [ "$NEED_GO" = true ]; then
    info "Installing Go 1.22..."
    if [ "$OS" = "Darwin" ]; then
        brew install go@1.22
        brew link --overwrite go@1.22
    else
        GO_ARCH="$(uname -m)"
        case "$GO_ARCH" in
            x86_64)  GO_ARCH="amd64" ;;
            aarch64) GO_ARCH="arm64" ;;
        esac
        GO_VERSION="1.22.12"
        GO_TARBALL="go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
        curl -fsSL "https://go.dev/dl/${GO_TARBALL}" -o "/tmp/${GO_TARBALL}"
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf "/tmp/${GO_TARBALL}"
        rm "/tmp/${GO_TARBALL}"
        # Add to PATH if not already
        if ! grep -q '/usr/local/go/bin' ~/.bashrc 2>/dev/null; then
            echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        fi
        export PATH=$PATH:/usr/local/go/bin
    fi
    ok "Go installed: $(go version)"
fi

# ── 5. Generate .env ────────────────────────────────────────────────────────
info "Checking .env file..."
if [ -f "$SCRIPT_DIR/.env" ]; then
    ok ".env already exists — skipping generation."
else
    info "Generating .env from .env.example..."

    # Generate random secrets
    JWT_SECRET="$(openssl rand -hex 32 2>/dev/null || head -c 64 /dev/urandom | xxd -p | tr -d '\n')"
    MYSQL_PASS="$(openssl rand -hex 16 2>/dev/null || head -c 32 /dev/urandom | xxd -p | tr -d '\n')"
    MYSQL_ROOT_PASS="$(openssl rand -hex 16 2>/dev/null || head -c 32 /dev/urandom | xxd -p | tr -d '\n')"

    cp "$SCRIPT_DIR/.env.example" "$SCRIPT_DIR/.env"

    # Fill in generated values
    sed -i "s/^MYSQL_ROOT_PASSWORD=.*/MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASS}/" "$SCRIPT_DIR/.env"
    sed -i "s/^MYSQL_PASSWORD=.*/MYSQL_PASSWORD=${MYSQL_PASS}/" "$SCRIPT_DIR/.env"
    sed -i "s/^DB_PASSWORD=.*/DB_PASSWORD=${MYSQL_PASS}/" "$SCRIPT_DIR/.env"
    sed -i "s/^JWT_SECRET=.*/JWT_SECRET=${JWT_SECRET}/" "$SCRIPT_DIR/.env"

    ok ".env generated with random secrets."
fi

# ── 5b. Google OAuth credentials ────────────────────────────────────────────
if [ -f "$SCRIPT_DIR/.env" ] && grep -q '^GOOGLE_CLIENT_ID=$' "$SCRIPT_DIR/.env"; then
    echo ""
    read -rp "$(echo -e "${BOLD}Google OAuth Client ID (leave blank to skip): ${NC}")" GOOGLE_ID
    read -rp "$(echo -e "${BOLD}Google OAuth Client Secret (leave blank to skip): ${NC}")" GOOGLE_SECRET

    if [ -n "$GOOGLE_ID" ] && [ -n "$GOOGLE_SECRET" ]; then
        sed -i "s|^GOOGLE_CLIENT_ID=.*|GOOGLE_CLIENT_ID=${GOOGLE_ID}|" "$SCRIPT_DIR/.env"
        sed -i "s|^GOOGLE_CLIENT_SECRET=.*|GOOGLE_CLIENT_SECRET=${GOOGLE_SECRET}|" "$SCRIPT_DIR/.env"
        ok "Google OAuth credentials saved."
    else
        warn "Skipping Google OAuth — set them later in .env for Drive integration."
    fi
fi

# ── 6. Domain & SSL (optional) ──────────────────────────────────────────────
SETUP_DOMAIN=false
DOMAIN_NAME=""
SETUP_EMAIL=""

if [ "$OS" = "Linux" ]; then
    echo ""
    read -rp "$(echo -e "${BOLD}Do you want to set up a domain with SSL? [y/N]: ${NC}")" DOMAIN_ANSWER
    if [[ "$DOMAIN_ANSWER" =~ ^[Yy]$ ]]; then
        SETUP_DOMAIN=true
        read -rp "$(echo -e "${BOLD}Enter your domain (e.g. drive.example.com): ${NC}")" DOMAIN_NAME
        read -rp "$(echo -e "${BOLD}Enter your email for Let's Encrypt (e.g. admin@example.com): ${NC}")" SETUP_EMAIL

        if [ -z "$DOMAIN_NAME" ] || [ -z "$SETUP_EMAIL" ]; then
            warn "Domain or email empty — skipping SSL setup."
            SETUP_DOMAIN=false
        fi
    fi
fi

if [ "$SETUP_DOMAIN" = true ]; then
    info "Setting up Nginx + SSL for ${DOMAIN_NAME}..."

    # Install Nginx
    if ! command -v nginx &>/dev/null; then
        info "Installing Nginx..."
        if [ "$PKG_MGR" = "apt" ]; then
            sudo apt-get update && sudo apt-get install -y nginx
        elif [ "$PKG_MGR" = "dnf" ] || [ "$PKG_MGR" = "yum" ]; then
            sudo $PKG_MGR install -y nginx
        fi
    fi
    ok "Nginx installed."

    # Write Nginx config (HTTP only first — Certbot needs port 80)
    sudo tee /etc/nginx/sites-available/routestorage > /dev/null <<NGINX_CONF
server {
    listen 80;
    server_name ${DOMAIN_NAME};

    # Frontend
    location / {
        proxy_pass         http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header   Upgrade \$http_upgrade;
        proxy_set_header   Connection "upgrade";
        proxy_set_header   Host \$host;
        proxy_set_header   X-Real-IP \$remote_addr;
        proxy_set_header   X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto \$scheme;
    }

    # Backend API
    location /api/ {
        proxy_pass         http://127.0.0.1:8081;
        proxy_http_version 1.1;
        proxy_set_header   Host \$host;
        proxy_set_header   X-Real-IP \$remote_addr;
        proxy_set_header   X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto \$scheme;

        # Upload size (256 MB chunks)
        client_max_body_size 300m;
    }
}
NGINX_CONF

    # Enable site
    sudo ln -sf /etc/nginx/sites-available/routestorage /etc/nginx/sites-enabled/routestorage
    sudo rm -f /etc/nginx/sites-enabled/default 2>/dev/null || true
    sudo nginx -t && sudo systemctl reload nginx
    ok "Nginx configured."

    # Pre-flight DNS check
    info "Checking DNS for ${DOMAIN_NAME}..."
    SERVER_IP="$(curl -s -4 ifconfig.me 2>/dev/null || curl -s -4 icanhazip.com 2>/dev/null || echo "")"
    DOMAIN_IP=""
    if command -v dig &>/dev/null; then
        DOMAIN_IP="$(dig +short "$DOMAIN_NAME" A 2>/dev/null | tail -n1)"
    elif command -v nslookup &>/dev/null; then
        DOMAIN_IP="$(nslookup "$DOMAIN_NAME" 2>/dev/null | awk '/^Address: / { print $2 }' | tail -n1)"
    fi

    if [ -n "$SERVER_IP" ] && [ -n "$DOMAIN_IP" ] && [ "$SERVER_IP" != "$DOMAIN_IP" ]; then
        warn "DNS mismatch: ${DOMAIN_NAME} resolves to ${DOMAIN_IP}, but this server's IP is ${SERVER_IP}."
        warn "Certbot will likely fail. Update your DNS A record, then re-run this script."
        read -rp "$(echo -e "${BOLD}Continue anyway? [y/N]: ${NC}")" CONTINUE_ANYWAY
        if [[ ! "$CONTINUE_ANYWAY" =~ ^[Yy]$ ]]; then
            SETUP_DOMAIN=false
        fi
    elif [ -n "$SERVER_IP" ] && [ -n "$DOMAIN_IP" ]; then
        ok "DNS OK: ${DOMAIN_NAME} → ${DOMAIN_IP}"
    else
        warn "Could not verify DNS — proceeding anyway."
    fi
fi

if [ "$SETUP_DOMAIN" = true ]; then
    # Install Certbot via snap (bundles its own Python — avoids system Python version conflicts)
    info "Installing Certbot..."
    if ! command -v snap &>/dev/null; then
        if [ "$PKG_MGR" = "apt" ]; then
            sudo apt-get install -y snapd
        elif [ "$PKG_MGR" = "dnf" ] || [ "$PKG_MGR" = "yum" ]; then
            sudo $PKG_MGR install -y snapd
            sudo systemctl enable --now snapd.socket
            sudo ln -sf /var/lib/snapd/snap /snap
        fi
    fi

    sudo snap install core 2>/dev/null || true
    sudo snap refresh core 2>/dev/null || true

    # Remove any conflicting apt-installed certbot first
    sudo apt-get remove -y certbot python3-certbot-nginx 2>/dev/null || true

    sudo snap install --classic certbot
    sudo ln -sf /snap/bin/certbot /usr/bin/certbot
    ok "Certbot installed (via snap)."

    # Obtain SSL certificate
    info "Requesting SSL certificate for ${DOMAIN_NAME}..."
    SSL_SUCCESS=false
    if sudo certbot --nginx -d "$DOMAIN_NAME" --non-interactive --agree-tos --email "$SETUP_EMAIL"; then
        SSL_SUCCESS=true
    else
        warn "Certbot failed — check DNS A record points to this server."
        warn "You can retry manually: sudo certbot --nginx -d ${DOMAIN_NAME}"
    fi

    if [ "$SSL_SUCCESS" = true ]; then
        # Set up auto-renewal
        if command -v systemctl &>/dev/null; then
            sudo systemctl enable certbot.timer 2>/dev/null || true
            sudo systemctl start certbot.timer 2>/dev/null || true
            ok "Certbot auto-renewal enabled."
        fi

        # Update .env with production URLs
        if [ -f "$SCRIPT_DIR/.env" ]; then
            sed -i "s|^FRONTEND_URL=.*|FRONTEND_URL=https://${DOMAIN_NAME}|" "$SCRIPT_DIR/.env"
            sed -i "s|^NUXT_PUBLIC_API_BASE=.*|NUXT_PUBLIC_API_BASE=https://${DOMAIN_NAME}|" "$SCRIPT_DIR/.env"
            sed -i "s|^GOOGLE_REDIRECT_URL=.*|GOOGLE_REDIRECT_URL=https://${DOMAIN_NAME}/api/accounts/oauth/callback|" "$SCRIPT_DIR/.env"
        fi

        ok "Domain + SSL setup complete: https://${DOMAIN_NAME}"
    else
        warn "SSL setup incomplete — site will run on HTTP only for now: http://${DOMAIN_NAME}"
        warn "Fix DNS, then retry: sudo certbot --nginx -d ${DOMAIN_NAME}"
        SETUP_DOMAIN=false   # so the final summary doesn't show a broken https:// URL
    fi
fi

# ── 7. Start services ───────────────────────────────────────────────────────
info "Starting services with Docker Compose..."
docker compose up --build -d

# Wait for health checks
info "Waiting for MySQL and Redis to become healthy..."
TRIES=0
MAX_TRIES=60
until docker compose ps mysql | grep -q "healthy" 2>/dev/null; do
    TRIES=$((TRIES + 1))
    if [ "$TRIES" -ge "$MAX_TRIES" ]; then
        warn "MySQL health check timed out — services may still be starting."
        break
    fi
    sleep 2
done

until docker compose ps redis | grep -q "healthy" 2>/dev/null; do
    TRIES=$((TRIES + 1))
    if [ "$TRIES" -ge "$MAX_TRIES" ]; then
        warn "Redis health check timed out — services may still be starting."
        break
    fi
    sleep 2
done

ok "All services are up."

# ── Summary ──────────────────────────────────────────────────────────────────
echo ""
if [ "$SETUP_DOMAIN" = true ]; then
    SITE_URL="https://${DOMAIN_NAME}"
elif [ -n "$DOMAIN_NAME" ]; then
    SITE_URL="http://${DOMAIN_NAME}"
else
    SITE_URL="http://localhost:3000"
fi

echo -e "${BOLD}╔══════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║         Setup Complete! 🚀               ║${NC}"
echo -e "${BOLD}╠══════════════════════════════════════════╣${NC}"
echo -e "${BOLD}║${NC}  Site      →  ${SITE_URL}       ${BOLD}║${NC}"
echo -e "${BOLD}║${NC}  Backend   →  http://localhost:8081      ${BOLD}║${NC}"
echo -e "${BOLD}║${NC}  MySQL     →  localhost:3306             ${BOLD}║${NC}"
echo -e "${BOLD}║${NC}  Redis     →  localhost:6379             ${BOLD}║${NC}"
echo -e "${BOLD}╠══════════════════════════════════════════╣${NC}"
echo -e "${BOLD}║${NC}  DB CLI    →  python3 db.py              ${BOLD}║${NC}"
echo -e "${BOLD}║${NC}  Logs      →  docker compose logs -f    ${BOLD}║${NC}"
echo -e "${BOLD}║${NC}  Stop      →  docker compose down       ${BOLD}║${NC}"
echo -e "${BOLD}╚══════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo -e "  1. Set ${BOLD}GOOGLE_CLIENT_ID${NC} and ${BOLD}GOOGLE_CLIENT_SECRET${NC} in .env"
if [ "$SETUP_DOMAIN" = true ]; then
    echo -e "  2. Open ${BOLD}https://${DOMAIN_NAME}${NC} and register an account"
elif [ -n "$DOMAIN_NAME" ]; then
    echo -e "  2. Open ${BOLD}http://${DOMAIN_NAME}${NC} and register an account"
    echo -e "     (SSL not active — run ${BOLD}sudo certbot --nginx -d ${DOMAIN_NAME}${NC} after DNS is pointed)"
else
    echo -e "  2. Open http://localhost:3000 and register an account"
fi
echo -e "  3. Run ${BOLD}python3 db.py${NC} to manage the database"
echo ""
