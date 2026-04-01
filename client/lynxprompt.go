package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	base  string
	token string
	hc    *http.Client
}

func New(base, token string) *Client {
	return &Client{
		base:  base,
		token: token,
		hc:    &http.Client{},
	}
}

func (c *Client) buildURL(path string, q url.Values) string {
	u, _ := url.Parse(c.base)
	u.Path = path
	if q != nil {
		u.RawQuery = q.Encode()
	}
	return u.String()
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
func (c *Client) ListBlueprints() ([]byte, error) {
	u := c.buildURL("/api/v1/blueprints", nil)
	req, _ := http.NewRequest("GET", u, nil)
	return c.do(req)
}

// GetBlueprint returns a single blueprint by ID (includes content).
func (c *Client) GetBlueprint(id string) ([]byte, error) {
	u := c.buildURL("/api/v1/blueprints/"+id, nil)
	req, _ := http.NewRequest("GET", u, nil)
	return c.do(req)
}

// ListHierarchies returns all hierarchies for the authenticated user.
func (c *Client) ListHierarchies() ([]byte, error) {
	u := c.buildURL("/api/v1/hierarchies", nil)
	req, _ := http.NewRequest("GET", u, nil)
	return c.do(req)
}

// GetHierarchy returns a single hierarchy by ID (includes tree).
func (c *Client) GetHierarchy(id string) ([]byte, error) {
	u := c.buildURL("/api/v1/hierarchies/"+id, nil)
	req, _ := http.NewRequest("GET", u, nil)
	return c.do(req)
}

// GetUser returns the authenticated user's info.
func (c *Client) GetUser() ([]byte, error) {
	u := c.buildURL("/api/v1/user", nil)
	req, _ := http.NewRequest("GET", u, nil)
	return c.do(req)
}

// ---------------------------------------------------------------------------
// Tools (public search + authenticated CRUD)
// ---------------------------------------------------------------------------

// SearchBlueprints searches the public marketplace.
func (c *Client) SearchBlueprints(query, category, bpType, tags string) ([]byte, error) {
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
	req, _ := http.NewRequest("GET", u, nil)
	return c.do(req)
}

// CreateBlueprint creates a new blueprint.
func (c *Client) CreateBlueprint(body map[string]any) ([]byte, error) {
	b, _ := json.Marshal(body)
	u := c.buildURL("/api/v1/blueprints", nil)
	req, _ := http.NewRequest("POST", u, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

// UpdateBlueprint updates an existing blueprint by ID.
func (c *Client) UpdateBlueprint(id string, body map[string]any) ([]byte, error) {
	b, _ := json.Marshal(body)
	u := c.buildURL("/api/v1/blueprints/"+id, nil)
	req, _ := http.NewRequest("PUT", u, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

// DeleteBlueprint deletes a blueprint by ID.
func (c *Client) DeleteBlueprint(id string) ([]byte, error) {
	u := c.buildURL("/api/v1/blueprints/"+id, nil)
	req, _ := http.NewRequest("DELETE", u, nil)
	return c.do(req)
}

// CreateHierarchy creates a new hierarchy.
func (c *Client) CreateHierarchy(body map[string]any) ([]byte, error) {
	b, _ := json.Marshal(body)
	u := c.buildURL("/api/v1/hierarchies", nil)
	req, _ := http.NewRequest("POST", u, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

// DeleteHierarchy deletes a hierarchy by ID.
func (c *Client) DeleteHierarchy(id string) ([]byte, error) {
	u := c.buildURL("/api/v1/hierarchies/"+id, nil)
	req, _ := http.NewRequest("DELETE", u, nil)
	return c.do(req)
}
