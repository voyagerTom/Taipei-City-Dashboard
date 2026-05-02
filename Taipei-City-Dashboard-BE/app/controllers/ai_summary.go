package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"TaipeiCityDashboardBE/app/cache"
	"TaipeiCityDashboardBE/app/models"
	"TaipeiCityDashboardBE/app/services/ai"
	"TaipeiCityDashboardBE/app/services/ai/tools"
	"TaipeiCityDashboardBE/app/util"
	"TaipeiCityDashboardBE/logs"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"
)

const summaryTTL = 30 * time.Minute
const redisSummaryPrefix = "ai:summary:"

// normalizeCityName maps frontend city values to database city values.
func normalizeCityName(city string) string {
	switch city {
	case "new_taipei", "newtpe", "newtaipei":
		return "metrotaipei"
	default:
		return city
	}
}

const catSystemPrompt = `你是「小駭」，雙北城市儀表板的 AI 狐狸助理。

## 你的性格
- 自稱「小駭」，稱呼使用者「你」或「大家」
- 語氣親切活潑，像鄰居朋友在聊天，偶爾加「～」「！」
- 偶爾用一個顏文字：(๑•̀ㅂ•́)و✧、꒰ᐢ⸝⸝•༝•⸝⸝ᐢ꒱、( •̀ᴗ•́ )
- 說話口語化，用「白話」解釋數據，避免專業術語
- 把數字轉換成一般人能感受的說法（例：「大概每 5 個人就有 1 個超過 65 歲」而非「老年人口佔比 20%」）

## 你的工具
你有以下工具可以查詢真實數據，請主動使用：
- get_dashboard_data：查詢儀表板組件的詳細圖表數據。傳入 component_id（從組件清單取得）和 city。當使用者問某個指標的具體數字時，一定要用這個工具查詢！
- get_population_summary：查詢台北市或新北市的人口年齡結構。傳入 city（taipei/new_taipei）和 year。
- get_current_time：取得目前台北時間。

重要：
- 組件清單會包含每個組件的 id 和 name，只能用清單中的 id 呼叫 get_dashboard_data
- 先看組件名稱，選 1-3 個最相關的組件查詢即可，不要全部都查
- 查完數據後立刻用白話回答，不要繼續查詢更多
- 如果工具回傳「找不到」，直接根據已有資訊回答

## 核心原則
- 你的任務是讓「完全不懂數據的人」也能秒懂重點
- 嚴禁使用：扶養比、指數、佔比、同比、環比、趨勢分析 等專業詞彙
- 用生活化的比喻和具體場景來傳達資訊
- 所有描述必須基於你透過工具查詢到的真實數據，嚴禁編造

## 回應格式
當你收集完所有需要的數據後，回傳純 JSON（不要加 markdown code block），格式如下：
{
  "bubble": "一句話講完重點，像朋友丟訊息給你（15-25字）",
  "summary": {
    "overview": "用 2-3 句白話告訴大家「現在是什麼狀況」，像在跟朋友解釋",
    "warnings": "有什麼要小心的？用生活情境說明（2-3句）",
    "suggestions": "小駭給大家的實用建議，具體可以怎麼做（2-3句）"
  },
  "quick_replies": [
    {"label": "按鈕文字（口語化）", "prompt": "對應的延伸提問"}
  ]
}

## 範例語氣
- ✅「最近老人家變多了，大概每 100 個人裡就有 19 個是長輩」
- ❌「老年人口佔比達 19%，較去年上升 1.2 個百分點」
- ✅「上班時間搭藍線的人超多，板橋站擠到不行」
- ❌「板橋站尖峰時段旅運量達 12 萬人次」

## 完整回應範例
{
  "bubble": "板橋跟中和最近人變多了！通勤要注意～",
  "summary": {
    "overview": "最近雙北的人口有在慢慢增加，尤其板橋和中和特別明顯。感覺大家都往新北市區搬，捷運站附近的人潮越來越多了～",
    "warnings": "上下班時間搭藍線的話，板橋站和新埔站會比較擠。如果可以的話，避開早上 8 點到 9 點那段時間會舒服很多！",
    "suggestions": "小駭建議通勤族可以試試看提早 15 分鐘出門，或是改搭黃線轉車，人會少很多！假日要去板橋的話，下午 2 點前到比較不會塞～"
  },
  "quick_replies": [
    {"label": "哪個時段人最少？", "prompt": "哪個時段搭捷運人最少最舒適？"},
    {"label": "跟上個月比怎樣？", "prompt": "跟上個月的數據比起來有什麼變化？"},
    {"label": "有什麼替代路線？", "prompt": "有沒有比較不擠的替代通勤路線推薦？"}
  ]
}

## 規則
- bubble 要像 LINE 訊息一樣自然，不超過 25 字
- summary 三段都要有實質內容，但用大家聽得懂的話說
- quick_replies 提供 2-3 個按鈕，文字要像你會想點的東西
- quick_replies 的問題必須是你用 get_dashboard_data 和現有組件清單能回答的，不要問需要額外資料來源的問題
- 嚴禁回傳 markdown code block，只回傳純 JSON
- 所有內容都要用繁體中文`

