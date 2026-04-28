package runtimeinstall

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallClaudeMergesExistingSettingsAndStatus(t *testing.T) {
	home := t.TempDir()
	claudeDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("mkdir claude: %v", err)
	}
	settings := `{
  "theme": "dark",
  "hooks": {
    "Stop": [{"hooks": [{"type": "command", "command": "/tmp/existing.sh"}]}]
  }
}`
	if err := os.WriteFile(filepath.Join(claudeDir, claudeSettings), []byte(settings), 0644); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	opts := Options{
		HomeDir:        home,
		ExecutablePath: "/usr/local/bin/knowns",
		LookPath: func(name string) (string, error) {
			if name == "claude" {
				return "/usr/local/bin/claude", nil
			}
			return "", os.ErrNotExist
		},
	}
	if err := Install("claude-code", opts); err != nil {
		t.Fatalf("install claude: %v", err)
	}

	body, err := os.ReadFile(filepath.Join(claudeDir, claudeSettings))
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, `"theme": "dark"`) {
		t.Fatalf("expected existing setting preserved, got:\n%s", text)
	}
	if !strings.Contains(text, `"SessionStart"`) {
		t.Fatalf("expected SessionStart hook added, got:\n%s", text)
	}
	status, err := StatusFor("claude-code", opts)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !status.Installed || status.State != StateInstalled {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestInstallClaudeWindowsQuotesExecutableForBashHooks(t *testing.T) {
	home := t.TempDir()
	claudeDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("mkdir claude: %v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, claudeSettings), []byte(`{"hooks":{}}`), 0644); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	exePath := `C:\Users\Admin\.knowns\bin\knowns.exe`
	opts := Options{
		HomeDir:        home,
		ExecutablePath: exePath,
		GOOS:           "windows",
		LookPath: func(name string) (string, error) {
			if name == "claude" {
				return `C:\Users\Admin\AppData\Local\Programs\Claude\claude.exe`, nil
			}
			return "", os.ErrNotExist
		},
	}
	if err := Install("claude-code", opts); err != nil {
		t.Fatalf("install claude windows: %v", err)
	}

	body, err := os.ReadFile(filepath.Join(claudeDir, claudeSettings))
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	text := string(body)
	expectedExe := strings.ReplaceAll(exePath, `\`, "/")
	if !strings.Contains(text, `"command": "\"`+expectedExe+`\"`) {
		t.Fatalf("expected slash-normalized quoted executable path, got:\n%s", text)
	}
	if !strings.Contains(text, `\"runtime-memory\" \"hook\" \"--runtime\" \"claude-code\" \"--event\" \"session-start\"`) {
		t.Fatalf("expected quoted runtime-memory hook args, got:\n%s", text)
	}
}

func TestInstallClaudeWindowsWritesExactSessionStartHookJSON(t *testing.T) {
	home := t.TempDir()
	claudeDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("mkdir claude: %v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, claudeSettings), []byte(`{"hooks":{}}`), 0644); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	exePath := `C:\Users\Admin\.knowns\bin\knowns.exe`
	opts := Options{
		HomeDir:        home,
		ExecutablePath: exePath,
		GOOS:           "windows",
		LookPath: func(name string) (string, error) {
			if name == "claude" {
				return `C:\Users\Admin\AppData\Local\Programs\Claude\claude.exe`, nil
			}
			return "", os.ErrNotExist
		},
	}
	if err := Install("claude-code", opts); err != nil {
		t.Fatalf("install claude windows: %v", err)
	}

	body, err := os.ReadFile(filepath.Join(claudeDir, claudeSettings))
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(body, &settings); err != nil {
		t.Fatalf("unmarshal settings: %v\n%s", err, string(body))
	}
	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatalf("expected hooks object, got: %#v", settings["hooks"])
	}
	sessionStart, ok := hooks["SessionStart"].([]any)
	if !ok || len(sessionStart) != 1 {
		t.Fatalf("expected one SessionStart hook group, got: %#v", hooks["SessionStart"])
	}
	group, ok := sessionStart[0].(map[string]any)
	if !ok {
		t.Fatalf("expected SessionStart group object, got: %#v", sessionStart[0])
	}
	groupHooks, ok := group["hooks"].([]any)
	if !ok || len(groupHooks) != 1 {
		t.Fatalf("expected one hook entry, got: %#v", group["hooks"])
	}
	hook, ok := groupHooks[0].(map[string]any)
	if !ok {
		t.Fatalf("expected hook object, got: %#v", groupHooks[0])
	}

	expectedCommand := `"` + strings.ReplaceAll(exePath, `\`, "/") + `" "runtime-memory" "hook" "--runtime" "claude-code" "--event" "session-start"`
	if got := hook["command"]; got != expectedCommand {
		t.Fatalf("unexpected command\nwant: %q\n got: %q", expectedCommand, got)
	}
	if got := hook["type"]; got != "command" {
		t.Fatalf("unexpected hook type: %#v", got)
	}
	if got := hook["statusMessage"]; got != managedStatus {
		t.Fatalf("unexpected statusMessage: %#v", got)
	}
}

func TestStaleHooksGetReplacedByFreshHooks(t *testing.T) {
	staleExe := "/var/folders/j7/wbwyp1sn0rv_34drdg4m34fw0000gn/T/go-build3766182753/b461/cli.test"
	freshExe := "/Users/datran/go/bin/knowns"

	// Simulate a settings.json with a stale go-build hook from a previous test run
	// and an unrelated existing hook.
	existing := []any{
		map[string]any{
			"statusMessage": managedStatus,
			"hooks": []any{
				map[string]any{
					"type":    "command",
					"command": staleExe + ` "runtime-memory" "hook" "--runtime" "claude-code" "--event" "session-start"`,
				},
			},
		},
		map[string]any{
			"hooks": []any{
				map[string]any{
					"type":    "command",
					"command": "/tmp/existing-hook.sh",
				},
			},
		},
	}

	freshPath := freshExe + ` "runtime-memory" "hook" "--runtime" "claude-code" "--event" "session-start"`
	result := ensureCommandHookGroup(existing, freshPath, managedStatus)

	// Debug: print all groups using JSON marshal (reliable for interface{} values)
	for i, g := range result {
		gg := g.(map[string]any)
		b, _ := json.MarshalIndent(gg, "", "  ")
		t.Logf("result[%d]: %s", i, string(b))
	}

	// Should end up with 2 groups: fresh Knowns hook + unrelated existing hook.
	if len(result) != 2 {
		t.Fatalf("expected 2 groups after stale cleanup, got %d: %+v", len(result), result)
	}
	// Identify by fresh command path presence (reliable when statusMessage is unreliable).
	var knownsGroup, otherGroup map[string]any
	for _, g := range result {
		gg := g.(map[string]any)
		hooks := gg["hooks"].([]any)
		if len(hooks) > 0 {
			cmd := hooks[0].(map[string]any)["command"].(string)
			if strings.Contains(cmd, freshExe) && strings.Contains(cmd, "runtime-memory") {
				knownsGroup = gg
			} else {
				otherGroup = gg
			}
		}
	}
	if knownsGroup == nil {
		t.Fatal("expected a Knowns-managed group in result")
	}
	if otherGroup == nil {
		t.Fatal("expected the unrelated existing hook to be preserved")
	}
	hooks0 := knownsGroup["hooks"].([]any)
	if len(hooks0) != 1 {
		t.Fatalf("expected 1 hook in Knowns group, got %d", len(hooks0))
	}
	if cmd, _ := hooks0[0].(map[string]any)["command"].(string); cmd != freshPath {
		t.Fatalf("expected fresh path, got %q", cmd)
	}
	hooks1 := otherGroup["hooks"].([]any)
	if cmd, _ := hooks1[0].(map[string]any)["command"].(string); cmd != "/tmp/existing-hook.sh" {
		t.Fatalf("expected unrelated existing hook preserved, got %q", cmd)
	}
}

func TestInstallCodexEnablesFeatureAndUninstallRemovesManagedHookOnly(t *testing.T) {
	home := t.TempDir()
	codexDir := filepath.Join(home, ".codex")
	if err := os.MkdirAll(codexDir, 0755); err != nil {
		t.Fatalf("mkdir codex: %v", err)
	}
	if err := os.WriteFile(filepath.Join(codexDir, codexConfig), []byte("model = \"gpt-5.4\"\n"), 0644); err != nil {
		t.Fatalf("seed config: %v", err)
	}
	hooks := `{
  "hooks": {
    "SessionStart": [
      {"hooks": [{"type": "command", "command": "/tmp/existing-hook.sh"}]}
    ]
  }
}`
	if err := os.WriteFile(filepath.Join(codexDir, codexHooksFile), []byte(hooks), 0644); err != nil {
		t.Fatalf("seed hooks: %v", err)
	}

	opts := Options{
		HomeDir:        home,
		ExecutablePath: "/usr/local/bin/knowns",
		LookPath: func(name string) (string, error) {
			if name == "codex" {
				return "/usr/local/bin/codex", nil
			}
			return "", os.ErrNotExist
		},
	}
	if err := Install("codex", opts); err != nil {
		t.Fatalf("install codex: %v", err)
	}
	configBody, err := os.ReadFile(filepath.Join(codexDir, codexConfig))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(configBody), "codex_hooks = true") {
		t.Fatalf("expected codex_hooks flag enabled, got:\n%s", string(configBody))
	}
	if err := Uninstall("codex", opts); err != nil {
		t.Fatalf("uninstall codex: %v", err)
	}
	hooksBody, err := os.ReadFile(filepath.Join(codexDir, codexHooksFile))
	if err != nil {
		t.Fatalf("read hooks after uninstall: %v", err)
	}
	text := string(hooksBody)
	if !strings.Contains(text, "/tmp/existing-hook.sh") {
		t.Fatalf("expected unrelated hook preserved, got:\n%s", text)
	}
	if strings.Contains(text, `"UserPromptSubmit"`) {
		t.Fatalf("expected legacy UserPromptSubmit hook removed, got:\n%s", text)
	}
	if strings.Contains(text, managedStatus) {
		t.Fatalf("expected managed hook removed, got:\n%s", text)
	}
}

func TestInstallOpenCodeCreatesPluginAndStatusInstalled(t *testing.T) {
	home := t.TempDir()
	opts := Options{
		HomeDir:        home,
		ExecutablePath: "/usr/local/bin/knowns",
		LookPath:       func(string) (string, error) { return "", os.ErrNotExist },
	}
	if err := Install("opencode", opts); err != nil {
		t.Fatalf("install opencode: %v", err)
	}
	pluginPath := filepath.Join(home, ".config", "opencode", "plugins", pluginFileName)
	plugin, err := os.ReadFile(pluginPath)
	if err != nil {
		t.Fatalf("read plugin: %v", err)
	}
	if !strings.Contains(string(plugin), "runtime-memory") {
		t.Fatalf("expected runtime-memory hook command in plugin, got:\n%s", string(plugin))
	}
	status, err := StatusFor("opencode", opts)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !status.Installed || status.State != StateInstalled {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestInstallKiroCreatesWorkspaceIDEHook(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	if err := os.MkdirAll(filepath.Join(project, ".kiro"), 0755); err != nil {
		t.Fatalf("mkdir project .kiro: %v", err)
	}
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(project); err != nil {
		t.Fatalf("chdir project: %v", err)
	}

	opts := Options{
		HomeDir:        home,
		ExecutablePath: "/new/bin/knowns",
		LookPath: func(name string) (string, error) {
			if name == "kiro" {
				return "/usr/local/bin/kiro", nil
			}
			return "", os.ErrNotExist
		},
	}
	if err := Install("kiro", opts); err != nil {
		t.Fatalf("install kiro: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(project, ".kiro", "hooks", managedName+".kiro.hook"))
	if err != nil {
		t.Fatalf("read IDE hook: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, `"type": "promptSubmit"`) {
		t.Fatalf("expected promptSubmit trigger, got:\n%s", text)
	}
	if !strings.Contains(text, `"type": "runCommand"`) {
		t.Fatalf("expected runCommand action, got:\n%s", text)
	}
	if !strings.Contains(text, "runtime-memory hook --runtime kiro --event promptsubmit") {
		t.Fatalf("expected runtime-memory hook command, got:\n%s", text)
	}
}

func TestRuntimePickerLabelIncludesAvailabilityForSupportedRuntimes(t *testing.T) {
	opts := Options{
		HomeDir:        t.TempDir(),
		ExecutablePath: "/usr/local/bin/knowns",
		LookPath: func(name string) (string, error) {
			if name == "codex" {
				return "/usr/local/bin/codex", nil
			}
			return "", os.ErrNotExist
		},
	}
	label := RuntimePickerLabel("codex", opts)
	if !strings.Contains(label, "Available") {
		t.Fatalf("expected availability in label, got %q", label)
	}
	desc := RuntimePickerDescription("opencode", opts)
	if !strings.Contains(desc, "installs OpenCode") && !strings.Contains(strings.ToLower(desc), "installs opencode") {
		t.Fatalf("expected auto-install description, got %q", desc)
	}
}
