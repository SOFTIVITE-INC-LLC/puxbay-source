const fs = require('fs');
let code = fs.readFileSync('/home/afari/Projects/development/softivite/puxbay/backend/internal/handlers/settings_handler.go', 'utf8');

code = code.replace(/type GlobalSettingsResponse struct \{[\s\S]*?\}/, `type GlobalSettingsResponse struct {
	Currency                string  \`json:"currency"\`
	Timezone                string  \`json:"timezone"\`
	DateFormat              string  \`json:"date_format"\`
	EnableEmailReceipts     bool    \`json:"enable_email_receipts"\`
	HardwareProxyURL        string  \`json:"hardware_proxy_url"\`
	EnableHardwareProxy     bool    \`json:"enable_hardware_proxy"\`
	AutoPrintReceipts       bool    \`json:"auto_print_receipts"\`
	EnableSMSNotifications  bool    \`json:"enable_sms_notifications"\`
	EnablePushNotifications bool    \`json:"enable_push_notifications"\`
	AdminNotificationEmail  string  \`json:"admin_notification_email"\`
	PromoThreshold          float64 \`json:"promo_threshold"\`
	PromoDiscountPercent    float64 \`json:"promo_discount_percent"\`
}`);

fs.writeFileSync('/home/afari/Projects/development/softivite/puxbay/backend/internal/handlers/settings_handler.go', code);
