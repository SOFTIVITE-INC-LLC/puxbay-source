const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/backend/internal/services/order_service.go', 'utf8');

const target2 = `BranchID:      func() uuid.UUID { if order.BranchID != nil { return *order.BranchID }; return product.TenantID }(),`;
const replace2 = `BranchID:      func() uuid.UUID { if order.BranchID != nil { return *order.BranchID }; var t models.Tenant; tx.First(&t); return t.ID }(),`;

code = code.replace(new RegExp(target2.replace(/[.*+?^$\{\}\(\)\|\[\]\\]/g, '\\$&'), 'g'), replace2);

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/backend/internal/services/order_service.go', code);
