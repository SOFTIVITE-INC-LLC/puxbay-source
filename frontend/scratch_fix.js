const fs = require('fs');
const path = require('path');
const dir = 'd:/Projects/development/softivite/puxbay/frontend/src/app/features/main';

function walk(d) {
  fs.readdirSync(d).forEach(f => {
    const p = path.join(d, f);
    if (fs.statSync(p).isDirectory()) {
      walk(p);
    } else if (p.endsWith('.ts')) {
      let c = fs.readFileSync(p, 'utf8');
      if (c.includes("templateUrl: '',")) {
        const name = path.basename(p, '.ts');
        c = c.replace(/templateUrl: '',/, `templateUrl: './${name}.html',`);
        fs.writeFileSync(p, c);
        console.log('Fixed ' + p);
      }
    }
  });
}
walk(dir);
