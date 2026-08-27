<template>
  <PageContainer>
    <ContentCard>
      <div class="section-stack">
        <div class="table-toolbar">
          <div class="table-toolbar__right">
            <div class="auto-refresh-group">
              <a-switch v-model:checked="autoRefresh" size="small" />
              <span class="auto-refresh-label">{{ t('runLog.autoRefresh.label') }}</span>
            </div>
            <a-button danger @click="openCleanModal">{{ t('runLog.clean.button') }}</a-button>
          </div>
        </div>

        <FilterBar>
          <div class="filter-field runlog-filter">
            <span class="filter-field__label">{{ t('runLog.search.recordId') }}</span>
            <a-input
              v-model:value="searchForm.id"
              allow-clear
              :placeholder="t('runLog.search.recordId.placeholder')"
            />
          </div>
          <div class="filter-field runlog-filter runlog-filter--wide">
            <span class="filter-field__label">{{ t('runLog.search.missionName') }}</span>
            <a-input
              v-model:value="searchForm.mission_name"
              allow-clear
              :placeholder="t('runLog.search.missionName.placeholder')"
            />
          </div>
          <div class="filter-field runlog-filter">
            <span class="filter-field__label">{{ t('runLog.search.status') }}</span>
            <a-select v-model:value="searchForm.status" allow-clear>
              <a-select-option v-for="item in getStatusOptions()" :key="item.value" :value="item.value">
                {{ item.label }}
              </a-select-option>
            </a-select>
          </div>
          <div class="filter-field filter-field--actions">
            <a-button :loading="loading" @click="handleSearch">
              {{ t('runLog.search.query') }}
            </a-button>
          </div>
        </FilterBar>

        <a-table
          class="app-shell-table"
          :columns="getColumns()"
          :data-source="tableData"
          :loading="loading"
          :pagination="pagination"
          :scroll="{ x: 'max-content', y: '100vh' }"
          @change="handleTableChange"
          row-key="id"
        >
        <template #bodyCell="{ column, record }">
          <!-- 状态列渲染为标签 -->
          <template v-if="column.key === 'status'">
            <a-tag :color="getStatusColor(record.status)">
              {{ getStatusText(record.status) }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'mission_name'">
            {{ record.task?.mission_name || "-" }}
          </template>
          <template v-else-if="column.key === 'message'">
            <div class="result-cell">
              <span>{{ formatResultText(record.message) }}</span>
              <a-button
                v-if="record.message"
                type="link"
                size="small"
                @click="showResultModal(record)"
              >
                {{ t("runLog.table.action.viewDetail") }}
              </a-button>
            </div>
          </template>
          <template v-else-if="column.key === 'duration'">
            {{ formatDuration(record.start_time, record.end_time) }}
          </template>
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button
                type="primary"
                size="small"
                @click="() => showParamsModal(record)"
              >
                {{ t("runLog.table.action.viewParams") }}
              </a-button>
              <a-button
                type="default"
                size="small"
                @click="showResultModal(record)"
              >
                {{ t("runLog.table.action.viewDetail") }}
              </a-button>
              <a-button
                type="primary"
                danger
                :disabled="record.status !== 0"
                size="small"
                @click="handleCancel(record)"
              >
                {{ t("runLog.table.action.cancel") }}
              </a-button>
              <a-button
                  type="primary"
                  size="small"
                  @click="showTaskFilesModal(record)"
              >
                {{ t("runLog.taskFiles.button") }}
              </a-button>
            </a-space>
          </template>
        </template>
        </a-table>
      </div>
    </ContentCard>

    <MissionConfigModal
      v-model:open="paramsModal.show"
      :title="paramsModal.title"
      :mode="paramsModal.mode"
      :id="paramsModal.id"
      :data="paramsModal.data"
      :record="paramsModal.record"
    />
 
  <a-modal
      v-model:open="taskFilesModal.visible"
      :title="t('runLog.taskFiles.modal.title')"
      :footer="null"
      :width="isNarrowScreen ? '96vw' : '80%'"
      @cancel="closeTaskFilesModal"
  >
    <a-table
        :columns="taskFileColumns()"
        :data-source="taskFilesModal.fileList"
        :loading="taskFilesModal.loading"
        :pagination="false"
        :scroll="{ y: '400px' }"
        row-key="id"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'size'">
          {{ formatFileSize(record.size) }}
        </template>
        <template v-else-if="column.key === 'action'">
          <a-space>
            <a-button type="primary" size="small" @click="downloadTaskFile(record)">
              {{ t('file.table.action.download') }}
            </a-button>
            <a-button type="primary" danger size="small" @click="handleDeleteTaskFile(record)">
              {{ t('file.table.action.delete') }}
            </a-button>
          </a-space>
        </template>
      </template>
    </a-table>
  </a-modal>
  <a-modal
      v-model:open="resultModal.visible"
      :title="t('runLog.detail.modal.title')"
      :footer="null"
      :width="isNarrowScreen ? '96vw' : '720px'"
  >
    <a-descriptions :column="1" bordered size="small">
      <a-descriptions-item :label="t('runLog.table.column.recordId')">
        {{ resultModal.record?.id || '-' }}
      </a-descriptions-item>
      <a-descriptions-item :label="t('runLog.table.column.missionName')">
        {{ resultModal.record?.task?.mission_name || '-' }}
      </a-descriptions-item>
      <a-descriptions-item :label="t('runLog.table.column.status')">
        {{ getStatusText(resultModal.record?.status) }}
      </a-descriptions-item>
      <a-descriptions-item :label="t('runLog.table.column.startTime')">
        {{ resultModal.record?.start_time || '-' }}
      </a-descriptions-item>
      <a-descriptions-item :label="t('runLog.table.column.endTime')">
        {{ resultModal.record?.end_time || '-' }}
      </a-descriptions-item>
      <a-descriptions-item :label="t('runLog.table.column.duration')">
        {{ formatDuration(resultModal.record?.start_time, resultModal.record?.end_time) }}
      </a-descriptions-item>
      <a-descriptions-item :label="t('runLog.detail.message')">
        <pre class="detail-pre">{{ resultModal.record?.message || t('common.empty') }}</pre>
      </a-descriptions-item>
      <a-descriptions-item :label="t('runLog.detail.payload')">
        <pre class="detail-pre">{{ resultModal.payload }}</pre>
      </a-descriptions-item>
    </a-descriptions>
  </a-modal>
  <a-modal
    v-model:open="cleanModal.visible"
    :title="t('runLog.clean.modal.title')"
    :confirm-loading="cleanModal.loading"
    :ok-text="t('runLog.clean.ok')"
    ok-type="danger"
    :cancel-text="t('runLog.clean.cancel')"
    @ok="handleClean"
  >
    <a-form layout="vertical" style="margin-top: 8px">
      <a-form-item :label="t('runLog.clean.status.label')">
        <a-select v-model:value="cleanModal.status" style="width: 100%">
          <a-select-option :value="-1">{{ t('runLog.clean.status.allFinished') }}</a-select-option>
          <a-select-option :value="1">{{ t('runLog.clean.status.successOnly') }}</a-select-option>
          <a-select-option :value="2">{{ t('runLog.clean.status.failedOnly') }}</a-select-option>
        </a-select>
      </a-form-item>
      <a-form-item :label="t('runLog.clean.before.label')">
        <a-select v-model:value="cleanModal.beforePreset" style="width: 100%" @change="(val: any) => handlePresetChange(val)">
          <a-select-option value="">{{ t('runLog.clean.before.all') }}</a-select-option>
          <a-select-option value="7">{{ t('runLog.clean.before.7days') }}</a-select-option>
          <a-select-option value="30">{{ t('runLog.clean.before.30days') }}</a-select-option>
          <a-select-option value="90">{{ t('runLog.clean.before.90days') }}</a-select-option>
          <a-select-option value="custom">{{ t('runLog.clean.before.custom') }}</a-select-option>
        </a-select>
      </a-form-item>
      <a-form-item v-if="cleanModal.beforePreset === 'custom'" :label="t('runLog.clean.before.custom.label')">
        <a-date-picker
          v-model:value="cleanModal.beforeDate"
          style="width: 100%"
          value-format="YYYY-MM-DD"
          :disabled-date="(d: any) => d && d.valueOf() > Date.now()"
        />
      </a-form-item>
      <a-alert
        type="warning"
        show-icon
        :message="t('runLog.clean.warning')"
        style="margin-top: 4px"
      />
    </a-form>
  </a-modal>
  </PageContainer>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, watch, computed } from "vue";
