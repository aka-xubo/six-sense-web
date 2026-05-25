package services

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"six-sense-web/backend/internal/models"
)

type AgentDetector struct {
	extraSearchPaths []string
	mu               sync.RWMutex
	resolvedCommands map[string]string
}

func NewAgentDetector() *AgentDetector {
	home, _ := os.UserHomeDir()
	return &AgentDetector{
		extraSearchPaths: []string{
			filepath.Join(home, ".local", "bin"),
			filepath.Join(home, ".npm-global", "bin"),
			"/opt/homebrew/bin",
			"/usr/local/bin",
		},
		resolvedCommands: map[string]string{},
	}
}

type supportedAgent struct {
	Name        string
	DisplayName string
	Command     string
	VersionFlag string
	EnvPath     string
}

var supportedAgents = []supportedAgent{
	{Name: "claude", DisplayName: "Claude Code", Command: "claude", VersionFlag: "--version", EnvPath: "CLAUDE_PATH"},
	{Name: "codex", DisplayName: "OpenAI Codex", Command: "codex", VersionFlag: "--version", EnvPath: "CODEX_PATH"},
}

func (d *AgentDetector) Detect(ctx context.Context) []models.AgentInfo {
	agents := make([]models.AgentInfo, 0, len(supportedAgents))
	for _, agent := range supportedAgents {
		command := d.resolveAgentCommand(agent)
		version := stringPointer("available")
		if command == "" {
			version = nil
		}
		available := command != ""
		agents = append(agents, models.AgentInfo{Name: agent.Name, DisplayName: agent.DisplayName, Version: version, Available: available})
	}
	return agents
}

func (d *AgentDetector) Resolve(ctx context.Context, command string) (string, error) {
	for _, agent := range supportedAgents {
		if agent.Command == command {
			resolved := d.resolveAgentCommand(agent)
			if resolved == "" {
				return "", os.ErrNotExist
			}
			return resolved, nil
		}
	}
	candidates := d.commandCandidates(command)
	if len(candidates) == 0 {
		return "", os.ErrNotExist
	}
	return candidates[0], nil
}

func (d *AgentDetector) findCommand(command string) string {
	for _, agent := range supportedAgents {
		if agent.Command == command {
			return d.resolveAgentCommand(agent)
		}
	}
	candidates := d.commandCandidates(command)
	if len(candidates) == 0 {
		return ""
	}
	return candidates[0]
}

func (d *AgentDetector) resolveAgentCommand(agent supportedAgent) string {
	d.mu.RLock()
	cached := d.resolvedCommands[agent.Command]
	d.mu.RUnlock()
	if cached != "" && isExecutableFile(cached) {
		return cached
	}

	resolved := d.findCommandForAgent(agent)
	if resolved != "" {
		d.mu.Lock()
		d.resolvedCommands[agent.Command] = resolved
		d.mu.Unlock()
	}
	return resolved
}

func (d *AgentDetector) findCommandForAgent(agent supportedAgent) string {
	if configured := strings.TrimSpace(os.Getenv(agent.EnvPath)); configured != "" {
		if isExecutableFile(configured) {
			return configured
		}
	}
	if agent.Command == "codex" || agent.Command == "claude" {
		home, _ := os.UserHomeDir()
		preferred := filepath.Join(home, ".npm-global", "bin", agent.Command)
		if isExecutableFile(preferred) {
			return preferred
		}
	}
	candidates := d.commandCandidates(agent.Command)
	if len(candidates) == 0 {
		return ""
	}
	return candidates[0]
}

func stringPointer(value string) *string {
	return &value
}

func (d *AgentDetector) commandCandidates(command string) []string {
	seen := map[string]bool{}
	candidates := []string{}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		addCandidate(&candidates, seen, filepath.Join(dir, command))
	}
	for _, dir := range d.extraSearchPaths {
		addCandidate(&candidates, seen, filepath.Join(dir, command))
	}
	if path, err := exec.LookPath(command); err == nil {
		addCandidate(&candidates, seen, path)
	}
	return candidates
}

func addCandidate(candidates *[]string, seen map[string]bool, candidate string) {
	if candidate == "" || seen[candidate] {
		return
	}
	if isExecutableFile(candidate) {
		*candidates = append(*candidates, candidate)
		seen[candidate] = true
	}
}

func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	return info.Mode()&0o111 != 0
}

type AgentAdapter struct {
	detector *AgentDetector
	template string
}

func NewAgentAdapter(detector *AgentDetector, templatePath string) (*AgentAdapter, error) {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, err
	}
	return &AgentAdapter{detector: detector, template: string(content)}, nil
}

func (a *AgentAdapter) AnalyzePage(ctx context.Context, agentName string, rawURL string, content string, onDelta func(string)) (string, error) {
	prompt := a.buildPrompt(rawURL, content)
	var cmd *exec.Cmd
	switch agentName {
	case "claude":
		command, err := a.detector.Resolve(ctx, "claude")
		if err != nil {
			return "", errors.New("Claude Code CLI 不可用，请确认 claude 命令已正确安装")
		}
		log.Printf("Using claude command: %s", command)
		cmd = exec.CommandContext(ctx, command, "--model", "sonnet")
	case "codex":
		command, err := a.detector.Resolve(ctx, "codex")
		if err != nil {
			return "", errors.New("Codex CLI 不可用，请确认 codex 命令已正确安装")
		}
		log.Printf("Using codex command: %s", command)
		cmd = exec.CommandContext(ctx, command, "exec", "--color", "never", "-")
	default:
		return "", errors.New("Unsupported agent: " + agentName)
	}

	cmd.Stdin = strings.NewReader(prompt)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return "", err
	}
	var full strings.Builder
	scanner := bufio.NewScanner(stdout)
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, 1024*1024)
	for scanner.Scan() {
		text := scanner.Text() + "\n"
		full.WriteString(text)
		onDelta(text)
	}
	if err := scanner.Err(); err != nil {
		_ = cmd.Wait()
		return full.String(), err
	}
	if err := cmd.Wait(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return full.String(), errors.New(message)
	}
	return full.String(), nil
}

func (a *AgentAdapter) buildPrompt(rawURL string, content string) string {
	if len(content) > 30000 {
		content = content[:30000]
	}
	return `请分析以下网页内容，严格按照 insights_prompt_template.md 中的规范生成结构化的 insights。

网页 URL: ` + rawURL + `
网页内容:
` + content + `

要求：
1. 输出纯 JSON 格式，无其他文字
2. 必须包含 summary、user_intent、key_points、value、next_action、type、keywords
3. key_points 必须是 3 到 5 条具体要点
4. keywords 必须恰好 3 个，无重复

输出格式：
{
  "summary": "...",
  "user_intent": "...",
  "key_points": ["...", "...", "..."],
  "value": "...",
  "next_action": "...",
  "type": "...",
  "keywords": ["...", "...", "..."]
}

模板规范：
` + a.template
}
