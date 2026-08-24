package logger

import (
	"context"
	"encoding/json"
	"time"

	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/websocket"
	"go.opentelemetry.io/otel/sdk/trace"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DBExporter struct {
	db  *gorm.DB
	hub *websocket.Hub
}

func NewDBExporter(db *gorm.DB, hub *websocket.Hub) *DBExporter {
	return &DBExporter{
		db:  db,
		hub: hub,
	}
}

func (e *DBExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	var logs []models.TelemetryLog
	var events []map[string]interface{}

	for _, span := range spans {
		attrs := make(map[string]interface{})
		for _, attr := range span.Attributes() {
			attrs[string(attr.Key)] = attr.Value.AsInterface()
		}
		attrJSON, _ := json.Marshal(attrs)

		spanEvents := make([]map[string]interface{}, 0)
		for _, event := range span.Events() {
			evtAttrs := make(map[string]interface{})
			for _, attr := range event.Attributes {
				evtAttrs[string(attr.Key)] = attr.Value.AsInterface()
			}
			spanEvents = append(spanEvents, map[string]interface{}{
				"name":       event.Name,
				"time":       event.Time,
				"attributes": evtAttrs,
			})
		}
		eventsJSON, _ := json.Marshal(spanEvents)

		status := "Unset"
		if span.Status().Code == 1 {
			status = "Ok"
		} else if span.Status().Code == 2 {
			status = "Error"
		}

		duration := span.EndTime().Sub(span.StartTime()).Seconds() * 1000.0

		logEntry := models.TelemetryLog{
			TraceID:    span.SpanContext().TraceID().String(),
			SpanID:     span.SpanContext().SpanID().String(),
			Name:       span.Name(),
			StartTime:  span.StartTime(),
			EndTime:    span.EndTime(),
			Duration:   duration,
			Status:     status,
			Attributes: datatypes.JSON(attrJSON),
			Events:     datatypes.JSON(eventsJSON),
			CreatedAt:  time.Now(),
		}
		logs = append(logs, logEntry)

		events = append(events, map[string]interface{}{
			"type": "telemetry",
			"data": logEntry,
		})
	}

	if len(logs) > 0 {
		// Persist to database in the public schema
		e.db.Create(&logs)

		// Broadcast to WebSocket clients listening on the admin telemetry channel
		if e.hub != nil {
			for _, event := range events {
				eventBytes, _ := json.Marshal(event)
				e.hub.BroadcastAdmin(eventBytes)
			}
		}
	}

	return nil
}

func (e *DBExporter) Shutdown(ctx context.Context) error {
	return nil
}
