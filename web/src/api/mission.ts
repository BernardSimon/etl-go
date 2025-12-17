import { request } from "../utils/request";
import type { ApiResponse } from "../types";
import {TypeData} from "@/src/types/mission.ts";





/**
 * 获取任务列表
 */
export const getTaskAll = () => {
  return request.post<ApiResponse<{ list: any[] }>>("/getTaskAll", {});
};

/**
 * 删除任务
 */
export const deleteTask = (data: { id: string }) => {
  return request.post<ApiResponse<string>>("/deleteTask", data);
};

/**
 * 新增任务
 */
export const addTask = (data: any) => {
  return request.post<ApiResponse<any>>("/addTask", data);
};

/**
 * 修改任务
 */
export const updateTask = (data: any) => {
  return request.post<ApiResponse<any>>("/updateTask", data);
};

/**
 * 启动任务
 */
export const runTask = (data: { id: string }) => {
  // 启动任务接口
  return request.post<ApiResponse<any>>("/runTask", data);
};

/**
 * 停止任务
 */
export const stopTask = (data: { id: string }) => {
  // 停止任务接口
  return request.post<ApiResponse<any>>("/stopTask", data);
};

/**
 * 手动执行一次任务
 */
export const runTaskOnce = (data: { id: string }) => {
  // 手动执行一次任务接口
  return request.post<ApiResponse<any>>("/runTaskOnce", data);
};

/**
 * 参数接口
 */
export const getTypeByComponent = (data: object = {}) => {
  return request.post<ApiResponse<TypeData>>("/getTypeByComponent", data);
};


