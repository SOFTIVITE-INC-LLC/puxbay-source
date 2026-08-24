import { Component, inject, OnInit } from '@angular/core';

import { PaymentMethodService } from '../../../core/services/payment-method.service';

@Component({
  selector: 'app-payment-methods',
  standalone: true,
  imports: [],
  templateUrl: './payment-methods.html',
  styles: `
    .glass-panel {
      background: rgba(255, 255, 255, 0.05);
      backdrop-filter: blur(10px);
      border: 1px solid rgba(255, 255, 255, 0.1);
    }
    .dark .glass-panel {
      background: rgba(0, 0, 0, 0.2);
    }
  `,
})
export class PaymentMethods implements OnInit {
  methodService = inject(PaymentMethodService);

  ngOnInit() {
    this.methodService.getMethods().subscribe();
  }
}
