import re

with open('frontend/src/app/features/main/landing/landing.html', 'r') as f:
    content = f.read()

# Remove the first <style>...</style> block
content = re.sub(r'<style>.*?</style>', '', content, count=1, flags=re.DOTALL)

# Remove the second <style>...</style> block
content = re.sub(r'<style>.*?</style>', '', content, count=1, flags=re.DOTALL)

# Also remove the Google Fonts link as we probably want it in index.html, but let's keep it here for now if needed.
# Actually, the user plan only said remove `<style>` blocks.

with open('frontend/src/app/features/main/landing/landing.html', 'w') as f:
    f.write(content)
