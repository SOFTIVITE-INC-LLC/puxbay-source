const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/backend/internal/handlers/inventory_handler.go', 'utf8');

const target = `\tfor _, h := range history {
\t\tformattedHistory = append(formattedHistory, gin.H{
\t\t\t"date":     h.CreatedAt,
\t\t\t"action":   h.Reason,
\t\t\t"quantity": h.Quantity,
\t\t\t"user":     "Admin", // Can be extended to fetch real user name
\t\t})
\t}`;

const replacement = `\tfor _, h := range history {
\t\t// Map backend reason to frontend expected action
\t\taction := "edit"
\t\tif h.Quantity > 0 {
\t\t\taction = "add"
\t\t} else if h.Quantity < 0 {
\t\t\taction = "remove"
\t\t}
\t\t
\t\tabsQty := h.Quantity
\t\tif absQty < 0 {
\t\t\tabsQty = -absQty
\t\t}
\t\t
\t\tformattedHistory = append(formattedHistory, gin.H{
\t\t\t"date":     h.CreatedAt,
\t\t\t"action":   action,
\t\t\t"quantity": absQty,
\t\t\t"user":     "System", // Can be extended to fetch real user name
\t\t\t"notes":    h.Reason,
\t\t})
\t}`;

code = code.replace(target, replacement);
fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/backend/internal/handlers/inventory_handler.go', code);
