package services

import (
	"encoding/json"
	"strings"
)

type ParsedInsights struct {
	Summary    string   `json:"summary"`
	Type       string   `json:"type"`
	Keywords   []string `json:"keywords"`
	UserIntent string   `json:"user_intent"`
	KeyPoints  []string `json:"key_points"`
	Value      string   `json:"value"`
	NextAction string   `json:"next_action"`
}

func ParseInsights(text string) (*ParsedInsights, bool) {
	for _, candidate := range extractJSONCandidates(text) {
		if insights, ok := parseInsightCandidate(candidate); ok {
			return insights, true
		}
		if insights, ok := parseInsightCandidate(repairJSONLike(candidate)); ok {
			return insights, true
		}
	}
	trimmed := strings.TrimSpace(text)
	if insights, ok := parseInsightCandidate(trimmed); ok {
		return insights, true
	}
	return parseInsightCandidate(repairJSONLike(trimmed))
}

func parseInsightCandidate(candidate string) (*ParsedInsights, bool) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(candidate), &raw); err != nil {
		return nil, false
	}
	summary, ok := raw["summary"].(string)
	if !ok || summary == "" {
		return nil, false
	}
	kind, ok := raw["type"].(string)
	if !ok || kind == "" {
		return nil, false
	}
	keywords := stringSlice(raw["keywords"])
	if len(keywords) != 3 {
		return nil, false
	}

	return &ParsedInsights{
		Summary:    summary,
		Type:       kind,
		Keywords:   keywords,
		UserIntent: stringValue(raw["user_intent"]),
		KeyPoints:  stringSlice(raw["key_points"]),
		Value:      stringValue(raw["value"]),
		NextAction: stringValue(raw["next_action"]),
	}, true
}

func extractJSONCandidates(text string) []string {
	candidates := []string{}
	start := -1
	depth := 0
	inString := false
	escape := false
	for index, r := range text {
		if inString {
			if escape {
				escape = false
			} else if r == '\\' {
				escape = true
			} else if r == '"' {
				inString = false
			}
			continue
		}
		if r == '"' {
			inString = true
			continue
		}
		if r == '{' {
			if depth == 0 {
				start = index
			}
			depth++
			continue
		}
		if r == '}' && depth > 0 {
			depth--
			if depth == 0 && start >= 0 {
				candidate := text[start : index+1]
				if strings.Contains(candidate, `"summary"`) && strings.Contains(candidate, `"keywords"`) {
					candidates = append(candidates, candidate)
				}
				start = -1
			}
		}
	}
	return candidates
}

func repairJSONLike(text string) string {
	var builder strings.Builder
	inString := false
	escape := false
	for index, r := range text {
		if !inString {
			builder.WriteRune(r)
			if r == '"' {
				inString = true
			}
			continue
		}

		if escape {
			builder.WriteRune(r)
			escape = false
			continue
		}
		if r == '\\' {
			builder.WriteRune(r)
			escape = true
			continue
		}
		if r == '"' {
			next := nextNonSpace(text[index+1:])
			if next == ':' || next == ',' || next == '}' || next == ']' || next == 0 {
				builder.WriteRune(r)
				inString = false
			} else {
				builder.WriteRune('\\')
				builder.WriteRune(r)
			}
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func nextNonSpace(text string) rune {
	for _, r := range text {
		if !strings.ContainsRune(" \n\r\t", r) {
			return r
		}
	}
	return 0
}

func stringSlice(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return []string{}
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok {
			out = append(out, text)
		}
	}
	return out
}

func stringValue(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}
