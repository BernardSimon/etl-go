import { request } from "../utils/request.ts";
import type {ApiResponse, Params} from "../types";

let dataSourceTypeListCache: { list: { type: string; params: Params[] }[] } | null = null;

/**
 * 获取数据源类型列表
 * {
 *   "code": 0,
 *   "message": "ok",
 *   "data": {
 *     "list": [
 *     ]
 *   }
 * }
 */
export const getDataSourceTypeList = async (forceRefresh = false) => {
  if (!forceRefresh && dataSourceTypeListCache) {
    return {
      code: 0,
      message: "ok",
      data: dataSourceTypeListCache,
    } as ApiResponse<{ list: { type: string; params: Params[] }[] }>;
  }

  const res = await request.get<ApiResponse<{ list: { type: string,params: Params[]  }[] }>>("/data-sources/types");
  dataSourceTypeListCache = res.data;
  return res;
};

/**
 * 新增/编辑数据源
 * @param {Object} data
 */
export const addDataSource = (data: {
  id?: string;
  name: string;
  type: string;
  data: {key: string,value: string}[];
  edit: boolean;
}) => {
  return request.post<ApiResponse<any>>("/data-sources", data);
};

export const testDataSource = (data: {
  id?: string;
  type: string;
  data: {key: string,value: string}[];
}) => {
  return request.post<ApiResponse<string>>("/data-sources/test", data);
};
/**
 * 删除数据源
 * @param {Object} data
 * {
 *   "id": string
 * }
 */
export const deleteDataSource = (data: { id: string }) => {
  return request.delete<ApiResponse<any>>(`/data-sources/${data.id}`);
};

/**
 * 获取数据源列表
 * {
 *   "code": 0,
 *   "message": "ok",
 *   "data": {
 *     "list": []
 *   }
 * }
 */
export const getDataSourceList = () => {
  return request.get<ApiResponse<{ list: any[] }>>("/data-sources");
};

export const getDataSourceById = (id: string) => {
  return request.get<ApiResponse<{id: string; name: string; type: string; data: {key: string, value: string}[]}>>(`/data-sources/${id}`);
};

export const getDataSourceSchema = (id: string) => {
  return request.get<ApiResponse<{ tables: { name: string; columns: { name: string; type: string; nullable: boolean }[] }[] }>>(`/data-sources/${id}/schema`);
};
