package services

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/google/uuid"
	"github.com/softivite/puxbay/internal/config"
	"github.com/softivite/puxbay/internal/models"
	"gorm.io/gorm"
)

type ReportService struct {
	db  *gorm.DB
	cfg config.SMTPConfig
}

func NewReportService(db *gorm.DB, cfg config.SMTPConfig) *ReportService {
	return &ReportService{db: db, cfg: cfg}
}

// DailyZReportData holds all aggregated metrics for a single business day's Z-Report.
type DailyZReportData struct {
	TenantName       string
	BranchName       string
	Date             string
	TotalSales       float64
	GrossProfit      float64
	GrossMarginPct   float64
	TotalOrders      int64
	TotalDiscounts   float64
	TotalTax         float64
	CashTotal        float64
	MoMoTotal        float64
	CardTotal        float64
	CreditTotal      float64
	GiftCardTotal    float64
	TopSellingItems  []ReportItemSummary
	CashDrawerFloats float64
	CashDrops        float64
}

type ReportItemSummary struct {
	ProductName string
	Quantity    float64
	Revenue     float64
}

// PLReportData holds aggregated revenue, COGS, expenses, and net profit metrics.
type PLReportData struct {
	TenantName    string
	PeriodLabel   string
	StartDate     string
	EndDate       string
	TotalRevenue  float64
	TotalCOGS     float64
	GrossProfit   float64
	GrossMargin   float64
	TotalExpenses float64
	NetProfit     float64
	NetMargin     float64
	DailyBreakdown []DailyPLRow
}

type DailyPLRow struct {
	Date        string
	Revenue     float64
	COGS        float64
	GrossProfit float64
}

// GenerateDailyZReport compiles the Z-Report for a given tenant on a specific date.
func (s *ReportService) GenerateDailyZReport(tenantID uuid.UUID, targetDate time.Time) (*DailyZReportData, error) {
	startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var tenant models.Tenant
	_ = s.db.Where("id = ?", tenantID).First(&tenant)
	tenantName := tenant.Name
	if tenantName == "" {
		tenantName = "Puxbay Merchant"
	}

	var orders []models.Order
	if err := s.db.Where("tenant_id = ? AND created_at >= ? AND created_at < ? AND status != 'cancelled'", tenantID, startOfDay, endOfDay).
		Preload("Items").Find(&orders).Error; err != nil {
		return nil, err
	}

	data := &DailyZReportData{
		TenantName:  tenantName,
		BranchName:  "All Branches",
		Date:        targetDate.Format("02 Jan 2006"),
		TotalOrders: int64(len(orders)),
	}

	var totalCost float64
	itemMap := make(map[string]*ReportItemSummary)

	for _, o := range orders {
		data.TotalSales += o.Total
		data.TotalDiscounts += o.Discount
		data.TotalTax += o.Tax

		switch o.PaymentMethod {
		case "cash":
			data.CashTotal += o.Total
		case "mobile", "momo":
			data.MoMoTotal += o.Total
		case "card":
			data.CardTotal += o.Total
		case "credit":
			data.CreditTotal += o.Total
		case "gift_card":
			data.GiftCardTotal += o.Total
		default:
			data.CashTotal += o.Total
		}

		for _, item := range o.Items {
			totalCost += item.CostPrice * item.Quantity
			pName := "Item"
			if item.ProductID != uuid.Nil {
				pName = fmt.Sprintf("Product %s", item.ProductID.String()[:6])
			}
			if existing, ok := itemMap[pName]; ok {
				existing.Quantity += item.Quantity
				existing.Revenue += item.Total
			} else {
				itemMap[pName] = &ReportItemSummary{
					ProductName: pName,
					Quantity:    item.Quantity,
					Revenue:     item.Total,
				}
			}
		}
	}

	data.GrossProfit = data.TotalSales - totalCost
	if data.TotalSales > 0 {
		data.GrossMarginPct = (data.GrossProfit / data.TotalSales) * 100
	}

	for _, v := range itemMap {
		data.TopSellingItems = append(data.TopSellingItems, *v)
		if len(data.TopSellingItems) >= 5 {
			break
		}
	}

	return data, nil
}

