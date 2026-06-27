# Algorithms Used in p2pchat

This document describes three core algorithms that power the p2pchat application. Each entry includes the theory behind it, how it is implemented in this project, and a pseudocode sketch written in Go style.

---

## 1. mDNS Peer Discovery with Heartbeat-Based Presence

### Theory

Multicast DNS (mDNS, RFC 6762) is a zero-configuration protocol that lets devices on a local network discover each other **without a central DNS server**. Each node multicasts its own service records to the `224.0.0.251` multicast group on port 5353. Any node that wants to find peers sends a DNS query on the same multicast address; all nodes that match the service type respond with their record. When combined with a periodic **heartbeat** (re-query + last-seen timestamp), the system can also detect when peers disappear from the network.

The algorithm has two interleaved phases:
1. **Announce** — Broadcast own service record so others can find us.
2. **Discover + Expire** — Periodically query for peers, upsert them in the local DB, and mark peers as offline when their `last_seen` timestamp crosses a threshold.

### Implementation Logic (this project)

| Component | File | Role |
|---|---|---|
| `MDNSService.StartServer` | `internal/discovery/mdns_service.go` | Registers the `_p2pchat._tcp.local.` service record with peer ID, username, and gRPC port embedded in TXT records. |
| `MDNSService.StartDiscovery` | same | Runs two tickers: **10 s** discovery sweep and **30 s** presence sweep. |
| `MDNSService.discoverPeers` | same | Issues an mDNS query, collects `ServiceEntry` responses over a channel, and calls `handleDiscoveredPeer` for each. |
| `MDNSService.handleDiscoveredPeer` | same | Parses TXT records, skips self, upserts peer in SQLite, then connects via gRPC and publishes a `peer_discovered` event to the event bus. |
| `MDNSService.updatePeerStatus` | same | Fetches all DB peers, sets `is_online = false` on any peer whose `last_seen` is older than **2 minutes**. |

### Pseudocode

```go
// Phase 1 — Announce
func (m *MDNSService) StartServer() error {
    ip := getLocalNonLoopbackIPv4()
    txtRecords := ["peer_id=" + myPeerID, "username=" + myUsername]
    service := mdns.NewService(myPeerID, "_p2pchat._tcp", "local.", ip, grpcPort, txtRecords)
    m.server = mdns.NewServer(service)
    return nil
}

// Phase 2 — Discover + Expire (runs in a goroutine)
func (m *MDNSService) StartDiscovery(ctx context.Context) {
    discoveryTick := time.NewTicker(10 * time.Second)
    presenceTick  := time.NewTicker(30 * time.Second)

    m.discoverPeers() // initial sweep

    for {
        select {
        case <-ctx.Done():
            return
        case <-discoveryTick.C:
            m.discoverPeers()
        case <-presenceTick.C:
            m.expireOfflinePeers()
        }
    }
}

func (m *MDNSService) discoverPeers() {
    entriesCh := make(chan *mdns.ServiceEntry, 10)
    go func() {
        for entry := range entriesCh {
            peerID, username := parseTXTRecords(entry.InfoFields)
            if peerID == m.myPeerID || peerID == "" {
                continue
            }
            address := formatAddress(entry.AddrV4 or entry.AddrV6, entry.Port)
            peer := db.UpsertPeer(peerID, username, address, isOnline=true, lastSeen=now())
            peerManager.ConnectToPeer(peerID, address, username) // gRPC dial
            eventBus.Publish("peer_discovered", peer)
        }
    }()
    mdns.Query("_p2pchat._tcp", "local.", timeout=5s, out=entriesCh)
    close(entriesCh)
}

func (m *MDNSService) expireOfflinePeers() {
    threshold := now().Add(-2 * time.Minute)
    for _, peer := range db.GetAllPeers() {
        if peer.IsOnline && peer.LastSeen.Before(threshold) {
            peer.IsOnline = false
            db.UpdatePeer(peer)
        }
    }
}
```

---

## 2. Fan-Out Publish/Subscribe Event Bus

### Theory

A **publish/subscribe (pub/sub)** system decouples event producers from consumers. A central **event bus** holds a list of subscriber channels. When a producer publishes an event, the bus **fans it out** to every registered channel. Subscribers process events independently and asynchronously.

The key design decisions in a concurrent pub/sub are:
- **Reader/writer locks** — multiple readers can subscribe/listen simultaneously; only one writer can mutate the subscriber list at a time.
- **Non-blocking send** — if a subscriber's channel is full, the event is dropped (fire-and-forget) to prevent a slow consumer from blocking the entire bus.
- **Buffered channels** — each subscriber gets a buffer (here, `100` events) so brief bursts do not cause immediate drops.

### Implementation Logic (this project)

