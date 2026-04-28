# Graphiti Go Client

A Go client library for the Graphiti HTTP API.

## Features

- Full coverage of Graphiti HTTP API endpoints
- **7 specialized advanced search methods** for different query patterns:
  - Temporal Window Search - search within time ranges
  - Entity Relationships Search - explore entity connections
  - Diverse Results Search - get non-redundant results using MMR
  - Episode Context Search - search conversation history
  - Successful Tools Search - find frequently mentioned techniques
  - Recent Context Search - get recent information with recency bias
  - Entity By Label Search - filter entities by type/label
- Optional Langfuse observation tracking for monitoring and debugging
- Configurable HTTP client and timeouts
- Type-safe request and response structures

## Installation

```bash
go get github.com/vxcontrol/graphiti-go-client
```

## Quick Start

See the complete working examples:
- **[Base Usage Example](./examples/base-usage-example/main.go)** - Basic operations and common patterns
- **[Advanced Search Example](./examples/advanced-search-example/main.go)** - All 7 advanced search methods

## Usage

### Creating a Client

```go
import (
    "time"
    graphiti "github.com/vxcontrol/graphiti-go-client"
)

// Create a client with default settings
client := graphiti.NewClient("http://localhost:8000")

// Create a client with custom timeout
client := graphiti.NewClient("http://localhost:8000",
    graphiti.WithTimeout(60 * time.Second))

// Create a client with a custom HTTP client
httpClient := &http.Client{
    Timeout: 30 * time.Second,
}
client := graphiti.NewClient("http://localhost:8000",
    graphiti.WithHTTPClient(httpClient))
```

### Langfuse Integration

The client supports optional Langfuse observation tracking for monitoring and debugging. You can attach an `Observation` object to any of the following operations:

- `Search` - Track search queries
- `GetMemory` - Track memory retrieval operations
- `AddMessages` - Track message ingestion
- `AddEntityNode` - Track entity node creation

**⚠️ Important:** The `Observation.ID` and `Observation.TraceID` must be valid UUIDs that correspond to actual observation and trace objects in your Langfuse instance. These IDs are used to link Graphiti operations to Langfuse traces for monitoring and debugging.

Example:

```go
import "github.com/google/uuid"

// Create an observation with UUIDs that exist in Langfuse
// These IDs should come from your Langfuse SDK after creating
// an observation/trace in your Langfuse instance
observation := &graphiti.Observation{
    ID:      "existing-observation-uuid-from-langfuse",
    TraceID: "existing-trace-uuid-from-langfuse",
    Time:    time.Now(),
}

// Use it in any supported operation
result, err := client.Search(graphiti.SearchQuery{
    Query:       "my search query",
    MaxFacts:    10,
    Observation: observation,
})
```

The observation tracking is completely optional - all operations work without it.

### Health Check

```go
health, err := client.HealthCheck()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Status: %s\n", health.Status)
```

### Search for Facts

```go
result, err := client.Search(graphiti.SearchQuery{
    Query:    "Tell me about user preferences",
    MaxFacts: 10,
})
if err != nil {
    log.Fatal(err)
}

for _, fact := range result.Facts {
    fmt.Printf("Fact: %s\n", fact.Fact)
}
```

### Search with Group Filtering and Observation Tracking

```go
groupIDs := []string{"group-123", "group-456"}

// Optional: Link to existing Langfuse observation
// IDs must correspond to actual observation/trace in Langfuse
observation := &graphiti.Observation{
    ID:      "existing-observation-uuid",
    TraceID: "existing-trace-uuid",
    Time:    time.Now(),
}

result, err := client.Search(graphiti.SearchQuery{
    GroupIDs:    &groupIDs,
    Query:       "user settings",
    MaxFacts:    5,
    Observation: observation, // Optional
})
```

### Add Messages

**⚠️ Important:** The `/messages` endpoint is asynchronous. Messages are queued and processed by a background worker. Data may not be immediately available after this call returns.

```go
messages := []graphiti.Message{
    {
        Content:   "Hello, how are you?",
        Author:    "User",
        Timestamp: time.Now(),
    },
    {
        Content:   "I'm doing great, thank you!",
        Author:    "Assistant",
        Timestamp: time.Now(),
    },
}

// Optional: Link to existing Langfuse observation
// IDs must correspond to actual observation/trace in Langfuse
observation := &graphiti.Observation{
    ID:      "existing-observation-uuid",
    TraceID: "existing-trace-uuid",
    Time:    time.Now(),
}

result, err := client.AddMessages(graphiti.AddMessagesRequest{
    GroupID:     "my-group-id",
    Messages:    messages,
    Observation: observation, // Optional
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("%s: %v\n", result.Message, result.Success)

// Wait for processing by polling for episodes
maxAttempts := 10
for attempt := 1; attempt <= maxAttempts; attempt++ {
    episodes, err := client.GetEpisodes("my-group-id", 10)
    if err == nil && len(episodes) > 0 {
        fmt.Println("Messages processed successfully!")
        break
    }
    time.Sleep(5 * time.Second)
}
```

