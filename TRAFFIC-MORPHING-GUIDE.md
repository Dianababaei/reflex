# Traffic Morphing Configuration Guide

## How It Works

Traffic Morphing makes your proxy traffic look like legitimate protocols to avoid detection. Users choose which protocol to mimic based on their usage pattern.

## Default Behavior

**If user doesn't specify a policy**: HTTP/2 API (most universal)

```go
// In morphing.go
GetProfileByName("")  // Returns HTTP2APIProfile (default)
```

This is best for general web browsing, APIs, and social media.

## User Configuration

In your server config, each user can choose their profile:

```json
{
  "protocol": "reflex",
  "settings": {
    "clients": [
      {
        "id": "uuid-user1",
        "email": "user1@example.com",
        "policy": "youtube"       // <-- User choice
      },
      {
        "id": "uuid-user2",
        "email": "user2@example.com",
        "policy": "zoom"          // <-- User choice
      },
      {
        "id": "uuid-user3",
        "email": "user3@example.com"
        // No policy specified → defaults to HTTP/2 API
      }
    ]
  }
}
```

## Available Profiles

### 1. YouTube (`"youtube"`)
**Best for**: Heavy downloading, video streaming, file transfers

**Pattern**:
- Packet sizes: 1400B (40%), 1200B (30%), 1000B (20%), 800B (10%)
- Delays: 10ms (50%), 20ms (30%), 30ms (20%)

**Why**: Video streaming uses large packets with small variable delays

**Use case**: User mainly watches videos or downloads files

---

### 2. Zoom (`"zoom"`)
**Best for**: Real-time communication, video calls, voice chat

**Pattern**:
- Packet sizes: 500B (30%), 600B (40%), 700B (30%)
- Delays: 30ms (40%), 40ms (40%), 50ms (20%)

**Why**: Video conferencing uses small packets with consistent delays

**Use case**: User primarily uses video calls or voice chat

---

### 3. HTTP/2 API (`"http2-api"` or empty)
**Best for**: Web browsing, APIs, social media (DEFAULT)

**Pattern**:
- Packet sizes: 200B (20%), 500B (30%), 1000B (30%), 1500B (20%)
- Delays: 5ms (30%), 10ms (40%), 15ms (30%)

**Why**: Web traffic is varied with quick responses

**Use case**: General browsing, APIs, mixed usage

---

## How to Choose

| User Type | Recommended | Why |
|-----------|-------------|-----|
| Heavy video watcher | youtube | Large consistent packets |
| Video call user | zoom | Small consistent packets |
| General browser | http2-api | Mixed, varied traffic |
| Don't know usage | http2-api | Works for everyone |

## Implementation

### In Code (encoding/morphing.go)

Get profile from user:
```go
// User's configured policy (from Account.Policy)
userPolicy := account.Policy

// Get the corresponding profile
profile := encoding.GetProfileByName(userPolicy)
// If userPolicy is empty or unknown → defaults to HTTP2APIProfile

// Use when writing frames
frameEncoder.WriteFrameWithMorphing(conn, frame, profile)
```

### How Morphing Works

For each frame:
1. Get random packet size from profile distribution
2. Add random padding to reach that size
3. Get random delay from profile
4. Send frame
5. Wait delay

**Result**: Observer sees traffic pattern matching the mimicked protocol

---

## Examples

### Example 1: Heavy Downloader
Config:
```json
{
  "id": "user-heavy-download",
  "policy": "youtube"
}
```

Traffic looks like: YouTube video streaming
- Large packets (1400B most common)
- Small delays (10-30ms)

### Example 2: Video Call User
Config:
```json
{
  "id": "user-zoom-calls",
  "policy": "zoom"
}
```

Traffic looks like: Zoom video conference
- Small packets (500-700B)
- Consistent delays (30-50ms)

### Example 3: General User
Config:
```json
{
  "id": "user-general",
  "policy": ""
}
```

Traffic looks like: HTTP/2 API requests (default)
- Mixed packets (200-1500B)
- Variable delays (5-15ms)

---

## How to Verify It Works

Check that frames are morphed correctly:

```bash
# Run tests
go test ./proxy/reflex/encoding/ -v

# Look for:
# ✅ TestMorphingIntegration - Morphing applied
# ✅ TestPacketSize - Sizes match distribution
# ✅ TestDelay - Delays applied
```

---

## No Policy Specified?

If user doesn't have a policy field in config:

```json
{
  "id": "user-no-policy",
  "email": "user@example.com"
  // policy field missing
}
```

**Behavior**:
```go
GetProfileByName("")  // Returns HTTP2APIProfile automatically
```

User gets default HTTP/2 API profile ✅

---

## Common Questions

**Q: Can I change profile mid-session?**
A: Current implementation uses one profile per user. Changing requires reconnection.

**Q: What if I pick wrong profile?**
A: Slightly lower obfuscation, but still better than no morphing. Can reconfigure and reconnect.

**Q: Is API safe if morphing disabled?**
A: No morphing = uniform packet sizes = easy to detect. Always enable.

**Q: Can profiles overlap?**
A: Yes, but distributions differ. Can't perfectly mimic multiple profiles simultaneously.

---

## Configuration Example (Full)

```json
{
  "inbounds": [{
    "port": 8443,
    "protocol": "reflex",
    "settings": {
      "clients": [
        {
          "id": "b831381d-6324-4d53-ad4f-8cda48b30811",
          "email": "heavy-user@example.com",
          "policy": "youtube"
        },
        {
          "id": "c942492e-7435-5e64-be5g-9deb59b41922",
          "email": "call-user@example.com",
          "policy": "zoom"
        },
        {
          "id": "d053503f-8546-6f75-cf6h-0efc60c52033",
          "email": "general-user@example.com",
          "policy": "http2-api"
        },
        {
          "id": "e164614g-9657-7g86-dg7i-1fgd71d63144",
          "email": "default-user@example.com"
          // No policy → auto defaults to http2-api
        }
      ]
    }
  }],
  "outbounds": [{
    "protocol": "freedom"
  }]
}
```

---

**Summary**: Users choose their profile based on usage, or get HTTP/2 API default. Morphing activates automatically based on their choice.
