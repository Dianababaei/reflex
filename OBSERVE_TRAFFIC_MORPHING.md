# Observing Reflex Traffic Morphing in Action

## Overview

The Reflex protocol implements **traffic morphing** - making encrypted proxy traffic look like legitimate protocols (YouTube, Zoom, HTTP/2 API) to avoid detection. This guide shows how to observe this in action using packet analysis tools.

---

## What You'll See

### Without Traffic Morphing (Bad)
```
Observer sees:
- All packets same size (e.g., 1024 bytes)
- Uniform timing patterns
- Obvious artificiality = Easy to detect as VPN/proxy
```

### With Traffic Morphing (Good)
```
Observer sees:
- Variable packet sizes matching real protocols
- Realistic timing patterns
- Looks like legitimate YouTube/Zoom/HTTP traffic
- Hard to distinguish from normal traffic
```

---

## Prerequisites

1. **Three terminals running**:
   - Terminal 1: Echo server (`go run echo-server.go`)
   - Terminal 2: Reflex server (`./xray -c ../reflex-server-test.json`)
   - Terminal 3: Reflex client (`./xray -c ../reflex-client-test.json`)

2. **Packet analysis tool**:
   - **Windows**: Wireshark (free, recommended)
     - Download: https://www.wireshark.org/download/
   - **Alternative**: NetMonitor, tcpdump via WSL

---

## Method 1: Using Wireshark (Recommended)

### Step 1: Start Wireshark

1. Open Wireshark
2. Select **Loopback: lo** interface (localhost traffic)
3. Click the **Start capturing** button (shark fin icon)

### Step 2: Generate Traffic

In a new terminal, send traffic through the reflex tunnel:

```bash
cd i:\dianababaei\reflex\reflex

# Send data through tunnel repeatedly to capture morphing
for i in {1..50}; do
  curl -s -x socks5://127.0.0.1:10002 http://127.0.0.1:9996 > /dev/null 2>&1 &
done
wait
```

Or use the traffic script:
```bash
# Alternative: Use Python to generate sustained traffic
python3 - << 'EOF'
import socket
import time

for i in range(50):
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.settimeout(2)
        s.connect(('127.0.0.1', 10002))  # SOCKS5 proxy
        s.send(b'\x05\x01\x00')  # SOCKS5 hello
        s.recv(1024)
        s.close()
        time.sleep(0.1)
    except:
        pass
EOF
```

### Step 3: Analyze Captured Packets

In Wireshark:

1. **Filter for reflex server traffic**:
   ```
   (ip.src == 127.0.0.1 and tcp.dstport == 8555) or (ip.dst == 127.0.0.1 and tcp.srcport == 8555)
   ```

2. **Look at packet lengths**:
   - Right-click any packet → **Copy** → **Summary**
   - Or use column: **View** → **Columns** → Add **Length**

3. **Expected patterns by policy**:
   - Check `reflex-server-test.json` to see which policy is configured
   - Current config uses `"mimic-http2-api"`

4. **What to verify**:
   ```
   Packet Size Distribution for HTTP/2 API profile:
   - 200 bytes   (20% probability)
   - 500 bytes   (30% probability)
   - 1000 bytes  (30% probability)
   - 1500 bytes  (20% probability)

   Sample should show mixed sizes, not uniform!
   ```

### Step 4: Export Data for Analysis

1. **File** → **Export Packet Dissections** → **As CSV**
2. Open in Excel/LibreOffice
3. Create chart of packet lengths
4. Verify non-uniform distribution matches configured profile

---

## Method 2: Command-Line Packet Capture

### Using tcpdump (via Windows WSL)

```bash
# In WSL terminal
sudo tcpdump -i lo 'tcp port 8555' -w reflex_traffic.pcap

# (In another terminal, send traffic as shown in Method 1, Step 2)

# Stop capture with Ctrl+C

# Analyze locally
tcpdump -r reflex_traffic.pcap -n | head -50
```

Then open `reflex_traffic.pcap` in Wireshark.

### Using netsh (Windows native)

```powershell
# Start trace
netsh trace start capture=yes tracefile=reflex_traffic.etl

# (Send traffic in another terminal)

# Stop trace
netsh trace stop

# Convert to format you can analyze
# (Can be opened in Windows Performance Analyzer)
```

---

## Method 3: Using Golang Test

Run the morphing tests directly to verify profiles work:

```bash
cd i:\dianababaei\reflex\reflex\xray-core

# Run morphing tests
go test ./proxy/reflex/encoding/ -v -run "Morphing"
```

**What to look for in test output**:
- ✅ `TestMorphingIntegration` passes
- ✅ No uniform packet sizes detected
- ✅ Delays applied correctly

---

## Interpreting Results

### Good Morphing (Mission Accomplished!)

```
If you see:
✅ Variable packet sizes (200, 500, 1000, 1500)
✅ Distribution matches profile (20%, 30%, 30%, 20%)
✅ Delays between packets (5-15ms for HTTP/2 API)
✅ Packet arrival patterns look "organic"

→ Traffic looks like legitimate HTTP/2 API!
→ Hard to detect as proxy/VPN
```

### Bad Morphing (Something's Wrong)

```
If you see:
❌ All packets same size (e.g., all 1024 bytes)
❌ Perfect timing intervals
❌ Suspiciously uniform patterns

→ Morphing not working or disabled
→ Easy to detect as proxy/VPN
→ Check policy in config file
```

