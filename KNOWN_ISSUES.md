# Known Issues and Troubleshooting

## 🎵 Music System Issues

### Current Status (2025-09-22)

#### ✅ RESOLVED Issues
- **Interaction timeouts**: Fixed with deferred responses
- **yt-dlp environment**: Fixed with proper environment variable handling
- **EOF streaming errors**: Partially resolved with enhanced error recovery and stream validation
- **Logging noise**: Reduced to essential logs with file-based output
- **DCA encoder compatibility**: Enhanced with better FFmpeg options and retry logic

#### ⚠️ CURRENT Issues

##### **Music Playback Intermittent Failures**
**Status**: **Active Issue** - Latest investigation (2025-09-22)
**Description**: Music system shows inconsistent behavior with frequent EOF (End of File) errors during streaming, causing playback to fail or end prematurely.

**Evidence from Recent Logs (2025-09-22 13:02:18)**:
```
[INFO] Starting playback: Joji - SLOW DANCING IN THE DARK (Guild: 897867284238991421)
[INFO] Playback started: Joji - SLOW DANCING IN THE DARK
[INFO] Playback started: Joji - SLOW DANCING IN THE DARK
[ERROR] Streaming error for Joji - SLOW DANCING IN THE DARK: stream ended unexpectedly - source may be unavailable: EOF
[INFO] Playback finished: Joji - SLOW DANCING IN THE DARK
```

**Analysis - Updated Investigation Findings**:
- ✅ yt-dlp service working (successfully starts and provides stream URLs)
- ✅ Voice connection established successfully (bot joins voice channels)
- ✅ DCA encoder creates without initial errors
- ❌ **NEW FINDING**: EOF errors occurring during actual streaming phase
- ❌ **CONFIRMED**: Streaming fails with "stream ended unexpectedly" errors

**Updated Potential Root Causes (Based on Investigation)**:
1. **YouTube Stream URL Expiration**: YouTube stream URLs may be short-lived and expire during playback
2. **Network/HTTP Issues**: Stream connections being closed by YouTube's CDN during playback
3. **DCA Encoder Stream Handling**: EOF errors suggest the input stream is terminating unexpectedly
4. **YouTube Rate Limiting**: Possible throttling of stream requests from the yt-dlp service
5. **FFmpeg Stream Reconnection**: Built-in reconnection options may not be handling YouTube's stream behavior
6. **Audio Format/Container Issues**: Opus/WebM formats from YouTube may have compatibility issues with DCA

**Priority Investigation Steps**:
1. **Stream URL Lifetime Testing**: Test how long YouTube stream URLs remain valid after extraction
2. **Alternative Audio Sources**: Test with direct MP3/OGG URLs to isolate YouTube-specific issues
3. **DCA Reconnection Enhancement**: Implement custom reconnection logic for YouTube streams
4. **Error Pattern Analysis**: Log and analyze specific EOF error timing patterns
5. **yt-dlp Format Selection**: Test with different audio format preferences (mp4a, aac vs opus)
6. **Stream Validation**: Add real-time stream health checking during playback

### 🔧 Implemented Fixes

#### DCA Encoder Enhancements
- **Buffer size increased**: 200 frames (from 100) for stability
- **Built-in reconnect options**: FFmpeg automatic reconnection enabled
- **Enhanced audio filter**: Format normalization and resampling
- **Retry logic**: 3 attempts with exponential backoff
- **Stream validation**: URL accessibility checks before encoding

#### Voice Connection Improvements
- **Health monitoring**: Periodic connection validation during streaming
- **Retry logic**: Up to 3 attempts for voice operations
- **State validation**: Guild/Channel ID verification
- **OpusSend responsiveness**: Channel availability testing

#### Error Recovery System
- **Circuit breaker pattern**: Robust → Basic streaming fallback
- **Frame-by-frame processing**: Individual frame error handling
- **EOF recovery**: Stream health validation on EOF errors
- **Consecutive error tracking**: Maximum 5 errors before abort

#### Logging Improvements
- **File-based logging**: All logs written to `logs/bot-YYYY-MM-DD.log`
- **Log level control**: Command line `--log-level` parameter
- **Essential logs only**: Reduced noise, kept important information
- **Structured logging**: Clear error classification and debugging info

## 🚀 Working Components

### Infrastructure
- ✅ **Bot connection**: Successfully connects to Discord
- ✅ **Command registration**: Slash commands register properly
- ✅ **yt-dlp service**: Starts automatically and provides stream URLs
- ✅ **Voice connections**: Can join/leave voice channels
- ✅ **Error handling**: Comprehensive error recovery and user feedback

### Services
- ✅ **YouTube integration**: Search and URL parsing working
- ✅ **Metadata extraction**: Title, duration, thumbnail retrieval
- ✅ **Service management**: Auto-start/stop with health monitoring
- ✅ **HTTP client**: Robust request handling with retry logic

## 📋 TODO: Next Steps

1. **Audio Pipeline Investigation**
   - Test with minimal audio configuration
   - Verify actual audio data reaches Discord
   - Debug DCA encoder output format

2. **Permission Verification**
   - Check bot permissions in voice channels
   - Verify Discord application settings
   - Test with different Discord servers

3. **Alternative Approaches**
   - Test direct FFmpeg → Discord streaming
   - Evaluate alternative audio libraries
   - Consider DCA alternatives if needed

4. **Debug Tooling**
   - Add voice data transmission logging
   - Implement audio pipeline state monitoring
   - Create diagnostic commands for troubleshooting

## 🛠️ Development Environment

### Requirements Met
- ✅ **Go 1.25.1**: Language runtime
- ✅ **discordgo v0.29.0**: Latest Discord library
- ✅ **yt-dlp**: Audio extraction service
- ✅ **FFmpeg**: Audio processing (via DCA)
- ✅ **Python 3**: yt-dlp service runtime

### Configuration
- ✅ **Environment variables**: Properly loaded from .env
- ✅ **Service integration**: yt-dlp HTTP service operational
- ✅ **Logging system**: File-based with level control
- ✅ **Error recovery**: Comprehensive retry and fallback logic