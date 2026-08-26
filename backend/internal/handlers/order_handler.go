package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/middleware"
	"github.com/softivite/puxbay/internal/models"
	"github.com/softivite/puxbay/internal/services"
	"github.com/softivite/puxbay/internal/utils"
	"github.com/softivite/puxbay/internal/websocket"
	"gorm.io/gorm"
)

type OrderHandler struct {
	db         *gorm.DB
	hub        *websocket.Hub
	smsService *services.SMSService
	rootDomain string
}

func NewOrderHandler(db *gorm.DB, hub *websocket.Hub, sms *services.SMSService, rootDomain string) *OrderHandler {
	return &OrderHandler{db: db, hub: hub, smsService: sms, rootDomain: rootDomain}
}

func (h *OrderHandler) service(c *gin.Context) *services.OrderService {
	return services.NewOrderService(getDB(c, h.db), h.smsService)
}

func (h *OrderHandler) tenantDB(c *gin.Context) *gorm.DB {
	return getDB(c, h.db)
}

func (h *OrderHandler) List(c *gin.Context) {
	p := utils.GetPagination(c)

	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))
	status := c.Query("status")
	orderType := c.Query("order_type")
	productID := c.Query("product_id")
	customerID := c.Query("customer_id")
	search := c.Query("q")

	cashierID := ""
	if role, exists := c.Get(middleware.ContextKeyRole); exists {
		if roleStr, ok := role.(string); ok && strings.ToLower(roleStr) == "cashier" {
			if userID, exists := c.Get(middleware.ContextKeyUserID); exists {
				if idStr, ok := userID.(string); ok {
					cashierID = idStr
				} else if idUUID, ok := userID.(uuid.UUID); ok {
					cashierID = idUUID.String()
				} else {
					cashierID = fmt.Sprintf("%v", userID)
				}
			}
		}
	}

	params := services.OrderListParams{
		BranchID:   branchID,
		Status:     status,
		OrderType:  orderType,
		ProductID:  productID,
		CustomerID: customerID,
		CashierID:  cashierID,
		Search:     search,
		Limit:      p.Limit,
		Offset:     p.Offset,
	}

	orders, total, err := h.service(c).ListOrders(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  orders,
		"total": total,
		"page":  p.Page,
		"limit": p.Limit,
	})
}