### Add an Entity Node

```go
uuid := "entity-uuid-123"

// Optional: Link to existing Langfuse observation
// IDs must correspond to actual observation/trace in Langfuse
observation := &graphiti.Observation{
    ID:      "existing-observation-uuid",
    TraceID: "existing-trace-uuid",
    Time:    time.Now(),
}

node, err := client.AddEntityNode(graphiti.AddEntityNodeRequest{
    UUID:        uuid,
    GroupID:     "my-group-id",
    Name:        "User Preferences",
    Summary:     "Contains user's preferred settings",
    Observation: observation, // Optional
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created node: %s\n", node.UUID)
```

### Get Memory from Messages

```go
messages := []graphiti.Message{
    {
        Content:   "What were my settings?",
        Author:    "User",
        Timestamp: time.Now(),
    },
}

// Optional: Link to existing Langfuse observation
// IDs must correspond to actual observation/trace in Langfuse
observation := &graphiti.Observation{
    ID:      "existing-observation-uuid",
    TraceID: "existing-trace-uuid",
    Time:    time.Now(),
}

response, err := client.GetMemory(graphiti.GetMemoryRequest{
    GroupID:     "my-group-id",
    MaxFacts:    10,
    Messages:    messages,
    Observation: observation, // Optional
})
if err != nil {
    log.Fatal(err)
}

for _, fact := range response.Facts {
    fmt.Printf("Fact: %s (from %s)\n", fact.Fact, fact.Name)
}
```

### Get Episodes

```go
episodes, err := client.GetEpisodes("my-group-id", 5)
if err != nil {
    log.Fatal(err)
}

for _, episode := range episodes {
    fmt.Printf("Episode: %s - %s\n", episode.Name, episode.Content)
}
```

### Get a Specific Entity Edge

```go
fact, err := client.GetEntityEdge("edge-uuid-123")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Fact: %s\n", fact.Fact)
```

### Advanced Search Methods

The client provides specialized search methods for different use cases:

#### Temporal Window Search

Search for context within a specific time window:

```go
result, err := client.TemporalWindowSearch(graphiti.TemporalSearchRequest{
    Query:      "user activities",
    GroupID:    &groupID,
    TimeStart:  time.Now().Add(-24 * time.Hour),
    TimeEnd:    time.Now(),
    MaxResults: 10,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Found %d edges, %d nodes, %d episodes\n",
    len(result.Edges), len(result.Nodes), len(result.Episodes))
```

#### Entity Relationships Search

Find relationships and related entities from a center node:

```go
result, err := client.EntityRelationshipsSearch(graphiti.EntityRelationshipSearchRequest{
    Query:          "related entities",
    GroupID:        &groupID,
    CenterNodeUUID: "entity-uuid-123",
    MaxDepth:       2,
    NodeLabels:     &[]string{"PERSON", "ORGANIZATION"},
    MaxResults:     20,
})
```

#### Diverse Results Search

Get diverse, non-redundant results using Maximal Marginal Relevance (MMR):

```go
result, err := client.DiverseResultsSearch(graphiti.DiverseSearchRequest{
    Query:          "user interests and hobbies",
    GroupID:        &groupID,
    DiversityLevel: "medium", // "low", "medium", or "high"
    MaxResults:     10,
})
```

#### Episode Context Search

Search through agent responses and conversation context:

```go
result, err := client.EpisodeContextSearch(graphiti.EpisodeContextSearchRequest{
    Query:      "tool execution results",
    GroupID:    &groupID,
    MaxResults: 5,
})
```

#### Successful Tools Search

Find frequently mentioned successful tools or techniques:

```go
result, err := client.SuccessfulToolsSearch(graphiti.SuccessfulToolsSearchRequest{
    Query:       "successful exploits",
    GroupID:     &groupID,
    MinMentions: 2,
    MaxResults:  10,
})
```

#### Recent Context Search

Get most recent relevant context with recency bias:

```go
result, err := client.RecentContextSearch(graphiti.RecentContextSearchRequest{
    Query:         "recent discoveries",
    GroupID:       &groupID,
    RecencyWindow: "6h", // e.g., "1h", "24h", "7d"
    MaxResults:    10,
})
```

#### Entity By Label Search

Search for entities by their labels/types:

```go
result, err := client.EntityByLabelSearch(graphiti.EntityByLabelSearchRequest{
    Query:      "vulnerable services",
    GroupID:    &groupID,
    NodeLabels: []string{"SERVICE", "VULNERABILITY"},
    MaxResults: 20,
})
```

### Delete Operations

```go
// Delete an entity edge
result, err := client.DeleteEntityEdge("edge-uuid-123")

// Delete an episode
result, err := client.DeleteEpisode("episode-uuid-123")

// Delete a group
result, err := client.DeleteGroup("group-id-123")

// Clear all data (use with caution!)
result, err := client.Clear()
```

