import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterModule } from '@angular/router';
import { ApiService } from '../../core/services/api.service';
import { Title } from '@angular/platform-browser';
import { ImageUrlPipe } from '../../core/pipes/image-url.pipe';

@Component({
  selector: 'app-public-receipt',
  standalone: true,
  imports: [CommonModule, RouterModule, ImageUrlPipe],
  templateUrl: './public-receipt.html',
})
export class PublicReceiptComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private api = inject(ApiService);
  private titleService = inject(Title);
  private imageUrlPipe = inject(ImageUrlPipe);

  order = signal<any | null>(null);
  isLoading = signal(true);
  error = signal<string | null>(null);
  tenantName = signal<string>('Store');
  logoUrl = signal<string | null>(null);
  branchName = signal<string | null>(null);
  branchAddress = signal<string | null>(null);
  branchPhone = signal<string | null>(null);
  receiptHeader = signal<string | null>(null);
  receiptFooter = signal<string | null>(null);

  ngOnInit() {
    const token = this.route.snapshot.paramMap.get('token');
    if (!token) {
      this.error.set('Invalid receipt link');
      this.isLoading.set(false);
      return;
    }

    // Call the backend API (passing Accept: application/json to get the JSON representation)
    this.api.get(`/public/receipts/${token}`, { headers: { 'Accept': 'application/json' } }).subscribe({
      next: (res: any) => {
        this.order.set(res);
        this.titleService.setTitle(`Receipt ${res.order_number}`);
        
        // Use hostname for tenant name fallback
        this.tenantName.set(window.location.hostname.split('.')[0].toUpperCase());

        if (res.branch) {
          this.branchName.set(res.branch.name);
          this.branchAddress.set(res.branch.address);
          this.branchPhone.set(res.branch.phone);
          this.receiptHeader.set(res.branch.receipt_header);
          this.receiptFooter.set(res.branch.receipt_footer);
          if (res.branch.logo) {
            this.logoUrl.set(this.imageUrlPipe.transform(res.branch.logo, false));
          }
        }
        this.isLoading.set(false);
      },
      error: () => {
        this.error.set('Receipt not found or expired.');
        this.isLoading.set(false);
      }
    });
  }

  printReceipt() {
    window.print();
  }
}
