const fs = require('fs');
const glob = require('glob');

const files = glob.sync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/**/*.ts');

for (const file of files) {
  let content = fs.readFileSync(file, 'utf8');
  let original = content;

  // Add ToastService import if it's missing but we need it
  let needsToast = false;

  // Regex to match alert('...') or alert(`...`)
  const alertRegex = /alert\((['"`])(.*?)(\1)\)/g;
  
  if (alertRegex.test(content)) {
    content = content.replace(alertRegex, (match, quote, msg) => {
      needsToast = true;
      if (msg.toLowerCase().includes('fail') || msg.toLowerCase().includes('error') || msg.toLowerCase().includes('invalid')) {
        return `this.toastService.showError(${quote}${msg}${quote})`;
      } else {
        return `this.toastService.showSuccess(${quote}${msg}${quote})`;
      }
    });
  }

  const alertVarRegex = /alert\(([^'"`)]+)\)/g;
  if (alertVarRegex.test(content) && !content.match(/alert\(this\./)) {
    content = content.replace(alertVarRegex, (match, variable) => {
        if(variable.includes('message')) {
            needsToast = true;
            return `this.toastService.showSuccess(${variable})`;
        }
        return match;
    });
  }

  if (needsToast && original !== content) {
    if (!content.includes('ToastService')) {
      content = `import { ToastService } from '../../../core/services/toast';\n` + content;
    }
    if (!content.includes('toastService = inject(') && !content.includes('toastService: ToastService')) {
        // Try to inject it
        content = content.replace(/export class (\w+) (implements \w+ )?{/, (match) => {
            return match + `\n  toastService = inject(ToastService);`;
        });
        
        if (!content.includes('import { inject }') && !content.includes('import { Component, inject }') && !content.includes(', inject')) {
            content = content.replace(/import { (Component|Injectable|Directive) }/, 'import { $1, inject }');
        }
    }
    
    // Fix imports if depth is wrong
    const depth = file.split('/').length - 10; 
    const importPath = depth === 0 ? './core/services/toast' : 
                      depth === 1 ? '../core/services/toast' :
                      depth === 2 ? '../../core/services/toast' :
                      depth === 3 ? '../../../core/services/toast' :
                      depth === 4 ? '../../../../core/services/toast' : '../../../../../core/services/toast';
                      
    content = content.replace(/import { ToastService } from '.*';/, `import { ToastService } from '${importPath}';`);

    fs.writeFileSync(file, content);
    console.log(`Updated ${file}`);
  }
}
