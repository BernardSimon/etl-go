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
}) => {
  return request.get<ApiResponse<{ list: any[]; total: number }>>("/task-records", {
    params: data,
  });
};

/**
 * 中止运行记录
 */
export const cancelTaskRecord = (data: { id: string }) => {
  return request.post<ApiResponse<any>>(`/task-records/${data.id}/cancel`);
};
