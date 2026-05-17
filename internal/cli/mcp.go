package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/datran93/knowns/internal/mcp"
)

// --- mcp platform registry ---

// mcpPlatform represents a platform that can have MCP configuration.
type mcpPlatform struct {
	Name    string
	Configs []mcpConfigTarget
}

// mcpConfigTarget represents a single config file for a platform.
type mcpConfigTarget struct {
	Path   string // absolute or relative to project root
	Global bool   // if true, path is home-relative
}

// mcpSetupFunc is a function that sets up MCP for a platform.
type mcpSetupFunc func(projectRoot, cmd string, args []string) error

var mcpPlatformRegistry = map[string]struct {
	Name  string
	Setup mcpSetupFunc
}{
	"claude-code": {
		Name: "Claude Code",
		Setup: func(projectRoot, cmd string, args []string) error {
			return setupClaudeCode(projectRoot, cmd, args)
		},
	},
	"kiro": {
		Name: "Kiro",
		Setup: func(projectRoot, cmd string, args []string) error {
			return setupKiro(projectRoot, cmd, args)
		},
	},
	"opencode": {
		Name: "OpenCode",
		Setup: func(projectRoot, cmd string, args []string) error {
			return setupOpenCode(projectRoot, cmd, args)
		},
	},
	"cursor": {
		Name: "Cursor",
		Setup: func(projectRoot, cmd string, args []string) error {
			return setupCursor(projectRoot, cmd, args)
		},
	},
	"codex": {
		Name: "Codex",
		Setup: func(projectRoot, cmd string, args []string) error {
			return setupCodex(projectRoot, cmd, args)
		},
	},
	"cline": {
		Name: "Cline",
		Setup: func(projectRoot, cmd string, args []string) error {
			return setupCline(projectRoot, cmd, args)
		},
	},
	"continue": {
		Name: "Continue",
		Setup: func(projectRoot, cmd string, args []string) error {
			return setupContinue(projectRoot, cmd, args)
		},
	},
	"claude-desktop": {
		Name: "Claude Desktop",
		Setup: func(projectRoot, cmd string, args []string) error {
			return setupClaudeDesktop(cmd, args)
		},
	},
	"antigravity": {
		Name: "Antigravity",
		Setup: func(projectRoot, cmd string, args []string) error {
			return setupAntigravity(cmd, args)
		},
	},
	"gemini": {
		Name: "Gemini CLI",
		Setup: func(projectRoot, cmd string, args []string) error {
			return setupGeminiCLI(cmd, args)
		},
	},
}

// --- helper functions ---

func getKnownExecutable() string {
	execPath, err := os.Executable()
	if err != nil {
		return "knowns"
	}
	return execPath
}

func ensureMCPServerEntry(configPath string, cmd string, args []string) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	var config map[string]any
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("parse config: %w", err)
		}
	} else {
		config = make(map[string]any)
	}

	mcpServers, ok := config["mcpServers"].(map[string]any)
	if !ok {
		mcpServers = make(map[string]any)
		config["mcpServers"] = mcpServers
	}

	mcpServers["knowns"] = map[string]any{
		"command": cmd,
		"args":   args,
	}

	out, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(configPath, append(out, '\n'), 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// --- platform-specific setup functions ---

func setupClaudeCode(projectRoot, cmd string, args []string) error {
	configPath := filepath.Join(projectRoot, ".claude", "settings.json")
	if err := ensureMCPServerEntry(configPath, cmd, args); err != nil {
		return err
	}
	fmt.Printf("  %s %s\n", RenderSuccess("✓"), configPath)
	return nil
}

func setupKiro(projectRoot, cmd string, args []string) error {
	configPath := filepath.Join(projectRoot, ".kiro", "settings", "mcp.json")
	if err := ensureMCPServerEntry(configPath, cmd, args); err != nil {
		return err
	}
	fmt.Printf("  %s %s\n", RenderSuccess("✓"), configPath)
	return nil
}

func setupOpenCode(projectRoot, cmd string, args []string) error {
	configPath := filepath.Join(projectRoot, "opencode.json")
	if err := ensureMCPServerEntry(configPath, cmd, args); err != nil {
		return err
	}
	fmt.Printf("  %s %s\n", RenderSuccess("✓"), configPath)
	return nil
}

func setupCursor(projectRoot, cmd string, args []string) error {
	configPath := filepath.Join(projectRoot, ".cursor", "mcp.json")
	if err := ensureMCPServerEntry(configPath, cmd, args); err != nil {
		return err
	}
	fmt.Printf("  %s %s\n", RenderSuccess("✓"), configPath)
	return nil
}

func setupCodex(projectRoot, cmd string, args []string) error {
	configPath := filepath.Join(projectRoot, ".codex", "config.toml")
	return setupTomlConfig(configPath, cmd, args)
}

func setupCline(projectRoot, cmd string, args []string) error {
	configPath := filepath.Join(projectRoot, ".cline", "mcp_settings.json")
	if err := ensureMCPServerEntry(configPath, cmd, args); err != nil {
		return err
	}
	fmt.Printf("  %s %s\n", RenderSuccess("✓"), configPath)
	return nil
}

func setupContinue(projectRoot, cmd string, args []string) error {
	configPath := filepath.Join(projectRoot, ".continue", "config.json")
	if err := ensureMCPServerEntry(configPath, cmd, args); err != nil {
		return err
	}
	fmt.Printf("  %s %s\n", RenderSuccess("✓"), configPath)
	return nil
}

func setupClaudeDesktop(cmd string, args []string) error {
	home, _ := os.UserHomeDir()
	var configPath string
	if runtime.GOOS == "windows" {
		configPath = filepath.Join(home, "AppData", "Roaming", "Claude", "claude_desktop_config.json")
	} else {
		configPath = filepath.Join(home, ".claude", "claude_desktop_config.json")
	}
	if err := ensureMCPServerEntry(configPath, cmd, args); err != nil {
		return err
	}
	fmt.Printf("  %s %s\n", RenderSuccess("✓"), configPath)
	return nil
}

func setupAntigravity(cmd string, args []string) error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".gemini", "antigravity", "mcp_config.json")
	if err := ensureMCPServerEntry(configPath, cmd, args); err != nil {
		return err
	}
	fmt.Printf("  %s %s\n", RenderSuccess("✓"), configPath)
	return nil
}

