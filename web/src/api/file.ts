import type { AxiosProgressEvent } from "axios";
import { request } from '../utils/request';
import {
    GetFileListRequest,
    GetFileListResponse,
    DeleteFileRequest, ApiResponse, FileInfo,
    InitUploadSessionRequest, UploadSessionInfo, UploadSessionStatus, ChunkUploadResult,
} from '../types';

/** Files smaller than this are uploaded via the original single-request path. */
export const LARGE_FILE_THRESHOLD = 10 * 1024 * 1024; // 10 MB

/** Each chunk is 5 MB. */
export const CHUNK_SIZE = 5 * 1024 * 1024;

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
export const uploadFile = (
    data: FormData,
    options?: {
        signal?: AbortSignal;
        onUploadProgress?: (event: AxiosProgressEvent) => void;
    }
) => {
    return request.post<ApiResponse<any>>('/files', data, {
        timeout: 0,
        signal: options?.signal,
        onUploadProgress: options?.onUploadProgress,
        maxBodyLength: Infinity,
        maxContentLength: Infinity,
    });
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

// ── Chunked upload API ────────────────────────────────────────────────────────

/** Compute a SHA-256 hex digest of a Blob using the Web Crypto API. */
export async function sha256Hex(blob: Blob): Promise<string> {
    const buffer = await blob.arrayBuffer();
    const hashBuffer = await crypto.subtle.digest('SHA-256', buffer);
    return Array.from(new Uint8Array(hashBuffer))
        .map(b => b.toString(16).padStart(2, '0'))
        .join('');
}

export const initUploadSession = (data: InitUploadSessionRequest) =>
    request.post<ApiResponse<UploadSessionInfo>>('/files/upload/init', data);

export const uploadChunk = (
    sessionId: string,
    chunkIndex: number,
    chunkBlob: Blob,
    sha256Hash: string,
    options?: {
        signal?: AbortSignal;
        onUploadProgress?: (event: AxiosProgressEvent) => void;
    }
) =>
    request.put<ApiResponse<ChunkUploadResult>>(
        `/files/upload/${sessionId}/chunk/${chunkIndex}`,
        chunkBlob,
        {
            headers: {
                'Content-Type': 'application/octet-stream',
                ...(sha256Hash ? { 'X-Chunk-SHA256': sha256Hash } : {}),
            },
            timeout: 0,
            signal: options?.signal,
            onUploadProgress: options?.onUploadProgress,
            maxBodyLength: Infinity,
            maxContentLength: Infinity,
        }
    );

export const completeUploadSession = (sessionId: string) =>
    request.post<ApiResponse<FileInfo>>(`/files/upload/${sessionId}/complete`);

export const cancelUploadSession = (sessionId: string) =>
    request.delete<ApiResponse<string>>(`/files/upload/${sessionId}`);

export const getUploadStatus = (sessionId: string) =>
    request.get<ApiResponse<UploadSessionStatus>>(`/files/upload/${sessionId}`);

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
