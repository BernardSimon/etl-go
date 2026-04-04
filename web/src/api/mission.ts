import { request } from "../utils/request";
import type { ApiResponse } from "../types";
import {TypeData} from "@/src/types/mission.ts";





/**
 * 获取任务列表
 */
export const getTaskAll = () => {
  return request.get<ApiResponse<any[]>>("/tasks");
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
export const getTypeByComponent = () => {
  return request.get<ApiResponse<TypeData>>("/components");
};
