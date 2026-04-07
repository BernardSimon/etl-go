// 登录请求参数
export interface LoginRequest {
  username: string;
  password: string;
}

// 登录响应数据（直接登录或 2FA 挑战）
export interface LoginResponse {
  code: number;
  message: string;
  data: {
    token?: string;
    refresh_token?: string;
    requires_2fa?: boolean;
    pre_auth_token?: string;
  };
}

// 两步验证请求
export interface VerifyTwoFactorRequest {
  pre_auth_token: string;
  code: string;
}

// API 响应通用格式
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
}

export interface ApiFieldError {
  field: string;
  message: string;
}

export interface ApiErrorData {
  errors?: ApiFieldError[];
  [key: string]: any;
}



// 订阅信息
export interface SubscriptionInfo {}

// 分页参数
export interface PaginationParams {
  page: number;
  pageSize: number;
}

// 分页响应
export interface PaginationResponse<T> {
  list: T[];
  data?: T[];
  total: number;
  page: number;
  pageSize: number;
}

// 侧边栏菜单项
export interface SidebarItem {
  index: string;
  title: string;
  icon?: any;
  children?: SidebarItem[];
}

// 在 types/datasource.ts 或相应类型文件中添加
export interface GetFileListRequest {
  page_size?: number;
  page_no?: number;
  keyword?: string;
  ids?: string;
}

export interface GetFileListResponse {
  total: number;
  page_no?: number;
  page_size?: number;
  list: FileInfo[];
}

export interface FileInfo {
  id: string;
  name: string;
  size: number;
  created_at: string;
  path: string;
  ex_name : string;
}

export interface DeleteFileRequest {
  id: string;
}

export interface Params {
  key:string;
  required:boolean;
  defaultValue:string;
  description: string;
  placeholder?: string;
  example?: string;
  type?: string;
}