// GeneratePLReport compiles the Profit & Loss summary for a custom date window.
func (s *ReportService) GeneratePLReport(tenantID uuid.UUID, periodLabel string, start, end time.Time) (*PLReportData, error) {
	var tenant models.Tenant
	_ = s.db.Where("id = ?", tenantID).First(&tenant)
	tenantName := tenant.Name
	if tenantName == "" {
		tenantName = "Puxbay Merchant"
	}

	var orders []models.Order
	if err := s.db.Where("tenant_id = ? AND created_at >= ? AND created_at < ? AND status != 'cancelled'", tenantID, start, end).
		Preload("Items").Find(&orders).Error; err != nil {
		return nil, err
	}

	// Fetch operating expenses in period
	var expenses []models.Expense
	s.db.Where("tenant_id = ? AND expense_date >= ? AND expense_date < ?", tenantID, start, end).Find(&expenses)

	var totalExpenses float64
	for _, e := range expenses {
		totalExpenses += e.Amount
	}

	data := &PLReportData{
		TenantName:    tenantName,
		PeriodLabel:   periodLabel,
		StartDate:     start.Format("02 Jan 2006"),
		EndDate:       end.Format("02 Jan 2006"),
		TotalExpenses: totalExpenses,
	}

	dailyMap := make(map[string]*DailyPLRow)

	for _, o := range orders {
		data.TotalRevenue += o.Total
		dayKey := o.CreatedAt.Format("02 Jan")

		if _, exists := dailyMap[dayKey]; !exists {
			dailyMap[dayKey] = &DailyPLRow{Date: dayKey}
		}
		dailyMap[dayKey].Revenue += o.Total

		for _, item := range o.Items {
			cogs := item.CostPrice * item.Quantity
			data.TotalCOGS += cogs
			dailyMap[dayKey].COGS += cogs
		}
	}

	data.GrossProfit = data.TotalRevenue - data.TotalCOGS
	if data.TotalRevenue > 0 {
		data.GrossMargin = (data.GrossProfit / data.TotalRevenue) * 100
	}

	data.NetProfit = data.GrossProfit - data.TotalExpenses
	if data.TotalRevenue > 0 {
		data.NetMargin = (data.NetProfit / data.TotalRevenue) * 100
	}

	for _, row := range dailyMap {
		row.GrossProfit = row.Revenue - row.COGS
		data.DailyBreakdown = append(data.DailyBreakdown, *row)
	}

	return data, nil
}

