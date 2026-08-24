const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.ts', 'utf8');
code = code.replace("Math = Math;", "Math = Math;\n  current_date = new Date();");
fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.ts', code);
