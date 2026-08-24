export interface Notification {
  user_id: string;
  title: string;
  message: string;
  link?: string;
  is_read: boolean;
  notification_type: string;
  category: string;
  user?: any | null;
}

export interface NotificationSetting {
  user_id: string;
  email_notifications: boolean;
  low_stock_alerts: boolean;
  sales_reports: boolean;
  security_alerts: boolean;
  system_alerts: boolean;
}