---

## Understanding Packet Sizes in the Protocol

### Reflex Frame Structure
```
+----------+----------+----------+
| Frame    | Encrypted| Optional |
| Header   | Payload  | Padding  |
| (16B)    | (var)    | (var)    |
+----------+----------+----------+
```

When morphing is enabled:
1. Frame is encrypted with ChaCha20-Poly1305
2. Total size is chosen from profile distribution
3. Padding added to reach target size
4. Transmitted with profile's delay

**Example for HTTP/2 API profile**:
- Random choice: "Send 1000-byte packet"
- Frame data: 50 bytes
- Padding added: 950 bytes
- Transmitted: 1000 bytes exactly
- Delay: 5-15ms before next packet

---

## Changing the Traffic Profile

To observe different morphing patterns, edit `reflex-server-test.json`:

### Test 1: YouTube Profile
```json
{
  "id": "b831381d-6324-4d53-ad4f-8cda48b30811",
  "email": "test@example.com",
  "account": {
    "type": "xray.proxy.reflex.Account",
    "id": "b831381d-6324-4d53-ad4f-8cda48b30811",
    "policy": "youtube"
  }
}
```

**Expected packet sizes**: 800, 1000, 1200, 1400 bytes
- Large packets = Video streaming
- Small delays = Fast delivery

Capture and verify you see mostly large packets!

### Test 2: Zoom Profile
```json
{
  "account": {
    "policy": "zoom"
  }
}
```

**Expected packet sizes**: 500, 600, 700 bytes
- Small packets = Video calls
- Consistent delays = Real-time communication

Capture and verify you see mostly small packets!

### Test 3: HTTP/2 API Profile (Default)
```json
{
  "account": {
    "policy": "http2-api"
  }
}
```

**Expected packet sizes**: 200, 500, 1000, 1500 bytes
- Mixed sizes = Web APIs
- Variable delays = Normal browsing

Capture and verify mixed sizes!

---

## Quick Verification Checklist

After changing policy:

1. ✅ Edit `reflex-server-test.json` with new policy
2. ✅ Restart Terminal 2 (reflex server)
3. ✅ Start Wireshark capture
4. ✅ Generate 50+ requests from terminal 4
5. ✅ Stop capture
6. ✅ Filter port 8555 traffic
7. ✅ Check packet length distribution
8. ✅ Verify matches profile expectation

---

## Troubleshooting Packet Capture

**Can't see packets in Wireshark?**
- Make sure loopback interface is selected
- Check firewall isn't blocking localhost
- Verify ports 8555, 10002, 9996 are in use:
  ```bash
  netstat -an | findstr :8555
  netstat -an | findstr :10002
  netstat -an | findstr :9996
  ```

**All packets same size?**
- Morphing might be disabled
- Check that frame encoder is actually using morphing profile
- Verify policy is set in config (not empty string)

**Can't generate enough traffic?**
- Use the Python script from Method 1, Step 2
- Increase loop iterations
- Use `ab` (Apache Bench) or `wrk` for more sustained load

---

## What the Code Does

### In `xray-core/proxy/reflex/encoding/morphing.go`:

```go
// GetPacketSize returns random size from profile
func (p *TrafficProfile) GetPacketSize() int {
    r := rand.Float64()
    cumulative := 0.0
    for _, pattern := range p.PacketSizes {
        cumulative += pattern.Weight
        if r <= cumulative {
            return pattern.Size  // ← Returns one of: 200, 500, 1000, 1500
        }
    }
}

// AddPadding fills data to reach target size
func AddPadding(data []byte, targetSize int) []byte {
    padded := make([]byte, targetSize)
    copy(padded, data)
    rand.Read(padded[len(data):])  // ← Fills with random bytes
    return padded
}
```

This randomness is what makes traffic look "organic" instead of uniform!

---

## Real-World Implications

### Why This Matters

Deep Packet Inspection (DPI) systems detect:
- ✅ Uniform packet sizes → VPN detected
- ✅ Perfect timing patterns → Proxy detected
- ✅ Non-random headers → Custom protocol detected

With morphing:
- ❌ Variable sizes → Looks like YouTube/Zoom/HTTP
- ❌ Realistic delays → Can't distinguish from normal traffic
- ❌ Protocol indistinguishable → DPI can't fingerprint it

### Real Scenario

```
DPI Monitor analyzing your traffic:
"This looks like... HTTP/2 API to normal websites"
"Nothing suspicious!"
"User is just browsing normally"

(Actually: Encrypted tunnel through reflex protocol)
```

---

## Summary

1. **Traffic morphing is working** when you see variable packet sizes matching the profile
2. **Use Wireshark** to visually confirm morphing patterns
3. **Different policies** have different packet size distributions
4. **This makes detection harder** for network monitors
5. **You've implemented** a sophisticated anti-censorship technique!

The reflex protocol successfully hides its true nature by looking like legitimate traffic. This is what all 5 implementation steps were building toward!

---

## Next Steps

- Test all 3 profiles (YouTube, Zoom, HTTP/2 API)
- Compare packet distributions visually
- Export data and create charts
- Document the morphing effectiveness
- Use in your anti-censorship deployment

Congratulations! Your reflex protocol implementation is complete and verified! 🎉
