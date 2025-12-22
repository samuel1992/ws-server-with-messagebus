# WebSocket Server with Message Bus

A Go-based WebSocket server demonstrating an **in-memory pub/sub (event bus) pattern** for decoupled message handling between WebSocket clients and backend services.

## Overview

This project showcases how to build a scalable WebSocket architecture where clients and services communicate through a centralized message bus, achieving complete decoupling. Services and clients never reference each other directly—they only interact through topic-based message passing.

## Architecture

```
WebSocket Client ←→ WS Client Handler ←→ Message Bus ←→ Backend Service
```

### Key Components

1. **MessageBus** (`messagebus/`)
   - Interface defining `Subscribe`, `Publish`, and `Unsubscribe` operations
   - `InMemoryMessageBus`: Concurrent-safe implementation using Go channels
   - Topics act as the communication contract between producers and consumers

2. **WebSocket Client** (`ws/client.go`)
   - Bridges WebSocket protocol with the message bus
   - Runs concurrent read/write loops using goroutines
   - Automatically manages subscriptions and cleanup

3. **Service Registry** (`services/registry.go`)
   - Manages service lifecycle with reference counting
   - Creates services on-demand, stops them when no longer needed
   - Enables service reuse across multiple WebSocket connections

4. **Example Services** (`services/`)
   - **PingService**: Echoes messages back to clients
   - **TimeNowService**: Broadcasts current time every 2 seconds

## Message Flow Example

Here's how a message travels through the system for the Ping service:

```
1. Client sends "hello" via WebSocket
   ↓
2. Client readLoop publishes to "ping:from-ws-to-service"
   ↓
3. MessageBus routes to all subscribers on that topic
   ↓
4. PingService receives message
   ↓
5. PingService publishes "hello" to "ping:from-service-to-ws"
   ↓
6. MessageBus routes to subscribed clients
   ↓
7. Client writeLoop receives and sends "hello" back via WebSocket
```

## Key Features

- **Complete Decoupling**: Services and clients interact only through topics, no direct references
- **In-Memory**: Fast message handling using Go channels (buffered, capacity: 256 bytes)
- **Thread-Safe**: Uses `sync.RWMutex` for concurrent access to the message bus
- **Non-Blocking**: Drops messages on full channels to prevent deadlocks
- **Lifecycle Management**: Services start on first client connection, stop when last client disconnects
- **Concurrent**: Goroutines handle WebSocket read/write independently

## Running the Project

### Build and Run

```bash
go run main.go
```

The server starts on `localhost:8080`.

### Test with WebSocket Clients

**Ping Service** (echo messages):
```bash
# Using websocat or similar WebSocket client
websocat ws://localhost:8080/ws/ping
> hello
< hello
```

**TimeNow Service** (receive time updates every 2 seconds):
```bash
websocat ws://localhost:8080/ws/timenow
< 2025-12-22T10:30:00Z
< 2025-12-22T10:30:02Z
< 2025-12-22T10:30:04Z
...
```

## Project Structure

```
.
├── main.go                # Application entry point
├── messagebus/
│   ├── messagebus.go     # MessageBus interface
│   └── inmemory.go       # In-memory implementation
├── ws/
│   └── client.go         # WebSocket client handler
├── services/
│   ├── registry.go       # Service lifecycle manager
│   ├── ping.go           # Ping service implementation
│   └── timenow.go        # TimeNow service implementation
└── handlers/
    ├── handlers.go       # Handler setup
    ├── ping.go           # Ping endpoint handler
    └── timenow.go        # TimeNow endpoint handler
```

## Why This Pattern?

This architecture is ideal for scenarios where:
- Multiple clients need to receive the same events (fan-out)
- Services should be reusable and not tied to specific clients
- You want to add new services without modifying existing code
- In-memory performance is sufficient (no persistence needed)

## Dependencies

- Go 1.25.5+
- [gorilla/websocket](https://github.com/gorilla/websocket)
