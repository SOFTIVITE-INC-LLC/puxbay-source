const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/core/services/settings.service.ts', 'utf8');

code = code.replace(/export interface GlobalSettings \{[\s\S]*?\}/, `export interface GlobalSettings {
  currency: string;
  timezone: string;
  date_format: string;
  enable_email_receipts: boolean;
  hardware_proxy_url: string;
  enable_hardware_proxy: boolean;
  auto_print_receipts: boolean;
  enable_sms_notifications: boolean;
  enable_push_notifications: boolean;
  admin_notification_email: string;
  promo_threshold: number;
  promo_discount_percent: number;
}`);

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/frontend/src/app/core/services/settings.service.ts', code);
