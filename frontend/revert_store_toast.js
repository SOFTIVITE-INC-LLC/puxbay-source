const fs = require('fs');
const glob = require('glob');
const path = require('path');

const files = glob.sync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/**/*.ts');

for (const file of files) {
  // If the file is in 'core/store' or 'features/store' OR it's one of the files complaining about 'show' not existing.
  if (file.includes('/store/')) {
    let content = fs.readFileSync(file, 'utf8');
    let original = content;

    if (content.includes('import { ToastService } from')) {
      const dir = path.dirname(file);
      let relative = path.relative(dir, '/home/afari/Projects/development/softivite/puxbay/frontend/src/app/core/store/services/toast.service');
      if (!relative.startsWith('.')) relative = './' + relative;
      
      content = content.replace(/import { ToastService } from '.*';/, `import { ToastService } from '${relative}';`);
    }

    if (original !== content) {
      fs.writeFileSync(file, content);
      console.log(`Reverted import in ${file}`);
    }
  }
}
