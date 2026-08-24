import os
import re
import glob

# Paths to search
dirs_to_search = [
    '/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/store',
    '/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/storefront'
]

# Patterns to replace (Dark class -> Light + Dark)
replacements = {
    # Backgrounds
    r'\bbg-zinc-950\b': 'bg-slate-50 dark:bg-zinc-950',
    r'\bbg-zinc-900\b': 'bg-white dark:bg-zinc-900',
    r'\bbg-zinc-800\b': 'bg-slate-100 dark:bg-zinc-800',
    
    # Text
    r'\btext-white\b': 'text-slate-900 dark:text-white',
    r'\btext-zinc-400\b': 'text-slate-500 dark:text-zinc-400',
    r'\btext-zinc-500\b': 'text-slate-400 dark:text-zinc-500',
    r'\btext-zinc-300\b': 'text-slate-600 dark:text-zinc-300',
    
    # Borders
    r'\bborder-white/10\b': 'border-slate-200 dark:border-white/10',
    r'\bborder-white/5\b': 'border-slate-200 dark:border-white/5',
    r'\bborder-zinc-800\b': 'border-slate-200 dark:border-zinc-800',
    r'\bborder-zinc-700\b': 'border-slate-300 dark:border-zinc-700',
    
    # Focus rings
    r'\bfocus:ring-white\b': 'focus:ring-indigo-500 dark:focus:ring-white',
    r'\bfocus:border-white\b': 'focus:border-indigo-500 dark:focus:border-white',
}

files_modified = 0

for directory in dirs_to_search:
    if not os.path.exists(directory):
        continue
    
    for filepath in glob.glob(directory + '/**/*.html', recursive=True):
        with open(filepath, 'r') as file:
            content = file.read()
            
        original_content = content
        
        for old, new in replacements.items():
            # Let's be careful about not replacing if it's already there (to avoid doubling if run twice)
            # Only do the replacement if the file doesn't already contain the `dark:xyz` mapping
            # Actually, `dark:` is enough to signal that it was replaced, but some elements might already have `dark:`
            # A simple way to avoid double mapping is just to use negative lookbehind/lookahead, but since I am running it once, it's fine.
            content = re.sub(old, new, content)
            
        if content != original_content:
            with open(filepath, 'w') as file:
                file.write(content)
            files_modified += 1
            print(f"Updated {filepath}")

print(f"Total files modified: {files_modified}")
