package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"google.golang.org/api/option"
	"gorm.io/gorm"

	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/models"
)

type CopilotService struct {
	db     *gorm.DB
	client *genai.Client
	model  *genai.GenerativeModel
}

func NewCopilotService(cfg *config.Config, db *gorm.DB) *CopilotService {
	if cfg.Google.GeminiAPIKey == "" {
		log.Println("⚠️ GEMINI_API_KEY is not set. Copilot will be disabled.")
		return &CopilotService{db: db}
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.Google.GeminiAPIKey))
	if err != nil {
		log.Printf("⚠️ Failed to initialize Gemini client: %v", err)
		return &CopilotService{db: db}
	}

	model := client.GenerativeModel("gemini-1.5-flash")
	model.SystemInstruction = genai.NewUserContent(genai.Text("You are Puxbay Copilot, an AI assistant for an ERP/POS system. You help business owners manage their store. You have access to tools to query their database. Keep answers concise, helpful, and formatted in Markdown. If you don't know the answer, say so."))

	// Define tools
	model.Tools = []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        "get_inventory_status",
					Description: "Returns the current inventory levels for products, especially those that are low in stock.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"critical_only": {
								Type:        genai.TypeBoolean,
								Description: "If true, only return products that are at or below their reorder level.",
							},
						},
						Required: []string{"critical_only"},
					},
				},
				{
					Name:        "get_sales_summary",
					Description: "Returns total sales revenue and order counts for a given period.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"period": {
								Type:        genai.TypeString,
								Description: "The time period (e.g. 'today', 'this_week', 'this_month')",
							},
						},
						Required: []string{"period"},
					},
				},
			},
		},
	}

	return &CopilotService{
		db:     db,
		client: client,
		model:  model,
	}
}

// ChatMessage represents a single message in the conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (s *CopilotService) Chat(ctx context.Context, tenantID uuid.UUID, history []ChatMessage) (string, error) {
	if s.client == nil {
		return "I'm sorry, my AI brain is currently offline.", nil
	}

	session := s.model.StartChat()

	// Pre-fill history
	for i, msg := range history {
		if i == len(history)-1 {
			break // The last message is the current prompt
		}
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		session.History = append(session.History, &genai.Content{
			Role:  role,
			Parts: []genai.Part{genai.Text(msg.Content)},
		})
	}

	lastMsg := history[len(history)-1].Content

	resp, err := session.SendMessage(ctx, genai.Text(lastMsg))
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	return s.handleResponse(ctx, session, tenantID, resp)
}

func (s *CopilotService) handleResponse(ctx context.Context, session *genai.ChatSession, tenantID uuid.UUID, resp *genai.GenerateContentResponse) (string, error) {
	if len(resp.Candidates) == 0 {
		return "I couldn't generate a response.", nil
	}

	part := resp.Candidates[0].Content.Parts[0]

	// Check if it's a function call
	if funcCall, ok := part.(genai.FunctionCall); ok {
		// Execute the tool
		var toolResult map[string]interface{}

		switch funcCall.Name {
		case "get_inventory_status":
			criticalOnly := false
			if val, exists := funcCall.Args["critical_only"]; exists {
				if b, ok := val.(bool); ok {
					criticalOnly = b
				}
			}
			toolResult = s.toolGetInventoryStatus(tenantID, criticalOnly)

		case "get_sales_summary":
			period := "today"
			if val, exists := funcCall.Args["period"]; exists {
				if str, ok := val.(string); ok {
					period = str
				}
			}
			toolResult = s.toolGetSalesSummary(tenantID, period)

		default:
			toolResult = map[string]interface{}{"error": "Unknown function"}
		}

		// Send result back to model
		resultJSON, _ := json.Marshal(toolResult)
		followupResp, err := session.SendMessage(ctx, genai.FunctionResponse{
			Name: funcCall.Name,
			Response: map[string]any{
				"result": string(resultJSON),
			},
		})
		if err != nil {
			return "", fmt.Errorf("failed to send function result: %w", err)
		}

		return s.handleResponse(ctx, session, tenantID, followupResp)
	}

	// It's a text response
	if txt, ok := part.(genai.Text); ok {
		return string(txt), nil
	}

	return "I don't know how to respond to that.", nil
}

// -- Tool Implementations --

func (s *CopilotService) toolGetInventoryStatus(tenantID uuid.UUID, criticalOnly bool) map[string]interface{} {
	var products []models.Product

	query := s.db.Where("tenant_id = ? AND is_active = true", tenantID).
		Select("name", "current_stock", "reorder_level")

	if criticalOnly {
		query = query.Where("current_stock <= reorder_level")
	}

	query.Order("current_stock ASC").Limit(15).Find(&products)

	var count int64
	s.db.Model(&models.Product{}).Where("tenant_id = ? AND is_active = true", tenantID).Count(&count)

	result := make([]map[string]interface{}, len(products))
	for i, p := range products {
		result[i] = map[string]interface{}{
			"name":          p.Name,
			"stock":         p.CurrentStock,
			"reorder_level": p.ReorderLevel,
			"status":        "critical",
		}
		if p.CurrentStock > p.ReorderLevel {
			result[i]["status"] = "healthy"
		}
	}

	return map[string]interface{}{
		"total_active_products": count,
		"items_returned":        len(result),
		"items":                 result,
	}
}

func (s *CopilotService) toolGetSalesSummary(tenantID uuid.UUID, period string) map[string]interface{} {
	var start time.Time
	now := time.Now()

	switch period {
	case "today":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "this_week":
		start = now.AddDate(0, 0, -int(now.Weekday()))
	case "this_month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	case "this_year":
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	default: // default to today
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		period = "today"
	}

	var result struct {
		TotalRevenue float64
		OrderCount   int64
	}

	s.db.Model(&models.Order{}).
		Where("tenant_id = ? AND created_at >= ? AND status = 'completed'", tenantID, start).
		Select("COALESCE(SUM(total), 0) as total_revenue, COUNT(*) as order_count").
		Scan(&result)

	return map[string]interface{}{
		"period":      period,
		"revenue":     result.TotalRevenue,
		"order_count": result.OrderCount,
	}
}
