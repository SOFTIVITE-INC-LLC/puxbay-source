export interface AuditLog {
  tenant_id?: string;
  user_id?: string;
  action: string;
  model_name: string;
  object_id?: string | null;
  changes?: any;
  ip_address?: string | null;
  user_agent?: string | null;
}

export interface APIRequestLog {
  tenant_id?: string;
  user_id?: string;
  method: string;
  endpoint: string;
  status_code: number;
  response_time_ms: number;
  ip_address?: string | null;
  user_agent?: string | null;
  request_body?: any;
  response_body?: any;
}

export interface HoneypotAttempt {
  id: number;
  username?: string | null;
  password?: string | null;
  ip_address?: string | null;
  user_agent?: string | null;
  path: string;
  timestamp: string;
}

export interface CrossTenantAuditLog {
  user_id?: string;
  accessed_tenant_id?: string;
  user_home_tenant_id?: string;
  action_type: string;
  target_model?: string | null;
  target_object_id?: string | null;
  target_object_repr?: string | null;
  description?: string | null;
  ip_address?: string | null;
  user_agent?: string | null;
}

export interface ActivityLog {
  tenant_id: string;
  actor_id?: string;
  action_type: string;
  target_model?: string | null;
  target_object_id?: string | null;
  description: string;
  changes?: any;
  ip_address?: string | null;
}

export interface SystemLog {
  level: string;
  module: string;
  message: string;
  traceback?: string | null;
  path?: string | null;
}

