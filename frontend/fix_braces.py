import re
with open("src/app/core/layout/sidebar/sidebar.html") as f:
    text = f.read()

lines = text.split("\n")
stack = []

for i, line in enumerate(lines):
    if "</nav>" in line:
        # We reached </nav>. We should close all open blocks.
        if len(stack) > 0:
            print(f"Adding {len(stack)} braces before </nav>")
            # Insert the braces
            for j in range(len(stack)):
                lines.insert(i, "      }")
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

# Save it back
with open("src/app/core/layout/sidebar/sidebar.html", "w") as f:
    f.write("\n".join(lines))
