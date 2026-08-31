import { Component, OnInit, signal, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AppCurrencyPipe } from '../../../core/pipes/app-currency.pipe';

@Component({
  selector: 'app-cds',
  standalone: true,
  imports: [CommonModule, AppCurrencyPipe],
  template: `
    <div class="h-screen w-full bg-slate-900 text-white flex flex-col font-sans overflow-hidden">
      <!-- Header -->
      <div class="h-24 bg-slate-800 border-b border-slate-700 flex justify-between items-center px-8 shrink-0 shadow-xl z-10">
        <div class="flex items-center gap-4">
          <div class="w-12 h-12 bg-indigo-600 rounded-2xl flex items-center justify-center">
            <span class="material-symbols-outlined text-3xl">storefront</span>
          </div>
          <div>
            <h1 class="text-2xl font-black">Softivite Welcome!</h1>
            <p class="text-slate-400 font-bold">Please review your order</p>
          </div>
        </div>
        <div class="text-right">
          <p class="text-slate-400 uppercase tracking-widest text-sm font-bold mb-1">Total Amount</p>
          <p class="text-5xl font-black text-indigo-400 tracking-tighter">{{ state().total | appCurrency }}</p>
        </div>
      </div>

      <!-- Main Area -->
      <div class="flex-1 flex overflow-hidden">
        
        <!-- Left: Cart Items -->
        <div class="flex-1 flex flex-col border-r border-slate-800 bg-slate-900/50">
          <div *ngIf="state().cart.length === 0" class="flex-1 flex flex-col items-center justify-center opacity-50">
            <span class="material-symbols-outlined text-8xl mb-6 text-slate-700">shopping_cart</span>
            <h2 class="text-3xl font-bold text-slate-500">Your cart is empty</h2>
            <p class="text-slate-600 mt-2 text-lg">Items will appear here as they are scanned.</p>
          </div>
          
          <div *ngIf="state().cart.length > 0" class="flex-1 overflow-y-auto p-6 space-y-4">
            <div *ngFor="let item of state().cart" class="bg-slate-800 rounded-3xl p-4 flex gap-6 items-center shadow-lg border border-slate-700">
              <div class="w-24 h-24 rounded-2xl bg-slate-900 overflow-hidden shrink-0">
                <img [src]="item.product.image_url || '/images/default-product.png'" class="w-full h-full object-cover">
              </div>
              <div class="flex-1 min-w-0">
                <h3 class="text-2xl font-bold text-white truncate">{{ item.product.name }}</h3>
                <p class="text-slate-400 text-lg mt-1">{{ item.quantity }} x {{ (item.product.selling_price || 0) | appCurrency }}</p>
                <div *ngIf="item.discount > 0" class="mt-2 inline-block px-3 py-1 bg-emerald-500/20 text-emerald-400 font-bold rounded-lg text-sm border border-emerald-500/30">
                  Discount applied
                </div>
              </div>
              <div class="text-right shrink-0">
                <p class="text-3xl font-black text-white">{{ ((item.product.selling_price || 0) * item.quantity) | appCurrency }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Right: Order Summary & Ads -->
        <div class="w-[450px] bg-slate-800 flex flex-col">
          <div class="p-8 space-y-4 bg-slate-800/80 border-b border-slate-700 backdrop-blur-md">
            <div class="flex justify-between text-slate-400 text-xl font-bold">
              <span>Subtotal</span>
              <span>{{ state().subtotal | appCurrency }}</span>
            </div>
            <div *ngIf="state().discount > 0" class="flex justify-between text-emerald-400 text-xl font-bold">
              <span>Discount</span>
              <span>-{{ state().discount | appCurrency }}</span>
            </div>
            <div class="flex justify-between text-slate-400 text-xl font-bold">
              <span>Tax</span>
              <span>{{ state().tax | appCurrency }}</span>
            </div>
            <div class="h-px bg-slate-700 my-6"></div>
            <div class="flex justify-between items-end">
              <span class="text-xl font-black text-slate-500 uppercase tracking-widest">Total Due</span>
              <span class="text-6xl font-black text-indigo-400 tracking-tighter">{{ state().total | appCurrency }}</span>
            </div>
          </div>
          
          <div class="flex-1 p-6 relative overflow-hidden flex flex-col justify-end">
            <!-- Mock Ad / Promo Area -->
            <div class="w-full bg-indigo-500 rounded-3xl p-8 text-white shadow-2xl relative overflow-hidden group">
              <div class="absolute inset-0 bg-black/20 opacity-0 group-hover:opacity-100 transition-opacity duration-700"></div>
              <div class="relative z-10">
                <h3 class="text-3xl font-black mb-2">Join Our Rewards!</h3>
                <p class="text-indigo-100 text-lg mb-6">Earn points on every purchase and unlock exclusive discounts.</p>
                <div class="flex items-center gap-3">
                  <span class="material-symbols-outlined text-4xl text-amber-300">stars</span>
                  <span class="font-bold text-xl">Ask your cashier to sign up today!</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Thank You Overlay -->
      <div *ngIf="state().isSuccess" class="absolute inset-0 bg-indigo-600 z-50 flex flex-col items-center justify-center text-white animate-in slide-in-from-bottom duration-500">
        <div class="w-48 h-48 bg-white/20 rounded-full flex items-center justify-center mb-8 shadow-2xl backdrop-blur-md">
          <span class="material-symbols-outlined text-8xl">check_circle</span>
        </div>
        <h1 class="text-7xl font-black mb-4">Thank You!</h1>
        <p class="text-3xl font-bold text-indigo-200">Your transaction is complete.</p>
        <p class="text-2xl mt-8 text-indigo-300">Please take your receipt.</p>
      </div>
    </div>
  `
})
export class Cds implements OnInit, OnDestroy {
  state = signal<{ cart: any[], subtotal: number, tax: number, discount: number, total: number, isSuccess: boolean }>({
    cart: [], subtotal: 0, tax: 0, discount: 0, total: 0, isSuccess: false
  });
  
  private bc!: BroadcastChannel;

  ngOnInit() {
    this.bc = new BroadcastChannel('pos_sync_channel');
    this.bc.onmessage = (event) => {
      if (event.data) {
        this.state.set(event.data);
      }
    };
    this.bc.postMessage({ type: 'request_sync' });
  }

  ngOnDestroy() {
    if (this.bc) this.bc.close();
  }
}