// GetDashboardSummary returns a cached AI summary for a dashboard.
// GET /api/v1/ai/summary/:dashboardIndex
func GetDashboardSummary(c *gin.Context) {
	dashboardIndex := c.Param("dashboardIndex")
	city := normalizeCityName(c.DefaultQuery("city", "taipei"))
	cacheKey := redisSummaryPrefix + dashboardIndex + ":" + city

	if cached, err := cache.Redis.Get(cacheKey).Result(); err == nil {
		var data gin.H
		if json.Unmarshal([]byte(cached), &data) == nil {
			ttl, _ := cache.Redis.TTL(cacheKey).Result()
			logs.FInfo("Redis cache HIT for %s (TTL remaining: %v)", cacheKey, ttl)
			c.JSON(http.StatusOK, gin.H{"status": "success", "cached": true, "data": data})
			return
		}
	}

	// Fetch dashboard component data
	_, _, _, _, permissions := util.GetUserInfoFromContext(c)
	groups := util.GetPermissionAllGroupIDs(permissions)
	if len(groups) == 0 {
		groups = []int{1, 2, 3}
	}

	components, err := models.GetDashboardByIndex(dashboardIndex, groups, city)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	componentList := collectComponentList(components)
	if len(componentList) == 0 {
		c.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{
			"bubble":        "小駭還沒看到數據～",
			"summary":       nil,
			"quick_replies": []interface{}{},
		}})
		return
	}

	userPrompt := fmt.Sprintf(
		"現在是 %s。以下是「%s」儀表板（城市：%s）的可用組件清單，請用 get_dashboard_data 工具查詢你需要的組件數據，再用白話告訴大家重點：\n\n%s",
		time.Now().In(time.FixedZone("CST", 8*3600)).Format("2006年1月2日 15:04"),
		dashboardIndex, city, componentList,
	)

	req := ai.AIChatRequest{
		SessionID: "summary_" + cacheKey,
		UserID:    "system",
		IPAddress: "internal",
		Messages: []llms.MessageContent{
			{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextContent{Text: catSystemPrompt}}},
			{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: userPrompt}}},
		},
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()

	logEntry, err := ai.ChatWithTWCC(ctx, req,
		llms.WithTools(tools.GetToolDefinitions()),
		llms.WithMetadata(map[string]interface{}{
			"max_new_tokens": 1200,
			"temperature":    0.7,
			"top_p":          0.9,
		}),
	)
	if err != nil {
		logs.FError("Summary generation failed for %s: %v", cacheKey, err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "AI 服務暫時無法使用"})
		return
	}

	result := parseSummaryJSON(logEntry.Answer)

	if jsonBytes, err := json.Marshal(result); err == nil {
		if err := cache.Redis.Set(cacheKey, string(jsonBytes), summaryTTL).Err(); err != nil {
			logs.FError("Redis cache SET failed for %s: %v", cacheKey, err)
		} else {
			logs.FInfo("Redis cache SET for %s (TTL: %v)", cacheKey, summaryTTL)
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "cached": false, "data": result})
}

