import re

with open("src/app/core/layout/sidebar/sidebar.html") as f:
    text = f.read()

lines = text.split("\n")
stack = []

for i, line in enumerate(lines):
    if "</nav>" in line:
        print(f"Reached </nav> at line {i+1}. Current stack:")
        for item in stack:
            print(f" - {item}")
        break

    # We only care about @if and }
    # Find all @if
    if "@if" in line:
        matches = re.findall(r'@if\s*\(.*?\)\s*\{', line)
        for m in matches:
            stack.append(f"Line {i+1}: {m}")
    
    # Find all }
    if "}" in line and not "{{" in line:
        # crude count of }
        c = line.count("}")
        for _ in range(c):
            if stack:
                stack.pop()
            else:
                print(f"Unexpected }} at line {i+1}")
