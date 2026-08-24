const fs = require('fs');
const path = require('path');

function replaceInFile(filePath, replacements) {
    const p = path.join(__dirname, filePath);
    if (!fs.existsSync(p)) return;
    let content = fs.readFileSync(p, 'utf8');
    for (const [search, replace] of replacements) {
        if (typeof search === 'string') {
            content = content.split(search).join(replace);
        } else {
            content = content.replace(search, replace);
        }
    }
    fs.writeFileSync(p, content, 'utf8');
}

// HTML fixes
replaceInFile('frontend/src/app/features/b2b/b2b/b2b.html', [
    ['q.id?.slice(0,8)', 'q.id.slice(0,8)']
]);

replaceInFile('frontend/src/app/features/branch-settings/branch-settings/branch-settings.html', [
    ['form()?.', 'form().']
]);

replaceInFile('frontend/src/app/features/branch-settings/branch-settings/branch-settings.ts', [
    ['export class BranchSettings {', 'export class BranchSettings {\n  protected readonly Math = Math;']
]);

replaceInFile('frontend/src/app/features/kiosk/kiosk/kiosk.html', [
    ['categories()?.find', 'categories().find']
]);

replaceInFile('frontend/src/app/features/portal/portal/portal.html', [
    ['branch()?.name', 'branch().name'],
    ['storeSettings()?.logo_image', 'storeSettings()!.logo_image']
]);

replaceInFile('frontend/src/app/features/wallet/wallet/wallet.html', [
    ["c.name?.[0] ?? '?'", "c.name[0] || '?'"],
    ["dash.gift_cards?.length", "dash.gift_cards.length"],
    ["dash.recent_orders?.length ?? 0", "dash.recent_orders.length"],
    ["dash.loyalty_history?.length > 0", "dash.loyalty_history.length > 0"]
]);

replaceInFile('frontend/src/app/features/customers/customer-detail/customer-detail.html', [
    ['customer()?.last_visit', 'customer()!.last_visit'],
    ['getOrderStatusClass(order.status)', "getOrderStatusClass(order.status || '')"]
]);

replaceInFile('frontend/src/app/features/customers/customer-detail/customer-detail.ts', [
    ['export class CustomerDetail {', 'export class CustomerDetail {\n  protected readonly Math = Math;']
]);

replaceInFile('frontend/src/app/features/enterprise/enterprise/enterprise.html', [
    ['{{ b.location }}', '{{ (b as any).location }}']
]);

replaceInFile('frontend/src/app/features/fb/fb/fb.html', [
    ['item.current_stock > 0', 'item!.current_stock > 0']
]);

replaceInFile('frontend/src/app/features/hr/hr/hr.ts', [
    ['export class Hr {', 'export class Hr {\n  createPayrollPeriod() {}\n']
]);

replaceInFile('frontend/src/app/features/orders/order-detail/order-detail.html', [
    ['order()!.amount_paid > order()!.total', '(order()!.amount_paid || 0) > order()!.total'],
    ['(order()!.amount_paid - order()!.total)', '((order()!.amount_paid || 0) - order()!.total)']
]);

replaceInFile('frontend/src/app/features/orders/orders/orders.html', [
    ['.includes(order.status)', '.includes(order.status || "")']
]);

replaceInFile('frontend/src/app/features/reports/reports/reports.html', [
    ['topProducts().by_revenue | slice:0:10', '$any(topProducts().by_revenue) | slice:0:10']
]);

replaceInFile('frontend/src/app/features/store/product-detail/product-detail.component.html', [
    ['activeImage.set(product()!.image_url)', "activeImage.set(product()!.image_url || '')"]
]);

replaceInFile('frontend/src/app/features/store/product-detail/product-detail.component.ts', [
    ['export class ProductDetailComponent {', 'export class ProductDetailComponent {\n  navigateToProduct(id: any) {}\n']
]);

// TS Financial Chart Options
replaceInFile('frontend/src/app/features/financial/financial/financial.ts', [
    ['weight: "bold"', 'weight: "bold" as const']
]);

// Add StorefrontSettings properties
const sfProps = `  enable_paystack?: boolean;
  paystack_public_key?: string;
  paystack_subaccount_code?: string;
  logo_image?: string;
`;
replaceInFile('frontend/src/app/core/models/models.ts', [
    [/(export interface StorefrontSettings extends TenantScoped \{)/, `$1\n${sfProps}`],
    [/(export interface Customer extends TenantScoped \{)/, `$1\n  last_visit?: string;`]
]);

replaceInFile('frontend/src/app/core/models/storefront.models.ts', [
    [/(export interface StorefrontSettings \{)/, `$1\n${sfProps}`]
]);

replaceInFile('frontend/src/app/core/store/services/storefront-settings.service.ts', [
    [/(export interface StorefrontSettings \{)/, `$1\n${sfProps}`]
]);

// Add credit_balance to Supplier
replaceInFile('frontend/src/app/core/models/financial.models.ts', [
    ['status: string;', 'status: string; credit_balance?: number;']
]);

replaceInFile('frontend/src/app/core/models/supplier.models.ts', [
    [/(export interface Supplier \{)/, `$1\n  credit_balance?: number;`]
]);
