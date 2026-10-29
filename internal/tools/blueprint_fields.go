package tools

import (
	"fmt"
	"strings"

	"github.com/geiserx/lynxprompt-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// Field options shared by create_blueprint and update_blueprint.

func nameProperty(desc string, opts ...mcp.PropertyOption) mcp.ToolOption {
	return mcp.WithString("name", append([]mcp.PropertyOption{
		mcp.Description(desc), mcp.MinLength(3), mcp.MaxLength(200),
	}, opts...)...)
}

func contentProperty(desc string, opts ...mcp.PropertyOption) mcp.ToolOption {
	return mcp.WithString("content", append([]mcp.PropertyOption{
		mcp.Description(desc), mcp.MinLength(10),
	}, opts...)...)
}

var (
	typeProperty = mcp.WithString("type",
		mcp.Description("File format the blueprint is for. AGENTS_MD is AGENTS.md, CLAUDE_MD is CLAUDE.md, "+
			"CURSOR_MDC a .cursor/rules/*.mdc file, *_COMMAND a slash command, AGENT_SKILL a skill's SKILL.md, "+
			"AGENT_DEFINITION a subagent such as .claude/agents/<name>.md, MCP_CONFIG an .mcp.json, "+
			"CUSTOM anything else."),
		mcp.Enum(blueprintTypes...),
	)
	descriptionProperty = mcp.WithString("description",
		mcp.Description("One or two sentences on what the blueprint is for."),
		mcp.MaxLength(1000),
	)
	visibilityProperty = mcp.WithString("visibility",
		mcp.Description("PRIVATE (only you), TEAM (your team) or PUBLIC. PUBLIC publishes the content "+
			"to the public marketplace where anyone can read it."),
		mcp.Enum(visibilities...),
	)
	teamIDProperty = mcp.WithString("team_id",
		mcp.Description("Team to share a TEAM blueprint with. Needed only when you belong to more than "+
			"one team; with one team, that team is used."),
		mcp.Pattern(`^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`),
	)
	categoryProperty = mcp.WithString("category",
		mcp.Description("Category slug, e.g. web, devops, saas, mobile, data or other."),
		mcp.MaxLength(50),
	)
	tagsProperty = mcp.WithArray("tags",
		mcp.Description("Up to 10 tags of at most 30 characters, e.g. [\"react\", \"typescript\"]."),
		mcp.WithStringItems(mcp.MinLength(1), mcp.MaxLength(30)),
		mcp.MaxItems(10),
	)
)

// commonBlueprintFields copies the optional fields both write tools accept
// into body, checking enums and lengths the way the API will.
func commonBlueprintFields(req mcp.CallToolRequest, body map[string]any) error {
	if v, ok := optString(req, "name"); ok {
		if len([]rune(v)) < 3 {
			return fmt.Errorf("name must be at least 3 characters")
		}
		body["name"] = v
	}
	if v, ok := optString(req, "type"); ok {
		v = strings.ToUpper(v)
		if err := oneOf("type", v, blueprintTypes); err != nil {
			return err
		}
		body["type"] = v
	}
	if v, ok := optString(req, "description"); ok {
		body["description"] = v
	}
	if v, ok := optString(req, "visibility"); ok {
		v = strings.ToUpper(v)
		if err := oneOf("visibility", v, visibilities); err != nil {
			return err
		}
		body["visibility"] = v
	}
	if v, ok := optString(req, "category"); ok {
		body["category"] = v
	}
	tags, err := optStringList(req, "tags")
	if err != nil {
		return err
	}
	if len(tags) > 10 {
		return fmt.Errorf("tags takes at most 10 items (got %d)", len(tags))
	}
	for _, t := range tags {
		if len([]rune(t)) > 30 {
			return fmt.Errorf("tag %q is longer than 30 characters", t)
		}
	}
	if len(tags) > 0 {
		body["tags"] = tags
	}
	return nil
}

// contentField validates blueprint text and puts it in body as content_base64.
func contentField(content string, body map[string]any) error {
	if len(strings.TrimSpace(content)) < 10 {
		return fmt.Errorf("content must be at least 10 characters")
	}
	encodeContent(body, content)
	return nil
}

// teamIDField copies team_id into body. The API reads it only for TEAM
// visibility, to pick one of the user's teams.
func teamIDField(req mcp.CallToolRequest, body map[string]any) error {
	if v, ok := optString(req, "team_id"); ok {
		if err := client.ValidateID(v); err != nil {
			return err
		}
		body["team_id"] = v
	}
	return nil
}
