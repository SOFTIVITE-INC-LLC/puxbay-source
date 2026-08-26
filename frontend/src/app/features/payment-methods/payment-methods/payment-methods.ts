import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { PaymentMethodService, PaymentMethod, PaystackSubaccount } from '../../../core/services/payment-method.service';
import { ToastService } from '../../../core/services/toast';

export interface ProviderPreset {
  provider: string;
  name: string;
  category: 'in_person' | 'mobile' | 'online' | 'bank' | 'split';
  icon: string;
  description: string;
  color: string;
  bg: string;
  badge: string;
}

@Component({
  selector: 'app-payment-methods',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './payment-methods.html',
  styles: `
    .glass-panel {
      background: rgba(255, 255, 255, 0.05);
      backdrop-filter: blur(12px);
      border: 1px solid rgba(255, 255, 255, 0.08);
    }
    .dark .glass-panel {
      background: rgba(15, 23, 42, 0.4);
      border-color: rgba(255, 255, 255, 0.06);
    }
  `,
})
export class PaymentMethods implements OnInit {
  methodService = inject(PaymentMethodService);
  private toast = inject(ToastService);

  // Filters & State
  searchQuery = signal<string>('');
  selectedFilter = signal<'all' | 'active' | 'in_person' | 'mobile' | 'online' | 'paystack'>('all');

  // Modal State
  isModalOpen = signal<boolean>(false);
  isSaving = signal<boolean>(false);
  editingId = signal<string | null>(null);

  // Paystack Subaccount Live Verification State
  paystackSubaccounts = signal<PaystackSubaccount[]>([]);
  isLoadingSubaccounts = signal<boolean>(false);
  isVerifyingSubaccount = signal<boolean>(false);
  verifiedSubaccount = signal<PaystackSubaccount | null>(null);
  subaccountInputMode = signal<'dropdown' | 'manual'>('dropdown');

  // Delete Confirmation Modal State
  deleteConfirmItem = signal<PaymentMethod | null>(null);

  // Form Model
  form = signal<{
    name: string;
    provider: string;
    is_active: boolean;
    api_key_hint: string;
    paystack_subaccount_code: string;
  }>({
    name: '',
    provider: 'cash',
    is_active: true,
    api_key_hint: '',
    paystack_subaccount_code: ''
  });

  // Built-in presets for quick onboarding
  readonly presets: ProviderPreset[] = [
    {
      provider: 'paystack_subaccount',
      name: 'Paystack Subaccount',
      category: 'split',
      icon: 'hub',
      description: 'Route store payouts directly into your Paystack merchant subaccount.',
      color: 'text-cyan-400',
      bg: 'bg-cyan-500/10 border-cyan-500/20',
      badge: 'Direct Split / Payout'
    },
    {
      provider: 'cash',
      name: 'Cash Payment',
      category: 'in_person',
      icon: 'payments',
      description: 'Accept physical cash payments at checkout or on delivery.',
      color: 'text-emerald-500',
      bg: 'bg-emerald-500/10 border-emerald-500/20',
      badge: 'Point of Sale'
    },
    {
      provider: 'mobile',
      name: 'MTN Mobile Money',
      category: 'mobile',
      icon: 'phone_android',
      description: 'Receive instant MoMo payments from MTN subscribers.',
      color: 'text-amber-500',
      bg: 'bg-amber-500/10 border-amber-500/20',
      badge: 'Mobile Money'
    },
    {
      provider: 'mobile',
      name: 'Telecel Cash',
      category: 'mobile',
      icon: 'smartphone',
      description: 'Direct mobile wallet payments for Telecel customers.',
      color: 'text-rose-500',
      bg: 'bg-rose-500/10 border-rose-500/20',
      badge: 'Mobile Money'
    },
    {
      provider: 'card',
      name: 'POS Card Reader',
      category: 'in_person',
      icon: 'credit_card',
      description: 'Chip & PIN, tap-to-pay Visa and Mastercard payments.',
      color: 'text-indigo-500',
      bg: 'bg-indigo-500/10 border-indigo-500/20',
      badge: 'Card / Terminal'
    },
    {
      provider: 'paystack',
      name: 'Paystack Gateway',
      category: 'online',
      icon: 'bolt',
      description: 'Unified online checkout for cards, MoMo, and QR.',
      color: 'text-cyan-500',
      bg: 'bg-cyan-500/10 border-cyan-500/20',
      badge: 'Online Gateway'
    },
    {
      provider: 'stripe',
      name: 'Stripe Global',
      category: 'online',
      icon: 'public',
      description: 'Accept international cards, Apple Pay, and Google Pay.',
      color: 'text-violet-500',
      bg: 'bg-violet-500/10 border-violet-500/20',
      badge: 'International'
    },
    {
      provider: 'bank_transfer',
      name: 'Direct Bank Wire',
      category: 'bank',
      icon: 'account_balance',
      description: 'B2B invoicing and automated bank settlement.',
      color: 'text-blue-500',
      bg: 'bg-blue-500/10 border-blue-500/20',
      badge: 'Bank Wire'
    }
  ];

