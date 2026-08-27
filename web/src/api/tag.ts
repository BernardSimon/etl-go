import { request } from "../utils/request";
import type { ApiResponse } from "../types";

export interface Tag {
  id: string;
  name: string;
  color: string;
  created_at?: string;
  updated_at?: string;
}

/**
 * 获取标签列表
 */
export const getTagList = (params?: { search?: string }) => {
  return request.get<ApiResponse<Tag[]>>("/tags", { params });
};

/**
 * 新增标签
 */
export const addTag = (data: { name: string; color?: string }) => {
  return request.post<ApiResponse<Tag>>("/tags", data);
};

/**
 * 更新标签
 */
export const updateTag = (id: string, data: { name: string; color?: string }) => {
  return request.put<ApiResponse<string>>(`/tags/${id}`, data);
};

/**
 * 删除标签
 */
export const deleteTag = (id: string) => {
  return request.delete<ApiResponse<string>>(`/tags/${id}`);
};
