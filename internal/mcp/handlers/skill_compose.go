package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/datran93/knowns/internal/storage"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterSkillComposerTool registers the skill composition MCP tool.
// It allows running composite skills built from multiple skill steps with
// variable interpolation ($1, $2, etc.).
func RegisterSkillComposerTool(s *server.MCPServer, getStore func() *storage.Store) {
	s.AddTool(
		mcp.NewTool("skill_compose",
			mcp.WithDescription("Skill composition and workflow operations. Use 'action' to specify: run, list."),
			mcp.WithString("action",
				mcp.Required(),
				mcp.Description("Action to perform"),
				mcp.Enum("run", "list"),
			),
			mcp.WithString("name",
				mcp.Description("Composite skill name (required for run)"),
			),
			mcp.WithArray("vars",
				mcp.Description("Variables for interpolation (positional: $1, $2, ... for run)"),
				mcp.WithStringItems(),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			action, err := req.RequireString("action")
			if err != nil {
				return errResult("action is required")
			}
			switch action {
			case "run":
				return handleSkillComposeRun(getStore, req)
			case "list":
				return handleSkillComposeList(getStore, req)
			default:
				return errResultf("unknown skill_compose action: %s", action)
			}
		},
	)
}

// CompositeSkill represents a skill made of multiple steps.
type CompositeSkill struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Steps       []SkillStep `json:"steps"`
}

// SkillStep is a single step in a composite skill.
type SkillStep struct {
	Skill string            `json:"skill"` // e.g., "kn-verify", "kn-doc"
	Args  map[string]any    `json:"args,omitempty"`
}

func handleSkillComposeRun(getStore func() *storage.Store, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	name, _ := stringArg(args, "name")
	if name == "" {
		return errResult("name is required for run")
	}

	varVars := stringArrayArg(args, "vars")
	// Build $1, $2, ... map from varVars
	interpolations := make(map[string]string)
	for i, v := range varVars {
		interpolations[fmt.Sprintf("$%d", i+1)] = v
	}

	skill := getBuiltInCompositeSkill(name)
	if skill == nil {
		return errResultf("composite skill %q not found", name)
	}

	// Run steps sequentially (simulated — actual skill execution would call MCP tools).
	results := make([]map[string]any, 0, len(skill.Steps))
	for i, step := range skill.Steps {
		// Interpolate args with $1, $2, ...
		interpolatedArgs := interpolateArgs(step.Args, interpolations)
		results = append(results, map[string]any{
			"step":     i + 1,
			"skill":    step.Skill,
			"args":     interpolatedArgs,
			"status":   "ok", // Would actually execute the skill here
			"output":   fmt.Sprintf("[simulated] %s with %v", step.Skill, interpolatedArgs),
		})
	}

	out, _ := json.MarshalIndent(map[string]any{
		"skill":  name,
		"steps":  len(skill.Steps),
		"results": results,
	}, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func handleSkillComposeList(getStore func() *storage.Store, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	skills := []map[string]any{
		{"name": "full-review", "description": "Review code changes, docs, and run validation", "steps": 3},
		{"name": "implement-and-test", "description": "Implement task and run tests", "steps": 4},
		{"name": "review-and-commit", "description": "Review changes and commit if approved", "steps": 2},
	}
	out, _ := json.MarshalIndent(skills, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

// getBuiltInCompositeSkill returns the built-in composite skill by name.
func getBuiltInCompositeSkill(name string) *CompositeSkill {
	switch name {
	case "full-review":
		return &CompositeSkill{
			Name:        "full-review",
			Description: "Review code changes, docs, and run validation",
			Steps: []SkillStep{
				{Skill: "kn-verify", Args: map[string]any{"scope": "all"}},
				{Skill: "kn-doc", Args: map[string]any{"action": "list"}},
				{Skill: "validate", Args: map[string]any{"scope": "all"}},
			},
		}
	case "implement-and-test":
		return &CompositeSkill{
			Name:        "implement-and-test",
			Description: "Implement task and run tests",
			Steps: []SkillStep{
				{Skill: "kn-plan", Args: map[string]any{"action": "current"}},
				{Skill: "kn-implement", Args: map[string]any{"action": "current"}},
				{Skill: "kn-verify", Args: map[string]any{"scope": "all"}},
				{Skill: "validate", Args: map[string]any{"scope": "all"}},
			},
		}
	case "review-and-commit":
		return &CompositeSkill{
			Name:        "review-and-commit",
			Description: "Review changes and commit if approved",
			Steps: []SkillStep{
				{Skill: "kn-verify", Args: map[string]any{}},
				{Skill: "kn-commit", Args: map[string]any{"auto": "true"}},
			},
		}
	}
	return nil
}

// interpolateArgs replaces $1, $2, ... in arg values with actual values.
func interpolateArgs(args map[string]any, interp map[string]string) map[string]any {
	if args == nil {
		return nil
	}
	result := make(map[string]any, len(args))
	for k, v := range args {
		if s, ok := v.(string); ok {
			result[k] = interpolateString(s, interp)
		} else {
			result[k] = v
		}
	}
	return result
}

// interpolateString replaces $1, $2, etc. in s with values from interp.
func interpolateString(s string, interp map[string]string) string {
	for k, v := range interp {
		s = replaceAll(s, k, v)
	}
	return s
}

func replaceAll(s, old, new string) string {
	if old == "" {
		return s
	}
	result := s
	for {
		idx := -1
		for i := 0; i <= len(result)-len(old); i++ {
			if result[i:i+len(old)] == old {
				idx = i
				break
			}
		}
		if idx < 0 {
			break
		}
		result = result[:idx] + new + result[idx+len(old):]
	}
	return result
}