  // Computed Stats
  totalMethods = computed(() => this.methodService.methods().length);
  activeMethodsCount = computed(() => this.methodService.methods().filter(m => m.is_active).length);
  inactiveMethodsCount = computed(() => this.totalMethods() - this.activeMethodsCount());

  // Filtered List
  filteredMethods = computed(() => {
    const list = this.methodService.methods();
    const query = this.searchQuery().toLowerCase().trim();
    const filter = this.selectedFilter();

    return list.filter(item => {
      // Search match
      const matchesSearch = !query ||
        item.name.toLowerCase().includes(query) ||
        item.provider.toLowerCase().includes(query) ||
        (item.paystack_subaccount_code && item.paystack_subaccount_code.toLowerCase().includes(query));

      // Status / Category match
      let matchesFilter = true;
      if (filter === 'active') {
        matchesFilter = item.is_active;
      } else if (filter === 'in_person') {
        matchesFilter = item.provider === 'cash' || item.provider === 'card';
      } else if (filter === 'mobile') {
        matchesFilter = item.provider === 'mobile';
      } else if (filter === 'online') {
        matchesFilter = item.provider === 'stripe' || item.provider === 'paystack' || item.provider === 'paystack_subaccount';
      } else if (filter === 'paystack') {
        matchesFilter = item.provider === 'paystack' || item.provider === 'paystack_subaccount';
      }

      return matchesSearch && matchesFilter;
    });
  });

  ngOnInit() {
    this.loadMethods();
  }

  loadMethods() {
    this.methodService.getMethods().subscribe({
      error: () => this.toast.showError('Failed to load payment methods.')
    });
  }

  getProviderInfo(provider: string) {
    switch (provider.toLowerCase()) {
      case 'paystack_subaccount':
        return { icon: 'hub', label: 'Paystack Subaccount', color: 'text-cyan-400', bg: 'bg-cyan-500/10 border-cyan-500/20' };
      case 'paystack':
        return { icon: 'bolt', label: 'Paystack', color: 'text-cyan-400', bg: 'bg-cyan-500/10 border-cyan-500/20' };
      case 'cash':
        return { icon: 'payments', label: 'Cash', color: 'text-emerald-400', bg: 'bg-emerald-500/10 border-emerald-500/20' };
      case 'mobile':
        return { icon: 'phone_android', label: 'Mobile Money', color: 'text-amber-400', bg: 'bg-amber-500/10 border-amber-500/20' };
      case 'card':
        return { icon: 'credit_card', label: 'Card / POS', color: 'text-indigo-400', bg: 'bg-indigo-500/10 border-indigo-500/20' };
      case 'stripe':
        return { icon: 'public', label: 'Stripe', color: 'text-violet-400', bg: 'bg-violet-500/10 border-violet-500/20' };
      case 'bank_transfer':
        return { icon: 'account_balance', label: 'Bank Wire', color: 'text-blue-400', bg: 'bg-blue-500/10 border-blue-500/20' };
      case 'crypto':
        return { icon: 'currency_bitcoin', label: 'Crypto', color: 'text-orange-400', bg: 'bg-orange-500/10 border-orange-500/20' };
      default:
        return { icon: 'tune', label: 'Custom', color: 'text-slate-400', bg: 'bg-slate-500/10 border-slate-500/20' };
    }
  }

