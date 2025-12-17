// types/mission.ts

/**
 * 任务类型枚举
 */
export type TaskType = 'scheduled' | 'manual';

/**
 * 任务模式枚举
 */
export type MissionMode = 'add' | 'edit' | 'read';

/**
 * 参数项接口
 */
export interface ParamItem {
  key: string;
  value: any;
  required?: boolean;
  description?: string;
  defaultValue?: any;
}

/**
 * 配置项接口 - 核心配置结构
 */
export interface ConfigItem {
  type?: string;
  data_source?: string | null;
  params: ParamItem[];
}

/**
 * 处理器配置接口（扩展ConfigItem）
 */
export interface ProcessorConfig extends ConfigItem {
  // 可以添加处理器特有的属性
  id?: string;
  order?: number;
}

/**
 * 任务数据接口
 */
export interface MissionData {
  before_execute: ConfigItem | null;
  source: ConfigItem;
  processors: ProcessorConfig[];
  sink: ConfigItem;
  after_execute: ConfigItem | null;
}

/**
 * 任务记录接口
 */
export interface MissionRecord {
  id: string;
  mission_name: string;
  cron: string;
  data: MissionData;
  status: number;
  last_run_time?: string;
  last_success_time?: string;
  last_end_time?: string;
  err_msg?: string;
  IsRunning: boolean;
  EntryID?: string | null;
  updated_at: string;
  created_at: string;
  DeletedAt?: string | null;
}

/**
 * 任务类型选项接口
 */
export interface TypeOption {
  type: string;
  data_source: Array<{
    id: string;
    name: string;
  }> | null;
  params: ParamItem[];
}

/**
 * 类型数据接口
 */
export interface TypeData {
  executor: TypeOption[];
  source: TypeOption[];
  sink: TypeOption[];
  processor: TypeOption[];
}

/**
 * 表单数据接口
 */
export interface MissionFormData {
  id?: string;
  mission_name: string;
  cron: string;
  before_execute: ConfigItem;
  source: ConfigItem;
  processors: ProcessorConfig[];
  sink: ConfigItem;
  after_execute: ConfigItem;
  task_type?: TaskType;
}

/**
 * 文件上传响应接口
 */
export interface FileUploadResponse {
  id: string;
  name: string;
  size: number;
  path: string;
  ex_name: string;
}