## Types

### Observation

```go
type Observation struct {
    ID      string    // Observation UUID from Langfuse
    TraceID string    // Trace UUID from Langfuse
    Time    time.Time // Observation timestamp
}
```

**Note:** The `Observation` type is used for integrating with Langfuse for tracking and observability. The `ID` and `TraceID` must be valid UUIDs corresponding to actual observation and trace objects in your Langfuse instance. This field is optional in all requests.

### Message

```go
type Message struct {
    Content           string    // The message content
    UUID              *string   // Optional UUID
    Name              string    // Optional name for episodic node
    Author            string    // The author/entity that created this message
    Timestamp         time.Time // Message timestamp
    SourceDescription string    // Optional source description
}
```

### SearchQuery

```go
type SearchQuery struct {
    GroupIDs    *[]string    // Optional group IDs to filter
    Query       string       // Search query text
    MaxFacts    int          // Maximum number of facts to return (default: 10)
    Observation *Observation // Optional Langfuse observation for tracking
}
```

### FactResult

```go
type FactResult struct {
    UUID      string     // Unique identifier
    Name      string     // Fact name
    Fact      string     // The actual fact text
    ValidAt   *time.Time // When fact became valid
    InvalidAt *time.Time // When fact became invalid
    CreatedAt time.Time  // Creation timestamp
    ExpiredAt *time.Time // Expiration timestamp
}
```

### GetMemoryRequest

```go
type GetMemoryRequest struct {
    GroupID        string       // Group ID
    MaxFacts       int          // Maximum number of facts to return
    CenterNodeUUID *string      // Optional center node UUID
    Messages       []Message    // Messages for context
    Observation    *Observation // Optional Langfuse observation for tracking
}
```

### AddMessagesRequest

```go
type AddMessagesRequest struct {
    GroupID     string       // Group ID
    Messages    []Message    // Messages to add
    Observation *Observation // Optional Langfuse observation for tracking
}
```

### AddEntityNodeRequest

```go
type AddEntityNodeRequest struct {
    UUID        string       // Entity UUID
    GroupID     string       // Group ID
    Name        string       // Entity name
    Summary     string       // Optional entity summary
    Observation *Observation // Optional Langfuse observation for tracking
}
```

### Advanced Search Types

#### NodeResult

```go
type NodeResult struct {
    UUID      string    // Node UUID
    Name      string    // Entity name
    Labels    []string  // Entity type labels (e.g., ["SERVICE", "WEB"])
    Summary   string    // Node summary/description
    CreatedAt time.Time // Creation timestamp
}
```

#### EdgeResult

```go
type EdgeResult struct {
    UUID           string     // Edge UUID
    Name           string     // Relationship name
    Fact           string     // The fact/relationship description
    SourceNodeUUID string     // Source entity UUID
    TargetNodeUUID string     // Target entity UUID
    ValidAt        *time.Time // When relationship became valid
    InvalidAt      *time.Time // When relationship became invalid
    CreatedAt      time.Time  // Creation timestamp
    ExpiredAt      *time.Time // Expiration timestamp
}
```

#### EpisodeResult

```go
type EpisodeResult struct {
    UUID              string    // Episode UUID
    Content           string    // Episode content (agent response, tool output, etc.)
    Source            string    // Source type (e.g., "tool", "agent")
    SourceDescription string    // Detailed source description
    CreatedAt         time.Time // Creation timestamp
    ValidAt           time.Time // When episode occurred
}
```

#### CommunityResult

```go
type CommunityResult struct {
    UUID      string    // Community UUID
    Name      string    // Community name
    Summary   string    // Community summary
    CreatedAt time.Time // Creation timestamp
}
```

## Error Handling

All client methods return an error as the last return value. Always check for errors:

```go
result, err := client.Search(query)
if err != nil {
    // Handle error
    log.Printf("Search failed: %v", err)
    return
}
// Use result
```

## Examples

Two complete working examples are available:

### [Base Usage Example](./examples/base-usage-example/main.go)

Demonstrates basic operations:
- Client initialization and health checks
- Adding messages with Langfuse observation tracking
- Basic search and memory retrieval
- Entity node creation
- Episode management
- Proper handling of asynchronous operations

### [Advanced Search Example](./examples/advanced-search-example/main.go)

Comprehensive demonstration of all 7 advanced search methods:
- **Temporal Window Search** - query within specific time ranges
- **Entity Relationships Search** - explore entity connections and graph traversal
- **Diverse Results Search** - get non-redundant results using MMR
- **Episode Context Search** - search through conversation history and tool outputs
- **Successful Tools Search** - find frequently mentioned successful techniques
- **Recent Context Search** - retrieve recent information with recency bias
- **Entity By Label Search** - filter entities by type/label

This example uses a realistic penetration testing scenario with detailed test data to demonstrate each search method's capabilities.
