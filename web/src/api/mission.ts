import { request } from "../utils/request";
import type { ApiResponse } from "../types";
import {TypeData} from "@/src/types/mission.ts";

let componentTypeCache: TypeData | null = null;




/**
 * 获取任务列表
 */
export const getTaskAll = (params?: {
  page_no?: number;
  page_size?: number;
  search?: string;
  mission_name?: string;
  status?: number;
  tasktypes?: "manual" | "scheduled";
  tag_id?: string;
}) => {
  return request.get<ApiResponse<{
    list: any[];
    total: number;
    page_no: number;
    page_size: number;
  }>>("/tasks", { params });
};

/**
 * 删除任务
 */
export const deleteTask = (data: { id: string }) => {
  return request.delete<ApiResponse<string>>(`/tasks/${data.id}`);
};

/**
 * 新增任务
 */
export const addTask = (data: any) => {
  return request.post<ApiResponse<any>>("/tasks", data);
};

/**
 * 修改任务
 */
export const updateTask = (data: any) => {
  return request.put<ApiResponse<any>>(`/tasks/${data.id}`, data);
};

/**
 * 启动任务
 */
export const runTask = (data: { id: string }) => {
  return request.post<ApiResponse<any>>(`/tasks/${data.id}/schedule`);
};

/**
 * 停止任务
 */
export const stopTask = (data: { id: string }) => {
  return request.post<ApiResponse<any>>(`/tasks/${data.id}/stop`);
};

/**
 * 手动执行一次任务
 */
export const runTaskOnce = (data: { id: string }) => {
  return request.post<ApiResponse<any>>(`/tasks/${data.id}/run`);
};

/**
 * 参数接口
 */
export const getTypeByComponent = async (forceRefresh = false) => {
  if (!forceRefresh && componentTypeCache) {
    return {
      code: 0,
      message: "ok",
      data: componentTypeCache,
    } as ApiResponse<TypeData>;
  }

  const res = await request.get<ApiResponse<TypeData>>("/components");
  componentTypeCache = res.data;
  return res;
};

export const getTaskTemplates = () => {
  return request.get<ApiResponse<{ list: any[] }>>("/task-templates");
};

export const saveTaskTemplate = (data: {
  id?: string;
  name: string;
  cron: string;
  tasktypes: "manual" | "scheduled";
  params: any;
}) => {
  return request.post<ApiResponse<any>>("/task-templates", data);
};

export const deleteTaskTemplate = (id: string) => {
  return request.delete<ApiResponse<any>>(`/task-templates/${id}`);
};

export const previewTask = (id: string) => {
  return request.post<ApiResponse<{ columns: string[]; rows: Record<string, any>[] }>>(`/tasks/${id}/preview`);
};
