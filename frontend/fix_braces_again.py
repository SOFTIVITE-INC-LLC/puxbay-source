import re
with open("src/app/core/layout/sidebar/sidebar.html") as f:
    text = f.read()

lines = text.split("\n")
stack = []

# First pass: trace open/close exactly as before, EXCEPT stop at the end to see how many we have.
for i, line in enumerate(lines):
    if "</nav>" in line:
        break
    if "@if" in line and not line.strip().startswith("<!--"):
        matches = re.findall(r'@if\s*\(.*?\)\s*\{', line)
        for m in matches:
            stack.append(f"Line {i+1}: {m}")
    if "}" in line and not "{{" in line and not line.strip().startswith("<!--"):
        c = line.count("}")
        for _ in range(c):
            if stack:
                stack.pop()

# If stack is negative, it means we have extra braces. Wait, stack can't be negative, we just pop. 
# Better: Let's just rewrite the end of the file properly.
# The `nav` is closed at the end. We know we need 0 extra braces before it.
# Let's count depth mathematically.
depth = 0
for i, line in enumerate(lines):
    if "</nav>" in line:
        # Check depth here
        print("Depth before </nav>:", depth)
        break
    if "@if" in line and not line.strip().startswith("<!--"):
        depth += len(re.findall(r'@if\s*\(.*?\)\s*\{', line))
    if "}" in line and not "{{" in line and not line.strip().startswith("<!--"):
        depth -= line.count("}")

