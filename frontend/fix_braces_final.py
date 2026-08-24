with open("src/app/core/layout/sidebar/sidebar.html") as f:
    lines = f.read().split("\n")

nav_idx = 0
for i, line in enumerate(lines):
    if "</nav>" in line:
        nav_idx = i
        break

# Remove the extra braces right above nav_idx
# The previous lines should be:
#       }
#       }
#       }
#       }
#   </nav>
removed = 0
for i in range(nav_idx - 1, -1, -1):
    if "}" in lines[i] and not "{{" in lines[i]:
        if removed < 3:
            lines[i] = ""
            removed += 1
        else:
            break

with open("src/app/core/layout/sidebar/sidebar.html", "w") as f:
    f.write("\n".join(x for x in lines if x != ""))

