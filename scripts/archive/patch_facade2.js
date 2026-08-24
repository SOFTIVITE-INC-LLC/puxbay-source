const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', 'utf8');

const newProps = `
  parkedSales = signal<{ cart: any[], customer: any, time: Date }[]>([]);
  isParkedSalesModalOpen = signal(false);

  isPrinterConnected = signal(false);
  printerPort: any = null;

  readonly quickCashOptions = computed(() => {
    const total = this.cartTotal();
    const options = [total];
    const rounded5 = Math.ceil(total / 5) * 5;
    if (rounded5 > total && rounded5 - total < 5) options.push(rounded5);
    const rounded10 = Math.ceil(total / 10) * 10;
    if (rounded10 > total && !options.includes(rounded10)) options.push(rounded10);
    const rounded50 = Math.ceil(total / 50) * 50;
    if (rounded50 > total && !options.includes(rounded50)) options.push(rounded50);
    return options.sort((a,b) => a-b);
  });

  parkSale() {
    if (this.cart().length === 0) return;
    this.parkedSales.update(sales => [
      ...sales,
      { cart: [...this.cart()], customer: this.selectedCustomer(), time: new Date() }
    ]);
    this.toastr.info('Sale parked successfully');
    this.clearCart();
  }

  resumeSale(index: number) {
    const sale = this.parkedSales()[index];
    if (sale) {
      this.cart.set(sale.cart);
      this.selectedCustomer.set(sale.customer);
      this.removeParkedSale(index);
      this.isParkedSalesModalOpen.set(false);
    }
  }

  removeParkedSale(index: number) {
    this.parkedSales.update(sales => sales.filter((_, i) => i !== index));
  }

  async connectPrinter() {
    if (!('serial' in navigator)) {
      this.toastr.error('Web Serial API is not supported in this browser');
      return;
    }
    try {
      const port = await (navigator as any).serial.requestPort();
      await port.open({ baudRate: 9600 });
      this.printerPort = port;
      this.isPrinterConnected.set(true);
      this.toastr.success('Receipt printer connected');
    } catch (e: any) {
      this.toastr.error('Failed to connect printer');
    }
  }

  async printReceipt(order: any = null) {
    const receiptOrder = order || this.checkoutSuccessOrder();
    if (!receiptOrder) return;
    
    if (this.printerPort && this.isPrinterConnected()) {
      try {
        const writer = this.printerPort.writable.getWriter();
        const encoder = new TextEncoder();
        const ESC = "\\x1B";
        const GS = "\\x1D";
        let text = ESC + "@" + ESC + "E1" + "Softivite POS\\n" + ESC + "E0";
        text += "Order #" + receiptOrder.order_number + "\\n";
        text += "Date: " + new Date().toLocaleString() + "\\n";
        text += "--------------------------------\\n";
        text += GS + "V\\x41\\x00"; // Cut paper
        await writer.write(encoder.encode(text));
        writer.releaseLock();
        this.toastr.success('Receipt sent to printer');
      } catch (e: any) {
        this.toastr.error('Failed to print receipt');
      }
    } else {
      window.print(); // Fallback to browser print
    }
  }
`;

code = code.replace('// Data', newProps + '\n  // Data');
fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', code);
