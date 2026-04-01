package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	base  string
	token string
	hc    *http.Client
}

func New(base, token string) *Client {
	return &Client{
		base:  strings.TrimRight(base, "/"),
		token: token,
		hc:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) buildURL(path string, q url.Values) string {
	u := c.base + path
	if q != nil && len(q) > 0 {
		u += "?" + q.Encode()
	}
	return u
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LynxPrompt error %d: %s", resp.StatusCode, string(b))
	}
	return io.ReadAll(resp.Body)
}

// ---------------------------------------------------------------------------
// Resources (GET, authenticated /api/v1)
// ---------------------------------------------------------------------------

// ListBlueprints returns all blueprints for the authenticated user.
func (c *Client) ListBlueprints(ctx context.Context) ([]byte, error) {
	u := c.buildURL("/api/v1/blueprints", nil)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// GetBlueprint returns a single blueprint by ID (includes content).
func (c *Client) GetBlueprint(ctx context.Context, id string) ([]byte, error) {
	u := c.buildURL("/api/v1/blueprints/"+id, nil)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// ListHierarchies returns all hierarchies for the authenticated user.
func (c *Client) ListHierarchies(ctx context.Context) ([]byte, error) {
	u := c.buildURL("/api/v1/hierarchies", nil)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// GetHierarchy returns a single hierarchy by ID (includes tree).
func (c *Client) GetHierarchy(ctx context.Context, id string) ([]byte, error) {
	u := c.buildURL("/api/v1/hierarchies/"+id, nil)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// GetUser returns the authenticated user's info.
func (c *Client) GetUser(ctx context.Context) ([]byte, error) {
	u := c.buildURL("/api/v1/user", nil)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// ---------------------------------------------------------------------------
// Tools (public search + authenticated CRUD)
// ---------------------------------------------------------------------------

// SearchBlueprints searches the public marketplace.
func (c *Client) SearchBlueprints(ctx context.Context, query, category, bpType, tags string) ([]byte, error) {
	q := url.Values{}
	q.Set("q", query)
	if category != "" {
		q.Set("category", category)
	}
	if bpType != "" {
		q.Set("type", bpType)
	}
	if tags != "" {
		q.Set("tags", tags)
	}
	u := c.buildURL("/api/blueprints", q)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// CreateBlueprint creates a new blueprint.
func (c *Client) CreateBlueprint(ctx context.Context, body map[string]any) ([]byte, error) {
	b, _ := json.Marshal(body)
	u := c.buildURL("/api/v1/blueprints", nil)
	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

// UpdateBlueprint updates an existing blueprint by ID.
func (c *Client) UpdateBlueprint(ctx context.Context, id string, body map[string]any) ([]byte, error) {
	b, _ := json.Marshal(body)
	u := c.buildURL("/api/v1/blueprints/"+id, nil)
	req, err := http.NewRequestWithContext(ctx, "PUT", u, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

// DeleteBlueprint deletes a blueprint by ID.
func (c *Client) DeleteBlueprint(ctx context.Context, id string) ([]byte, error) {
	u := c.buildURL("/api/v1/blueprints/"+id, nil)
	req, err := http.NewRequestWithContext(ctx, "DELETE", u, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// CreateHierarchy creates a new hierarchy.
func (c *Client) CreateHierarchy(ctx context.Context, body map[string]any) ([]byte, error) {
	b, _ := json.Marshal(body)
	u := c.buildURL("/api/v1/hierarchies", nil)
	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

// DeleteHierarchy deletes a hierarchy by ID.
func (c *Client) DeleteHierarchy(ctx context.Context, id string) ([]byte, error) {
	u := c.buildURL("/api/v1/hierarchies/"+id, nil)
	req, err := http.NewRequestWithContext(ctx, "DELETE", u, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}
