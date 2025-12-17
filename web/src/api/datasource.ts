import { request } from "../utils/request.ts";
import type {ApiResponse, Params} from "../types";

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
export const getDataSourceTypeList = () => {
  return request.post<ApiResponse<{ list: { type: string,params: Params[]  }[] }>>("/getDataSourceTypeList");
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
  edit: string;
}) => {
  return request.post<ApiResponse<any>>("/newDataSource", data);
};
/**
 * 删除数据源
 * @param {Object} data
 * {
 *   "id": string
 * }
 */
export const deleteDataSource = (data: { id: string }) => {
  return request.post<ApiResponse<any>>("/deleteDataSource", data);
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
  return request.post<ApiResponse<{ list: any[] }>>("/getDataSourceList", {});
};
