export interface Tenant { 
  id: string; 
  name?: string;
  domain?: string; 
  created_at?: string; 
  status?: string; 
  [key: string]: any; 
}
