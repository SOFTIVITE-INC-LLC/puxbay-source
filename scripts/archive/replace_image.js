const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'utf8');

code = code.replace(/'https:\/\/placehold.co\/100x100\/f8fafc\/94a3b8\?text=No\+Img'/g, "'/assets/images/default-product.png'");
code = code.replace(/'https:\/\/placehold.co\/400x400\/f8fafc\/94a3b8\?text=No\+Image'/g, "'/assets/images/default-product.png'");
code = code.replace(/'https:\/\/placehold.co\/100x100\/f8fafc\/94a3b8'/g, "'/assets/images/default-product.png'");

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', code);