// RenderDailyZReportHTML renders a styled, email-ready HTML Z-Report.
func (s *ReportService) RenderDailyZReportHTML(data *DailyZReportData) (string, error) {
	tmplStr := `
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; color: #1e293b; margin: 0; padding: 24px; }
  .container { max-width: 600px; margin: 0 auto; background: #ffffff; border-radius: 16px; overflow: hidden; box-shadow: 0 4px 20px rgba(0,0,0,0.06); border: 1px solid #e2e8f0; }
  .header { background: #041E42; color: #ffffff; padding: 28px; text-align: center; }
  .header h1 { margin: 0; font-size: 22px; font-weight: 800; letter-spacing: 0.5px; text-transform: uppercase; }
  .header p { margin: 6px 0 0; font-size: 13px; color: #F3A41D; font-weight: 600; }
  .content { padding: 24px; }
  .grid { display: table; width: 100%; margin-bottom: 20px; }
  .col { display: table-cell; width: 50%; padding: 12px; background: #f8fafc; border-radius: 12px; border: 1px solid #e2e8f0; }
  .col:first-child { margin-right: 8px; }
  .metric-label { font-size: 11px; text-transform: uppercase; font-weight: 700; color: #64748b; letter-spacing: 0.5px; }
  .metric-val { font-size: 22px; font-weight: 900; color: #041E42; margin-top: 4px; }
  .metric-sub { font-size: 12px; font-weight: 600; color: #10b981; }
  .section-title { font-size: 14px; font-weight: 800; text-transform: uppercase; color: #041E42; border-bottom: 2px solid #F3A41D; padding-bottom: 6px; margin: 24px 0 12px; }
  table { width: 100%; border-collapse: collapse; font-size: 13px; }
  th { text-align: left; padding: 8px; background: #f1f5f9; color: #475569; font-weight: 700; font-size: 11px; text-transform: uppercase; }
  td { padding: 10px 8px; border-bottom: 1px solid #f1f5f9; }
  .text-right { text-align: right; }
  .badge { background: #e0f2fe; color: #0369a1; padding: 3px 8px; border-radius: 6px; font-weight: 700; font-size: 11px; }
  .footer { background: #f8fafc; padding: 16px; text-align: center; font-size: 11px; color: #94a3b8; border-top: 1px solid #e2e8f0; }
</style>
</head>
<body>
<div class="container">
  <div class="header">
    <h1>{{.TenantName}}</h1>
    <p>DAILY END-OF-DAY Z-REPORT &bull; {{.Date}}</p>
  </div>
  <div class="content">
    <div style="margin-bottom: 16px; background: #f0fdf4; border: 1px solid #bbf7d0; border-radius: 12px; padding: 16px; text-align: center;">
      <span class="metric-label" style="color: #166534;">Total Net Revenue</span>
      <div style="font-size: 32px; font-weight: 900; color: #15803d; margin: 4px 0;">GHS {{printf "%.2f" .TotalSales}}</div>
      <div style="font-size: 12px; color: #166534; font-weight: 600;">{{.TotalOrders}} Completed Orders &bull; Gross Margin: {{printf "%.1f" .GrossMarginPct}}%</div>
    </div>

    <div class="section-title">Payment Methods Breakdown</div>
    <table>
      <tr><th>Payment Method</th><th class="text-right">Amount (GHS)</th></tr>
      <tr><td>💵 Cash</td><td class="text-right font-bold">{{printf "%.2f" .CashTotal}}</td></tr>
      <tr><td>📱 Mobile Money (MTN / Telecel)</td><td class="text-right font-bold">{{printf "%.2f" .MoMoTotal}}</td></tr>
      <tr><td>💳 Card / Paystack</td><td class="text-right font-bold">{{printf "%.2f" .CardTotal}}</td></tr>
      <tr><td>🤝 Store Credit / BNPL</td><td class="text-right font-bold">{{printf "%.2f" .CreditTotal}}</td></tr>
      <tr><td>🎁 Gift Cards</td><td class="text-right font-bold">{{printf "%.2f" .GiftCardTotal}}</td></tr>
      <tr style="background:#f8fafc; font-weight: bold;"><td>Total Collected</td><td class="text-right">{{printf "%.2f" .TotalSales}}</td></tr>
    </table>

    <div class="section-title">Audit & Profit Summary</div>
    <table>
      <tr><td>Estimated Gross Profit</td><td class="text-right" style="color:#10b981; font-weight:bold;">GHS {{printf "%.2f" .GrossProfit}}</td></tr>
      <tr><td>Discounts Granted</td><td class="text-right" style="color:#ef4444;">-GHS {{printf "%.2f" .TotalDiscounts}}</td></tr>
      <tr><td>Taxes Accrued</td><td class="text-right">GHS {{printf "%.2f" .TotalTax}}</td></tr>
    </table>
  </div>
  <div class="footer">
    Sent automatically by Puxbay Commerce System &bull; <a href="https://puxbay.com" style="color: #F3A41D; text-decoration: none; font-weight: bold;">puxbay.com</a>
  </div>
</div>
</body>
</html>`

	t, err := template.New("zreport").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderPLReportHTML renders a styled, email-ready HTML P&L summary.
func (s *ReportService) RenderPLReportHTML(data *PLReportData) (string, error) {
	tmplStr := `
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f8fafc; color: #1e293b; margin: 0; padding: 24px; }
  .container { max-width: 650px; margin: 0 auto; background: #ffffff; border-radius: 16px; overflow: hidden; border: 1px solid #e2e8f0; }
  .header { background: #041E42; color: #fff; padding: 28px; text-align: center; }
  .header h1 { margin: 0; font-size: 22px; font-weight: 800; text-transform: uppercase; }
  .header p { margin: 6px 0 0; font-size: 13px; color: #F3A41D; font-weight: 700; text-transform: uppercase; }
  .content { padding: 24px; }
  .summary-box { background: #f1f5f9; border-radius: 12px; padding: 18px; margin-bottom: 20px; }
  table { width: 100%; border-collapse: collapse; font-size: 13px; margin-top: 10px; }
  th { text-align: left; padding: 8px; background: #f8fafc; color: #64748b; font-size: 11px; text-transform: uppercase; }
  td { padding: 10px 8px; border-bottom: 1px solid #f1f5f9; }
  .text-right { text-align: right; }
  .font-bold { font-weight: bold; }
  .positive { color: #10b981; }
  .negative { color: #ef4444; }
  .footer { background: #f8fafc; padding: 16px; text-align: center; font-size: 11px; color: #94a3b8; border-top: 1px solid #e2e8f0; }
</style>
</head>
<body>
<div class="container">
  <div class="header">
    <h1>{{.TenantName}}</h1>
    <p>{{.PeriodLabel}} P&L FINANCIAL SUMMARY ({{.StartDate}} - {{.EndDate}})</p>
  </div>
  <div class="content">
    <div style="background: {{if ge .NetProfit 0.0}}#f0fdf4{{else}}#fef2f2{{end}}; border: 1px solid {{if ge .NetProfit 0.0}}#bbf7d0{{else}}#fecaca{{end}}; border-radius: 12px; padding: 20px; text-align: center; margin-bottom: 20px;">
      <div style="font-size: 11px; font-weight: 700; text-transform: uppercase; color: #64748b;">Net Profit / Margin</div>
      <div style="font-size: 34px; font-weight: 900; color: {{if ge .NetProfit 0.0}}#15803d{{else}}#b91c1c{{end}}; margin: 6px 0;">GHS {{printf "%.2f" .NetProfit}}</div>
      <div style="font-size: 13px; font-weight: 700; color: {{if ge .NetProfit 0.0}}#15803d{{else}}#b91c1c{{end}};">Net Profit Margin: {{printf "%.1f" .NetMargin}}%</div>
    </div>

    <table>
      <thead><tr><th>Metric</th><th class="text-right">Amount (GHS)</th></tr></thead>
      <tbody>
        <tr><td>Total Gross Revenue</td><td class="text-right font-bold">{{printf "%.2f" .TotalRevenue}}</td></tr>
        <tr><td>Cost of Goods Sold (COGS)</td><td class="text-right negative">-{{printf "%.2f" .TotalCOGS}}</td></tr>
        <tr style="background:#f8fafc; font-weight:bold;"><td>Gross Profit (Margin: {{printf "%.1f" .GrossMargin}}%)</td><td class="text-right positive">{{printf "%.2f" .GrossProfit}}</td></tr>
        <tr><td>Operating Expenses</td><td class="text-right negative">-{{printf "%.2f" .TotalExpenses}}</td></tr>
        <tr style="background:#e0f2fe; font-weight:bold; font-size:14px;"><td>Net Operating Profit</td><td class="text-right {{if ge .NetProfit 0.0}}positive{{else}}negative{{end}}">{{printf "%.2f" .NetProfit}}</td></tr>
      </tbody>
    </table>
  </div>
  <div class="footer">
    Generated automatically by Puxbay Financial Intelligence &bull; <a href="https://puxbay.com" style="color: #F3A41D; text-decoration: none; font-weight: bold;">puxbay.com</a>
  </div>
</div>
</body>
</html>`

	t, err := template.New("plreport").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
