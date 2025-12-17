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
    return request.post<ApiResponse<GetFileListResponse>>('/getFileList', data);
};

/**
 * 上传文件
 */
export const uploadFile = (data: FormData) => {
    return request.post<ApiResponse<any>>('/uploadFile', data, {
        headers: {
            'Content-Type': 'multipart/form-data',
        }
    });
};

/**
 * 删除文件
 */
export const deleteFile = (data: DeleteFileRequest) => {
    return request.post<ApiResponse<string>>('/deleteFile', data);
};

export const getFileListByTaskRecordID = (id: string ) => {
    return request.post<ApiResponse<FileInfo[]>>('/getFileListByTaskRecordID', { id: id });
};
