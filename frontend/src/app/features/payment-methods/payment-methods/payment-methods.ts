import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import {
  PaymentMethodService,
  PaymentMethod,
  PaystackSubaccount,
  PaystackCountry,
  PaystackBank
} from '../../../core/services/payment-method.service';
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

// ── Wizard Steps ───────────────────────────────────────────────────
type SubaccountStep = 1 | 2 | 3 | 4 | 5;

interface SubaccountWizard {
  step: SubaccountStep;
  // Step 1 – Country
  country: PaystackCountry | null;
  countries: PaystackCountry[];
  isLoadingCountries: boolean;
  // Step 2 – Bank
  bank: PaystackBank | null;
  banks: PaystackBank[];
  isLoadingBanks: boolean;
  bankSearch: string;
  // Step 3 – Account Number
  accountNumber: string;
  // Step 4 – Verify Name
  resolvedName: string;
  resolvedAccountNumber: string;
  isResolving: boolean;
  // Step 5 – Create & Save
  businessName: string;
  percentageCharge: number;
  description: string;
  contactEmail: string;
  localName: string;
  isActive: boolean;
  isCreating: boolean;
  createdSubaccount: PaystackSubaccount | null;
}

function freshWizard(): SubaccountWizard {
  return {
    step: 1,
    country: null,
    countries: [],
    isLoadingCountries: false,
    bank: null,
    banks: [],
    isLoadingBanks: false,
    bankSearch: '',
    accountNumber: '',
    resolvedName: '',
    resolvedAccountNumber: '',
    isResolving: false,
    businessName: '',
    percentageCharge: 0,
    description: '',
    contactEmail: '',
    localName: '',
    isActive: true,
    isCreating: false,
    createdSubaccount: null,
  };
}

@Component({
  selector: 'app-payment-methods',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './payment-methods.html',
})
export class PaymentMethods implements OnInit {
  methodService = inject(PaymentMethodService);
  private toast = inject(ToastService);

  // ── Filters & State ────────────────────────────────────────────
  searchQuery = signal<string>('');
  selectedFilter = signal<'all' | 'active' | 'in_person' | 'mobile' | 'online' | 'paystack'>('all');

  // ── Modal State ────────────────────────────────────────────────
  isModalOpen = signal<boolean>(false);
  isSaving = signal<boolean>(false);
  editingId = signal<string | null>(null);

  // ── Wizard (Paystack Subaccount Creation) ─────────────────────
  wizard = signal<SubaccountWizard>(freshWizard());

  // ── Legacy Subaccount Verify State (for edit) ─────────────────
  paystackSubaccounts = signal<PaystackSubaccount[]>([]);
  isLoadingSubaccounts = signal<boolean>(false);
  isVerifyingSubaccount = signal<boolean>(false);
  verifiedSubaccount = signal<PaystackSubaccount | null>(null);
  subaccountInputMode = signal<'dropdown' | 'manual'>('dropdown');

  // ── Delete Confirm ─────────────────────────────────────────────
  deleteConfirmItem = signal<PaymentMethod | null>(null);

  // ── Form for Non-Subaccount Methods ───────────────────────────
  formData = {
    name: '',
    provider: 'cash' as string,
    is_active: true,
    api_key_hint: '',
    paystack_subaccount_code: ''
  };
  form = () => this.formData;

