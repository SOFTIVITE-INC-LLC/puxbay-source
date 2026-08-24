const fs = require('fs');

let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'utf8');

code = code.replace('<!-- Main Products Area -->', `<!-- Main Products Area -->
  <div style="position: absolute; z-index: 9999; background: white; color: black; padding: 10px; border: 2px solid red;">
    DEBUG PRODUCT:
    <pre>{{ facade.filteredProducts()[0] | json }}</pre>
  </div>`);

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', code);
