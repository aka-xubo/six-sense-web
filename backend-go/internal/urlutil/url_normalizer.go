package urlutil

import (
	"net/url"
	"sort"
	"strings"
	"time"
)

var wechatCoreParams = []string{"__biz", "mid", "idx", "sn"}

var wechatVolatileParams = map[string]bool{
	"poc_token":   true,
	"pass_ticket": true,
	"chksm":       true,
	"scene":       true,
	"ascene":      true,
	"devicetype":  true,
	"version":     true,
	"lang":        true,
	"nettype":     true,
	"sessionid":   true,
	"subscene":    true,
	"clicktime":   true,
	"enterid":     true,
}

func NormalizePageURL(rawURL string, lastVisitTime string) (string, string) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		visitDate := dateKey(lastVisitTime)
		return rawURL, "url:" + rawURL + ":" + visitDate
	}

	visitDate := dateKey(lastVisitTime)
	host := strings.ToLower(parsed.Hostname())
	if host != "mp.weixin.qq.com" {
		if canonical, repoKey, ok := normalizeGitHubRepositoryURL(parsed); ok {
			return canonical, "github-repo:" + repoKey + ":" + visitDate
		}
		canonical := normalizeGenericURL(parsed)
		return canonical, "url:" + canonical + ":" + visitDate
	}

	articleURL := unwrapWechatTargetURL(rawURL)
	canonicalURL, articleKey := canonicalizeWechatArticleURL(articleURL)
	if articleKey != "" {
		return canonicalURL, "wechat:" + articleKey + ":" + visitDate
	}
	return canonicalURL, "wechat-url:" + canonicalURL + ":" + visitDate
}

func unwrapWechatTargetURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Path != "/mp/wappoc_appmsgcaptcha" {
		return rawURL
	}
	target := parsed.Query().Get("target_url")
	if target == "" {
		return rawURL
	}
	return target
}

func canonicalizeWechatArticleURL(rawURL string) (string, string) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL, ""
	}
	query := parsed.Query()
	values := make([]string, 0, len(wechatCoreParams))
	for _, key := range wechatCoreParams {
		value := query.Get(key)
		if value == "" {
			return canonicalizeWechatFallback(parsed), ""
		}
		values = append(values, value)
	}

	canonicalQuery := url.Values{}
	for index, key := range wechatCoreParams {
		canonicalQuery.Set(key, values[index])
	}
	parsed.RawQuery = canonicalQuery.Encode()
	parsed.Fragment = ""
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}
	return parsed.String(), strings.Join(values, ":")
}

func canonicalizeWechatFallback(parsed *url.URL) string {
	query := parsed.Query()
	stable := url.Values{}
	keys := make([]string, 0, len(query))
	for key := range query {
		if !wechatVolatileParams[key] && key != "target_url" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		for _, value := range query[key] {
			stable.Add(key, value)
		}
	}
	parsed.RawQuery = stable.Encode()
	parsed.Fragment = ""
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}
	return parsed.String()
}

func normalizeGenericURL(parsed *url.URL) string {
	parsed.Fragment = ""
	return parsed.String()
}

func normalizeGitHubRepositoryURL(parsed *url.URL) (string, string, bool) {
	host := strings.ToLower(parsed.Hostname())
	if host != "github.com" && host != "www.github.com" {
		return "", "", false
	}
	segments := pathSegments(parsed.Path)
	if len(segments) < 2 || isGitHubReservedPath(segments[0]) {
		return "", "", false
	}
	owner := segments[0]
	repo := strings.TrimSuffix(segments[1], ".git")
	if owner == "" || repo == "" {
		return "", "", false
	}
	scheme := parsed.Scheme
	if scheme == "" {
		scheme = "https"
	}
	canonical := (&url.URL{
		Scheme: scheme,
		Host:   "github.com",
		Path:   "/" + owner + "/" + repo,
	}).String()
	return canonical, strings.ToLower(owner) + "/" + strings.ToLower(repo), true
}

func pathSegments(path string) []string {
	rawSegments := strings.Split(path, "/")
	segments := make([]string, 0, len(rawSegments))
	for _, segment := range rawSegments {
		if segment != "" {
			segments = append(segments, segment)
		}
	}
	return segments
}

func isGitHubReservedPath(segment string) bool {
	switch strings.ToLower(segment) {
	case "about", "account", "apps", "blog", "codespaces", "collections", "contact", "customer-stories", "enterprise", "events", "explore", "features", "github", "issues", "login", "marketplace", "mobile", "new", "notifications", "orgs", "organizations", "pricing", "pulls", "readme", "resources", "search", "security", "settings", "signup", "site", "solutions", "sponsors", "team", "topics", "trending":
		return true
	default:
		return false
	}
}

func dateKey(value string) string {
	if len(value) >= 10 {
		return value[:10]
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.Format("2006-01-02")
	}
	return time.Now().Format("2006-01-02")
}
