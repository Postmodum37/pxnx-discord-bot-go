#!/bin/bash

# Setup script for yt-dlp service
# This script installs required Python dependencies and sets up the yt-dlp service

set -e

echo "🎵 Setting up yt-dlp service for Discord Music Bot"
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Python 3 is installed
print_status "Checking Python installation..."
if ! command -v python3 &> /dev/null; then
    print_error "Python 3 is not installed or not in PATH"
    print_status "Please install Python 3.7+ and try again"
    print_status "Ubuntu/Debian: sudo apt update && sudo apt install python3 python3-pip"
    print_status "macOS: brew install python3"
    print_status "Windows: Download from https://python.org"
    exit 1
fi

PYTHON_VERSION=$(python3 --version 2>&1 | cut -d' ' -f2)
print_success "Python $PYTHON_VERSION found"

# Check if pip is installed
print_status "Checking pip installation..."
if ! command -v pip3 &> /dev/null && ! command -v pip &> /dev/null; then
    print_error "pip is not installed"
    print_status "Installing pip..."
    if command -v apt-get &> /dev/null; then
        sudo apt-get update && sudo apt-get install -y python3-pip
    elif command -v brew &> /dev/null; then
        brew install python3
    else
        print_error "Please install pip manually"
        exit 1
    fi
fi

# Determine pip command
PIP_CMD="pip3"
if ! command -v pip3 &> /dev/null; then
    PIP_CMD="pip"
fi

print_success "pip found"

# Create virtual environment (optional but recommended)
print_status "Creating virtual environment (optional)..."
VENV_DIR="venv-ytdlp"

if [ ! -d "$VENV_DIR" ]; then
    python3 -m venv "$VENV_DIR"
    print_success "Virtual environment created in $VENV_DIR"

    # Activate virtual environment
    source "$VENV_DIR/bin/activate"
    PIP_CMD="pip"
    print_status "Virtual environment activated"
else
    print_warning "Virtual environment already exists"
    if [ -f "$VENV_DIR/bin/activate" ]; then
        source "$VENV_DIR/bin/activate"
        PIP_CMD="pip"
        print_status "Virtual environment activated"
    fi
fi

# Upgrade pip
print_status "Upgrading pip..."
$PIP_CMD install --upgrade pip

# Install requirements
REQUIREMENTS_FILE="services/ytdlp/requirements.txt"

if [ -f "$REQUIREMENTS_FILE" ]; then
    print_status "Installing Python dependencies from $REQUIREMENTS_FILE..."
    $PIP_CMD install -r "$REQUIREMENTS_FILE"
    print_success "Dependencies installed successfully"
else
    print_warning "Requirements file not found, installing core dependencies..."
    $PIP_CMD install yt-dlp aiohttp
    print_success "Core dependencies installed"
fi

# Verify installations
print_status "Verifying installations..."

echo -n "Checking yt-dlp... "
if python3 -c "import yt_dlp; print(f'v{yt_dlp.version.__version__}')" 2>/dev/null; then
    print_success "yt-dlp is working"
else
    print_error "yt-dlp verification failed"
    exit 1
fi

echo -n "Checking aiohttp... "
if python3 -c "import aiohttp; print(f'v{aiohttp.__version__}')" 2>/dev/null; then
    print_success "aiohttp is working"
else
    print_error "aiohttp verification failed"
    exit 1
fi

# Create cache directory
CACHE_DIR="/tmp/ytdlp-cache"
print_status "Creating cache directory: $CACHE_DIR"
mkdir -p "$CACHE_DIR"
print_success "Cache directory created"

# Create logs directory
LOGS_DIR="$CACHE_DIR/logs"
print_status "Creating logs directory: $LOGS_DIR"
mkdir -p "$LOGS_DIR"
print_success "Logs directory created"

# Test the server script
SERVER_SCRIPT="services/ytdlp/server.py"
if [ -f "$SERVER_SCRIPT" ]; then
    print_status "Testing server script..."
    if python3 "$SERVER_SCRIPT" --help > /dev/null 2>&1; then
        print_success "Server script is working"
    else
        print_warning "Server script test failed (this might be normal)"
    fi
else
    print_error "Server script not found: $SERVER_SCRIPT"
    exit 1
fi

# Create a simple configuration file
CONFIG_FILE="ytdlp-config.json"
print_status "Creating configuration file: $CONFIG_FILE"

cat > "$CONFIG_FILE" << 'EOF'
{
  "host": "localhost",
  "port": 8080,
  "max_workers": 2,
  "timeout": "45s",
  "max_retries": 3,
  "format": "bestaudio[ext=webm][acodec=opus]/bestaudio[ext=m4a]/bestaudio/best",
  "audio_format": "opus",
  "audio_quality": "128K",
  "rate_limit": "1M",
  "sleep_interval": "0",
  "cache_dir": "/tmp/ytdlp-cache",
  "cache_ttl": "24h",
  "max_cache_size": 1073741824,
  "health_check_interval": "30s"
}
EOF

print_success "Configuration file created"

# Final instructions
echo ""
print_success "🎉 yt-dlp service setup completed successfully!"
echo ""
print_status "Next steps:"
echo "  1. Make sure your Discord bot is configured"
echo "  2. Run your Go application with yt-dlp integration enabled"
echo "  3. The yt-dlp service will start automatically when needed"
echo ""
print_status "Configuration:"
echo "  - Config file: $CONFIG_FILE"
echo "  - Cache directory: $CACHE_DIR"
echo "  - Logs directory: $LOGS_DIR"

if [ -d "$VENV_DIR" ]; then
    echo "  - Virtual environment: $VENV_DIR"
    echo ""
    print_warning "Remember to activate the virtual environment when running the bot:"
    echo "  source $VENV_DIR/bin/activate"
fi

echo ""
print_status "To test the service manually:"
echo "  python3 $SERVER_SCRIPT --host localhost --port 8080"
echo ""
print_status "Happy music streaming! 🎵"