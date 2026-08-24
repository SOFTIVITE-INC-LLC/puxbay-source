const fs = require('fs');

let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'utf8');

// Remove the debug block I added earlier
code = code.replace(/<div style="position: absolute; z-index: 9999; background: white; color: black; padding: 10px; border: 2px solid red;">[\s\S]*?<\/div>/, '');

// Fix the CSS classes that are breaking the layout
code = code.replace('flex flex-col h-full relative', 'flex flex-col relative');
code = code.replace('w-full aspect-square rounded-xl bg-slate-50 dark:bg-slate-900 overflow-hidden mb-3 relative shrink-0', 'w-full aspect-square rounded-xl bg-slate-50 dark:bg-slate-900 overflow-hidden mb-3 relative');
code = code.replace('flex flex-col flex-1 min-h-0', 'flex flex-col flex-1');

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', code);
