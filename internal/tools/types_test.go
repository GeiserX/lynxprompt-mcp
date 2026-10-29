package tools

import (
	"os"
	"regexp"
	"slices"
	"testing"
)

// nonAgentTypes are TemplateType values that are not agent instruction files
// (repository boilerplate and the legacy CURSORRULES alias), so the write
// tools do not offer them.
var nonAgentTypes = []string{
	"LICENSE", "FUNDING", "CURSORRULES", "GITIGNORE", "README", "CONTRIBUTING",
	"CODE_OF_CONDUCT", "SECURITY", "DOCKERFILE", "DOCKER_COMPOSE",
	"GITHUB_ACTIONS", "GITLAB_CI", "EDITORCONFIG", "PRETTIER", "ESLINT", "PRE_COMMIT",
}

// testdata/template_type.prisma is the TemplateType enum copied from
// LynxPrompt's prisma/schema-users.prisma. When LynxPrompt adds a type, paste
// the new enum there: this test then fails until the type is either offered in
// blueprintTypes or listed in nonAgentTypes.
func TestBlueprintTypes_match_the_upstream_enum(t *testing.T) {
	raw, err := os.ReadFile("testdata/template_type.prisma")
	if err != nil {
		t.Fatal(err)
	}
	upstream := regexp.MustCompile(`(?m)^\s+([A-Z][A-Z0-9_]*)\b`).FindAllStringSubmatch(string(raw), -1)
	if len(upstream) < 40 {
		t.Fatalf("parsed only %d values from the enum copy", len(upstream))
	}
	seen := map[string]bool{}
	for _, m := range upstream {
		v := m[1]
		seen[v] = true
		if !slices.Contains(blueprintTypes, v) && !slices.Contains(nonAgentTypes, v) {
			t.Errorf("upstream type %s is neither in blueprintTypes nor in nonAgentTypes", v)
		}
	}
	for _, v := range blueprintTypes {
		if !seen[v] {
			t.Errorf("blueprintTypes has %s, which the upstream enum does not", v)
		}
	}
	for _, v := range nonAgentTypes {
		if !seen[v] {
			t.Errorf("nonAgentTypes has %s, which the upstream enum does not", v)
		}
	}
}