  // ── Presets ────────────────────────────────────────────────────
  readonly presets: ProviderPreset[] = [
    {
      provider: 'paystack_subaccount',
      name: 'Paystack Subaccount',
      category: 'split',
      icon: 'hub',
      description: 'Route store payouts into your Paystack merchant subaccount with real-time bank verification.',
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
      description: 'Chip & PIN, tap- Visa and Mastercard payments.',
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

  // ── Computed ───────────────────────────────────────────────────
  totalMethods = computed(() => this.methodService.methods().length);
  activeMethodsCount = computed(() => this.methodService.methods().filter(m => m.is_active).length);
  inactiveMethodsCount = computed(() => this.totalMethods() - this.activeMethodsCount());

  filteredMethods = computed(() => {
    const list = this.methodService.methods();
    const query = this.searchQuery().toLowerCase().trim();
    const filter = this.selectedFilter();

    return list.filter(item => {
      const matchesSearch = !query ||
        item.name.toLowerCase().includes(query) ||
        item.provider.toLowerCase().includes(query) ||
        (item.paystack_subaccount_code && item.paystack_subaccount_code.toLowerCase().includes(query));

      let matchesFilter = true;
      if (filter === 'active') matchesFilter = item.is_active;
      else if (filter === 'in_person') matchesFilter = item.provider === 'cash' || item.provider === 'card';
      else if (filter === 'mobile') matchesFilter = item.provider === 'mobile';
      else if (filter === 'online') matchesFilter = ['stripe', 'paystack', 'paystack_subaccount'].includes(item.provider);
      else if (filter === 'paystack') matchesFilter = item.provider === 'paystack' || item.provider === 'paystack_subaccount';

      return matchesSearch && matchesFilter;
    });
  });

  filteredBanks = computed(() => {
    const q = this.wizard().bankSearch.toLowerCase();
    if (!q) return this.wizard().banks;
    return this.wizard().banks.filter(b => b.name.toLowerCase().includes(q));
  });

  ngOnInit() {
    this.methodService.getMethods().subscribe({
      error: () => this.toast.showError('Failed to load payment methods.')
    });
  }

  getProviderInfo(provider: string) {
    switch (provider.toLowerCase()) {
      case 'paystack_subaccount': return { icon: 'hub', label: 'Paystack Subaccount' };
      case 'paystack':            return { icon: 'bolt', label: 'Paystack' };
      case 'cash':                return { icon: 'payments', label: 'Cash' };
      case 'mobile':              return { icon: 'phone_android', label: 'Mobile Money' };
      case 'card':                return { icon: 'credit_card', label: 'Card / POS' };
      case 'stripe':              return { icon: 'public', label: 'Stripe' };
      case 'bank_transfer':       return { icon: 'account_balance', label: 'Bank Wire' };
      case 'crypto':              return { icon: 'currency_bitcoin', label: 'Crypto' };
      default:                    return { icon: 'tune', label: 'Custom' };
    }
  }

  onToggleStatus(method: PaymentMethod, event: Event) {
    event.stopPropagation();
    const newStatus = !method.is_active;
    this.methodService.toggleMethod(method.id, newStatus).subscribe({
      next: () => this.toast.showSuccess(`${method.name} is now ${newStatus ? 'active' : 'disabled'}.`),
      error: () => {
        this.methodService.toggleMethod(method.id, !newStatus).subscribe();
        this.toast.showError(`Failed to update ${method.name}.`);
      }
    });
  }

  // ── Modal Open / Close ─────────────────────────────────────────

  openCreateModal() {
    this.editingId.set(null);
    this.formData = { name: '', provider: 'paystack_subaccount', is_active: true, api_key_hint: '', paystack_subaccount_code: '' };
    this.wizard.set(freshWizard());
    this.verifiedSubaccount.set(null);
    this.isModalOpen.set(true);

    // Kick off countries load for wizard
    this._loadCountries();
  }

  openEditModal(method: PaymentMethod, event: Event) {
    event.stopPropagation();
    this.editingId.set(method.id);
    this.wizard.set(freshWizard());
    this.verifiedSubaccount.set(null);
    this.formData = {
      name: method.name,
      provider: method.provider || 'custom',
      is_active: method.is_active,
      api_key_hint: method.api_key_hint || '',
      paystack_subaccount_code: method.paystack_subaccount_code || ''
    };
    if (method.provider === 'paystack_subaccount' && method.paystack_subaccount_code) {
      this._verifyForEdit(method.paystack_subaccount_code);
    }
    this.isModalOpen.set(true);
  }

  closeModal() {
    this.isModalOpen.set(false);
    this.editingId.set(null);
    this.verifiedSubaccount.set(null);
    this.wizard.set(freshWizard());
  }

  // ── Wizard: Step Navigation ───────────────────────────────────

  private _loadCountries() {
    this.wizard.update(w => ({ ...w, isLoadingCountries: true }));
    this.methodService.getPaystackCountries().subscribe({
      next: res => {
        const countries = res?.data ?? [];
        // Filter to only countries where Paystack supports subaccounts (those with integration_type === "full")
        this.wizard.update(w => ({ ...w, countries, isLoadingCountries: false }));
      },
      error: () => {
        this.wizard.update(w => ({ ...w, isLoadingCountries: false }));
        this.toast.showError('Failed to load supported countries from Paystack.');
      }
    });
  }

  wizardSelectCountry(country: PaystackCountry) {
    this.wizard.update(w => ({
      ...w,
      country,
      bank: null,
      banks: [],
      bankSearch: '',
      accountNumber: '',
      resolvedName: '',
      step: 2,
      isLoadingBanks: true
    }));
    this.methodService.getPaystackBanks(country.name.toLowerCase(), country.default_currency_code).subscribe({
      next: res => {
        const banks = (res?.data ?? []).filter((b: PaystackBank) => b.active);
        this.wizard.update(w => ({ ...w, banks, isLoadingBanks: false }));
      },
      error: () => {
        this.wizard.update(w => ({ ...w, isLoadingBanks: false }));
        this.toast.showError('Failed to load banks for this country.');
      }
    });
  }

  wizardSelectBank(bank: PaystackBank) {
    this.wizard.update(w => ({ ...w, bank, accountNumber: '', resolvedName: '', step: 3 }));
  }

  wizardGoToStep(step: SubaccountStep) {
    // Allow going back freely
    const current = this.wizard().step;
    if (step < current) {
      this.wizard.update(w => ({ ...w, step }));
    }
  }

  wizardResolveAccount() {
    const w = this.wizard();
    if (!w.bank || !w.accountNumber.trim()) {
      this.toast.showWarning('Please enter a valid account number.');
      return;
    }
    if (w.accountNumber.length < 8) {
      this.toast.showWarning('Account number must be at least 8 digits.');
      return;
    }
    this.wizard.update(w2 => ({ ...w2, isResolving: true, resolvedName: '' }));
    this.methodService.resolvePaystackAccount(w.accountNumber.trim(), w.bank.code).subscribe({
      next: res => {
        this.wizard.update(w2 => ({
          ...w2,
          isResolving: false,
          resolvedName: res.account_name,
          resolvedAccountNumber: res.account_number,
          step: 4,
          businessName: w2.businessName || res.account_name
        }));
      },
      error: (err) => {
        this.wizard.update(w2 => ({ ...w2, isResolving: false }));
        const msg = err?.error?.error || 'Could not resolve account. Please check the account number.';
        this.toast.showError(msg);
      }
    });
  }

  wizardConfirmName() {
    const w = this.wizard();
    if (!w.resolvedName) return;
    this.wizard.update(w2 => ({
      ...w2,
      step: 5,
      businessName: w2.businessName || w2.resolvedName
    }));
  }

  wizardCreate() {
    const w = this.wizard();
    if (!w.bank || !w.resolvedName || !w.businessName.trim()) {
      this.toast.showWarning('Please complete all required fields.');
      return;
    }

    this.wizard.update(w2 => ({ ...w2, isCreating: true }));

    this.methodService.createPaystackSubaccount({
      business_name: w.businessName.trim(),
      settlement_bank: w.bank.code,
      account_number: w.resolvedAccountNumber || w.accountNumber.trim(),
      percentage_charge: w.percentageCharge,
      description: w.description.trim(),
      primary_contact_email: w.contactEmail.trim(),
      local_name: w.localName.trim() || `Paystack – ${w.businessName.trim()} (${w.bank.name})`,
      is_active: w.isActive
    }).subscribe({
      next: res => {
        this.wizard.update(w2 => ({
          ...w2,
          isCreating: false,
          createdSubaccount: res.subaccount,
          step: 5
        }));
        this.toast.showSuccess(`Subaccount ${res.subaccount.subaccount_code} created successfully!`);
        // Close modal after 1.5s to let user see the success
        setTimeout(() => this.closeModal(), 1800);
      },
      error: (err) => {
        this.wizard.update(w2 => ({ ...w2, isCreating: false }));
        const msg = err?.error?.error || 'Failed to create Paystack subaccount.';
        this.toast.showError(msg);
      }
    });
  }

  // ── Edit-mode legacy verify ────────────────────────────────────

  private _verifyForEdit(code: string) {
    this.isVerifyingSubaccount.set(true);
    this.methodService.verifyPaystackSubaccount(code).subscribe({
      next: res => {
        this.isVerifyingSubaccount.set(false);
        if (res?.subaccount) this.verifiedSubaccount.set(res.subaccount);
      },
      error: () => this.isVerifyingSubaccount.set(false)
    });
  }

  verifyCode(codeToVerify?: string, showNotification = true) {
    const code = codeToVerify || this.formData.paystack_subaccount_code.trim();
    if (!code) {
      if (showNotification) this.toast.showWarning('Enter a Paystack subaccount code (ACCT_...)');
      return;
    }
    this.isVerifyingSubaccount.set(true);
    this.methodService.verifyPaystackSubaccount(code).subscribe({
      next: res => {
        this.isVerifyingSubaccount.set(false);
        if (res?.subaccount) {
          this.verifiedSubaccount.set(res.subaccount);
          this.formData.paystack_subaccount_code = res.subaccount.subaccount_code;
          if (!this.formData.name) {
            this.formData.name = `Paystack – ${res.subaccount.business_name} (${res.subaccount.settlement_bank})`;
          }
          if (showNotification) this.toast.showSuccess(`Verified: ${res.subaccount.business_name}`);
        }
      },
      error: (err) => {
        this.isVerifyingSubaccount.set(false);
        this.verifiedSubaccount.set(null);
        if (showNotification) this.toast.showError(err?.error?.error || 'Could not verify subaccount code.');
      }
    });
  }

  // ── Quick Add Preset ───────────────────────────────────────────

  quickAddPreset(preset: ProviderPreset) {
    if (preset.provider === 'paystack_subaccount') {
      this.openCreateModal();
      return;
    }
    const exists = this.methodService.methods().some(m => m.name.toLowerCase() === preset.name.toLowerCase());
    if (exists) { this.toast.showInfo(`${preset.name} is already configured.`); return; }
    this.isSaving.set(true);
    this.methodService.createMethod({ name: preset.name, provider: preset.provider, is_active: true }).subscribe({
      next: () => { this.isSaving.set(false); this.toast.showSuccess(`${preset.name} added!`); },
      error: () => { this.isSaving.set(false); this.toast.showError(`Failed to add ${preset.name}.`); }
    });
  }

  seedStandardSuite() {
    this.isSaving.set(true);
    const presets = [
      { name: 'Cash', provider: 'cash' },
      { name: 'MTN Mobile Money', provider: 'mobile' },
      { name: 'Telecel Cash', provider: 'mobile' },
      { name: 'Card (Visa / Mastercard)', provider: 'card' }
    ];
    let done = 0;
    const current = this.methodService.methods();
    presets.forEach(p => {
      if (current.some(m => m.name.toLowerCase() === p.name.toLowerCase())) {
        done++;
        if (done >= presets.length) { this.isSaving.set(false); this.toast.showSuccess('Suite ready.'); }
        return;
      }
      this.methodService.createMethod({ ...p, is_active: true }).subscribe({
        next: () => { done++; if (done >= presets.length) { this.isSaving.set(false); this.toast.showSuccess('Core payment methods configured!'); } }
      });
    });
  }

  // ── Save Form (non-subaccount / edit) ─────────────────────────

  saveForm() {
    const data = this.formData;
    if (!data.name.trim()) { this.toast.showWarning('Display name is required.'); return; }
    if (data.provider === 'paystack_subaccount' && !data.paystack_subaccount_code.trim()) {
      this.toast.showWarning('Please verify a Paystack subaccount code.'); return;
    }
    this.isSaving.set(true);
    const id = this.editingId();
    const payload: Partial<PaymentMethod> = {
      name: data.name.trim(),
      provider: data.provider,
      is_active: data.is_active,
      api_key_hint: data.api_key_hint.trim() || undefined,
      paystack_subaccount_code: data.paystack_subaccount_code.trim() || undefined
    };
    const obs = id
      ? this.methodService.updateMethod(id, payload)
      : this.methodService.createMethod(payload);
    obs.subscribe({
      next: () => { this.isSaving.set(false); this.closeModal(); this.toast.showSuccess(id ? 'Method updated.' : 'Method created.'); },
      error: () => { this.isSaving.set(false); this.toast.showError(id ? 'Failed to update.' : 'Failed to create.'); }
    });
  }

  // ── Delete ─────────────────────────────────────────────────────

  promptDelete(method: PaymentMethod, event: Event) {
    event.stopPropagation();
    this.deleteConfirmItem.set(method);
  }

  cancelDelete() { this.deleteConfirmItem.set(null); }

  confirmDelete() {
    const item = this.deleteConfirmItem();
    if (!item) return;
    this.methodService.deleteMethod(item.id).subscribe({
      next: () => { this.deleteConfirmItem.set(null); this.toast.showSuccess(`${item.name} removed.`); },
      error: () => { this.deleteConfirmItem.set(null); this.toast.showError(`Failed to delete.`); }
    });
  }
}