  onToggleStatus(method: PaymentMethod, event: Event) {
    event.stopPropagation();
    const newStatus = !method.is_active;
    this.methodService.toggleMethod(method.id, newStatus).subscribe({
      next: () => {
        this.toast.showSuccess(`${method.name} is now ${newStatus ? 'active' : 'disabled'}.`);
      },
      error: () => {
        // Revert on error
        this.methodService.toggleMethod(method.id, !newStatus).subscribe();
        this.toast.showError(`Failed to update ${method.name} status.`);
      }
    });
  }

  openCreateModal() {
    this.editingId.set(null);
    this.verifiedSubaccount.set(null);
    this.form.set({
      name: '',
      provider: 'paystack_subaccount',
      is_active: true,
      api_key_hint: '',
      paystack_subaccount_code: ''
    });
    this.loadPaystackSubaccountsList();
    this.isModalOpen.set(true);
  }

  openEditModal(method: PaymentMethod, event: Event) {
    event.stopPropagation();
    this.editingId.set(method.id);
    this.verifiedSubaccount.set(null);
    this.form.set({
      name: method.name,
      provider: method.provider || 'custom',
      is_active: method.is_active,
      api_key_hint: method.api_key_hint || '',
      paystack_subaccount_code: method.paystack_subaccount_code || ''
    });

    if (method.provider === 'paystack_subaccount') {
      this.loadPaystackSubaccountsList();
      if (method.paystack_subaccount_code) {
        this.verifyCode(method.paystack_subaccount_code, false);
      }
    }
    this.isModalOpen.set(true);
  }

  closeModal() {
    this.isModalOpen.set(false);
    this.editingId.set(null);
    this.verifiedSubaccount.set(null);
  }

  loadPaystackSubaccountsList() {
    this.isLoadingSubaccounts.set(true);
    this.methodService.getPaystackSubaccounts().subscribe({
      next: (res) => {
        const list = res.subaccounts || [];
        this.paystackSubaccounts.set(list);
        this.isLoadingSubaccounts.set(false);
        if (list.length > 0 && !this.form().paystack_subaccount_code) {
          this.onSelectSubaccount(list[0]);
        }
      },
      error: () => {
        this.isLoadingSubaccounts.set(false);
        this.subaccountInputMode.set('manual');
      }
    });
  }

  onSelectSubaccount(sub: PaystackSubaccount) {
    this.verifiedSubaccount.set(sub);
    this.form.update(f => ({
      ...f,
      paystack_subaccount_code: sub.subaccount_code,
      name: f.name || `Paystack - ${sub.business_name} (${sub.settlement_bank})`
    }));
  }

  verifyCode(codeToVerify?: string, showNotification = true) {
    const code = codeToVerify || this.form().paystack_subaccount_code.trim();
    if (!code) {
      if (showNotification) this.toast.showWarning('Please enter a Paystack subaccount code (e.g. ACCT_...)');
      return;
    }

    this.isVerifyingSubaccount.set(true);
    this.methodService.verifyPaystackSubaccount(code).subscribe({
      next: (res) => {
        this.isVerifyingSubaccount.set(false);
        if (res && res.subaccount) {
          this.verifiedSubaccount.set(res.subaccount);
          this.form.update(f => ({
            ...f,
            paystack_subaccount_code: res.subaccount.subaccount_code,
            name: f.name || `Paystack - ${res.subaccount.business_name} (${res.subaccount.settlement_bank})`
          }));
          if (showNotification) this.toast.showSuccess(`Verified: ${res.subaccount.business_name} (${res.subaccount.settlement_bank})`);
        }
      },
      error: (err) => {
        this.isVerifyingSubaccount.set(false);
        this.verifiedSubaccount.set(null);
        if (showNotification) {
          const errMsg = err.error?.error || 'Subaccount code could not be verified on Paystack.';
          this.toast.showError(errMsg);
        }
      }
    });
  }