import { useRoute } from "vue-router";
import { getTaskRecordList, cancelTaskRecord, getTaskRecordLogs, getTaskRecordParams, cleanTaskRecords } from "../api/run_log";
import { message, Modal } from "ant-design-vue";
import type { TablePaginationConfig } from "ant-design-vue";
import MissionConfigModal from "../components/MissionConfigModal.vue";
import { useI18n } from "vue-i18n";
import { buildFileDownloadUrl, deleteFile, getFileListByTaskRecordID } from "../api/file";
import {useUserStore} from "../../src/stores/user.ts"; // 新增引入

const { t } = useI18n(); // 初始化i18n实例
const route = useRoute();
const getStatusOptions =(): { value: number; label: string }[] =>{return [
  { value: -1, label: t("runLog.search.status.all") }, // 改为国际化文本
  { value: 0, label: t("runLog.table.status.running") },
  { value: 1, label: t("runLog.table.status.success") },
  { value: 2, label: t("runLog.table.status.failed") },
]};

const searchForm = reactive({
  id: "",
  mission_name: "",
  status: -1 as number,
  task_id: "",
});

const loading = ref(false);
const autoRefresh = ref(true);
const refreshTimer = ref<number | null>(null);
const tableData = ref<any[]>([]);
const screenWidth = ref(window.innerWidth);
const isNarrowScreen = computed(() => screenWidth.value < 768);
const pagination = reactive<TablePaginationConfig>({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total) => t("common.pagination.total", { total }),
});

