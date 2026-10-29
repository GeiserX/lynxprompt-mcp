package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// maxResponseBytes caps a successful upstream response. The largest
	// LynxPrompt payload is a single blueprint with its content, far below this.
	maxResponseBytes = 4 << 20
	// maxErrorBodyBytes caps how much of an error response is read to find the
	// JSON {"error","message"} fields. Nothing past it is ever shown. It must
	// hold the whole 422 sensitive-data answer, which lists every match (about
	// 250 bytes each), or the JSON is cut and fails to parse.
	maxErrorBodyBytes = 64 << 10
)

// idPattern accepts LynxPrompt API ids (bp_..., ha_..., usr_... or the bare
// database id). It rejects anything that could change the meaning of the URL
// path, such as ".", "..", "/", "?" or "%".
var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`)

// ValidateID reports whether id is safe to place in an API path.
func ValidateID(id string) error {
	if !idPattern.MatchString(id) {
		return fmt.Errorf("invalid id %q: expected letters, digits, '_' or '-' (for example bp_abc123 or ha_abc123)", id)
	}
	return nil
}

// APIError is a non-2xx answer from the LynxPrompt API, reduced to what is
// safe and useful to show a model: the status, upstream's own short message
// and, for a 409, the checksum the blueprint has now.
type APIError struct {
	Status          int
	Message         string
	CurrentChecksum string
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("LynxPrompt API returned %d: %s", e.Status, e.Message)
	if hint := statusHint(e.Status); hint != "" {
		msg += ". " + hint
	}
	if e.CurrentChecksum != "" {
		msg += " (current content_checksum: " + e.CurrentChecksum + ")"
	}
	return msg
}

func statusHint(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return "Check LYNXPROMPT_TOKEN; tokens are created under Settings > API tokens on your LynxPrompt instance"
	case http.StatusForbidden:
		return "The token lacks the permission or plan this call needs, or a firewall in front of the instance refused the request"
	case http.StatusNotFound:
		return "No item with that id is visible to this token"
	case http.StatusConflict:
		return "The blueprint changed since you read it; call get_blueprint again, merge, and retry with the new content_checksum"
	case http.StatusTooManyRequests:
		return "Rate limited; wait before retrying"
	}
	return ""
}

func newAPIError(status int, body []byte) *APIError {
	e := &APIError{Status: status}
	var parsed struct {
		Error           string `json:"error"`
		Message         string `json:"message"`
		CurrentChecksum string `json:"current_checksum"`
	}
	if json.Unmarshal(body, &parsed) == nil && (parsed.Error != "" || parsed.Message != "") {
		e.Message = oneLine(parsed.Error)
		if parsed.Message != "" {
			if e.Message != "" {
				e.Message += ": "
			}
			e.Message += oneLine(parsed.Message)
		}
		e.CurrentChecksum = oneLine(parsed.CurrentChecksum)
		return e
	}
	// Not the API's JSON error shape: an HTML page from a proxy or a WAF, or a
	// truncated body. Never echo it.
	e.Message = http.StatusText(status)
	if e.Message == "" {
		e.Message = "unexpected response"
	}
	e.Message += " (the response was not LynxPrompt API JSON)"
	return e
}

// oneLine keeps an upstream string short and on one line.
func oneLine(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 300 {
		s = s[:300] + "..."
	}
	return s
}

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
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	return u
}

// do sends an authenticated request.
func (c *Client) do(req *http.Request) ([]byte, error) {
	return c.send(req, true)
}

func (c *Client) send(req *http.Request, withAuth bool) ([]byte, error) {
	if withAuth && c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.hc.Do(req)
	if err != nil {
		// *url.Error carries the full request URL; keep only the cause.
		var ue *url.Error
		if errors.As(err, &ue) {
			err = ue.Err
		}
		return nil, fmt.Errorf("LynxPrompt API request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyBytes))
		return nil, newAPIError(resp.StatusCode, b)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("LynxPrompt API response could not be read: %v", err)
	}
	if len(b) > maxResponseBytes {
		return nil, fmt.Errorf("LynxPrompt API response is larger than %d bytes; refusing it", maxResponseBytes)
	}
	return b, nil
}

func (c *Client) get(ctx context.Context, path string, q url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL(path, q), nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) sendJSON(ctx context.Context, method, path string, body map[string]any) ([]byte, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.buildURL(path, nil), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

func (c *Client) itemPath(prefix, id string) (string, error) {
	if err := ValidateID(id); err != nil {
		return "", err
	}
	return prefix + url.PathEscape(id), nil
}

// ListOptions pages the owner-only list endpoints. Zero values leave the
// upstream defaults (limit 50, offset 0, every visibility).
type ListOptions struct {
	Limit      int
	Offset     int
	Visibility string // PRIVATE, TEAM, PUBLIC or "" for all (blueprints only)
}

func (o ListOptions) query() url.Values {
	q := url.Values{}
	if o.Limit > 0 {
		q.Set("limit", strconv.Itoa(o.Limit))
	}
	if o.Offset > 0 {
		q.Set("offset", strconv.Itoa(o.Offset))
	}
	if o.Visibility != "" {
		q.Set("visibility", o.Visibility)
	}
	return q
}

// ---------------------------------------------------------------------------
// Reads (authenticated /api/v1)
// ---------------------------------------------------------------------------

// ListBlueprints returns one page of the authenticated user's blueprints.
func (c *Client) ListBlueprints(ctx context.Context, opts ListOptions) ([]byte, error) {
	return c.get(ctx, "/api/v1/blueprints", opts.query())
}

// GetBlueprint returns a single blueprint by ID (includes content).
func (c *Client) GetBlueprint(ctx context.Context, id string) ([]byte, error) {
	p, err := c.itemPath("/api/v1/blueprints/", id)
	if err != nil {
		return nil, err
	}
	return c.get(ctx, p, nil)
}

// ListHierarchies returns one page of the authenticated user's hierarchies.
func (c *Client) ListHierarchies(ctx context.Context, opts ListOptions) ([]byte, error) {
	opts.Visibility = ""
	return c.get(ctx, "/api/v1/hierarchies", opts.query())
}

// GetHierarchy returns a single hierarchy by ID (includes tree).
func (c *Client) GetHierarchy(ctx context.Context, id string) ([]byte, error) {
	p, err := c.itemPath("/api/v1/hierarchies/", id)
	if err != nil {
		return nil, err
	}
	return c.get(ctx, p, nil)
}

// GetUser returns the authenticated user's info.
func (c *Client) GetUser(ctx context.Context) ([]byte, error) {
	return c.get(ctx, "/api/v1/user", nil)
}

// ---------------------------------------------------------------------------
// Public marketplace search
// ---------------------------------------------------------------------------

// SearchParams are the filters the public /api/blueprints endpoint honours.
type SearchParams struct {
	Query    string
	Category string
	Type     string // "configs" or "commands"
	Tags     []string
	Sort     string // popular, recent, downloads, favorites
	Page     int
	Limit    int
}

// SearchBlueprints searches the public marketplace. The endpoint is public,
// so the API token is not sent.
func (c *Client) SearchBlueprints(ctx context.Context, p SearchParams) ([]byte, error) {
	q := url.Values{}
	if p.Query != "" {
		q.Set("q", p.Query)
	}
	if p.Category != "" {
		q.Set("category", p.Category)
	}
	if p.Type != "" {
		q.Set("type", p.Type)
	}
	if len(p.Tags) > 0 {
		q.Set("tags", strings.Join(p.Tags, ","))
	}
	if p.Sort != "" {
		q.Set("sort", p.Sort)
	}
	if p.Page > 0 {
		q.Set("page", strconv.Itoa(p.Page))
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL("/api/blueprints", q), nil)
	if err != nil {
		return nil, err
	}
	return c.send(req, false)
}

// ---------------------------------------------------------------------------
// Writes (authenticated /api/v1)
// ---------------------------------------------------------------------------

// CreateBlueprint creates a new blueprint.
func (c *Client) CreateBlueprint(ctx context.Context, body map[string]any) ([]byte, error) {
	return c.sendJSON(ctx, http.MethodPost, "/api/v1/blueprints", body)
}

// UpdateBlueprint updates an existing blueprint by ID.
func (c *Client) UpdateBlueprint(ctx context.Context, id string, body map[string]any) ([]byte, error) {
	p, err := c.itemPath("/api/v1/blueprints/", id)
	if err != nil {
		return nil, err
	}
	return c.sendJSON(ctx, http.MethodPut, p, body)
}

// DeleteBlueprint deletes a blueprint by ID.
func (c *Client) DeleteBlueprint(ctx context.Context, id string) ([]byte, error) {
	p, err := c.itemPath("/api/v1/blueprints/", id)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.buildURL(p, nil), nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// CreateHierarchy creates a new hierarchy.
func (c *Client) CreateHierarchy(ctx context.Context, body map[string]any) ([]byte, error) {
	return c.sendJSON(ctx, http.MethodPost, "/api/v1/hierarchies", body)
}

// DeleteHierarchy deletes a hierarchy by ID.
func (c *Client) DeleteHierarchy(ctx context.Context, id string) ([]byte, error) {
	p, err := c.itemPath("/api/v1/hierarchies/", id)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.buildURL(p, nil), nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}
