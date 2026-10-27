# Usage

## Tools

| Tool | Read-only | Destructive | What it does |
|------|-----------|-------------|--------------|
| `search_blueprints` | yes | no | Search the public marketplace: `query`, `type` (`configs` or `commands`), `category`, `tags[]`, `sort` (`popular`, `recent`, `downloads`, `favorites`), `page`, `limit` (1 to 50). |
| `list_blueprints` | yes | no | Your blueprints without content: `limit` (1 to 100), `offset`, `visibility`. Page while `has_more` is true. |
| `get_blueprint` | yes | no | One of your blueprints with `content` and `content_checksum`. |
| `list_hierarchies` | yes | no | Your hierarchies: `limit`, `offset`. |
| `get_hierarchy` | yes | no | One hierarchy with its blueprints as a flat list and a tree. |
| `create_blueprint` | no | no | `name`, `content`, optional `type` (default `AGENTS_MD`), `description`, `visibility`, `category`, `tags[]`, `hierarchy_id`, `parent_id`, `repository_path`. |
| `update_blueprint` | no | yes | Changes only the fields passed. Pass `expected_checksum` (the `content_checksum` from `get_blueprint`); a 409 means someone else changed it first. |
| `delete_blueprint` | no | yes | Permanent. |
| `create_hierarchy` | no | no | `name`, `repository_root`, `description`. Returns the existing hierarchy for the same `repository_root`. |
| `delete_hierarchy` | no | yes | Deletes the hierarchy; its blueprints are kept, unlinked. |

Ids are `bp_...` for blueprints and `ha_...` for hierarchies. Results come back as `structuredContent` that matches each tool's output schema, plus the same JSON as text. Blueprint content is sent to the API base64-encoded (`content_base64`), so Cloudflare's firewall in front of lynxprompt.com does not reject shell or config snippets.

With `LYNXPROMPT_READ_ONLY=true` only the five read tools are registered.

## Resources

`lynxprompt://blueprints`, `lynxprompt://blueprint/{id}`, `lynxprompt://hierarchies`, `lynxprompt://hierarchy/{id}`, `lynxprompt://user`. The two lists return up to 100 items; the list tools page past that.

## Protocol versions

The server speaks MCP 2026-07-28, 2025-11-25, 2025-06-18, 2025-03-26 and 2024-11-05. A 2026-07-28 client calls `server/discover` and sends its version in each request's `_meta`; older clients use the `initialize` handshake. Both work on stdio and on the stateless HTTP endpoint (`/mcp`).
