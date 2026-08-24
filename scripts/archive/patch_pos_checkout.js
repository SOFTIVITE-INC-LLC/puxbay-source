const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/backend/internal/services/order_service.go', 'utf8');

const target = `\t\treturn nil, err\n\t}\n\n\treturn &order, nil\n}\n\nfunc generateOrderNumber()`;
const replacement = `\t\treturn nil, err\n\t}\n\n\t// Preload Branch and Tenant for the frontend receipt\n\ts.db.Preload("Branch").Preload("Tenant").First(&order, "id = ?", order.ID)\n\n\treturn &order, nil\n}\n\nfunc generateOrderNumber()`;

code = code.replace(target, replacement);

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/backend/internal/services/order_service.go', code);
