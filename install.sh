#!/bin/bash
# OneMCP Quick Install Script

set -e

echo "🚀 Installing OneMCP..."

# Detect platform
case "$(uname -s)" in
    Linux*)     platform="linux";;
    Darwin*)    platform="darwin";;
    CYGWIN*|MINGW*|MSYS*) platform="windows";;
    *)          echo "Unsupported platform"; exit 1;;
esac

# Download appropriate binary
if [ "$platform" = "linux" ]; then
    curl -L -o onemcp https://github.com/mdarshad-ai/OneMCP/releases/latest/download/onemcp_unix
elif [ "$platform" = "darwin" ]; then
    curl -L -o onemcp https://github.com/mdarshad-ai/OneMCP/releases/latest/download/onemcp_darwin
elif [ "$platform" = "windows" ]; then
    curl -L -o onemcp.exe https://github.com/mdarshad-ai/OneMCP/releases/latest/download/onemcp.exe
fi

# Make executable and install
chmod +x onemcp*
sudo mv onemcp* /usr/local/bin/ 2>/dev/null || mv onemcp* ~/bin/

echo "✅ OneMCP installed successfully!"
echo "Run 'onemcp --help' to get started"
