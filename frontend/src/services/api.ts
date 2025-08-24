import axios, { AxiosResponse } from 'axios';
import { message } from 'antd';
import {
  ApiResponse,
  PagedResponse,
  Task,
  TaskCreateRequest,
  TaskUpdateRequest,
  TaskListParams,
  TaskStats,
  Model,
  ModelStats,
  DashboardStats,
  HealthStatus,
  SystemInfo,
} from '../types';

// 创建 axios 实例
const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL || '/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 可以在这里添加认证 token
    // const token = localStorage.getItem('token');
    // if (token) {
    //   config.headers.Authorization = `Bearer ${token}`;
    // }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
api.interceptors.response.use(
  (response: AxiosResponse) => {
    return response;
  },
  (error) => {
    if (error.response) {
      const { status, data } = error.response;
      switch (status) {
        case 400:
          message.error(data.message || '请求参数错误');
          break;
        case 401:
          message.error('未授权访问');
          // 可以跳转到登录页
          break;
        case 403:
          message.error('禁止访问');
          break;
        case 404:
          message.error('请求的资源不存在');
          break;
        case 500:
          message.error('服务器内部错误');
          break;
        default:
          message.error(data.message || '网络错误');
      }
    } else if (error.request) {
      message.error('网络连接失败');
    } else {
      message.error('请求失败');
    }
    return Promise.reject(error);
  }
);

// 系统 API
export const systemApi = {
  // 健康检查
  health: (): Promise<ApiResponse<HealthStatus>> =>
    api.get('/system/health').then((res) => res.data),

  // 系统信息
  info: (): Promise<ApiResponse<SystemInfo>> =>
    api.get('/system/info').then((res) => res.data),
};

// 任务 API
export const taskApi = {
  // 创建任务
  create: (data: TaskCreateRequest): Promise<ApiResponse<Task>> =>
    api.post('/tasks', data).then((res) => res.data),

  // 获取任务列表
  list: (params: TaskListParams): Promise<PagedResponse<Task[]>> =>
    api.get('/tasks', { params }).then((res) => res.data),

  // 获取任务详情
  get: (id: number): Promise<ApiResponse<Task>> =>
    api.get(`/tasks/${id}`).then((res) => res.data),

  // 更新任务
  update: (id: number, data: TaskUpdateRequest): Promise<ApiResponse<Task>> =>
    api.put(`/tasks/${id}`, data).then((res) => res.data),

  // 取消任务
  cancel: (id: number): Promise<ApiResponse> =>
    api.delete(`/tasks/${id}`).then((res) => res.data),

  // 重试任务
  retry: (id: number): Promise<ApiResponse> =>
    api.post(`/tasks/${id}/retry`).then((res) => res.data),

  // 获取任务统计
  stats: (): Promise<ApiResponse<TaskStats>> =>
    api.get('/tasks/stats').then((res) => res.data),
};

// 模型 API
export const modelApi = {
  // 创建模型
  create: (data: Partial<Model>): Promise<ApiResponse<Model>> =>
    api.post('/models', data).then((res) => res.data),

  // 获取模型列表
  list: (params?: { type?: string; status?: string }): Promise<ApiResponse<Model[]>> =>
    api.get('/models', { params }).then((res) => res.data),

  // 获取可用模型
  available: (): Promise<ApiResponse<Model[]>> =>
    api.get('/models/available').then((res) => res.data),

  // 获取模型统计
  stats: (): Promise<ApiResponse<ModelStats[]>> =>
    api.get('/models/stats').then((res) => res.data),

  // 获取模型详情
  get: (id: number): Promise<ApiResponse<Model>> =>
    api.get(`/models/${id}`).then((res) => res.data),

  // 更新模型
  update: (id: number, data: Partial<Model>): Promise<ApiResponse<Model>> =>
    api.put(`/models/${id}`, data).then((res) => res.data),

  // 删除模型
  delete: (id: number): Promise<ApiResponse> =>
    api.delete(`/models/${id}`).then((res) => res.data),

  // 更新模型状态
  updateStatus: (id: number, status: string): Promise<ApiResponse> =>
    api.put(`/models/${id}/status`, { status }).then((res) => res.data),
};

// 统计 API
export const statsApi = {
  // Dashboard 统计
  dashboard: (): Promise<ApiResponse<DashboardStats>> =>
    api.get('/stats/dashboard').then((res) => res.data),

  // 按日期获取任务统计
  tasksByDate: (days: number = 7): Promise<ApiResponse<any[]>> =>
    api.get('/stats/tasks/date', { params: { days } }).then((res) => res.data),

  // 按模型获取任务统计
  tasksByModel: (): Promise<ApiResponse<any[]>> =>
    api.get('/stats/tasks/model').then((res) => res.data),

  // 按类型获取任务统计
  tasksByType: (): Promise<ApiResponse<any[]>> =>
    api.get('/stats/tasks/type').then((res) => res.data),
};

export default api;
