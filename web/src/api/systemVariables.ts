import { request } from "../utils/request";
import type { ApiResponse } from "../types";
import { Params } from "@/src/types";

// 定义变量类型列表项的接口
export interface VariableTypeListItem {
  type: string;
  params: Params[];
  datasource_list: {
    ID: string;
    Name: string;
  }[] | null;
}

/**
 * 获取系统变量类型列表
 */
export const getVariableTypeList = () => {
  return request.get<ApiResponse<{ list: VariableTypeListItem[] }>>("/variables/types");
};

/**
 * 获取系统变量列表
 */
export const getVariableList = () => {
  return request.get<ApiResponse<{ list: any[] }>>("/variables");
};

/**
 * 新增/编辑系统变量
 */
export const saveVariable = (data: {
  id?: string;
  name: string;
  type: string;
  datasource_id?: string | null;
  description: string;
  value: { key: string; value: string }[];
  edit: boolean;
}) => {
  return request.post<ApiResponse<any>>("/variables", data);
};

/**
 * 删除系统变量
 */
export const deleteVariable = (data: { id: string }) => {
  return request.delete<ApiResponse<any>>(`/variables/${data.id}`);
};

/**
 * 测试系统变量
 */
export const testVariable = (data: { id: string }) => {
  return request.post<ApiResponse<string>>(`/variables/${data.id}/test`);
};