const getColumns=(): any[] =>{return [
  {
    title: t("runLog.table.column.recordId"),
    dataIndex: "id",
    key: "id",
    align: "center",
    width: 200,
  },
  {
    title: t("runLog.table.column.missionName"),
    dataIndex: "mission_name",
    key: "mission_name",
    align: "center",
    width: 230,
  },
  {
    title: t("runLog.table.column.runner"),
    dataIndex: "run_by",
    key: "run_by",
    align: "center",
    width: 110,
  },
  {
    title: t("runLog.table.column.status"),
    dataIndex: "status",
    key: "status",
    align: "center",
    width: 110,
  },
  {
    title: t("runLog.table.column.result"),
    dataIndex: "message",
    key: "message",
    align: "center",
    width: 300,
  },
  {
    title: t("runLog.table.column.startTime"),
    dataIndex: "start_time",
    key: "start_time",
    align: "center",
    width: 150,
  },
  {
    title: t("runLog.table.column.endTime"),
    dataIndex: "end_time",
    key: "end_time",
    align: "center",
    width: 150,
  },
  {
    title: t("runLog.table.column.duration"),
    dataIndex: "duration",
    key: "duration",
    align: "center",
    width: 120,
  },
  {
    title: t("runLog.table.column.actions"),
    key: "action",
    align: "center",
    fixed: "right",
    width: 260,
  },
]};

const getStatusText = (status: number) => {
  switch (status) {
    case 0:
      return t("runLog.table.status.running");
    case 1:
      return t("runLog.table.status.success");
    case 2:
      return t("runLog.table.status.failed");
    default:
      return t("runLog.table.status.unknown");
  }
};

const getStatusColor = (status: number) => {
  switch (status) {
    case 0:
      return "processing";
    case 1:
      return "success";
    case 2:
      return "error";
    default:
      return "default";
  }
};

const formatResultText = (text?: string) => {
  if (!text) return "-";
  return text.length > 36 ? `${text.slice(0, 36)}...` : text;
};

