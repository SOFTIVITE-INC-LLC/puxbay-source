const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', 'utf8');

const target = 'iframe.src = `${environment.apiUrl}/orders/${order.id}/receipt`;';
const replace = 'iframe.src = `${environment.apiUrl}/orders/${order.id}/receipt?token=${localStorage.getItem(\'access_token\')}`;';

code = code.replace(target, replace);
fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', code);
