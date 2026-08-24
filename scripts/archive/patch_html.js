const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'utf8');

code = code.replace("10% Off Orders > $100", "{{ facade.promoPercent() }}% Off Orders > {{ facade.promoThreshold() | appCurrency }}");

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', code);
