package tools

import (
	"TaipeiCityDashboardBE/app/models"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ToolFunc defines the signature for a tool function
type ToolFunc func(ctx context.Context, args string) (string, error)

var registry = make(map[string]ToolFunc)

func init() {
	Register("get_current_time", GetCurrentTime)
	Register("get_population_summary", GetPopulationSummary)
	Register("get_dashboard_data", GetDashboardData)
}

// Register adds a tool to the registry
func Register(name string, fn ToolFunc) {
	registry[name] = fn
}

// Execute calls a registered tool with the given arguments
func Execute(ctx context.Context, name string, args string) (string, error) {
	fn, ok := registry[name]
	if !ok {
		return "", fmt.Errorf("tool %s not found", name)
	}
	return fn(ctx, args)
}

// PopulationArgs defines the arguments for the get_population_summary tool
type PopulationArgs struct {
	City string `json:"city"`
	Year int    `json:"year"`
}

// GetPopulationSummary queries the population age distribution from the dashboard database
func GetPopulationSummary(ctx context.Context, args string) (string, error) {
	var params PopulationArgs
	if err := parseArgs(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %v", err)
	}

	// Default to Taipei if not specified or unrecognized
	tableName := "population_age_distribution_tpe"
	cityName := "台北市"
	if params.City == "new_taipei" {
		tableName = "population_age_distribution_new_tpe"
		cityName = "新北市"
	}

	// Define result structure based on database schema
	var result struct {
		Year      int `gorm:"column:year"`
		Young     int `gorm:"column:young_population"`
		Working   int `gorm:"column:working_age_population"`
		Elderly   int `gorm:"column:elderly_population"`
		DataTime  time.Time `gorm:"column:data_time"`
	}

	// Query the dashboard database
	err := models.DBDashboard.Table(tableName).
		Where("year = ?", params.Year).
		Order("data_time DESC"). // Get the latest record for that year
		First(&result).Error

	if err != nil {
		return "", fmt.Errorf("找不到 %s %d 年的人口統計資料: %v", cityName, params.Year, err)
	}

	// Format the response for the LLM
	return fmt.Sprintf(
		"【%d年 %s 人口結構概況】\n- 幼年人口 (0-14歲)：%d 人\n- 青壯年人口 (15-64歲)：%d 人\n- 老年人口 (65歲以上)：%d 人\n- 總人口： %d 人\n- 數據更新時間：%s",
		result.Year, cityName, result.Young, result.Working, result.Elderly,
		result.Young+result.Working+result.Elderly,
		result.DataTime.Format("2006-01-02"),
	), nil
}

// GetCurrentTime is a demo tool that returns the current Taipei time
func GetCurrentTime(ctx context.Context, args string) (string, error) {
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		// Fallback to UTC if timezone data is missing
		return time.Now().Format(time.RFC3339), nil
	}
	return time.Now().In(loc).Format("2006-01-02 15:04:05"), nil
}

// GetDashboardData fetches chart data for a specific component by ID and city.
func GetDashboardData(ctx context.Context, args string) (string, error) {
	var params struct {
		ComponentID int    `json:"component_id"`
		City        string `json:"city"`
	}
	if err := parseArgs(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %v", err)
	}
	if params.ComponentID == 0 {
		return "", fmt.Errorf("component_id is required")
	}
	if params.City == "" {
		params.City = "taipei"
	}

	queryType, queryString, err := models.GetComponentChartDataQuery(params.ComponentID, params.City)
	if err != nil || queryString == "" {
		return fmt.Sprintf("找不到組件 ID %d 的數據", params.ComponentID), nil
	}

	var result interface{}
	switch queryType {
	case "two_d":
		data, err := models.GetTwoDimensionalData(&queryString, "", "")
		if err != nil {
			return fmt.Sprintf("查詢失敗: %v", err), nil
		}
		result = data
	case "three_d", "percent":
		data, categories, err := models.GetThreeDimensionalData(&queryString, "", "")
		if err != nil {
			return fmt.Sprintf("查詢失敗: %v", err), nil
		}
		result = map[string]interface{}{"categories": categories, "series": data}
	case "time":
		data, err := models.GetTimeSeriesData(&queryString, "", "")
		if err != nil {
			return fmt.Sprintf("查詢失敗: %v", err), nil
		}
		for i := range data {
			if len(data[i].Data) > 10 {
				data[i].Data = data[i].Data[len(data[i].Data)-10:]
			}
		}
		result = data
	case "map_legend":
		data, err := models.GetMapLegendData(&queryString, "", "")
		if err != nil {
			return fmt.Sprintf("查詢失敗: %v", err), nil
		}
		result = data
	default:
		return fmt.Sprintf("不支援的查詢類型: %s", queryType), nil
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return "資料序列化失敗", nil
	}
	return string(jsonBytes), nil
}

func parseArgs(args string, v interface{}) error {
	return json.Unmarshal([]byte(args), v)
}
