package services

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"
)

const maxContentLength = 50000
const maxTextLength = 30000

type PageFetcher struct {
	client *http.Client
}

func NewPageFetcher() *PageFetcher {
	return &PageFetcher{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (f *PageFetcher) FetchContent(ctx context.Context, rawURL string) (string, bool) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", false
	}
	if isGitHubURL(rawURL) {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	} else if isWechatURL(rawURL) {
		req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 MicroMessenger/8.0.47")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Referer", "https://mp.weixin.qq.com/")
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", false
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxContentLength*2))
	if err != nil {
		return "", false
	}
	content := string(body)
	if isWechatURL(rawURL) {
		if extracted := extractWechatArticleText(rawURL, content); extracted != "" {
			return limitContent(extracted), true
		}
	}
	if extracted := extractReadableText(rawURL, content); extracted != "" {
		return limitContent(extracted), true
	}
	return limitContent(content), true
}

func isWechatURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "mp.weixin.qq.com" || strings.HasSuffix(host, ".mp.weixin.qq.com")
}

func isGitHubURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "github.com" || host == "www.github.com"
}

func extractWechatArticleText(rawURL string, content string) string {
	body := extractElementByID(content, "div", "js_content")
	if body == "" && !strings.Contains(content, "js_content") {
		return ""
	}
	parts := []string{"URL: " + rawURL}
	if title := extractElementByID(content, "h1", "activity-name"); title != "" {
		parts = append(parts, "Title: "+title)
	}
	if author := extractElementByID(content, "a", "js_name"); author != "" {
		parts = append(parts, "Author: "+author)
	}
	if publishTime := extractElementByID(content, "em", "publish_time"); publishTime != "" {
		parts = append(parts, "Publish time: "+publishTime)
	}
	if body != "" {
		parts = append(parts, "Content:\n"+htmlToText(body))
	}
	return strings.Join(parts, "\n\n")
}

func extractReadableText(rawURL string, content string) string {
	parts := []string{"URL: " + rawURL}
	if title := firstRegex(content, `(?is)<title[^>]*>(.*?)</title>`); title != "" {
		parts = append(parts, "Title: "+htmlToText(title))
	}
	if desc := firstRegex(content, `(?is)<meta[^>]+(?:name|property)=["'](?:description|og:description)["'][^>]+content=["']([^"']*)["'][^>]*>`); desc != "" {
		parts = append(parts, "Description: "+decodeHTMLEntities(desc))
	}
	text := htmlToText(content)
	if text != "" {
		if len(text) > maxTextLength {
			text = text[:maxTextLength]
		}
		parts = append(parts, "Content:\n"+text)
	}
	return strings.Join(parts, "\n\n")
}

func extractElementByID(content string, tag string, id string) string {
	pattern := `(?is)<` + regexp.QuoteMeta(tag) + `\b(?=[^>]*\bid=["']` + regexp.QuoteMeta(id) + `["'])[^>]*>(.*?)</` + regexp.QuoteMeta(tag) + `>`
	return htmlToText(firstRegex(content, pattern))
}

func firstRegex(content string, pattern string) string {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(content)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func htmlToText(content string) string {
	replacerPatterns := []string{
		`(?is)<script[^>]*>.*?</script>`,
		`(?is)<style[^>]*>.*?</style>`,
		`(?is)<noscript[^>]*>.*?</noscript>`,
		`(?is)<svg[^>]*>.*?</svg>`,
	}
	text := content
	for _, pattern := range replacerPatterns {
		text = regexp.MustCompile(pattern).ReplaceAllString(text, " ")
	}
	text = regexp.MustCompile(`(?i)<br\s*/?>|</p>|</div>|</section>|</article>|</li>|</h[1-3]>`).ReplaceAllString(text, "\n")
	text = regexp.MustCompile(`(?s)<[^>]+>`).ReplaceAllString(text, " ")
	text = decodeHTMLEntities(text)
	lines := strings.Split(text, "\n")
	out := bytes.Buffer{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(collapseSpace(line))
		if len([]rune(trimmed)) > 1 {
			out.WriteString(trimmed)
			out.WriteByte('\n')
		}
	}
	return strings.TrimSpace(out.String())
}

func decodeHTMLEntities(text string) string {
	replacer := strings.NewReplacer("&nbsp;", " ", "&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", `"`, "&#39;", "'")
	return replacer.Replace(text)
}

func collapseSpace(text string) string {
	var b strings.Builder
	previousSpace := false
	for _, r := range text {
		if unicode.IsSpace(r) {
			if !previousSpace {
				b.WriteRune(' ')
				previousSpace = true
			}
			continue
		}
		b.WriteRune(r)
		previousSpace = false
	}
	return b.String()
}

func limitContent(content string) string {
	if len(content) > maxContentLength {
		return content[:maxContentLength]
	}
	return content
}
