with open('internal/handlers/order_handler.go', 'r', encoding='utf-8', errors='ignore') as f:
    lines = f.readlines()
with open('internal/handlers/order_handler.go', 'w', encoding='utf-8') as f:
    f.writelines(lines[:176])

with open('internal/handlers/inventory_handler.go', 'r', encoding='utf-8', errors='ignore') as f:
    lines = f.readlines()
with open('internal/handlers/inventory_handler.go', 'w', encoding='utf-8') as f:
    f.writelines(lines[:174])
