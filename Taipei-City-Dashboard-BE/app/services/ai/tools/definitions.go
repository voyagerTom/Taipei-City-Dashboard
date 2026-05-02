package tools

import "github.com/tmc/langchaingo/llms"

// GetToolDefinitions returns the OpenAPI-style tool schemas for all registered tools.
// These are passed to the LLM so it knows what tools are available and how to call them.
func GetToolDefinitions() []llms.Tool {
	return []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_dashboard_data",
				Description: "查詢指定儀表板組件的圖表數據。需要提供組件 ID 和城市。可以用來查看特定指標的詳細數據，例如交通流量、人口統計、環境監測等。",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"component_id": map[string]interface{}{
							"type":        "integer",
							"description": "儀表板組件的 ID（從組件清單中取得）",
						},
						"city": map[string]interface{}{
							"type":        "string",
							"description": "城市代碼：taipei（台北市）或 metrotaipei（新北市/雙北）",
							"enum":        []string{"taipei", "metrotaipei"},
						},
					},
					"required": []string{"component_id"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_current_time",
				Description: "取得目前的台北時間（Asia/Taipei 時區）。當需要知道現在的日期或時間來回答問題時使用。",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_population_summary",
				Description: "查詢台北市或新北市的人口年齡結構統計（幼年、青壯年、老年人口數）。需要提供城市和年份。",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"city": map[string]interface{}{
							"type":        "string",
							"description": "城市代碼：taipei（台北市）或 metrotaipei（新北市/雙北）",
							"enum":        []string{"taipei", "metrotaipei"},
						},
						"year": map[string]interface{}{
							"type":        "integer",
							"description": "要查詢的年份，例如 2024",
						},
					},
					"required": []string{"city", "year"},
				},
			},
		},
	}
}