func (h *OrderHandler) Summary(c *gin.Context) {
	branchID := middleware.ResolveBranchID(c, c.Query("branch_id"))

	cashierID := ""
	if role, exists := c.Get(middleware.ContextKeyRole); exists {
		if roleStr, ok := role.(string); ok && strings.ToLower(roleStr) == "cashier" {
			if userID, exists := c.Get(middleware.ContextKeyUserID); exists {
				if idStr, ok := userID.(string); ok {
					cashierID = idStr
				} else if idUUID, ok := userID.(uuid.UUID); ok {
					cashierID = idUUID.String()
				} else {
					cashierID = fmt.Sprintf("%v", userID)
				}
			}
		}
	}

	stats, err := h.service(c).GetOrderSummaryStats(branchID, cashierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get order summary"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *OrderHandler) Get(c *gin.Context) {
	id := c.Param("id")

	order, err := h.service(c).GetOrder(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

type OrderCreateRequest struct {
	BranchID      *uuid.UUID                   `json:"branch_id"`
	CustomerID    *uuid.UUID                   `json:"customer_id"`
	Subtotal      float64                      `json:"subtotal"`
	Tax           float64                      `json:"tax"`
	Discount      float64                      `json:"discount"`
	Total         float64                      `json:"total"`
	AmountPaid    float64                      `json:"amount_paid"`
	PaymentMethod string                       `json:"payment_method"`
	Payments      []services.OrderPaymentInput `json:"payments"`
	OrderType     string                       `json:"order_type"`
	Notes         string                       `json:"notes"`
	Items         []OrderItemParam             `json:"items" binding:"required,min=1"`
}

type OrderItemParam struct {
	ProductID uuid.UUID  `json:"product_id" binding:"required"`
	VariantID *uuid.UUID `json:"variant_id"`
	Quantity  float64    `json:"quantity" binding:"required,gt=0"`
	UnitPrice float64    `json:"unit_price" binding:"required,gte=0"`
	Discount  float64    `json:"discount"`
	Total     float64    `json:"total" binding:"required,gte=0"`
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req OrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	var items []services.OrderItemInput
	for _, item := range req.Items {
		items = append(items, services.OrderItemInput{
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Discount:  item.Discount,
			Total:     item.Total,
		})
	}

	payments := req.Payments
	if len(payments) == 0 {
		payments = []services.OrderPaymentInput{
			{Method: req.PaymentMethod, Amount: req.AmountPaid},
		}
	}

	var branchID *uuid.UUID
	if req.BranchID != nil {
		if bid := middleware.ResolveBranchID(c, req.BranchID.String()); bid != "" {
			if parsed, err := uuid.Parse(bid); err == nil {
				branchID = &parsed
			}
		}
	}

	if branchID == nil {
		if ctxBranchID, ok := middleware.GetBranchID(c); ok {
			branchID = ctxBranchID
		}
	}

	userID, _ := c.Get("user_id")
	var cashierID *uuid.UUID
	if uid, ok := userID.(uuid.UUID); ok {
		cashierID = &uid
	}

	input := services.OrderCreateInput{
		BranchID:   branchID,
		CustomerID: req.CustomerID,
		CashierID:  cashierID,
		Subtotal:   req.Subtotal,
		Tax:        req.Tax,
		Discount:   req.Discount,
		Total:      req.Total,
		AmountPaid: req.AmountPaid,
		Payments:   payments,
		OrderType:  req.OrderType,
		Notes:      req.Notes,
		Items:      items,
	}

	order, err := h.service(c).CreateOrder(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	if order.OrderType == "dine_in" || order.OrderType == "takeout" {
		if tenantIDRaw, exists := c.Get("tenant_id"); exists {
			h.hub.BroadcastMessage(tenantIDRaw.(string), []byte(`{"type": "KDS_UPDATE"}`))
		}
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) POS(c *gin.Context) {
	var req OrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	var cashierID *uuid.UUID
	if uid, ok := userID.(uuid.UUID); ok {
		cashierID = &uid
	}

	var items []services.OrderItemInput
	for _, item := range req.Items {
		items = append(items, services.OrderItemInput{
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Discount:  item.Discount,
			Total:     item.Total,
		})
	}

	payments := req.Payments
	if len(payments) == 0 {
		payments = []services.OrderPaymentInput{
			{Method: req.PaymentMethod, Amount: req.AmountPaid},
		}
	}

	var branchID *uuid.UUID
	if ctxBranchID, ok := middleware.GetBranchID(c); ok {
		branchID = ctxBranchID
	}

	input := services.OrderCreateInput{
		BranchID:   branchID,
		CustomerID: req.CustomerID,
		Subtotal:   req.Subtotal,
		Tax:        req.Tax,
		Discount:   req.Discount,
		Total:      req.Total,
		AmountPaid: req.AmountPaid,
		Payments:   payments,
		OrderType:  "in_store",
		Notes:      req.Notes,
		Items:      items,
	}

	order, err := h.service(c).ProcessPOSCheckout(input, cashierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "POS checkout failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) VoidOrder(c *gin.Context) {
	id := c.Param("id")
	if err := h.service(c).VoidOrder(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Audit log — fire-and-forget
	tenantRaw, _ := c.Get(middleware.ContextKeyTenantID)
	userRaw, _ := c.Get(middleware.ContextKeyUserID)
	tenantID, _ := tenantRaw.(uuid.UUID)
	userID, _ := userRaw.(uuid.UUID)
	
	metadata := map[string]interface{}{"order_id": id}
	if overrideMgrID, exists := c.Get("override_manager_id"); exists {
		metadata["override_manager_id"] = overrideMgrID
	}

	auditAsync(getDB(c, h.db), tenantID, userID, "VOID_ORDER", "orders", c.ClientIP(), metadata)

	c.JSON(http.StatusOK, gin.H{"status": "voided"})
}

func (h *OrderHandler) CompleteOrder(c *gin.Context) {
	id := c.Param("id")
	if err := h.service(c).CompleteOrder(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "completed"})
}

func (h *OrderHandler) GetReceipt(c *gin.Context) {
	id := c.Param("id")
	order, err := h.service(c).GetReceipt(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if strings.Contains(c.GetHeader("Accept"), "text/html") || c.Query("token") != "" {
		// --- Dynamic tenant info from the public schema ---
		db := h.tenantDB(c)
		var tenant models.Tenant
		db.Table("public.tenants").First(&tenant)

		// Build the public receipt URL using the subdomain and receipt_token
		subdomain := c.GetHeader("X-Tenant-Subdomain")
		if subdomain == "" {
			subdomain = c.Query("tenant")
		}
		if subdomain == "" && tenant.Subdomain != "" {
			subdomain = tenant.Subdomain
		}
		// Build base URL: in dev use localhost, in prod use puxbay.com
		host := c.Request.Host
		publicReceiptURL := fmt.Sprintf("http://%s/public/receipts/%s", host, order.ReceiptToken)
		if subdomain != "" && !strings.Contains(host, subdomain) {
			// Reconstruct proper subdomain URL for the QR
			protocol := "https"
			if strings.Contains(host, "localhost") {
				protocol = "http"
			}
			publicReceiptURL = fmt.Sprintf("%s://%s.%s/public/receipts/%s", protocol, subdomain, h.rootDomain, order.ReceiptToken)
		}

		// Prefer branch logo → tenant logo
		logoURL := ""
		if order.Branch != nil && order.Branch.Logo != nil && *order.Branch.Logo != "" {
			logoURL = *order.Branch.Logo
		} else if tenant.Logo != nil && *tenant.Logo != "" {
			logoURL = *tenant.Logo
		}

		// Company name: tenant name
		storeName := tenant.Name
		if storeName == "" {
			storeName = c.GetHeader("X-Tenant-Subdomain")
		}

		// Branch info
		branchName := "Main Branch"
		branchAddress := ""
		branchPhone := ""
		receiptHeader := ""
		receiptFooter := "Thank you for shopping with us!"

		if order.Branch != nil {
			if order.Branch.Name != "" {
				branchName = order.Branch.Name
			}
			if order.Branch.Address != nil {
				branchAddress = *order.Branch.Address
			}
			if order.Branch.Phone != nil {
				branchPhone = *order.Branch.Phone
			}
			if order.Branch.ReceiptHeader != nil && *order.Branch.ReceiptHeader != "" {
				receiptHeader = *order.Branch.ReceiptHeader
			}
			if order.Branch.ReceiptFooter != nil && *order.Branch.ReceiptFooter != "" {
				receiptFooter = *order.Branch.ReceiptFooter
			}
		}

		// Logo HTML block
		logoHTML := ""
		if logoURL != "" {
			logoHTML = fmt.Sprintf(`<img src="%s" alt="%s" style="max-width:100px;max-height:80px;object-fit:contain;margin:0 auto 8px;display:block;">`, logoURL, storeName)
		}

		// Branch sub-header lines
		branchInfoHTML := fmt.Sprintf(`<div>%s</div>`, branchName)
		if branchAddress != "" {
			branchInfoHTML += fmt.Sprintf(`<div style="font-size:11px;margin-top:3px;">%s</div>`, branchAddress)
		}
		if branchPhone != "" {
			branchInfoHTML += fmt.Sprintf(`<div style="font-size:11px;">Tel: %s</div>`, branchPhone)
		}
		if receiptHeader != "" {
			branchInfoHTML += fmt.Sprintf(`<div style="font-size:11px;margin-top:4px;font-style:italic;">%s</div>`, receiptHeader)
		}

		itemsHtml := ""
		for _, item := range order.Items {
			productName := "Unknown Product"
			if item.Product != nil {
				productName = item.Product.Name
			}
			discountRow := ""
			if item.Discount > 0 {
				discountRow = fmt.Sprintf(`<tr><td colspan="3" style="font-size:11px;color:#555;padding-left:8px;">Discount</td><td class="text-right" style="font-size:11px;color:#555;">-%.2f</td></tr>`, item.Discount)
			}
			itemsHtml += fmt.Sprintf(`
			<tr>
				<td>%s</td>
				<td class="text-right">%.1f</td>
				<td class="text-right">%.2f</td>
				<td class="text-right">%.2f</td>
			</tr>%s`, productName, item.Quantity, item.UnitPrice, item.Total, discountRow)
		}

		// Totals section
		totalsHTML := fmt.Sprintf(`<div class="total-row"><span>Subtotal</span><span>%.2f</span></div>`, order.Subtotal)
		if order.Tax > 0 {
			totalsHTML += fmt.Sprintf(`<div class="total-row" style="font-weight:normal;font-size:13px;"><span>Tax</span><span>%.2f</span></div>`, order.Tax)
		}
		if order.Discount > 0 {
			totalsHTML += fmt.Sprintf(`<div class="total-row" style="font-weight:normal;font-size:13px;"><span>Discount</span><span>-%.2f</span></div>`, order.Discount)
		}
		totalsHTML += fmt.Sprintf(`<div class="total-row grand-total"><span>TOTAL</span><span>%.2f</span></div>`, order.Total)
		totalsHTML += fmt.Sprintf(`<div class="total-row" style="font-weight:normal;"><span>Paid (%s)</span><span>%.2f</span></div>`, strings.ToUpper(order.PaymentMethod), order.AmountPaid)
		if change := order.AmountPaid - order.Total; change > 0 {
			totalsHTML += fmt.Sprintf(`<div class="total-row" style="font-weight:normal;"><span>Change</span><span>%.2f</span></div>`, change)
		}

		html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Receipt %s – %s</title>
	<style>
		body {
			font-family: 'Courier New', Courier, monospace;
			font-size: 14px;
			color: #000;
			background: #fff;
			padding: 20px;
			max-width: 350px;
			margin: 0 auto;
		}
		.header {
			text-align: center;
			margin-bottom: 16px;
			border-bottom: 1px dashed #000;
			padding-bottom: 12px;
		}
		.store-name {
			font-size: 20px;
			font-weight: bold;
			margin-bottom: 4px;
			text-transform: uppercase;
			letter-spacing: 1px;
		}
		.meta {
			margin-bottom: 15px;
			font-size: 12px;
		}
		.meta div {
			display: flex;
			justify-content: space-between;
		}
		table {
			width: 100%%;
			border-collapse: collapse;
			margin-bottom: 15px;
		}
		th {
			text-align: left;
			border-bottom: 1px solid #000;
			padding: 5px 0;
		}
		td {
			padding: 5px 0;
			vertical-align: top;
		}
		.text-right { text-align: right; }
		.totals {
			border-top: 1px dashed #000;
			padding-top: 10px;
			margin-top: 10px;
		}
		.total-row {
			display: flex;
			justify-content: space-between;
			font-weight: bold;
			font-size: 14px;
			margin-top: 4px;
		}
		.grand-total {
			font-size: 17px;
			border-top: 1px dashed #000;
			border-bottom: 1px dashed #000;
			padding: 6px 0;
			margin: 6px 0;
		}
		.footer {
			text-align: center;
			margin-top: 20px;
			font-size: 12px;
			border-top: 1px dashed #000;
			padding-top: 10px;
		}
		@media print {
			@page { margin: 0; size: auto; }
			body { padding: 5px; margin: 0; }
			.no-print { display: none; }
		}
		.btn-print {
			display: block;
			width: 100%%;
			padding: 10px;
			background: #000;
			color: #fff;
			text-align: center;
			text-decoration: none;
			margin-bottom: 20px;
			font-family: sans-serif;
			border-radius: 4px;
			box-sizing: border-box;
			cursor: pointer;
			border: none;
			font-size: 14px;
		}
	</style>
</head>
<body>
	<button onclick="window.print();" class="btn-print no-print">🖨 Print Receipt</button>

	<div class="header">
		%s
		<div class="store-name">%s</div>
		%s
	</div>

	<div class="meta">
		<div><span>Order #:</span><span>%s</span></div>
		<div><span>Date:</span><span>%s</span></div>
		<div><span>Payment:</span><span style="text-transform:capitalize;">%s</span></div>
	</div>

	<table>
		<thead>
			<tr>
				<th style="width:45%%;">Item</th>
				<th class="text-right">Qty</th>
				<th class="text-right">Price</th>
				<th class="text-right">Total</th>
			</tr>
		</thead>
		<tbody>
			%s
		</tbody>
	</table>

	<div class="totals">
		%s
	</div>

	<div class="footer">
		<p>%s</p>
		<p style="font-size:10px;color:#666;">Please keep this receipt for returns.</p>
		<div style="margin-top:15px;padding:12px;border:1px solid #eee;border-radius:6px;">
			<div style="font-weight:bold;margin-bottom:4px;">Join MyWallet</div>
			<div style="font-size:10px;color:#666;margin-bottom:8px;">Track points &amp; receipts on your phone</div>
			<img src="https://api.qrserver.com/v1/create-qr-code/?size=100x100&data=%s" alt="QR Code" style="width:100px;height:100px;margin:0 auto;display:block;cursor:pointer;" onclick="window.open('%s','_blank')">
		</div>
	</div>

	<script>
		window.onload = function() { setTimeout(function() { window.print(); }, 600); };
	</script>
</body>
</html>`,
			order.OrderNumber, storeName,
			logoHTML, storeName, branchInfoHTML,
			order.OrderNumber, order.CreatedAt.Format("02/01/2006 15:04"), order.PaymentMethod,
			itemsHtml,
			totalsHTML,
			receiptFooter,
			publicReceiptURL, publicReceiptURL)

		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetPublicReceipt(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing token"})
		return
	}

	// Orders live in per-tenant PostgreSQL schemas (e.g. "tenant_abc.orders").
	// We use LATERAL + set_config to search all tenant schemas in one query.
	type schemaHit struct {
		SchemaName string `gorm:"column:schema_name"`
		OrderID    string `gorm:"column:order_id"`
	}
	var hit schemaHit

	err := h.db.Raw(`
		SELECT t.schema_name, r.order_id
		FROM public.tenants t,
		LATERAL (
			SELECT o.id::text AS order_id
			FROM (SELECT set_config('search_path', t.schema_name, true)) _sc,
				orders o
			WHERE o.receipt_token = @token
			  AND o.deleted_at IS NULL
			LIMIT 1
		) r
		WHERE t.schema_name IS NOT NULL
		  AND t.schema_name != ''
		LIMIT 1
	`, map[string]interface{}{"token": token}).Scan(&hit).Error

	if err != nil || hit.SchemaName == "" || hit.OrderID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "receipt not found"})
		return
	}

	// Fetch the full order (with preloads) using the correct tenant schema.
	var order models.Order
	fetchErr := h.db.Transaction(func(tx *gorm.DB) error {
		tx.Exec(fmt.Sprintf("SET LOCAL search_path TO %s", hit.SchemaName))
		return tx.Where("receipt_token = ? AND deleted_at IS NULL", token).
			Preload("Items.Product").
			Preload("Branch").
			Preload("Customer").
			First(&order).Error
	})

	if fetchErr != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "receipt not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}
