import re

with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', 'r') as f:
    facade = f.read()

# Fix duplicate clearCart
facade = re.sub(r'  clearCart\(\) \{\n    this\.cart\.set\(\[\]\);\n  \}\n', '', facade, count=1)

# Fix orderData -> newOrder
facade = facade.replace('this.offlineService.addSyncRequest(\'/orders\', \'POST\', orderData);', 'this.offlineService.addSyncRequest(\'/orders\', \'POST\', newOrder);')
facade = facade.replace('this.orderService.createOrder(orderData).subscribe({', 'this.orderService.createOrder(newOrder as any).subscribe({')
facade = facade.replace('if (this.orderNote()) orderData.notes = this.orderNote();', 'if (this.orderNote()) (newOrder as any).notes = this.orderNote();')
facade = facade.replace('this.placeOrder();', 'this.checkout(\'completed\');')

missing = """
  currentUser = { name: 'Admin', username: 'admin' };
  branch = { name: 'Main Branch' };
"""
facade = facade.replace("showMobileCart = signal(false);", missing + "\n  showMobileCart = signal(false);")

with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', 'w') as f:
    f.write(facade)


with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'r') as f:
    html = f.read()

html = html.replace('branch.name', 'facade.branch.name')
html = html.replace('facade.selectedCategory.set(\\\'all\\\')', 'facade.selectedCategory.set(\'all\')')
html = html.replace('facade.selectedCategory === \'all\'', 'facade.selectedCategory() === \'all\'')
html = html.replace('facade.isLargeScreen ?', 'facade.isLargeScreen() ?')
html = html.replace('(!facade.isLargeScreen', '(!facade.isLargeScreen()')
html = html.replace('(click)="toggleFullscreen"', '(click)="facade.toggleFullscreen()"')

with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'w') as f:
    f.write(html)