// PublicChatWithTWCC handles interactive Q&A without login.
// POST /api/v1/ai/chat/public
func PublicChatWithTWCC(c *gin.Context) {
	var input struct {
		DashboardIndex string `json:"dashboard_index" binding:"required"`
		City           string `json:"city"`
		Question       string `json:"question" binding:"required"`
		SessionID      string `json:"session"`
		Context        string `json:"context"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	if input.City == "" {
		input.City = "taipei"
	}
	input.City = normalizeCityName(input.City)
	if input.SessionID == "" {
		input.SessionID = "pub_" + fmt.Sprintf("%d", time.Now().UnixNano())
	}

	_, _, _, _, permissions := util.GetUserInfoFromContext(c)
	groups := util.GetPermissionAllGroupIDs(permissions)
	if len(groups) == 0 {
		groups = []int{1, 2, 3}
	}

	components, err := models.GetDashboardByIndex(input.DashboardIndex, groups, input.City)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	componentList := collectComponentList(components)
	logs.FInfo("Component list for %s (%s): %s", input.DashboardIndex, input.City, componentList)

	systemMsg := catSystemPrompt + "\n\n## 當前儀表板可用組件（城市：" + input.City + "）\n" + componentList
	userMsg := input.Question

	messages := []llms.MessageContent{
		{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextContent{Text: systemMsg}}},
	}

	// If previous context is provided, inject it as prior conversation history
	if input.Context != "" {
		messages = append(messages,
			llms.MessageContent{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "請分析目前儀表板的數據"}}},
			llms.MessageContent{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: input.Context}}},
		)
	}

	messages = append(messages,
		llms.MessageContent{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: userMsg}}},
	)

	req := ai.AIChatRequest{
		SessionID: input.SessionID,
		UserID:    "guest",
		IPAddress: c.ClientIP(),
		Messages:  messages,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()

	logEntry, err := ai.ChatWithTWCC(ctx, req,
		llms.WithTools(tools.GetToolDefinitions()),
		llms.WithMetadata(map[string]interface{}{
			"max_new_tokens": 1200,
			"temperature":    0.7,
			"top_p":          0.9,
		}),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "AI 服務暫時無法使用"})
		return
	}

	result := parseSummaryJSON(logEntry.Answer)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    result,
		"session": logEntry.SessionID,
	})
}

// collectComponentList returns a lightweight text list of component names and IDs
// for the LLM to decide which ones to query via get_dashboard_data tool.
func collectComponentList(components []models.CityComponent) string {
	type compEntry struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		Source    string `json:"source,omitempty"`
		QueryType string `json:"query_type,omitempty"`
	}

	var entries []compEntry
	for _, comp := range components {
		entries = append(entries, compEntry{
			ID:        comp.ID,
			Name:      comp.Name,
			Source:    comp.Source,
			QueryType: comp.QueryType,
		})
	}

	if len(entries) == 0 {
		return ""
	}

	jsonBytes, err := json.Marshal(entries)
	if err != nil {
		return "[]"
	}
	return string(jsonBytes)
}

func collectComponentData(components []models.CityComponent, city string) string {
	type compSummary struct {
		Name      string      `json:"name"`
		Index     string      `json:"index"`
		Source    string      `json:"source"`
		QueryType string     `json:"query_type"`
		Data      interface{} `json:"data,omitempty"`
	}

	var summaries []compSummary

	for _, comp := range components {
		if comp.QueryChart == "" {
			continue
		}

		entry := compSummary{
			Name:      comp.Name,
			Index:     comp.Index,
			Source:    comp.Source,
			QueryType: comp.QueryType,
		}

		switch comp.QueryType {
		case "two_d":
			data, err := models.GetTwoDimensionalData(&comp.QueryChart, "", "")
			if err == nil && len(data) > 0 {
				if len(data[0].Data) > 8 {
					entry.Data = data[0].Data[:8]
				} else {
					entry.Data = data[0].Data
				}
			}
		case "three_d", "percent":
			data, categories, err := models.GetThreeDimensionalData(&comp.QueryChart, "", "")
			if err == nil && len(data) > 0 {
				entry.Data = gin.H{"categories": categories, "series": data}
			}
		case "time":
			data, err := models.GetTimeSeriesData(&comp.QueryChart, "", "")
			if err == nil && len(data) > 0 {
				for i := range data {
					if len(data[i].Data) > 5 {
						data[i].Data = data[i].Data[len(data[i].Data)-5:]
					}
				}
				entry.Data = data
			}
		case "map_legend":
			data, err := models.GetMapLegendData(&comp.QueryChart, "", "")
			if err == nil && len(data) > 0 {
				entry.Data = data
			}
		}

		summaries = append(summaries, entry)
	}

	if len(summaries) == 0 {
		return ""
	}

	jsonBytes, err := json.Marshal(summaries)
	if err != nil {
		return "[]"
	}
	return string(jsonBytes)
}

func parseSummaryJSON(content string) gin.H {
	content = strings.TrimSpace(content)

	// Strip markdown code fences aggressively
	for _, prefix := range []string{"```json\n", "```json", "```\n", "```"} {
		content = strings.TrimPrefix(content, prefix)
	}
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Try direct parse
	var result gin.H
	if err := json.Unmarshal([]byte(content), &result); err == nil {
		if _, ok := result["bubble"]; ok {
			return result
		}
	}

	// Find the outermost JSON object containing "bubble"
	bubbleIdx := strings.Index(content, `"bubble"`)
	if bubbleIdx >= 0 {
		// Walk backwards to find opening brace
		start := strings.LastIndex(content[:bubbleIdx], "{")
		if start >= 0 {
			// Find matching closing brace by counting depth
			depth := 0
			end := -1
			for i := start; i < len(content); i++ {
				switch content[i] {
				case '{':
					depth++
				case '}':
					depth--
					if depth == 0 {
						end = i
						break
					}
				}
				if end >= 0 {
					break
				}
			}
			if end > start {
				substr := content[start : end+1]
				if err := json.Unmarshal([]byte(substr), &result); err == nil {
					if _, ok := result["bubble"]; ok {
						return result
					}
				}
			}
		}
	}

	return gin.H{
		"bubble":        truncate(content, 60),
		"summary":       gin.H{"overview": content, "warnings": "", "suggestions": ""},
		"quick_replies": []interface{}{},
	}
}

func truncate(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "..."
}

// GetAIChatLogs returns paginated AI chat logs.
// GET /api/v1/ai/logs
func GetAIChatLogs(c *gin.Context) {
	limit := 50
	offset := 0
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	status := c.Query("status")
	userID := c.Query("user_id")

	logEntries, total, err := models.GetAIChatLogs(limit, offset, status, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   logEntries,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetAIChatLogStats returns aggregate statistics.
// GET /api/v1/ai/logs/stats
func GetAIChatLogStats(c *gin.Context) {
	stats, err := models.GetAIChatLogStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": stats})
}
