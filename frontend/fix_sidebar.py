import re

with open("src/app/core/layout/sidebar/sidebar.html") as f:
    text = f.read()

# Replace userRole === with userRole() ===
text = re.sub(r'userRole ===', 'userRole() ===', text)
text = re.sub(r'userRole !==', 'userRole() !==', text)

lines = text.split('\n')
new_lines = []
skip_next = False

for i, line in enumerate(lines):
    if "<!-- Overview Operations -->" in line:
        new_lines.append(line)
        new_lines.append("        @if (userRole() === 'admin' || userRole() === 'manager' || userRole() === 'superadmin') {")
    elif "<!-- Sales Section -->" in line:
        new_lines.append("        }")
        new_lines.append("")
        new_lines.append(line)
    elif "<!-- Stock Section -->" in line:
        new_lines.append(line)
        new_lines.append("        @if (userRole() === 'admin' || userRole() === 'manager' || userRole() === 'superadmin') {")
    elif "<!-- Management Section -->" in line:
        new_lines.append("        }")
        new_lines.append("")
        new_lines.append(line)
        new_lines.append("        @if (userRole() === 'admin' || userRole() === 'manager' || userRole() === 'superadmin') {")
    elif "<!-- Analytics & Growth Section -->" in line:
        new_lines.append("        }")
        new_lines.append("")
        new_lines.append(line)
        new_lines.append("        @if (userRole() === 'admin' || userRole() === 'manager' || userRole() === 'superadmin') {")
    elif "<!-- Account Section -->" in line and i > 500: # Only the one in branch mode
        new_lines.append("        }")
        new_lines.append("")
        new_lines.append(line)
    else:
        new_lines.append(line)

with open("src/app/core/layout/sidebar/sidebar.html", "w") as f:
    f.write("\n".join(new_lines))

