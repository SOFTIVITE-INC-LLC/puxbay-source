package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

func RegisterAuditCallbacks(db *gorm.DB) {
	db.Callback().Create().After("gorm:create").Register("audit:create", auditCallback("CREATE"))
	db.Callback().Update().Before("gorm:update").Register("audit:before_update", beforeUpdateCallback())
	db.Callback().Update().After("gorm:update").Register("audit:update", auditCallback("UPDATE"))
	db.Callback().Delete().After("gorm:delete").Register("audit:delete", auditCallback("DELETE"))
	db.Callback().Query().After("gorm:query").Register("audit:query", auditCallback("GET"))
}

func beforeUpdateCallback() func(*gorm.DB) {
	return func(db *gorm.DB) {
		if db.Statement == nil || db.Statement.Schema == nil || db.Statement.ReflectValue.IsValid() == false {
			return
		}

		table := db.Statement.Schema.Table
		if table == "audit_logs" || table == "api_request_logs" || table == "activity_logs" || table == "system_logs" || table == "cross_tenant_audit_logs" {
			return
		}

		primaryFields := db.Statement.Schema.PrimaryFields
		if len(primaryFields) == 0 {
			return
		}

		conditions := map[string]interface{}{}
		for _, field := range primaryFields {
			val, isZero := field.ValueOf(db.Statement.Context, db.Statement.ReflectValue)
			if !isZero {
				conditions[field.DBName] = val
			}
		}

		if len(conditions) == 0 {
			return
		}

		var oldData map[string]interface{}
		if err := db.Session(&gorm.Session{NewDB: true, SkipHooks: true}).Table(table).Where(conditions).Take(&oldData).Error; err == nil {
			if oldJSON, err := json.Marshal(oldData); err == nil {
				db.InstanceSet("audit:old_values", string(oldJSON))
			}
		}
	}
}

func auditCallback(action string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		if db.Error != nil || db.Statement == nil || db.Statement.Schema == nil {
			return
		}

		table := db.Statement.Schema.Table
		// Prevent infinite loops and ignore noisy tables
		if table == "audit_logs" || table == "api_request_logs" || table == "activity_logs" || table == "system_logs" || table == "cross_tenant_audit_logs" {
			return
		}

		var tenantIDPtr *uuid.UUID
		var userIDPtr *uuid.UUID
		var ipPtr *string

		if db.Statement.Context != nil {
			if tID, ok := db.Statement.Context.Value("tenant_id").(uuid.UUID); ok {
				tenantIDPtr = &tID
			}
			if uID, ok := db.Statement.Context.Value("user_id").(uuid.UUID); ok {
				userIDPtr = &uID
			}
		}

		sql := db.Statement.SQL.String()
		vars := db.Statement.Vars

		changesData := map[string]interface{}{
			"sql":  sql,
			"vars": fmt.Sprintf("%v", vars),
		}

		if action == "UPDATE" {
			if old, ok := db.InstanceGet("audit:old_values"); ok && old != "" {
				changesData["old_values"] = json.RawMessage(old.(string))
			}
			if destJSON, err := json.Marshal(db.Statement.Dest); err == nil {
				changesData["new_values"] = json.RawMessage(destJSON)
			}
		} else if action != "GET" {
			// Try to serialize the dest object
			if destJSON, err := json.Marshal(db.Statement.Dest); err == nil {
				changesData["data"] = json.RawMessage(destJSON)
			}
		}

		changesBytes, _ := json.Marshal(changesData)

		entry := models.AuditLog{
			Action:    action,
			ModelName: table,
			TenantID:  tenantIDPtr,
			UserID:    userIDPtr,
			IPAddress: ipPtr,
		}
		_ = entry.Changes.UnmarshalJSON(changesBytes)

		// Extract schema name to use in the background routine
		schemaName := "public"
		if db.Statement.Context != nil {
			if schema, ok := db.Statement.Context.Value(TenantSchemaKey).(string); ok && schema != "" {
				schemaName = schema
			}
		}

		// Async insert to not block the response
		go func(auditEntry models.AuditLog, schema string) {
			if DB != nil {
				bgCtx := context.Background()
				DB.WithContext(bgCtx).Transaction(func(tx *gorm.DB) error {
					if schema != "public" && schema != "" {
						// Need to use raw SQL here to avoid triggering callbacks infinitely
						tx.Exec(fmt.Sprintf("SET LOCAL search_path TO %q", schema))
					} else {
						tx.Exec("SET LOCAL search_path TO public")
					}
					// Use SkipHooks so we don't trigger the create hook again
					return tx.Session(&gorm.Session{SkipHooks: true}).Create(&auditEntry).Error
				})
			}
		}(entry, schemaName)
	}
}
