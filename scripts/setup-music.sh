#!/bin/bash

# Discord Bot Music Setup Script
# Sets up Python dependencies and yt-dlp for music functionality

set -e  # Exit on any error

echo "🎵 Setting up Discord Bot Music Dependencies..."
echo

# Check if Python3 is installed
if ! command -v python3 &> /dev/null; then
    echo "❌ Python3 is required but not installed."
    echo "Please install Python3 and try again."
    exit 1
fi

echo "✅ Python3 found: $(python3 --version)"

# Check if pip is installed
if ! command -v pip3 &> /dev/null; then
    echo "❌ pip3 is required but not installed."
    echo "Please install pip3 and try again."
    exit 1
fi

echo "✅ pip3 found: $(pip3 --version)"
echo

# Install Python dependencies
echo "📦 Installing Python dependencies for yt-dlp service..."
if [ -f "services/ytdlp/requirements.txt" ]; then
    pip3 install -r services/ytdlp/requirements.txt
    echo "✅ Python dependencies installed successfully!"
else
    echo "❌ requirements.txt not found at services/ytdlp/requirements.txt"
    exit 1
fi

echo

# Verify yt-dlp installation
echo "🔍 Verifying yt-dlp installation..."
if python3 -c "import yt_dlp; print('yt-dlp version:', yt_dlp.version.__version__)" 2>/dev/null; then
    echo "✅ yt-dlp is working correctly!"
else
    echo "❌ yt-dlp verification failed!"
    exit 1
fi

echo

# Test yt-dlp service functionality
echo "🧪 Testing yt-dlp service functionality..."
python3 -c "
import asyncio
import sys
sys.path.append('services/ytdlp')
from server import YTDLPService

async def test():
    service = YTDLPService()
    try:
        # Test that the service can be created and basic functionality works
        print('✅ YTDLPService can be instantiated')
        return True
    except Exception as e:
        print(f'❌ YTDLPService test failed: {e}')
        return False

result = asyncio.run(test())
sys.exit(0 if result else 1)
"

if [ $? -eq 0 ]; then
    echo "✅ yt-dlp service test passed!"
else
    echo "❌ yt-dlp service test failed!"
    exit 1
fi

echo

# Create cache directories
echo "📁 Creating cache directories..."
mkdir -p /tmp/ytdlp-cache/logs
echo "✅ Cache directories created!"

echo

# Check if FFmpeg is available (optional but recommended)
if command -v ffmpeg &> /dev/null; then
    echo "✅ FFmpeg found: $(ffmpeg -version | head -n1)"
    echo "   This will enable high-quality audio processing."
else
    echo "⚠️  FFmpeg not found (optional but recommended for best audio quality)"
    echo "   Install FFmpeg for enhanced audio processing capabilities."
fi

echo
echo "🎉 Music system setup complete!"
echo
echo "Next steps:"
echo "1. Configure your .env file with Discord and OpenWeather tokens"
echo "2. Register bot commands: go run main.go --register-commands"
echo "3. Start the bot: go run main.go"
echo
echo "Available music commands:"
echo "- /join  : Connect bot to your voice channel"
echo "- /play  : Search and play music from YouTube"
echo "- /leave : Disconnect bot from voice channel"
echo
echo "🎵 Ready to rock! 🤖"