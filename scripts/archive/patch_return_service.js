const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/backend/internal/services/return_service.go', 'utf8');

const target = `\t\tfor _, item := range items {
\t\t\tif item.ProductID != nil {
\t\t\t\tif err := tx.Model(&models.Product{}).Where("id = ?", item.ProductID).
\t\t\t\t\tUpdateColumn("current_stock", gorm.Expr("current_stock + ?", item.Quantity)).Error; err != nil {
\t\t\t\t\treturn err
\t\t\t\t}
\t\t\t}
\t\t}`;

const replacement = `\t\tfor _, item := range items {
\t\t\tif item.ProductID != nil {
\t\t\t\tvar product models.Product
\t\t\t\tif err := tx.Where("id = ?", item.ProductID).First(&product).Error; err != nil {
\t\t\t\t\treturn err
\t\t\t\t}
\t\t\t\t
\t\t\t\tif err := tx.Model(&models.Product{}).Where("id = ?", item.ProductID).
\t\t\t\t\tUpdateColumn("current_stock", gorm.Expr("current_stock + ?", item.Quantity)).Error; err != nil {
\t\t\t\t\treturn err
\t\t\t\t}
\t\t\t\t
\t\t\t\tretIDStr := ret.ID.String()
\t\t\t\tmovement := models.StockMovement{
\t\t\t\t\tTenantID:      func() uuid.UUID { var t models.Tenant; tx.First(&t); return t.ID }(),
\t\t\t\t\tBranchID:      func() uuid.UUID { if ret.BranchID != nil { return *ret.BranchID }; var t models.Tenant; tx.First(&t); return t.ID }(),
\t\t\t\t\tProductID:     product.ID,
\t\t\t\t\tQuantity:      item.Quantity,
\t\t\t\t\tPreviousStock: product.CurrentStock,
\t\t\t\t\tNewStock:      product.CurrentStock + item.Quantity,
\t\t\t\t\tReason:        "return",
\t\t\t\t\tReferenceID:   &retIDStr,
\t\t\t\t}
\t\t\t\tif err := tx.Create(&movement).Error; err != nil {
\t\t\t\t\treturn err
\t\t\t\t}
\t\t\t}
\t\t}`;

code = code.replace(target, replacement);
fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/backend/internal/services/return_service.go', code);
