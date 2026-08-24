export interface APIKey {
  name: string;
  key_prefix: string;
  is_active: boolean;
  is_sandbox: boolean;
  last_used_at?: string;
}

export interface ExternalSystem {
  developer_id: string;
  name: string;
  description?: string | null;
  client_id: string;
  redirect_uris?: any;
  webhook_url?: string | null;
  icon: string;
  is_public: boolean;
  is_active: boolean;
}

export interface WebhookEndpoint {
  external_system_id?: string;
  url: string;
  is_active: boolean;
  events: any;
}

export interface WebhookEvent {
  endpoint_id: string;
  event_type: string;
  payload: any;
  signature?: string | null;
  status_code?: number;
  response_body?: string | null;
  error_message?: string | null;
}

