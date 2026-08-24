const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/backend/internal/services/order_service.go', 'utf8');

const target = `\terr := s.db.Transaction(func(tx *gorm.DB) error {
\t\tfor _, itemReq := range input.Items {
\t\t\tvar product models.Product
\t\t\tif err := tx.Where("id = ?", itemReq.ProductID).First(&product).Error; err != nil {
\t\t\t\treturn fmt.Errorf("invalid product: %s", itemReq.ProductID)
\t\t\t}

\t\t\torderItem := models.OrderItem{
\t\t\t\tProductID: itemReq.ProductID,
\t\t\t\tVariantID: itemReq.VariantID,
\t\t\t\tQuantity:  itemReq.Quantity,
\t\t\t\tUnitPrice: itemReq.UnitPrice,
\t\t\t\tDiscount:  itemReq.Discount,
\t\t\t\tTotal:     itemReq.Total,
\t\t\t\tCostPrice: product.CostPrice,
\t\t\t}
\t\t\torder.Items = append(order.Items, orderItem)

\t\t\t// Gap #11: Prevent negative stock with WHERE guard
\t\t\tif product.TrackInventory {
\t\t\t\tresult := tx.Model(&models.Product{}).
\t\t\t\t\tWhere("id = ? AND current_stock >= ?", product.ID, itemReq.Quantity).
\t\t\t\t\tUpdate("current_stock", gorm.Expr("current_stock - ?", itemReq.Quantity))
\t\t\t\tif result.RowsAffected == 0 {
\t\t\t\t\treturn fmt.Errorf("insufficient stock for product %s (available: %.2f, requested: %.2f)",
\t\t\t\t\t\tproduct.Name, product.CurrentStock, itemReq.Quantity)
\t\t\t\t}
\t\t\t}
\t\t}

\t\tif err := tx.Create(&order).Error; err != nil {
\t\t\treturn err
\t\t}`;

const replacement = `\terr := s.db.Transaction(func(tx *gorm.DB) error {
\t\tvar movements []models.StockMovement
\t\t
\t\tfor _, itemReq := range input.Items {
\t\t\tvar product models.Product
\t\t\tif err := tx.Where("id = ?", itemReq.ProductID).First(&product).Error; err != nil {
\t\t\t\treturn fmt.Errorf("invalid product: %s", itemReq.ProductID)
\t\t\t}

\t\t\torderItem := models.OrderItem{
\t\t\t\tProductID: itemReq.ProductID,
\t\t\t\tVariantID: itemReq.VariantID,
\t\t\t\tQuantity:  itemReq.Quantity,
\t\t\t\tUnitPrice: itemReq.UnitPrice,
\t\t\t\tDiscount:  itemReq.Discount,
\t\t\t\tTotal:     itemReq.Total,
\t\t\t\tCostPrice: product.CostPrice,
\t\t\t}
\t\t\torder.Items = append(order.Items, orderItem)

\t\t\t// Gap #11: Prevent negative stock with WHERE guard
\t\t\tif product.TrackInventory {
\t\t\t\tresult := tx.Model(&models.Product{}).
\t\t\t\t\tWhere("id = ? AND current_stock >= ?", product.ID, itemReq.Quantity).
\t\t\t\t\tUpdate("current_stock", gorm.Expr("current_stock - ?", itemReq.Quantity))
\t\t\t\tif result.RowsAffected == 0 {
\t\t\t\t\treturn fmt.Errorf("insufficient stock for product %s (available: %.2f, requested: %.2f)",
\t\t\t\t\t\tproduct.Name, product.CurrentStock, itemReq.Quantity)
\t\t\t\t}
\t\t\t\t
\t\t\t\t// Queue stock movement for history
\t\t\t\tmovements = append(movements, models.StockMovement{
\t\t\t\t\tTenantID:      product.TenantID,
\t\t\t\t\tBranchID:      *order.BranchID,
\t\t\t\t\tProductID:     product.ID,
\t\t\t\t\tVariantID:     itemReq.VariantID,
\t\t\t\t\tQuantity:      -itemReq.Quantity,
\t\t\t\t\tPreviousStock: product.CurrentStock,
\t\t\t\t\tNewStock:      product.CurrentStock - itemReq.Quantity,
\t\t\t\t\tReason:        "sale",
\t\t\t\t\tUserID:        cashierID,
\t\t\t\t})
\t\t\t}
\t\t}

\t\tif err := tx.Create(&order).Error; err != nil {
\t\t\treturn err
\t\t}
\t\t
\t\t// Assign ReferenceID as Order ID and create movements
\t\torderIDStr := order.ID.String()
\t\tfor i := range movements {
\t\t\tmovements[i].ReferenceID = &orderIDStr
\t\t}
\t\tif len(movements) > 0 {
\t\t\tif err := tx.Create(&movements).Error; err != nil {
\t\t\t\treturn err
\t\t\t}
\t\t}`;

code = code.replace(target, replacement);
fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/backend/internal/services/order_service.go', code);