const formatDuration = (start?: string, end?: string) => {
  if (!start || !end) return "-";
  const startAt = new Date(start).getTime();
  const endAt = new Date(end).getTime();
  if (Number.isNaN(startAt) || Number.isNaN(endAt) || endAt < startAt) {
    return "-";
  }
  const totalSeconds = Math.floor((endAt - startAt) / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return minutes > 0 ? `${minutes}m ${seconds}s` : `${seconds}s`;
};

const hasRunningRecords = () => tableData.value.some((item) => item.status === 0);
const handleResize = () => {
  screenWidth.value = window.innerWidth;
};

const stopAutoRefresh = () => {
  if (refreshTimer.value) {
    window.clearInterval(refreshTimer.value);
    refreshTimer.value = null;
  }
};

const startAutoRefresh = () => {
  stopAutoRefresh();
  if (!autoRefresh.value) {
    return;
  }
  refreshTimer.value = window.setInterval(() => {
    if (hasRunningRecords()) {
      fetchData();
    }
  }, 10000);
};

// 获取运行记录
const fetchData = async () => {
  loading.value = true;
  try {
    const res = await getTaskRecordList({
      page_no: pagination.current || 1,
      page_size: pagination.pageSize || 10,
      mission_name: searchForm.mission_name || "",
      status: searchForm.status ?? -1,
      id: searchForm.id ?? "",
      task_id: searchForm.task_id || "",
    });
    if (res && res.code === 0) {
      tableData.value = res.data.list || [];
      pagination.total = res.data.total || 0;
      startAutoRefresh();
    } else {
      tableData.value = [];
      pagination.total = 0;
      stopAutoRefresh();
    }
  } catch (error) {
    console.error("获取运行记录失败", error);
    stopAutoRefresh();
  } finally {
    loading.value = false;
  }
};

// 查询
const handleSearch = () => {
  pagination.current = 1;
  fetchData();
};

// 表格分页
const handleTableChange = (pag: TablePaginationConfig) => {
  pagination.current = pag.current;
  pagination.pageSize = pag.pageSize;
  fetchData();
};

// 中止任务
const handleCancel = (record: any) => {
  Modal.confirm({
    title: t("runLog.modal.confirmCancel.title"),
    content: t("runLog.modal.confirmCancel.content"),
    async onOk() {
      try {
        const res = await cancelTaskRecord({ id: record.id });
        if (res && res.code === 0) {
          message.success(t("runLog.cancel.success"));
          fetchData();
        }
      } catch (error) {
        console.error("中止任务失败：", error);
      }
    },
  });
};

// 运行参数弹窗
const paramsModal = ref<any>({
  show: false,
  title: t("runLog.table.action.viewTitle"),
  mode: "read",
  id: "",
  data: null,
  record: null,
});

const resultModal = reactive({
  visible: false,
  record: null as any,
  payload: "",
});

// 显示运行参数弹窗
const showParamsModal = (record: any) => {
  getTaskRecordParams(record.id).then((res) => {
    paramsModal.value.id = record.id;
    paramsModal.value.data = res.data?.params || record.data || null;
    paramsModal.value.record = {
      mission_name: res.data?.mission_name || record.task?.mission_name || "-",
      cron: record.task?.cron,
    };
    paramsModal.value.show = true;
  });
};

const showResultModal = (record: any) => {
  getTaskRecordLogs(record.id).then((res) => {
    resultModal.visible = true;
    resultModal.record = {
      ...record,
      status: res.data?.status ?? record.status,
      message: res.data?.message ?? record.message,
      start_time: res.data?.start_time ?? record.start_time,
      end_time: res.data?.end_time ?? record.end_time,
      task: {
        ...(record.task || {}),
        mission_name: res.data?.mission_name || record.task?.mission_name,
      },
    };
    resultModal.payload = JSON.stringify(record.data || {}, null, 2);
  });
};

// 任务文件模态框状态
const taskFilesModal = reactive({
  visible: false,
  loading: false,
  fileList: [] as any[],
  recordId: ''
});

// 任务文件表格列定义
const taskFileColumns = (): any[] => {
  return [
    {
      title: t("file.table.column.id"),
      dataIndex: 'id',
      key: 'id',
      align: 'center',
      width: 200
    },
    {
      title: t('file.table.column.name'),
      dataIndex: 'name',
      key: 'name',
      align: 'center'
    },
    {
      title: t('file.table.column.path'),
      dataIndex: 'path',
      key: 'path',
      align: 'center',
      width: 200
    },
    {
      title: t('file.table.column.size'),
      dataIndex: 'size',
      key: 'size',
      align: 'center'
    },
    {
      title: t('file.table.column.createdAt'),
      dataIndex: 'created_at',
      key: 'created_at',
      align: 'center'
    },
    {
      title: t('file.table.column.actions'),
      key: 'action',
      align: 'center',
      fixed: 'right'
    }
  ];
};

// 显示任务文件模态框
const showTaskFilesModal = async (record: any) => {
  taskFilesModal.visible = true;
  taskFilesModal.recordId = record.id;
  taskFilesModal.loading = true;

  try {
    const res = await getFileListByTaskRecordID(record.id);
    if (res && res.code === 0) {
      taskFilesModal.fileList = res.data || [];
    } else {
      taskFilesModal.fileList = [];
    }
  } catch (error) {
    taskFilesModal.fileList = [];
  } finally {
    taskFilesModal.loading = false;
  }
};

// 关闭任务文件模态框
const closeTaskFilesModal = () => {
  taskFilesModal.visible = false;
  taskFilesModal.fileList = [];
  taskFilesModal.recordId = '';
};

// 下载任务文件
const downloadTaskFile = (record: any) => {
  const link = document.createElement('a');
  const token = useUserStore().token;
  link.href = buildFileDownloadUrl(record, token);
  link.download = record.name;
  link.click();
};

// 删除任务文件
const handleDeleteTaskFile = (record: any) => {
  Modal.confirm({
    title: t('file.delete.confirm.title'),
    content: t('runLog.taskFiles.deletePhysicalWarning'),
    onOk: async () => {
      try {
        const res = await deleteFile({ id: record.id });
        if (res && res.code === 0) {
          message.success(t('file.delete.success'));
          // 重新加载当前任务的文件列表
          const index = taskFilesModal.fileList.findIndex(file => file.id === record.id);
          if (index > -1) {
            taskFilesModal.fileList.splice(index, 1);
          }
        }
      } catch (error) {
        console.error('删除文件失败:', error);
      }
    }
  });
};
const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return t("common.fileSize.zero");
  const k = 1024;
  const sizes = [
    t("common.fileSize.bytes"),
    t("common.fileSize.kb"),
    t("common.fileSize.mb"),
    t("common.fileSize.gb"),
  ];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};
 
