export interface User {
  id: number
  phone: string
  user_type: 1 | 2 | 3
  company_name: string
  credit_score: number
  status: 0 | 1
  created_at: string
  updated_at: string
}

export interface Project {
  id: number
  user_id: number
  project_type: string
  title: string
  budget_min: number
  budget_max: number
  status: 0 | 1 | 2 | 3 | 4
  publish_time?: string
  deadline?: string
  created_at: string
  updated_at: string
}

export interface Order {
  id: number
  project_id: number
  supplier_id: number
  amount: number
  status: 0 | 1 | 2 | 3 | 4
  signed_at?: string
  completed_at?: string
  created_at: string
  updated_at: string
}

export interface Deliverable {
  id: number
  order_id: number
  milestone: string
  file_url: string
  version: number
  status: 0 | 1 | 2
  created_at: string
}

export interface Payout {
  id: number
  user_id: number
  order_id?: number
  amount: number
  type: 'recharge' | 'withdraw' | 'payment' | 'refund'
  status: 0 | 1 | 2
  created_at: string
}