| Component | File | Role |
|---|---|---|
| `EventBus` | `internal/events/bus.go` | Holds `[]chan Event` protected by `sync.RWMutex`. |
| `Subscribe()` | same | Creates a `chan Event` (cap 100), appends it, returns it to caller. |
| `Publish()` | same | Takes `RLock`, iterates subscribers, does a `select`/`default` non-blocking send. |
| `Unsubscribe()` | same | Takes full `Lock`, finds and removes the channel, closes it. |
| Producers | `chat_service.go`, `mdns_service.go`, `peer/manager.go` | Call `eventBus.Publish()` on message receipt, peer discovery, peer connect. |
| Consumer | `api/handlers.go` (`StreamEvents`) | Subscribes and forwards events over WebSocket to the browser. |

### Pseudocode

```go
type Event struct {
    Type string
    Data any
}

type EventBus struct {
    subscribers []chan Event
    mu          sync.RWMutex
}

func (eb *EventBus) Subscribe() chan Event {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    ch := make(chan Event, 100) // buffered — absorbs bursts
    eb.subscribers = append(eb.subscribers, ch)
    return ch
}

func (eb *EventBus) Unsubscribe(ch chan Event) {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    for i, sub := range eb.subscribers {
        if sub == ch {
            close(ch)
            // Splice out: replace with last element, shrink slice
            eb.subscribers[i] = eb.subscribers[len(eb.subscribers)-1]
            eb.subscribers = eb.subscribers[:len(eb.subscribers)-1]
            break
        }
    }
}

func (eb *EventBus) Publish(event Event) {
    eb.mu.RLock() // multiple publishers can read the list concurrently
    defer eb.mu.RUnlock()
    for _, sub := range eb.subscribers {
        select {
        case sub <- event: // non-blocking fan-out
        default:           // channel full — drop to avoid back-pressure
        }
    }
}
```

---

## 3. WebSocket Real-Time Event Streaming (Server → Browser Push)

### Theory

> **Note:** The project currently uses Server-Sent Events (SSE) for the streaming transport. This document describes the WebSocket version that is being adopted.

A **WebSocket** connection is a full-duplex, persistent TCP channel established via an HTTP `Upgrade` handshake. For server→client push scenarios (real-time notifications, live peer lists, incoming messages) the server holds the connection open and writes framed messages whenever an internal event fires.

The algorithm is a **fan-in select loop**:
1. Upgrade the HTTP request to a WebSocket connection.
2. Subscribe to the internal event bus to receive a channel of events.
3. Start a **keep-alive ticker** (heartbeat ping frames) so the OS TCP stack and any intermediate proxies do not time out the idle connection.
4. Block on a `select` over three channels: `ctx.Done` (client disconnect), the event-bus channel (new event to push), and the keep-alive ticker (write a ping/comment frame).
5. For each event, serialize it to JSON and write it as a WebSocket text frame.
6. On disconnect (or error), unsubscribe from the event bus and return.

### Implementation Logic (this project)

| Component | File | Role |
|---|---|---|
| `StreamEvents` handler | `internal/api/handlers.go` | Currently SSE; will be upgraded to WebSocket. Owns the select loop that bridges the event bus to the browser. |
| `eventBus.Subscribe()` | `internal/events/bus.go` | Returns the channel that the handler ranges over. |
| `eventBus.Unsubscribe()` | same | Called via `defer` so cleanup is guaranteed even on panic/disconnect. |
| Event routing | `handlers.go` lines 202–216 | `message_received`/`message_sent` → forward as-is; `peer_discovered`/`peer_joined` → re-query DB and push full peer list. |

### Pseudocode

```go
func (h *APIHandler) StreamEvents(conn *websocket.Conn, r *http.Request) {
    // 1. Subscribe to the internal event bus
    eventCh := h.eventBus.Subscribe()
    defer h.eventBus.Unsubscribe(eventCh)

    // 2. Send initial "connected" frame
    writeJSON(conn, "connected", map[string]string{"status": "connected"})

    // 3. Keep-alive ticker — prevents proxy/OS timeout on idle connections
    keepAlive := time.NewTicker(30 * time.Second)
    defer keepAlive.Stop()

    ctx := r.Context()

    // 4. Fan-in select loop
    for {
        select {
        case <-ctx.Done():
            // Client disconnected — cleanup happens via defer
            return

        case event, ok := <-eventCh:
            if !ok {
                return // event bus shut down
            }
            // 5. Route event type → JSON frame
            switch event.Type {
            case "message_received", "message_sent":
                if err := writeJSON(conn, event.Type, event.Data); err != nil {
                    return
                }
            case "peer_discovered", "peer_joined":
                peers, err := h.database.GetAllPeers()
                if err != nil {
                    continue
                }
                if err := writeJSON(conn, "peers_update", peers); err != nil {
                    return
                }
            }

        case <-keepAlive.C:
            // 6. Ping frame to keep the connection alive
            if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

// helper — marshals payload and writes a typed WebSocket text frame
func writeJSON(conn *websocket.Conn, eventType string, data any) error {
    frame := struct {
        Type string `json:"type"`
        Data any    `json:"data"`
    }{Type: eventType, Data: data}
    return conn.WriteJSON(frame)
}
```
