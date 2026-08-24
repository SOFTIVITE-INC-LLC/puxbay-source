const fs = require('fs');
const glob = require('glob');
const path = require('path');

const files = glob.sync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/**/*.ts');
const toastPath = '/home/afari/Projects/development/softivite/puxbay/frontend/src/app/core/services/toast.ts';

for (const file of files) {
  let content = fs.readFileSync(file, 'utf8');
  let original = content;

  if (content.includes('import { ToastService } from')) {
    const dir = path.dirname(file);
    let relative = path.relative(dir, '/home/afari/Projects/development/softivite/puxbay/frontend/src/app/core/services/toast');
    if (!relative.startsWith('.')) relative = './' + relative;
    
    content = content.replace(/import { ToastService } from '.*';/, `import { ToastService } from '${relative}';`);
  }
  
  if (content.includes('toastService = inject(ToastService)') && !content.includes('import { inject }') && !content.includes('import { Component, inject }') && !content.includes(', inject')) {
      content = content.replace(/import { (Component|Injectable|Directive) }/, 'import { $1, inject }');
      if (!content.includes('inject }')) {
          content = `import { inject } from '@angular/core';\n` + content;
      }
  }

  if (original !== content) {
    fs.writeFileSync(file, content);
    console.log(`Fixed ${file}`);
  }
}
