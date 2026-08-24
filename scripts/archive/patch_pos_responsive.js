const fs = require('fs');

let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'utf8');

// 1. Root Flex
code = code.replace(
  '<div id="pos-app" class="h-screen w-full flex bg-slate-50 dark:bg-slate-900 overflow-hidden font-sans text-slate-800 dark:text-slate-200">',
  '<div id="pos-app" class="h-screen w-full flex flex-col lg:flex-row bg-slate-50 dark:bg-slate-900 overflow-hidden font-sans text-slate-800 dark:text-slate-200">'
);

// 2. Hide Sidebar on mobile
code = code.replace(
  '<div class="w-[80px] bg-slate-900 flex flex-col items-center py-6 shrink-0 shadow-2xl relative z-30">',
  '<div class="hidden lg:flex w-[80px] bg-slate-900 flex-col items-center py-6 shrink-0 shadow-2xl relative z-30">'
);

// 3. Make cart 50vh on mobile and side-by-side on desktop
code = code.replace(
  '<div class="w-full md:w-[420px] lg:w-[480px] bg-white dark:bg-slate-800 border-l border-slate-200 dark:border-slate-700/50 flex flex-col h-full shrink-0 shadow-2xl relative z-20">',
  '<div class="w-full lg:w-[420px] xl:w-[480px] bg-white dark:bg-slate-800 border-t lg:border-t-0 lg:border-l border-slate-200 dark:border-slate-700/50 flex flex-col h-[50vh] lg:h-full shrink-0 shadow-2xl relative z-20">'
);

// 4. Adjust the product grid columns for mobile
code = code.replace(
  '<div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 pb-24">',
  '<div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-4 pb-24">'
);

// 5. Hide customer selector partially or adjust padding on mobile
code = code.replace(
  '<div class="p-5 border-b border-slate-100 dark:border-slate-700/50 relative bg-slate-50/50 dark:bg-slate-900/20">',
  '<div class="p-3 lg:p-5 border-b border-slate-100 dark:border-slate-700/50 relative bg-slate-50/50 dark:bg-slate-900/20">'
);

// 6. Adjust checkout button area padding on mobile
code = code.replace(
  '<div class="p-5 bg-slate-50/50 dark:bg-slate-900/20 border-t border-slate-200 dark:border-slate-700/50">',
  '<div class="p-3 lg:p-5 bg-slate-50/50 dark:bg-slate-900/20 border-t border-slate-200 dark:border-slate-700/50">'
);


fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', code);
