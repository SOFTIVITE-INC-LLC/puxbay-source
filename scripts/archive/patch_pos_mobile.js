const fs = require('fs');

let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', 'utf8');

// --- 1. Redesign Search Bar ---
const oldHeader = `<header class="h-20 bg-white/80 dark:bg-slate-800/80 backdrop-blur-xl border-b border-slate-200 dark:border-slate-700/50 flex items-center justify-between px-6 shrink-0 z-10">
      <div class="flex items-center gap-4">
        <div *ngIf="facade.isOffline()" class="px-3 py-1.5 bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 rounded-lg flex items-center gap-2 text-sm font-bold animate-pulse">
          <span class="material-symbols-outlined text-sm">wifi_off</span> Offline Mode
        </div>
        <div *ngIf="!facade.isOffline()" class="px-3 py-1.5 bg-emerald-50 dark:bg-emerald-900/10 text-emerald-600 dark:text-emerald-400 rounded-lg flex items-center gap-2 text-sm font-bold">
          <span class="material-symbols-outlined text-sm">cloud_done</span> Online
        </div>
      </div>
      
      <div class="flex-1 max-w-xl mx-8 relative group">
        <span class="material-symbols-outlined absolute left-4 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-indigo-600 transition-colors">search</span>
        <input type="text" placeholder="Search or scan barcode (F4)"
               [ngModel]="facade.searchQuery()" (ngModelChange)="facade.searchQuery.set($event)"
               class="w-full bg-slate-100 dark:bg-slate-900/50 border border-transparent focus:bg-white dark:focus:bg-slate-800 focus:border-indigo-500 focus:ring-4 focus:ring-indigo-500/10 rounded-2xl py-3.5 pl-12 pr-4 text-sm font-bold text-slate-800 dark:text-white outline-none transition-all placeholder:text-slate-400 placeholder:font-medium shadow-inner">
        
        <button *ngIf="facade.searchQuery()" (click)="facade.searchQuery.set('')" class="absolute right-4 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200">
          <span class="material-symbols-outlined text-sm">close</span>
        </button>
      </div>

      <div class="flex items-center gap-3">
        <button (click)="facade.isCustomItemModalOpen.set(true)" class="h-10 px-4 flex items-center gap-2 rounded-xl bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-400 hover:bg-indigo-100 dark:hover:bg-indigo-900/40 font-bold text-sm transition-colors">
          <span class="material-symbols-outlined text-[18px]">add_circle</span> Custom Item
        </button>
      </div>
    </header>`;

const newHeader = `<div class="p-4 lg:p-6 shrink-0 z-10 w-full relative z-40">
      <div class="flex flex-col sm:flex-row items-center justify-between gap-4 w-full">
        <div class="w-full sm:w-auto flex items-center justify-between gap-4 order-2 sm:order-1">
          <div *ngIf="facade.isOffline()" class="px-3 py-1.5 bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 rounded-lg flex items-center gap-2 text-sm font-bold shadow-sm">
            <span class="material-symbols-outlined text-sm">wifi_off</span> Offline
          </div>
          <button (click)="facade.isCustomItemModalOpen.set(true)" class="sm:hidden h-12 px-4 flex-1 flex items-center justify-center gap-2 rounded-xl bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-400 font-bold text-sm transition-colors shadow-sm">
            <span class="material-symbols-outlined text-[18px]">add_circle</span> Custom Item
          </button>
        </div>
        
        <div class="flex-1 w-full max-w-2xl relative group order-1 sm:order-2 shadow-sm rounded-2xl">
          <span class="material-symbols-outlined absolute left-4 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-indigo-600 transition-colors">search</span>
          <input type="text" placeholder="Search or scan barcode (F4)"
                 [ngModel]="facade.searchQuery()" (ngModelChange)="facade.searchQuery.set($event)"
                 class="w-full bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 focus:border-indigo-500 focus:ring-4 focus:ring-indigo-500/10 rounded-2xl py-4 pl-12 pr-12 text-base font-bold text-slate-800 dark:text-white outline-none transition-all placeholder:text-slate-400 placeholder:font-medium shadow-sm">
          <button *ngIf="facade.searchQuery()" (click)="facade.searchQuery.set('')" class="absolute right-4 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200">
            <span class="material-symbols-outlined text-sm">close</span>
          </button>
        </div>

        <button (click)="facade.isCustomItemModalOpen.set(true)" class="hidden sm:flex h-14 px-5 items-center gap-2 rounded-xl bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 hover:border-indigo-300 dark:hover:border-indigo-500/50 text-indigo-600 dark:text-indigo-400 font-bold text-sm transition-all shadow-sm order-3">
          <span class="material-symbols-outlined text-[18px]">add_circle</span> Custom Item
        </button>
      </div>
    </div>`;

code = code.replace(oldHeader, newHeader);


// --- 2. Remove legacy receipt component block (backend integration handled in ts) ---
code = code.replace('<!-- PRINT RECEIPT BLOCK -->\n  <app-receipt [order]="facade.checkoutSuccessOrder()"></app-receipt>', '');