func setupGeminiCLI(cmd string, args []string) error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".gemini", "cli", "mcp_config.json")
	if err := ensureMCPServerEntry(configPath, cmd, args); err != nil {
		return err
	}
	fmt.Printf("  %s %s\n", RenderSuccess("✓"), configPath)
	return nil
}

func setupTomlConfig(configPath, cmd string, args []string) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	var config map[string]any
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("parse config: %w", err)
		}
	} else {
		config = make(map[string]any)
	}

	mcpServers, ok := config["mcp_servers"].(map[string]any)
	if !ok {
		mcpServers = make(map[string]any)
		config["mcp_servers"] = mcpServers
	}

	mcpServers["knowns"] = map[string]any{
		"command": cmd,
		"args":   args,
	}

	out, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(configPath, append(out, '\n'), 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	fmt.Printf("  %s %s\n", RenderSuccess("✓"), configPath)
	return nil
}

// --- main MCP command ---

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP (Model Context Protocol) server",
	Long: `Start the Knowns MCP server, which exposes project management tools
to AI agents via the Model Context Protocol.

Use --stdio to communicate over stdin/stdout (default for MCP clients).`,
	RunE:         runMCP,
	SilenceUsage: true,
}

func runMCP(cmd *cobra.Command, args []string) error {
	projectRoot, _ := cmd.Flags().GetString("project")
	if projectRoot == "" {
		projectRoot = os.Getenv("KNOWNS_PROJECT")
	}

	s := mcp.NewMCPServer(projectRoot)
	return s.Start()
}

// --- mcp setup ---

var mcpSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up MCP for AI platforms",
	Long:  "Set up Knowns MCP server configuration for one or more AI platforms.",
	RunE:  runMCPSetup,
}

func runMCPSetup(cmd *cobra.Command, args []string) error {
	platforms := args
	if len(platforms) == 0 {
		platforms = []string{"claude-code"}
	}

	execPath := getKnownExecutable()
	argsList := []string{"mcp", "--stdio"}

	var lastErr error
	for _, platform := range platforms {
		setup, ok := mcpPlatformRegistry[platform]
		if !ok {
			fmt.Printf("  %s unknown platform: %s\n", RenderWarning("⚠"), platform)
			continue
		}
		fmt.Printf("  %s Setting up %s...\n", StyleInfo.Render("→"), setup.Name)
		projectRoot, _ := os.Getwd()
		if err := setup.Setup(projectRoot, execPath, argsList); err != nil {
			lastErr = err
			continue
		}
	}

	if lastErr != nil {
		return lastErr
	}
	fmt.Println(RenderSuccess("MCP setup complete."))
	return nil
}

func init() {
	mcpCmd.Flags().Bool("stdio", false, "Use stdio transport (for MCP clients)")
	mcpCmd.Flags().String("project", "", "Project root directory (auto-detected from cwd if not set)")

	// Add platform subcommands
	for platformKey := range mcpPlatformRegistry {
		platformCmd := &cobra.Command{
			Use:   platformKey,
			Short: fmt.Sprintf("Set up MCP for %s", mcpPlatformRegistry[platformKey].Name),
			RunE: func(cmd *cobra.Command, args []string) error {
				execPath := getKnownExecutable()
				argsList := []string{"mcp", "--stdio"}
				projectRoot, _ := os.Getwd()
				fmt.Printf("  %s Setting up %s...\n", StyleInfo.Render("→"), mcpPlatformRegistry[platformKey].Name)
				if err := mcpPlatformRegistry[platformKey].Setup(projectRoot, execPath, argsList); err != nil {
					return err
				}
				fmt.Println(RenderSuccess("MCP setup complete for "+mcpPlatformRegistry[platformKey].Name+"."))
				return nil
			},
		}
		mcpSetupCmd.AddCommand(platformCmd)
	}

	mcpCmd.AddCommand(mcpSetupCmd)
	rootCmd.AddCommand(mcpCmd)
}
