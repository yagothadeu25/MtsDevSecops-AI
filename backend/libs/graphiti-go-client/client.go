package graphiti

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client represents a Graphiti API client
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// ClientOption is a functional option for configuring the Client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// NewClient creates a new Graphiti API client
func NewClient(baseURL string, opts ...ClientOption) *Client {
	client := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// do performs an HTTP request and decodes the response
func (c *Client) do(method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	reqURL := c.baseURL + path
	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// HealthCheck performs a health check on the API
func (c *Client) HealthCheck() (*HealthCheckResponse, error) {
	var result HealthCheckResponse
	if err := c.do(http.MethodGet, "/healthcheck", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Search searches for facts in the graph
func (c *Client) Search(query SearchQuery) (*SearchResults, error) {
	var result SearchResults
	if err := c.do(http.MethodPost, "/search", query, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetEntityEdge retrieves a specific entity edge by UUID
func (c *Client) GetEntityEdge(uuid string) (*FactResult, error) {
	var result FactResult
	path := fmt.Sprintf("/entity-edge/%s", url.PathEscape(uuid))
	if err := c.do(http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetEpisodes retrieves episodes for a group
func (c *Client) GetEpisodes(groupID string, lastN int) ([]Episode, error) {
	var result []Episode
	path := fmt.Sprintf("/episodes/%s?last_n=%d", url.PathEscape(groupID), lastN)
	if err := c.do(http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetMemory retrieves memory based on messages
func (c *Client) GetMemory(request GetMemoryRequest) (*GetMemoryResponse, error) {
	var result GetMemoryResponse
	if err := c.do(http.MethodPost, "/get-memory", request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AddMessages adds messages to the graph (asynchronous operation)
func (c *Client) AddMessages(request AddMessagesRequest) (*Result, error) {
	var result Result
	if err := c.do(http.MethodPost, "/messages", request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AddEntityNode adds an entity node to the graph
func (c *Client) AddEntityNode(request AddEntityNodeRequest) (*EntityNode, error) {
	var result EntityNode
	if err := c.do(http.MethodPost, "/entity-node", request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteEntityEdge deletes an entity edge by UUID
func (c *Client) DeleteEntityEdge(uuid string) (*Result, error) {
	var result Result
	path := fmt.Sprintf("/entity-edge/%s", url.PathEscape(uuid))
	if err := c.do(http.MethodDelete, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteGroup deletes a group by ID
func (c *Client) DeleteGroup(groupID string) (*Result, error) {
	var result Result
	path := fmt.Sprintf("/group/%s", url.PathEscape(groupID))
	if err := c.do(http.MethodDelete, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteEpisode deletes an episode by UUID
func (c *Client) DeleteEpisode(uuid string) (*Result, error) {
	var result Result
	path := fmt.Sprintf("/episode/%s", url.PathEscape(uuid))
	if err := c.do(http.MethodDelete, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Clear clears all data from the graph
func (c *Client) Clear() (*Result, error) {
	var result Result
	if err := c.do(http.MethodPost, "/clear", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Advanced Search Methods

// TemporalWindowSearch searches for context within a specific time window
func (c *Client) TemporalWindowSearch(request TemporalSearchRequest) (*TemporalSearchResponse, error) {
	var result TemporalSearchResponse
	if err := c.do(http.MethodPost, "/search/temporal-window", request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// EntityRelationshipsSearch finds relationships and related entities from a center node
func (c *Client) EntityRelationshipsSearch(request EntityRelationshipSearchRequest) (*EntityRelationshipSearchResponse, error) {
	var result EntityRelationshipSearchResponse
	if err := c.do(http.MethodPost, "/search/entity-relationships", request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DiverseResultsSearch gets diverse, non-redundant results using MMR
func (c *Client) DiverseResultsSearch(request DiverseSearchRequest) (*DiverseSearchResponse, error) {
	var result DiverseSearchResponse
	if err := c.do(http.MethodPost, "/search/diverse-results", request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// EpisodeContextSearch searches through agent responses and tool execution records
func (c *Client) EpisodeContextSearch(request EpisodeContextSearchRequest) (*EpisodeContextSearchResponse, error) {
	var result EpisodeContextSearchResponse
	if err := c.do(http.MethodPost, "/search/episode-context", request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SuccessfulToolsSearch finds successful tool executions and attack patterns
func (c *Client) SuccessfulToolsSearch(request SuccessfulToolsSearchRequest) (*SuccessfulToolsSearchResponse, error) {
	var result SuccessfulToolsSearchResponse
	if err := c.do(http.MethodPost, "/search/successful-tools", request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RecentContextSearch retrieves recent relevant context
func (c *Client) RecentContextSearch(request RecentContextSearchRequest) (*RecentContextSearchResponse, error) {
	var result RecentContextSearchResponse
	if err := c.do(http.MethodPost, "/search/recent-context", request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// EntityByLabelSearch searches for entities by label/type with optional edge filtering
func (c *Client) EntityByLabelSearch(request EntityByLabelSearchRequest) (*EntityByLabelSearchResponse, error) {
	var result EntityByLabelSearchResponse
	if err := c.do(http.MethodPost, "/search/entity-by-label", request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
