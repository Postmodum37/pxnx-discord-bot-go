# Music System Debugging TODO

## Current Issue Summary
The `/play` command infrastructure is fully implemented and working, but actual audio playback fails with EOF (End of File) errors during streaming. Investigation shows that the issue occurs during the DCA encoding/streaming phase, not in the yt-dlp service or URL extraction.

## Recent Investigation Findings (2025-09-22)

### ✅ Working Components
- yt-dlp service starts successfully
- YouTube URL extraction and metadata retrieval
- Voice channel connections
- DCA encoder initialization
- Command parsing and embed generation

### ❌ Failing Component
- Audio streaming to Discord voice channels
- **Key Error**: `"stream ended unexpectedly - source may be unavailable: EOF"`

## Priority Action Items

### Immediate (P0)
1. **[ ] Stream URL Lifetime Test**
   - Test how long YouTube stream URLs remain valid
   - Add timestamp logging to track URL usage time
   - File: `music/providers/youtube_ytdlp.go`

2. **[ ] Alternative Audio Source Test**
   - Test with direct MP3/OGG URLs (not YouTube)
   - Isolate if issue is YouTube-specific or DCA-general
   - File: `debug_fallback_test.go` (enhance existing test)

3. **[ ] DCA Error Analysis**
   - Add detailed logging to DCA streaming process
   - Capture exact timing of EOF errors
   - File: `music/player/player.go:streamToVoice*` functions

### Short Term (P1)
4. **[ ] YouTube Format Preference Testing**
   - Test mp4a/aac formats instead of opus
   - Modify format selection in yt-dlp provider
   - File: `services/ytdlp/server.py` and `music/providers/youtube_ytdlp.go`

5. **[ ] Stream Validation Enhancement**
   - Add real-time stream health checking during playback
   - Implement stream URL re-extraction on failure
   - File: `music/player/player.go`

6. **[ ] FFmpeg Reconnection Options**
   - Test custom reconnection parameters
   - Override DCA default options
   - File: `music/player/player.go:createEncoder`

### Medium Term (P2)
7. **[ ] Alternative DCA Configuration**
   - Test minimal DCA options
   - Reduce buffer sizes and complexity
   - File: `music/player/player.go:createEncoder`

8. **[ ] Voice Connection Debugging**
   - Add voice connection state monitoring
   - Verify Discord permissions and capabilities
   - File: `music/manager/manager.go`

## Investigation Commands

```bash
# Test specific components
go run debug_fallback_test.go           # Test non-YouTube sources
go test ./music/player -v              # Test audio player
go test ./music/providers -v           # Test yt-dlp provider

# Run with enhanced logging
go run main.go --log-level debug       # Maximum logging detail

# Check yt-dlp service directly
curl http://localhost:8080/health      # Service health check
```

## Log Files to Monitor
- `logs/bot-2025-09-22.log` - Main bot logs
- `services/ytdlp/server.py` stdout - yt-dlp service logs
- DCA internal logs (if available)

## Success Metrics
- [ ] Audio plays successfully in Discord voice channel
- [ ] No EOF errors during streaming
- [ ] Consistent playback across different videos
- [ ] Proper error handling for unavailable videos

## Notes
- The infrastructure is solid - this is primarily a streaming/format compatibility issue
- Focus on the YouTube → DCA → Discord audio pipeline
- Consider that YouTube may be implementing new anti-bot measures