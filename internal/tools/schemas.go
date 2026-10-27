package tools

import (
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
)

// Output schemas mirror the LynxPrompt API responses. They list the fields an
// agent needs and leave additional properties open, so a new upstream field
// never turns a good answer into a schema error.

const (
	nullableString = `{"type":["string","null"]}`

	blueprintSchema = `{
  "type": "object",
  "properties": {
    "id": {"type": "string", "description": "Blueprint id, e.g. bp_clx1abc"},
    "name": {"type": "string"},
    "description": ` + nullableString + `,
    "content": {"type": "string", "description": "The blueprint text (only on single-blueprint reads and updates)"},
    "type": {"type": "string", "description": "File format, e.g. AGENTS_MD or CLAUDE_MD"},
    "tier": {"type": ["string", "null"]},
    "category": ` + nullableString + `,
    "tags": {"type": "array", "items": {"type": "string"}},
    "visibility": {"type": "string", "description": "PRIVATE, TEAM or PUBLIC"},
    "hierarchy_id": ` + nullableString + `,
    "parent_id": ` + nullableString + `,
    "repository_path": ` + nullableString + `,
    "content_checksum": {"type": ["string", "null"], "description": "Pass as expected_checksum to update_blueprint"},
    "created_at": {"type": "string"},
    "updated_at": {"type": "string"}
  }
}`

	hierarchySchema = `{
  "type": "object",
  "properties": {
    "id": {"type": "string", "description": "Hierarchy id, e.g. ha_clx1abc"},
    "name": {"type": "string"},
    "description": ` + nullableString + `,
    "repository_root": {"type": "string"},
    "blueprint_count": {"type": "integer"},
    "created_at": {"type": "string"},
    "updated_at": {"type": "string"}
  }
}`

	pageProps = `
    "total": {"type": "integer", "description": "Items matching, across all pages"},
    "limit": {"type": "integer"},
    "offset": {"type": "integer"},
    "has_more": {"type": "boolean", "description": "True when another page exists: call again with offset + limit"}`
)

func rawSchema(s string) mcp.ToolOption {
	if !json.Valid([]byte(s)) {
		panic("invalid output schema: " + s)
	}
	return mcp.WithRawOutputSchema(json.RawMessage(s))
}

var (
	listBlueprintsOutput = rawSchema(`{"type":"object","properties":{
    "blueprints": {"type": "array", "items": ` + blueprintSchema + `},` + pageProps + `}}`)

	blueprintOutput = rawSchema(`{"type":"object","properties":{
    "success": {"type": "boolean"},
    "blueprint": ` + blueprintSchema + `}}`)

	listHierarchiesOutput = rawSchema(`{"type":"object","properties":{
    "hierarchies": {"type": "array", "items": ` + hierarchySchema + `},` + pageProps + `}}`)

	hierarchyOutput = rawSchema(`{"type":"object","properties":{
    "success": {"type": "boolean"},
    "message": {"type": "string"},
    "hierarchy": ` + hierarchySchema + `,
    "blueprints": {"type": "array", "items": ` + blueprintSchema + `, "description": "Every blueprint in the hierarchy, flat"},
    "tree": {"type": "array", "items": {"type": "object"}, "description": "The same blueprints nested by parent_id"},
    "total_blueprints": {"type": "integer"}}}`)

	searchOutput = rawSchema(`{"type":"object","properties":{
    "templates": {"type": "array", "items": {"type": "object", "properties": {
      "id": {"type": "string"},
      "name": {"type": "string"},
      "description": ` + nullableString + `,
      "author": {"type": "string"},
      "type": {"type": "string"},
      "category": ` + nullableString + `,
      "tier": ` + nullableString + `,
      "tags": {"type": "array", "items": {"type": "string"}},
      "downloads": {"type": "number"},
      "likes": {"type": "number"}}}},
    "popularTags": {"type": "array"},
    "total": {"type": "integer"},
    "hasMore": {"type": "boolean", "description": "True when another page exists: call again with page + 1"}}}`)

	deleteOutput = rawSchema(`{"type":"object","properties":{
    "success": {"type": "boolean"},
    "message": {"type": "string"}}}`)
)
