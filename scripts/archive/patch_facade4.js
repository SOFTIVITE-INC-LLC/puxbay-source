const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', 'utf8');

code = code.replace("import { CategoryService } from '../../core/services/category.service';", "import { CategoryService } from '../../core/services/category.service';\nimport { SettingsService } from '../../core/services/settings.service';");

code = code.replace("private toastr = inject(ToastrService);", "private toastr = inject(ToastrService);\n  private settingsService = inject(SettingsService);");

code = code.replace("  // PROMO ENGINE: Auto 10% off over $100 subtotal\n  readonly autoPromoDiscount = computed(() => {\n     return this.cartSubtotal() > 100 ? this.cartSubtotal() * 0.1 : 0;\n  });", "  // PROMO ENGINE: Dynamic settings\n  readonly promoThreshold = computed(() => this.settingsService.settings()?.promo_threshold || 100);\n  readonly promoPercent = computed(() => this.settingsService.settings()?.promo_discount_percent || 10);\n  readonly autoPromoDiscount = computed(() => {\n     const threshold = this.promoThreshold();\n     const percent = this.promoPercent();\n     return (this.cartSubtotal() > threshold && percent > 0) ? this.cartSubtotal() * (percent / 100) : 0;\n  });");

code = code.replace("this.categoryService.getCategories().subscribe();", "this.categoryService.getCategories().subscribe();\n    this.settingsService.getSettings().subscribe();");

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/features/pos/pos.facade.ts', code);
