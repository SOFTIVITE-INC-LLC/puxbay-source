with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', 'r') as f:
    facade = f.read()

missing = """
  addSplitPayment() {
    this.splitPayments.update(p => [...p, {method: 'cash', amount: 0}]);
  }

  removeSplitPayment(index: number) {
    this.splitPayments.update(p => {
      const copy = [...p];
      copy.splice(index, 1);
      return copy;
    });
  }
"""

facade = facade.replace("showMobileCart = signal(false);", missing + "\n  showMobileCart = signal(false);")

with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', 'w') as f:
    f.write(facade)

with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'r') as f:
    html = f.read()

html = html.replace("facade.splitPayments.update(p => { p.splice(index, 1); return p; })", "facade.removeSplitPayment(index)")
html = html.replace("facade.splitPayments.update(p => { p.push({method: 'cash', amount: 0}); return p; })", "facade.addSplitPayment()")

with open('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'w') as f:
    f.write(html)