// 清理日志
const cleanModal = reactive({
  visible: false,
  loading: false,
  status: -1 as number,
  beforePreset: "" as string,
  beforeDate: "" as string,
});

const openCleanModal = () => {
  cleanModal.status = -1;
  cleanModal.beforePreset = "30";
  cleanModal.beforeDate = "";
  cleanModal.visible = true;
};

const handlePresetChange = (val: any) => {
  if (val !== "custom") {
    cleanModal.beforeDate = "";
  }
};

const handleClean = async () => {
  const params: { status?: number; before?: string } = {};
  if (cleanModal.status !== -1) {
    params.status = cleanModal.status;
  }
  if (cleanModal.beforePreset === "custom") {
    if (!cleanModal.beforeDate) {
      message.warning(t("runLog.clean.missingDate"));
      return;
    }
    params.before = cleanModal.beforeDate;
  } else if (cleanModal.beforePreset !== "") {
    const d = new Date();
    d.setDate(d.getDate() - Number(cleanModal.beforePreset));
    params.before = d.toISOString().slice(0, 10);
  }
  cleanModal.loading = true;
  try {
    const res = await cleanTaskRecords(params);
    if (res && res.code === 0) {
      message.success(t("runLog.clean.success", { count: res.data?.deleted ?? 0 }));
      cleanModal.visible = false;
      fetchData();
    }
  } catch (e) {
    console.error("清理失败", e);
  } finally {
    cleanModal.loading = false;
  }
};

onMounted(() => {
  searchForm.task_id = String(route.query.task_id || "");
  searchForm.mission_name = String(route.query.mission_name || "");
  window.addEventListener("resize", handleResize);
  fetchData();
});

watch(autoRefresh, () => {
  startAutoRefresh();
});

onUnmounted(() => {
  window.removeEventListener("resize", handleResize);
  stopAutoRefresh();
});
</script>

<style scoped lang="scss">
.table-toolbar {
  margin-bottom: -4px;
}

.table-toolbar__right {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.runlog-filter {
  width: 190px;
}

.runlog-filter--wide {
  width: 260px;
}

.auto-refresh-group {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  white-space: nowrap;
}

.auto-refresh-label {
  display: inline-flex;
  align-items: center;
  color: var(--app-text-soft);
  font-size: 14px;
}

.result-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  text-align: left;
}

.detail-pre {
  margin: 0;
  max-height: 320px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
}

@media (max-width: 768px) {
  .auto-refresh-group {
    justify-content: flex-start;
  }

  .runlog-filter {
    width: 180px;
  }

  .runlog-filter--wide {
    width: 240px;
  }

  .result-cell {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
