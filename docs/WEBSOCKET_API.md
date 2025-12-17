# Kite WebSocket API Documentation

Real-time WebSocket API for receiving live updates from the Kite platform.

## Connection

Connect to the WebSocket endpoint:

```
ws://localhost:8080/ws
wss://api.kite.example.com/ws (production)
```

### JavaScript Example

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('Connected to Kite WebSocket');

  // Subscribe to scraping events
  ws.send(JSON.stringify({
    type: 'subscribe',
    room: 'scrape:all'
  }));
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Disconnected from Kite WebSocket');
};
```

### Python Example

```python
import websocket
import json

def on_message(ws, message):
    data = json.loads(message)
    print(f"Received: {data}")

def on_open(ws):
    print("Connected to Kite WebSocket")

    # Subscribe to case events
    ws.send(json.dumps({
        'type': 'subscribe',
        'room': 'cases:all'
    }))

ws = websocket.WebSocketApp(
    "ws://localhost:8080/ws",
    on_message=on_message,
    on_open=on_open
)

ws.run_forever()
```

## Message Format

All messages follow this structure:

```json
{
  "type": "message.type",
  "data": {
    "key": "value"
  },
  "timestamp": "2023-12-16T10:00:00Z"
}
```

## Message Types

### Connection Messages

#### welcome
Sent immediately after connection.

```json
{
  "type": "welcome",
  "data": {
    "client_id": "uuid-client-id",
    "message": "Connected to Kite WebSocket API"
  },
  "timestamp": "2023-12-16T10:00:00Z"
}
```

#### ping / pong
Heartbeat messages to keep connection alive.

```json
{
  "type": "ping",
  "data": {
    "client_id": "uuid-client-id"
  },
  "timestamp": "2023-12-16T10:00:00Z"
}
```

#### subscribe
Subscribe to a room to receive specific events.

**Client sends:**
```json
{
  "type": "subscribe",
  "room": "scrape:all"
}
```

**Server responds:**
```json
{
  "type": "subscribe",
  "data": {
    "client_id": "uuid-client-id",
    "room": "scrape:all",
    "status": "subscribed"
  },
  "timestamp": "2023-12-16T10:00:00Z"
}
```

#### unsubscribe
Unsubscribe from a room.

**Client sends:**
```json
{
  "type": "unsubscribe",
  "room": "scrape:all"
}
```

## Rooms

Subscribe to specific rooms to receive targeted events:

### Scraping Rooms

- `scrape:all` - All scraping events
- `scrape:{job_id}` - Specific scraping job events

### Case Rooms

- `cases:all` - All case events (created, updated, deleted)
- `cases:{jurisdiction}` - Cases from specific jurisdiction
- `case:{case_id}` - Events for a specific case

### Search Rooms

- `search:{search_id}` - Results from specific search

### Worker Rooms

- `workers:all` - Worker pool status updates
- `queue:all` - Job queue updates

### Alert Rooms

- `alerts:all` - All alerts
- `alerts:system` - System alerts
- `alerts:quality` - Data quality alerts

### Metric Rooms

- `metrics:all` - All metric updates
- `metrics:{metric_name}` - Specific metric updates

## Event Types

### Scraping Events

#### scrape.started

```json
{
  "type": "scrape.started",
  "data": {
    "job_id": "job_abc123",
    "jurisdiction": "Australia",
    "status": "started"
  },
  "timestamp": "2023-12-16T10:00:00Z"
}
```

#### scrape.progress

```json
{
  "type": "scrape.progress",
  "data": {
    "job_id": "job_abc123",
    "current": 45,
    "total": 100,
    "percent": 45.0,
    "message": "Scraping page 45 of 100"
  },
  "timestamp": "2023-12-16T10:01:30Z"
}
```

#### scrape.complete

```json
{
  "type": "scrape.complete",
  "data": {
    "job_id": "job_abc123",
    "cases_scraped": 100,
    "duration_seconds": 145.3,
    "status": "completed"
  },
  "timestamp": "2023-12-16T10:05:00Z"
}
```

#### scrape.error

```json
{
  "type": "scrape.error",
  "data": {
    "job_id": "job_abc123",
    "error": "Failed to fetch page: timeout",
    "status": "failed"
  },
  "timestamp": "2023-12-16T10:02:00Z"
}
```

### Search Events

#### search.results

```json
{
  "type": "search.results",
  "data": {
    "search_id": "search_xyz789",
    "results": [
      {
        "id": "cth/HCA/2023/15",
        "case_name": "Smith v Jones",
        "case_number": "[2023] HCA 15",
        "court": "High Court of Australia",
        "jurisdiction": "Australia"
      }
    ],
    "result_count": 1,
    "total": 42
  },
  "timestamp": "2023-12-16T10:00:00Z"
}
```

### Case Events

#### case.created

```json
{
  "type": "case.created",
  "data": {
    "case_id": "cth/HCA/2023/15",
    "case_name": "Smith v Jones",
    "case_number": "[2023] HCA 15",
    "court": "High Court of Australia",
    "jurisdiction": "Australia"
  },
  "timestamp": "2023-12-16T10:00:00Z"
}
```

#### case.updated

```json
{
  "type": "case.updated",
  "data": {
    "case_id": "cth/HCA/2023/15",
    "case_name": "Smith v Jones (Updated)",
    "case_number": "[2023] HCA 15"
  },
  "timestamp": "2023-12-16T10:05:00Z"
}
```

#### case.deleted

```json
{
  "type": "case.deleted",
  "data": {
    "case_id": "cth/HCA/2023/15"
  },
  "timestamp": "2023-12-16T10:10:00Z"
}
```

### Validation Events

#### validation.complete

```json
{
  "type": "validation.complete",
  "data": {
    "case_id": "cth/HCA/2023/15",
    "valid": true,
    "quality_score": 0.92
  },
  "timestamp": "2023-12-16T10:00:00Z"
}
```

#### quality.alert

```json
{
  "type": "quality.alert",
  "data": {
    "severity": "warning",
    "message": "High duplicate rate detected",
    "details": {
      "duplicate_rate": 0.25,
      "threshold": 0.20
    }
  },
  "timestamp": "2023-12-16T10:00:00Z"
}
```

### Worker Events

#### worker.status

```json
{
  "type": "worker.status",
  "data": {
    "active_workers": 3,
    "total_workers": 4,
    "queue_size": 152,
    "utilization": 0.75
  },
  "timestamp": "2023-12-16T10:00:00Z"
}
```

#### queue.update

```json
{
  "type": "queue.update",
  "data": {
    "pending": 98,
    "running": 12,
    "completed": 1453,
    "failed": 31,
    "total": 110
  },
  "timestamp": "2023-12-16T10:00:00Z"
}
```

### System Events

#### system.alert

```json
{
  "type": "system.alert",
  "data": {
    "severity": "critical",
    "component": "database",
    "message": "High connection usage detected"
  },
  "timestamp": "2023-12-16T10:00:00Z"
}
```

#### metric.update

```json
{
  "type": "metric.update",
  "data": {
    "metric": "http_requests_total",
    "value": 1523,
    "labels": {
      "method": "GET",
      "status": "200"
    }
  },
  "timestamp": "2023-12-16T10:00:00Z"
}
```

## Complete Example: Scraping Job Monitor

```javascript
class ScrapeMonitor {
  constructor(jobId) {
    this.jobId = jobId;
    this.ws = new WebSocket('ws://localhost:8080/ws');

    this.ws.onopen = () => this.onConnect();
    this.ws.onmessage = (event) => this.onMessage(event);
    this.ws.onerror = (error) => this.onError(error);
    this.ws.onclose = () => this.onClose();
  }

