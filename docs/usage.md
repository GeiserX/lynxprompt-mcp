# Usage

| Type          | What for                                                   | MCP URI / Tool id                |
|---------------|------------------------------------------------------------|----------------------------------|
| **Resources** | Browse blueprints, hierarchies, and user info read-only    | `lynxprompt://blueprints`<br>`lynxprompt://blueprint/{id}`<br>`lynxprompt://hierarchies`<br>`lynxprompt://hierarchy/{id}`<br>`lynxprompt://user` |
| **Tools**     | Create, update, delete blueprints and manage hierarchies   | `search_blueprints`<br>`create_blueprint`<br>`update_blueprint`<br>`delete_blueprint`<br>`create_hierarchy`<br>`delete_hierarchy` |

Everything is exposed over a single JSON-RPC endpoint (`/mcp`).
LLMs / Agents can: `initialize` -> `readResource` -> `listTools` -> `callTool` ... and so on.
