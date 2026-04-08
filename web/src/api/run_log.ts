import { request } from "../utils/request";
import type { ApiResponse } from "../types";

/**
 * 运行记录列表
 */
export const getTaskRecordList = (data: {
  page_no: number;
  page_size: number;
  mission_name?: string;
  status?: number;
  id?: string;
  task_id?: string;
}) => {
  return request.get<ApiResponse<{ list: any[]; total: number; page_no: number; page_size: number }>>("/task-records", {
    params: data,
  });
};

/**
 * 中止运行记录
 */
export const cancelTaskRecord = (data: { id: string }) => {
  return request.post<ApiResponse<any>>(`/task-records/${data.id}/cancel`);
};

export const getTaskRecordParams = (id: string) => {
  return request.get<ApiResponse<{ id: string; task_id: string; mission_name: string; params: any }>>(
    `/task-records/${id}/params`
  );
};

export const getTaskRecordLogs = (id: string) => {
  return request.get<ApiResponse<{
    id: string;
    task_id: string;
    mission_name: string;
    status: number;
    start_time: string;
    end_time: string;
    message: string;
    log: string;
  }>>(`/task-records/${id}/logs`);
};

/**
 * 清理运行记录（运行中的记录不会被清理）
 * status: 1=成功 2=失败 不传=全部已完成
 * before: YYYY-MM-DD，清理该日期之前的记录，不传则不限时间
 */
export const cleanTaskRecords = (params: { status?: number; before?: string }) => {
  return request.delete<ApiResponse<{ deleted: number }>>("/task-records", { params });
};