  onConnect() {
    console.log('Connected to Kite WebSocket');

    // Subscribe to this specific job
    this.subscribe(`scrape:${this.jobId}`);
  }

  subscribe(room) {
    this.ws.send(JSON.stringify({
      type: 'subscribe',
      room: room
    }));
  }

  onMessage(event) {
    const message = JSON.parse(event.data);

    switch (message.type) {
      case 'scrape.started':
        console.log('Scraping started:', message.data);
        break;

      case 'scrape.progress':
        const { current, total, percent } = message.data;
        console.log(`Progress: ${current}/${total} (${percent.toFixed(1)}%)`);
        this.updateProgressBar(percent);
        break;

      case 'scrape.complete':
        console.log('Scraping completed:', message.data);
        this.onComplete(message.data);
        break;

      case 'scrape.error':
        console.error('Scraping error:', message.data);
        this.onError(message.data);
        break;
    }
  }

  updateProgressBar(percent) {
    // Update UI progress bar
    document.getElementById('progress').style.width = `${percent}%`;
  }

  onComplete(data) {
    console.log(`Scraped ${data.cases_scraped} cases in ${data.duration_seconds}s`);
    this.ws.close();
  }

  onError(error) {
    console.error('Error:', error);
  }

  onClose() {
    console.log('Disconnected from Kite WebSocket');
  }
}

// Usage
const monitor = new ScrapeMonitor('job_abc123');
```

## Best Practices

1. **Handle Reconnection**: Implement automatic reconnection with exponential backoff
2. **Subscribe Selectively**: Only subscribe to rooms you need
3. **Heartbeat**: Respond to ping messages to keep connection alive
4. **Error Handling**: Always handle connection errors gracefully
5. **Message Validation**: Validate incoming messages before processing
6. **Unsubscribe**: Clean up subscriptions when no longer needed
7. **Rate Limiting**: Be aware of message rate limits

## Connection Limits

- Maximum connections per client: 10
- Maximum subscriptions per connection: 50
- Message rate limit: 100 messages/second
- Maximum message size: 512 KB

## Error Handling

Connection errors will close the WebSocket connection. Common errors:

- **Upgrade Required (426)**: Not a WebSocket upgrade request
- **Too Many Connections (429)**: Client has too many open connections
- **Unauthorized (401)**: Authentication required (if auth is enabled)

## Support

- Documentation: https://docs.kite.example.com/websocket
- Examples: https://github.com/gongahkia/kite/tree/main/examples/websocket
- Issues: https://github.com/gongahkia/kite/issues
