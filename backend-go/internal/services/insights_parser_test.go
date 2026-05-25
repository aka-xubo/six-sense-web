package services

import "testing"

func TestParseInsightsRepairsUnescapedQuotes(t *testing.T) {
	raw := `{"summary":"腾讯上线AI应用生成平台"吐司"，用户可生成安卓App。","user_intent":"了解产品动态","key_points":["产品定位聚焦"灵感实现"与"共创"","支持一键打包APK","公测期间免费"],"value":"适合竞品分析。","next_action":"搜索"吐司AI"体验。","type":"新闻","keywords":["吐司","vibe coding","AI应用生成"]}`

	insights, ok := ParseInsights(raw)
	if !ok {
		t.Fatal("expected parser to repair Claude-style JSON")
	}
	if insights.Summary == "" || len(insights.Keywords) != 3 || len(insights.KeyPoints) != 3 {
		t.Fatalf("unexpected parsed insights: %#v", insights)
	}
}
