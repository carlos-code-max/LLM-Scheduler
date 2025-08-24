// API 响应基础结构
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data?: T;
}

// 分页响应结构
export interface PagedResponse<T = any> {
  code: number;
  message: string;
  data?: T;
  total: number;
  page: number;
  size: number;
}

// 任务相关类型
export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
export type TaskPriority = 1 | 2 | 3; // 1-低，2-中，3-高

export interface Task {
  id: number;
  model_id: number;
  type: string;
  input: string;
  output?: string;
  status: TaskStatus;
  priority: TaskPriority;
  retry_count: number;
  max_retries: number;
  error_message?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
  model?: Model;
  logs?: TaskLog[];
}

export interface TaskCreateRequest {
  model_id: number;
  type: string;
  input: string;
  priority?: TaskPriority;
}

export interface TaskUpdateRequest {
  priority?: TaskPriority;
  status?: TaskStatus;
}

export interface TaskListParams {
  model_id?: number;
  status?: TaskStatus;
  type?: string;
  priority?: TaskPriority;
  page?: number;
  page_size?: number;
  order_by?: string;
  order?: 'asc' | 'desc';
}

export interface TaskStats {
  total_tasks: number;
  pending_tasks: number;
  running_tasks: number;
  completed_tasks: number;
  failed_tasks: number;
  cancelled_tasks: number;
  success_rate: number;
  avg_processing_ms: number;
}

// 模型相关类型
export type ModelType = 'openai' | 'local' | 'custom';
export type ModelStatus = 'online' | 'offline' | 'maintenance';

export interface ModelConfig {
  [key: string]: any;
}

export interface Model {
  id: number;
  name: string;
  type: ModelType;
  config: ModelConfig;
  status: ModelStatus;
  max_workers: number;
  current_workers: number;
  total_requests: number;
  success_requests: number;
  created_at: string;
  updated_at: string;
}

export interface ModelStats extends Model {
  pending_tasks: number;
  running_tasks: number;
  success_rate: number;
  avg_response_ms: number;
}

// 任务日志类型
export type LogLevel = 'debug' | 'info' | 'warn' | 'error';

export interface TaskLog {
  id: number;
  task_id: number;
  level: LogLevel;
  message: string;
  data?: { [key: string]: any };
  created_at: string;
}

// 队列状态
export interface QueueStatus {
  high_priority_count: number;
  medium_priority_count: number;
  low_priority_count: number;
  processing_count: number;
  delayed_count: number;
  total_count: number;
}

// Worker 状态
export interface WorkerStatus {
  worker_id: string;
  model_id: number;
  model_name: string;
  status: string;
  current_task_id?: number;
  start_time: string;
  last_heartbeat: string;
}

// 系统统计
export interface SystemStats {
  id: number;
  stat_date: string;
  total_tasks: number;
  completed_tasks: number;
  failed_tasks: number;
  avg_processing_time_ms: number;
  queue_length: number;
  active_models: number;
  created_at: string;
}

// Dashboard 统计数据
export interface DashboardStats {
  task_stats: TaskStats;
  model_stats: ModelStats[];
  queue_status: QueueStatus;
  worker_status: WorkerStatus[];
  system_stats: SystemStats;
  recent_tasks: Task[];
}

// 图表数据类型
export interface ChartData {
  date: string;
  total: number;
  completed: number;
  failed: number;
  avg_processing_ms?: number;
}

// 系统健康状态
export interface HealthStatus {
  status: 'ok' | 'error';
  database: 'ok' | 'error';
  redis: 'ok' | 'error';
  queue: 'ok' | 'error';
  database_error?: string;
  redis_error?: string;
  queue_error?: string;
  queue_status?: QueueStatus;
}

// 系统信息
export interface SystemInfo {
  version: string;
  environment: string;
  database_stats?: { [key: string]: any };
  redis_info?: { [key: string]: any };
  queue_status?: QueueStatus;
}

// 表格分页参数
export interface PaginationParams {
  current: number;
  pageSize: number;
  total?: number;
}
