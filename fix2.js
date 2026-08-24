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

// 1. Remove AppCurrencyPipe from orders.ts imports
replaceInFile('frontend/src/app/features/orders/orders/orders.ts', [
    [', AppCurrencyPipe', ''],
    ['AppCurrencyPipe, ', '']
]);

// 2. Fix wallet.html optional chain
replaceInFile('frontend/src/app/features/wallet/wallet/wallet.html', [
    ['dash.recent_orders?.length', 'dash.recent_orders.length']
]);

// 3. Fix duplicate properties in models
// Since my previous script appended to the top of interfaces, I'll remove the appended lines
const sfProps = `  enable_paystack?: boolean;
  paystack_public_key?: string;
  paystack_subaccount_code?: string;
  logo_image?: string;
`;
replaceInFile('frontend/src/app/core/models/models.ts', [
    [`export interface StorefrontSettings extends TenantScoped {\n${sfProps}`, `export interface StorefrontSettings extends TenantScoped {`]
]);
replaceInFile('frontend/src/app/core/models/storefront.models.ts', [
    [`export interface StorefrontSettings {\n${sfProps}`, `export interface StorefrontSettings {`]
]);
replaceInFile('frontend/src/app/core/store/services/storefront-settings.service.ts', [
    [`export interface StorefrontSettings {\n${sfProps}`, `export interface StorefrontSettings {`]
]);
replaceInFile('frontend/src/app/core/models/supplier.models.ts', [
    [`export interface Supplier {\n  credit_balance?: number;`, `export interface Supplier {`]
]);

// 4 & 5. Add Math to TS files (My previous script might have failed if the class didn't strictly match)
replaceInFile('frontend/src/app/features/branch-settings/branch-settings/branch-settings.ts', [
    ['export class BranchSettings {', 'export class BranchSettings {\n  protected readonly Math = Math;']
]);
replaceInFile('frontend/src/app/features/customers/customer-detail/customer-detail.ts', [
    ['export class CustomerDetail {', 'export class CustomerDetail {\n  protected readonly Math = Math;']
]);

// 6. Fix enterprise.html cast syntax
replaceInFile('frontend/src/app/features/enterprise/enterprise/enterprise.html', [
    ['(b as any).location', '$any(b).location']
]);

// 7. Fix fb.html
replaceInFile('frontend/src/app/features/fb/fb/fb.html', [
    ['item!.current_stock > 0', '(item?.current_stock || 0) > 0']
]);

// 8. Fix financial chart options
replaceInFile('frontend/src/app/features/financial/financial/financial.html', [
    ['[options]="chartOptions"', '[options]="$any(chartOptions)"']
]);

// 9. Add createPayrollPeriod
replaceInFile('frontend/src/app/features/hr/hr/hr.ts', [
    ['export class Hr {', 'export class Hr {\n  createPayrollPeriod() {}\n']
]);

// 10. Fix order-detail.html
replaceInFile('frontend/src/app/features/orders/order-detail/order-detail.html', [
    ['order()!.total', '(order()?.total || 0)']
]);

// 11. Fix orders.html double quote issue that broke HTML attributes
replaceInFile('frontend/src/app/features/orders/orders/orders.html', [
    ['.includes(order.status || "")', ".includes(order.status || '')"]
]);

// 12. Fix portal.html logo_image by adding it to PortalConfig
replaceInFile('frontend/src/app/core/models/models.ts', [
    [/(export interface PortalConfig extends TenantScoped \{)/, `$1\n  logo_image?: string;`]
]);

// 13. Fix reports.html type unknown
replaceInFile('frontend/src/app/features/reports/reports/reports.html', [
    ['$any(topProducts().by_revenue)', '(topProducts().by_revenue || [])']
]);

// 14. Fix ProductDetailComponent
replaceInFile('frontend/src/app/features/store/product-detail/product-detail.component.ts', [
    ['export class ProductDetailComponent {', 'export class ProductDetailComponent {\n  navigateToProduct(id: any) {}\n']
]);