  quickAddPreset(preset: ProviderPreset) {
    if (preset.provider === 'paystack_subaccount') {
      this.openCreateModal();
      return;
    }

    const exists = this.methodService.methods().some(m => m.name.toLowerCase() === preset.name.toLowerCase());
    if (exists) {
      this.toast.showInfo(`${preset.name} is already configured.`);
      return;
    }

    this.isSaving.set(true);
    this.methodService.createMethod({
      name: preset.name,
      provider: preset.provider,
      is_active: true
    }).subscribe({
      next: () => {
        this.isSaving.set(false);
        this.toast.showSuccess(`${preset.name} added successfully!`);
      },
      error: () => {
        this.isSaving.set(false);
        this.toast.showError(`Failed to add ${preset.name}.`);
      }
    });
  }

  seedStandardSuite() {
    this.isSaving.set(true);
    const standardPresets = [
      { name: 'Cash', provider: 'cash' },
      { name: 'MTN Mobile Money', provider: 'mobile' },
      { name: 'Telecel Cash', provider: 'mobile' },
      { name: 'Card (Visa / Mastercard)', provider: 'card' }
    ];

    let completed = 0;
    const current = this.methodService.methods();

    standardPresets.forEach(preset => {
      const exists = current.some(m => m.name.toLowerCase() === preset.name.toLowerCase());
      if (!exists) {
        this.methodService.createMethod({ ...preset, is_active: true }).subscribe({
          next: () => {
            completed++;
            if (completed >= standardPresets.length) {
              this.isSaving.set(false);
              this.toast.showSuccess('Core payment methods configured!');
            }
          }
        });
      } else {
        completed++;
        if (completed >= standardPresets.length) {
          this.isSaving.set(false);
          this.toast.showSuccess('Payment setup updated.');
        }
      }
    });
  }

  saveForm() {
    const data = this.form();
    if (!data.name.trim()) {
      this.toast.showWarning('Payment method name is required.');
      return;
    }

    if (data.provider === 'paystack_subaccount' && !data.paystack_subaccount_code.trim()) {
      this.toast.showWarning('Please select or verify a Paystack subaccount.');
      return;
    }

    this.isSaving.set(true);
    const editId = this.editingId();

    const payload: Partial<PaymentMethod> = {
      name: data.name.trim(),
      provider: data.provider,
      is_active: data.is_active,
      api_key_hint: data.api_key_hint.trim() || undefined,
      paystack_subaccount_code: data.paystack_subaccount_code.trim() || undefined
    };

    if (editId) {
      this.methodService.updateMethod(editId, payload).subscribe({
        next: () => {
          this.isSaving.set(false);
          this.closeModal();
          this.toast.showSuccess('Payment method updated.');
        },
        error: () => {
          this.isSaving.set(false);
          this.toast.showError('Failed to update payment method.');
        }
      });
    } else {
      this.methodService.createMethod(payload).subscribe({
        next: () => {
          this.isSaving.set(false);
          this.closeModal();
          this.toast.showSuccess('Payment method created.');
        },
        error: () => {
          this.isSaving.set(false);
          this.toast.showError('Failed to create payment method.');
        }
      });
    }
  }

  promptDelete(method: PaymentMethod, event: Event) {
    event.stopPropagation();
    this.deleteConfirmItem.set(method);
  }

  cancelDelete() {
    this.deleteConfirmItem.set(null);
  }

  confirmDelete() {
    const item = this.deleteConfirmItem();
    if (!item) return;

    this.methodService.deleteMethod(item.id).subscribe({
      next: () => {
        this.deleteConfirmItem.set(null);
        this.toast.showSuccess(`${item.name} removed.`);
      },
      error: () => {
        this.deleteConfirmItem.set(null);
        this.toast.showError(`Failed to delete ${item.name}.`);
      }
    });
  }
}
