// ---- Status literal unions ----

export type ServiceStatus = 'draft' | 'creating' | 'ready' | 'degraded' | 'archived';

export type ReleaseStatus =
  | 'created'
  | 'reviewing'
  | 'approved'
  | 'deploying'
  | 'verifying'
  | 'observing'
  | 'succeeded'
  | 'failed'
  | 'rollback_pending'
  | 'rolling_back'
  | 'rolled_back';

export type ReleaseStrategy = 'direct' | 'canary' | 'blue_green';

// ---- Core entities ----

export interface Service {
  id: string;
  name: string;
  description: string;
  status: ServiceStatus;
  tech_stack: string;
  owner: string;
  repo_url: string;
  branch: string;
  language: string;
  framework: string;
  created_at: string;
  updated_at: string;
}

export interface Environment {
  id: string;
  service_id: string;
  name: string;
  cluster: string;
  namespace: string;
  replicas: number;
  status: string;
  created_at: string;
}

export interface ReleaseEvent {
  id: string;
  release_id: string;
  event_type: string;
  message: string;
  actor: string;
  created_at: string;
}

export interface Release {
  id: string;
  service_id: string;
  service_name: string;
  version: string;
  status: ReleaseStatus;
  strategy: ReleaseStrategy;
  environment: string;
  commit_sha: string;
  commit_message: string;
  image_tag: string;
  created_by: string;
  events: ReleaseEvent[];
  created_at: string;
  updated_at: string;
}

// ---- API envelope ----

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
  requestId: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  offset: number;
  limit: number;
}

// ---- Create inputs ----

export interface CreateServiceInput {
  name: string;
  description: string;
  tech_stack: string;
  owner: string;
  repo_url: string;
  branch: string;
  language: string;
  framework: string;
}

export interface CreateReleaseInput {
  service_id: string;
  environment: string;
  strategy: ReleaseStrategy;
  version: string;
  commit_sha: string;
  commit_message: string;
  image_tag: string;
}

// ---- AI ----

export interface AIPlan {
  risk_level: string;
  confidence: number;
  editable: boolean;
  estimated_duration: string;
  steps: Array<{
    order: number;
    action: string;
    description: string;
    estimated_duration: string;
    status: string;
  }>;
  recommendations: string[];
  summary: string;
}

export interface RiskAnalysis {
  risk_level: string;
  risk_score: number;
  confidence: number;
  reason: string;
  factors: string[];
  recommendations: string[];
}