// --- 3. Update Cart Sidebar to act as a Bottom Sheet on mobile ---
const oldCartSidebarStart = `<div class="w-full lg:w-[420px] xl:w-[480px] bg-white dark:bg-slate-800 border-t lg:border-t-0 lg:border-l border-slate-200 dark:border-slate-700/50 flex flex-col h-[50vh] lg:h-full shrink-0 shadow-2xl relative z-20">`;
const newCartSidebarStart = `
  <!-- Mobile Cart Overlay Background -->
  <div *ngIf="facade.isMobileCartOpen()" (click)="facade.isMobileCartOpen.set(false)" class="lg:hidden fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-[90] animate-in fade-in duration-200"></div>

  <!-- Cart Sidebar / Mobile Bottom Sheet -->
  <div class="fixed lg:static inset-x-0 bottom-0 bg-white dark:bg-slate-800 border-t lg:border-t-0 lg:border-l border-slate-200 dark:border-slate-700/50 flex flex-col h-[85vh] lg:h-full lg:w-[420px] xl:w-[480px] shrink-0 shadow-2xl z-[100] lg:z-20 transform transition-transform duration-300 ease-out lg:translate-y-0 rounded-t-3xl lg:rounded-none" [class.translate-y-full]="!facade.isMobileCartOpen()">
    
    <!-- Mobile Sheet Handle -->
    <div class="w-full flex justify-center py-3 lg:hidden shrink-0 cursor-pointer" (click)="facade.isMobileCartOpen.set(false)">
      <div class="w-12 h-1.5 bg-slate-300 dark:bg-slate-600 rounded-full"></div>
    </div>
`;
code = code.replace(oldCartSidebarStart, newCartSidebarStart);

// Let's also adjust the Grid padding to account for the bottom nav on mobile
code = code.replace('<div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-4 pb-24">', '<div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-4 pb-40 lg:pb-24">');

// --- 4. Add Bottom Nav and Floating Cart Button ---
const bottomElements = `
  <!-- Mobile Floating Cart Button -->
  <div class="fixed bottom-[80px] left-4 right-4 z-40 lg:hidden">
    <button (click)="facade.isMobileCartOpen.set(true)" class="w-full h-14 bg-indigo-600 rounded-2xl shadow-xl shadow-indigo-900/20 text-white flex items-center justify-between px-5 transition-transform active:scale-95">
      <div class="flex items-center gap-3">
        <div class="w-8 h-8 bg-white/20 rounded-lg flex items-center justify-center font-black">
          {{ facade.cart().length }}
        </div>
        <span class="font-black tracking-wide">View Cart</span>
      </div>
      <span class="font-black text-lg">{{ facade.cartTotal() | appCurrency }}</span>
    </button>
  </div>

  <!-- Mobile Bottom Navigation -->
  <div class="fixed bottom-0 inset-x-0 h-[65px] bg-white dark:bg-slate-900 border-t border-slate-200 dark:border-slate-800 shadow-[0_-10px_20px_rgba(0,0,0,0.05)] z-40 lg:hidden flex items-center justify-around px-2">
    <button class="flex flex-col items-center justify-center gap-1 text-indigo-600 w-16">
      <span class="material-symbols-outlined text-[24px]">dashboard</span>
      <span class="text-[9px] font-bold tracking-wider">SALE</span>
    </button>
    <button (click)="facade.isParkedSalesModalOpen.set(true)" class="flex flex-col items-center justify-center gap-1 text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-white w-16 relative">
      <span class="material-symbols-outlined text-[24px]">pause_presentation</span>
      <span class="text-[9px] font-bold tracking-wider">HOLD</span>
      <div *ngIf="facade.parkedSales().length > 0" class="absolute top-0 right-2 w-3.5 h-3.5 bg-red-500 rounded-full text-[8px] font-black text-white flex items-center justify-center">{{ facade.parkedSales().length }}</div>
    </button>
    <button (click)="facade.isShiftModalOpen.set(true)" class="flex flex-col items-center justify-center gap-1 text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-white w-16 relative">
      <span class="material-symbols-outlined text-[24px]">local_activity</span>
      <span class="text-[9px] font-bold tracking-wider">SHIFT</span>
      <div *ngIf="facade.shiftStatus() === 'open'" class="absolute top-0 right-2 w-2 h-2 bg-emerald-500 rounded-full animate-pulse"></div>
    </button>
    <button (click)="facade.isHardwareModalOpen.set(true)" class="flex flex-col items-center justify-center gap-1 text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-white w-16">
      <span class="material-symbols-outlined text-[24px]">devices</span>
      <span class="text-[9px] font-bold tracking-wider">HUB</span>
    </button>
  </div>
</div>
`;

// Replace the closing div of #pos-app with the bottom elements + closing div
code = code.replace(/<\/div>[\s\n]*$/g, bottomElements);

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos/pos.html', code);
