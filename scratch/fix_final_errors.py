import re

with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'r') as f:
    html = f.read()

# Fix two-way bindings for signals
html = re.sub(r'\[\(ngModel\)\]="facade\.([a-zA-Z0-Signal]+)\(\)"', r'[ngModel]="facade.\1()" (ngModelChange)="facade.\1.set($event)"', html)
html = re.sub(r'\[\(ngModel\)\]="facade\.([a-zA-Z0-9]+)"', r'[ngModel]="facade.\1()" (ngModelChange)="facade.\1.set($event)"', html)
# Fix assignments to signals
html = re.sub(r'facade\.selectedCategory = \'([^\']+)\'', r'facade.selectedCategory.set(\'\1\')', html)
html = re.sub(r'facade\.selectedCategory = ([a-zA-Z0-9.]+)', r'facade.selectedCategory.set(\1)', html)
html = re.sub(r'facade\.showCustomerDropdown = !facade\.showCustomerDropdown', r'facade.showCustomerDropdown.set(!facade.showCustomerDropdown())', html)
html = html.replace('!(facade.showCustomerDropdown)', '!(facade.showCustomerDropdown())')

html = html.replace('facade.selectedCustomer().name', 'facade.selectedCustomer()?.name')
html = html.replace('@if (item.batchNumber)', '*ngIf="item.batchNumber"')
html = html.replace('facade.facade.clearCart()', 'facade.clearCart()')

html = html.replace('facade.batchProduct?.batches', 'facade.batchProduct()?.batches')
html = html.replace('facade.batchProduct?.name', 'facade.batchProduct()?.name')

html = html.replace('facade.splitPayments.push({method: \'cash\', amount: 0})', 'facade.splitPayments.update(p => { p.push({method: \'cash\', amount: 0}); return p; })')
html = html.replace('facade.splitPayments.splice(index, 1)', 'facade.splitPayments.update(p => { p.splice(index, 1); return p; })')
html = html.replace('facade.issueGCCode = this.Math.random().toString(36).substring(2, 10).toUpperCase()', 'facade.issueGCCode.set(this.Math.random().toString(36).substring(2, 10).toUpperCase())')

with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'w') as f:
    f.write(html)

with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', 'r') as f:
    facade = f.read()

missing = """
  isLargeScreen = signal(window.innerWidth > 1024);
  showCustomerDropdown = signal(false);
  customerSearch = signal('');
"""

facade = facade.replace("showMobileCart = signal(false);", missing + "\n  showMobileCart = signal(false);")

with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', 'w') as f:
    f.write(facade)
