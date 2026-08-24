export interface Base {
  id: string;
  created_at: string;
  updated_at: string;
}

export interface TenantScoped {

}

export interface BranchScoped {
  branch_id?: string;
}

