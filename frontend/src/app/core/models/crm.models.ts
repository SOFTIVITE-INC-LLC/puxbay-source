export interface Campaign { 
  id: string; 
  name: string; 
  target_audience?: string; 
  budget?: number; 
  type?: string; 
  [key: string]: any; 
}
export interface Customer { 
  id: string; 
  user_id: string; 
  branch_id: string; 
  tier_id: string; 
  [key: string]: any; 
}
