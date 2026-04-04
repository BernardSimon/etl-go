import { request } from '../utils/request';
import {
    GetFileListRequest,
    GetFileListResponse,
    DeleteFileRequest, ApiResponse, FileInfo
} from '../types';

/**
 * 获取文件列表
 */
export const getFileList = (data: GetFileListRequest) => {
    return request.get<ApiResponse<GetFileListResponse>>('/files', {
        params: data,
    });
};

/**
 * 上传文件
 */
export const uploadFile = (data: FormData) => {
    return request.post<ApiResponse<any>>('/files', data);
};

/**
 * 删除文件
 */
export const deleteFile = (data: DeleteFileRequest) => {
    return request.delete<ApiResponse<string>>(`/files/${data.id}`);
};

export const getFileListByTaskRecordID = (id: string ) => {
    return request.get<ApiResponse<FileInfo[]>>(`/task-records/${id}/files`);
};

export const buildFileDownloadUrl = (
    file: Pick<FileInfo, 'id' | 'path' | 'ex_name'>,
    token?: string
) => {
    const baseUrl = String(import.meta.env.VITE_API_BASE_URL || "").replace(/\/$/, "");
    const pathParts = String(file.path || "")
        .split("/")
        .filter(Boolean)
        .map((part) => encodeURIComponent(part));
    const fileName = `${encodeURIComponent(file.id)}${file.ex_name || ""}`;
    const url = new URL(`${baseUrl}/file/${[...pathParts, fileName].join("/")}`, window.location.origin);

    if (token) {
        url.searchParams.set("token", token);
    }

    return url.toString();
};
