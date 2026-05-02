package models

import (
	"time"
)

// AIChatLog defines the model for AI chat logs as specified in the system design.
type AIChatLog struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID    string    `gorm:"type:varchar(100);not null;index:idx_ai_chatlog_session" json:"session"`
	UserID       string    `gorm:"type:varchar(100);index:idx_ai_chatlog_user" json:"user_id"`
	Provider     string    `gorm:"type:varchar(50);not null;default:'twcc'" json:"provider"`
	Model        string    `gorm:"type:varchar(100)" json:"model"`
	Question     string    `gorm:"type:text;not null" json:"question"`
	Answer       string    `gorm:"type:text" json:"answer"`
	ToolUsed     bool      `gorm:"default:false" json:"tool_used"`
	Tools        string    `gorm:"type:jsonb" json:"tools"` // Stored as JSONB in DB
	InputTokens  int       `gorm:"default:0" json:"input_tokens"`
	OutputTokens int       `gorm:"default:0" json:"output_tokens"`
	TotalTokens  int       `gorm:"default:0" json:"total_tokens"`
	LatencyMS    int       `json:"latency_ms"`
	Status       string    `gorm:"type:varchar(30);not null;default:'success'" json:"status"`
	ErrorCode    string    `gorm:"type:varchar(100)" json:"error_code"`
	ErrorMessage string    `gorm:"type:text" json:"error_message"`
	IPAddress    string    `gorm:"type:varchar(45);not null" json:"ip_address"`
	CreatedAt    time.Time `gorm:"not null;default:now()" json:"created_at"`
}

// TableName overrides the table name used by AIChatLog to `ai_chatlog`
func (AIChatLog) TableName() string {
	return "ai_chatlog"
}

// CreateAIChatLog inserts a new AI chat log into the database.
func CreateAIChatLog(log *AIChatLog) error {
	return DBManager.Create(log).Error
}

// GetAIChatLogs retrieves AI chat logs with pagination and optional filters.
func GetAIChatLogs(limit, offset int, status, userID string) (logs []AIChatLog, total int64, err error) {
	query := DBManager.Model(&AIChatLog{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	err = query.Count(&total).Error
	if err != nil {
		return
	}

	err = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error
	return
}

// GetAIChatLogStats returns aggregate stats for AI usage.
func GetAIChatLogStats() (stats map[string]interface{}, err error) {
	stats = make(map[string]interface{})

	var totalCount int64
	DBManager.Model(&AIChatLog{}).Count(&totalCount)
	stats["total_requests"] = totalCount

	var successCount int64
	DBManager.Model(&AIChatLog{}).Where("status = ?", "success").Count(&successCount)
	stats["success_count"] = successCount

	var errorCount int64
	DBManager.Model(&AIChatLog{}).Where("status = ?", "error").Count(&errorCount)
	stats["error_count"] = errorCount

	var avgLatency float64
	DBManager.Model(&AIChatLog{}).Where("status = ?", "success").Select("COALESCE(AVG(latency_ms), 0)").Scan(&avgLatency)
	stats["avg_latency_ms"] = int(avgLatency)

	var totalTokens int64
	DBManager.Model(&AIChatLog{}).Select("COALESCE(SUM(total_tokens), 0)").Scan(&totalTokens)
	stats["total_tokens"] = totalTokens

	return
}
