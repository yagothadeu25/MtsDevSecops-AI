package graphiti

import "time"

// Observation represents Langfuse observation object to link
type Observation struct {
	ID      string    `json:"id"`
	TraceID string    `json:"trace_id"`
	Time    time.Time `json:"time"`
}

// Message represents a message in the system
type Message struct {
	Content           string    `json:"content"`
	UUID              *string   `json:"uuid,omitempty"`
	Name              string    `json:"name,omitempty"`
	Author            string    `json:"author"`
	Timestamp         time.Time `json:"timestamp"`
	SourceDescription string    `json:"source_description,omitempty"`
}

// Result represents a generic result response
type Result struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// HealthCheckResponse represents the health check response
type HealthCheckResponse struct {
	Status string `json:"status"`
}

// SearchQuery represents a search query request
type SearchQuery struct {
	GroupIDs    *[]string    `json:"group_ids,omitempty"`
	Query       string       `json:"query"`
	MaxFacts    int          `json:"max_facts,omitempty"`
	Observation *Observation `json:"observation,omitempty"`
}

// FactResult represents a fact result from the graph
type FactResult struct {
	UUID      string     `json:"uuid"`
	Name      string     `json:"name"`
	Fact      string     `json:"fact"`
	ValidAt   *time.Time `json:"valid_at,omitempty"`
	InvalidAt *time.Time `json:"invalid_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiredAt *time.Time `json:"expired_at,omitempty"`
}

// SearchResults represents the results of a search query
type SearchResults struct {
	Facts []FactResult `json:"facts"`
}

// GetMemoryRequest represents a request to get memory
type GetMemoryRequest struct {
	GroupID        string       `json:"group_id"`
	MaxFacts       int          `json:"max_facts,omitempty"`
	CenterNodeUUID *string      `json:"center_node_uuid"`
	Messages       []Message    `json:"messages"`
	Observation    *Observation `json:"observation,omitempty"`
}

// GetMemoryResponse represents the response from getting memory
type GetMemoryResponse struct {
	Facts []FactResult `json:"facts"`
}

// AddMessagesRequest represents a request to add messages
type AddMessagesRequest struct {
	GroupID     string       `json:"group_id"`
	Messages    []Message    `json:"messages"`
	Observation *Observation `json:"observation,omitempty"`
}

// AddEntityNodeRequest represents a request to add an entity node
type AddEntityNodeRequest struct {
	UUID        string       `json:"uuid"`
	GroupID     string       `json:"group_id"`
	Name        string       `json:"name"`
	Summary     string       `json:"summary,omitempty"`
	Observation *Observation `json:"observation,omitempty"`
}

// EntityNode represents an entity node in the graph
type EntityNode struct {
	UUID      string                 `json:"uuid"`
	GroupID   string                 `json:"group_id"`
	Name      string                 `json:"name"`
	Summary   string                 `json:"summary,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	Labels    []string               `json:"labels,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Episode represents an episode in the graph
type Episode struct {
	UUID              string                 `json:"uuid"`
	GroupID           string                 `json:"group_id"`
	Name              string                 `json:"name"`
	Content           string                 `json:"content"`
	Source            string                 `json:"source"`
	SourceDescription string                 `json:"source_description,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	ValidAt           time.Time              `json:"valid_at"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// Advanced Search Types

// NodeResult represents a node result from search
type NodeResult struct {
	UUID       string                 `json:"uuid"`
	Name       string                 `json:"name"`
	Labels     []string               `json:"labels"`
	Summary    string                 `json:"summary"`
	CreatedAt  time.Time              `json:"created_at"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// EdgeResult represents an edge result from search
type EdgeResult struct {
	UUID           string     `json:"uuid"`
	Name           string     `json:"name"`
	Fact           string     `json:"fact"`
	SourceNodeUUID string     `json:"source_node_uuid"`
	TargetNodeUUID string     `json:"target_node_uuid"`
	ValidAt        *time.Time `json:"valid_at,omitempty"`
	InvalidAt      *time.Time `json:"invalid_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiredAt      *time.Time `json:"expired_at,omitempty"`
}

// EpisodeResult represents an episode result from search
type EpisodeResult struct {
	UUID              string    `json:"uuid"`
	Content           string    `json:"content"`
	Source            string    `json:"source"`
	SourceDescription string    `json:"source_description"`
	CreatedAt         time.Time `json:"created_at"`
	ValidAt           time.Time `json:"valid_at"`
}

// CommunityResult represents a community result from search
type CommunityResult struct {
	UUID      string    `json:"uuid"`
	Name      string    `json:"name"`
	Summary   string    `json:"summary"`
	CreatedAt time.Time `json:"created_at"`
}

// TimeWindow represents a time window
type TimeWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// TemporalSearchRequest represents a temporal window search request
type TemporalSearchRequest struct {
	Query       string       `json:"query"`
	GroupID     *string      `json:"group_id,omitempty"`
	TimeStart   time.Time    `json:"time_start"`
	TimeEnd     time.Time    `json:"time_end"`
	MaxResults  int          `json:"max_results,omitempty"`
	Observation *Observation `json:"observation,omitempty"`
}

// TemporalSearchResponse represents a temporal window search response
type TemporalSearchResponse struct {
	Edges         []EdgeResult    `json:"edges"`
	EdgeScores    []float64       `json:"edge_scores"`
	Nodes         []NodeResult    `json:"nodes"`
	NodeScores    []float64       `json:"node_scores"`
	Episodes      []EpisodeResult `json:"episodes"`
	EpisodeScores []float64       `json:"episode_scores"`
	TimeWindow    TimeWindow      `json:"time_window"`
}

// EntityRelationshipSearchRequest represents an entity relationships search request
type EntityRelationshipSearchRequest struct {
	Query          string       `json:"query"`
	GroupID        *string      `json:"group_id,omitempty"`
	CenterNodeUUID string       `json:"center_node_uuid"`
	MaxDepth       int          `json:"max_depth,omitempty"`
	NodeLabels     *[]string    `json:"node_labels,omitempty"`
	EdgeTypes      *[]string    `json:"edge_types,omitempty"`
	MaxResults     int          `json:"max_results,omitempty"`
	Observation    *Observation `json:"observation,omitempty"`
}

// EntityRelationshipSearchResponse represents an entity relationships search response
type EntityRelationshipSearchResponse struct {
	Edges         []EdgeResult `json:"edges"`
	EdgeDistances []float64    `json:"edge_distances"`
	Nodes         []NodeResult `json:"nodes"`
	NodeDistances []float64    `json:"node_distances"`
	CenterNode    *NodeResult  `json:"center_node,omitempty"`
}

// DiverseSearchRequest represents a diverse results search request
type DiverseSearchRequest struct {
	Query          string       `json:"query"`
	GroupID        *string      `json:"group_id,omitempty"`
	DiversityLevel string       `json:"diversity_level,omitempty"`
	MaxResults     int          `json:"max_results,omitempty"`
	Observation    *Observation `json:"observation,omitempty"`
}

// DiverseSearchResponse represents a diverse results search response
type DiverseSearchResponse struct {
	Edges              []EdgeResult      `json:"edges"`
	EdgeMMRScores      []float64         `json:"edge_mmr_scores"`
	Nodes              []NodeResult      `json:"nodes"`
	NodeMMRScores      []float64         `json:"node_mmr_scores"`
	Episodes           []EpisodeResult   `json:"episodes"`
	EpisodeScores      []float64         `json:"episode_scores"`
	Communities        []CommunityResult `json:"communities"`
	CommunityMMRScores []float64         `json:"community_mmr_scores"`
}

// EpisodeContextSearchRequest represents an episode context search request
type EpisodeContextSearchRequest struct {
	Query       string       `json:"query"`
	GroupID     *string      `json:"group_id,omitempty"`
	MaxResults  int          `json:"max_results,omitempty"`
	Observation *Observation `json:"observation,omitempty"`
}

// EpisodeContextSearchResponse represents an episode context search response
type EpisodeContextSearchResponse struct {
	Episodes            []EpisodeResult `json:"episodes"`
	RerankerScores      []float64       `json:"reranker_scores"`
	MentionedNodes      []NodeResult    `json:"mentioned_nodes"`
	MentionedNodeScores []float64       `json:"mentioned_node_scores"`
}

// SuccessfulToolsSearchRequest represents a successful tools search request
type SuccessfulToolsSearchRequest struct {
	Query       string       `json:"query"`
	GroupID     *string      `json:"group_id,omitempty"`
	MinMentions int          `json:"min_mentions,omitempty"`
	MaxResults  int          `json:"max_results,omitempty"`
	Observation *Observation `json:"observation,omitempty"`
}

// SuccessfulToolsSearchResponse represents a successful tools search response
type SuccessfulToolsSearchResponse struct {
	Edges             []EdgeResult    `json:"edges"`
	EdgeMentionCounts []float64       `json:"edge_mention_counts"`
	Nodes             []NodeResult    `json:"nodes"`
	NodeMentionCounts []float64       `json:"node_mention_counts"`
	Episodes          []EpisodeResult `json:"episodes"`
	EpisodeScores     []float64       `json:"episode_scores"`
}

// RecentContextSearchRequest represents a recent context search request
type RecentContextSearchRequest struct {
	Query         string       `json:"query"`
	GroupID       *string      `json:"group_id,omitempty"`
	RecencyWindow string       `json:"recency_window,omitempty"`
	MaxResults    int          `json:"max_results,omitempty"`
	Observation   *Observation `json:"observation,omitempty"`
}

// RecentContextSearchResponse represents a recent context search response
type RecentContextSearchResponse struct {
	Edges         []EdgeResult    `json:"edges"`
	EdgeScores    []float64       `json:"edge_scores"`
	Nodes         []NodeResult    `json:"nodes"`
	NodeScores    []float64       `json:"node_scores"`
	Episodes      []EpisodeResult `json:"episodes"`
	EpisodeScores []float64       `json:"episode_scores"`
	TimeWindow    TimeWindow      `json:"time_window"`
}

// EntityByLabelSearchRequest represents an entity by label search request
type EntityByLabelSearchRequest struct {
	Query       string       `json:"query"`
	GroupID     *string      `json:"group_id,omitempty"`
	NodeLabels  []string     `json:"node_labels"`
	EdgeTypes   *[]string    `json:"edge_types,omitempty"`
	MaxResults  int          `json:"max_results,omitempty"`
	Observation *Observation `json:"observation,omitempty"`
}

// EntityByLabelSearchResponse represents an entity by label search response
type EntityByLabelSearchResponse struct {
	Nodes      []NodeResult `json:"nodes"`
	NodeScores []float64    `json:"node_scores"`
	Edges      []EdgeResult `json:"edges"`
	EdgeScores []float64    `json:"edge_scores"`
}
