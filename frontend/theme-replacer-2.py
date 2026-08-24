import os
import re
import glob

# Paths to search
dirs_to_search = [
    '/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/store',
    '/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/storefront'
]

replacements = {
    r'\bfrom-zinc-950/90\b': 'from-slate-50/90 dark:from-zinc-950/90',
    r'\bvia-zinc-950/60\b': 'via-slate-50/60 dark:via-zinc-950/60',
    r'\bfrom-zinc-950\b': 'from-slate-50 dark:from-zinc-950',
    r'\bvia-zinc-950/40\b': 'via-slate-50/40 dark:via-zinc-950/40',
    r'\bfrom-zinc-950/80\b': 'from-slate-50/80 dark:from-zinc-950/80',
    r'\bto-\[\#050505\]\b': 'to-slate-100 dark:to-[#050505]',
    
    # Let's fix the footer background:
    r'bg-gradient-to-b from-zinc-950 to-\[\#050505\]': 'bg-gradient-to-b from-slate-50 to-slate-200 dark:from-zinc-950 dark:to-[#050505]',
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
            content = re.sub(old, new, content)
            
        if content != original_content:
            with open(filepath, 'w') as file:
                file.write(content)
            files_modified += 1
            print(f"Updated {filepath}")

print(f"Total files modified: {files_modified}")
