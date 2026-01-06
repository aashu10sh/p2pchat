# P2P Chat
> A discord/slack like peer to peer chat application made to chat in a local subnet.
The application will discover new hosts in a subnet using the mDNS protocol and maintin a client collection with them along with getting their info.


┌─────────────────────────────────────────────────────────────┐
│                    Peer A (192.168.1.10)                     │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  HTTP REST API Server (Port 8080)                      │ │
│  │  ├─ GET  /api/profile                                  │ │
│  │  ├─ POST /api/profile                                  │ │
│  │  ├─ GET  /api/peers                                    │ │
│  │  ├─ GET  /api/messages?peer_id=xxx                     │ │
│  │  ├─ POST /api/messages                                 │ │
│  │  ├─ GET  /api/messages/stream (SSE)                    │ │
│  │  └─ Static files (Svelte app at /)                     │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  gRPC Server (Port 50051) - P2P Only                   │ │
│  │  ├─ ReceiveMessage(Message) → MessageAck              │ │
│  │  ├─ Ping(PingRequest) → PingResponse                  │ │
│  │  └─ GetPeerInfo() → PeerInfo                          │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  mDNS Service                                          │ │
│  │  └─ Broadcasts: _p2pchat._tcp.local                   │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Peer Manager                                          │ │
│  │  └─ Maintains gRPC clients to other nodes             │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  SQLite Database                                       │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
         ▲                                    ║
         │ REST API                           ║ gRPC
         │ (JSON over HTTP)                   ║ (Protocol Buffers)
         │                                    ║
    ┌────┴─────┐                         ┌───▼────────────────┐
    │ Browser  │                         │  Peer B            │
    │ Svelte   │                         │  (192.168.1.20)    │
    └──────────┘                         │  gRPC Server :50052│
                                         └────────────────────┘


## 🎯 **Summary**

| Layer | Technology | Purpose |
|-------|------------|---------|
| **Frontend ↔ Backend** | REST API (JSON) | Browser makes HTTP requests |
| **Node ↔ Node** | gRPC (Protobuf) | Direct P2P communication |
| **Discovery** | mDNS | Auto-discover peers on subnet |
| **Real-time Updates** | Server-Sent Events | Push messages to browser |

**Benefits:**
- ✅ Simple frontend (just fetch API)
- ✅ Efficient P2P (gRPC between nodes)
- ✅ Real-time (SSE for message streaming)
- ✅ Type-safe (Protobuf for P2P, TypeScript for frontend)
- ✅ No desktop framework issues

This is the perfect architecture! 🎉