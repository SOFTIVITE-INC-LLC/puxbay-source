import re

with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', 'r') as f:
    facade = f.read()

missing = """
  showMobileCart = signal(false);
  loading = signal(false);
  loadingMessage = signal('');
  giftCardCode = signal('');
  batchProduct = signal<any>(null);
  splitPayments = signal<any[]>([]);
  issueGCAmount = signal(0);
  issueGCCode = signal('');
  lastOrderId = signal('');

  toggleMobileCart() {
    this.showMobileCart.set(!this.showMobileCart());
  }
"""

facade = facade.replace("sendEmailReceipt(email: string): void {", missing + "\n  sendEmailReceipt(email: string): void {")

with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', 'w') as f:
    f.write(facade)

with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'r') as f:
    html = f.read()

html = html.replace('clearCart()', 'facade.clearCart()')
html = html.replace('Math.random()', 'this.Math.random()') # Fix issueGCCode assignment in template

with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'w') as f:
    f.write(html)
