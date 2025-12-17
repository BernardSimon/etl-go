<template>
  <div>
  <div class="p-5">
    <a-card :bordered="false">
      <div class="flex flex-wrap gap-2 items-center justify-between mb-4">
        <!-- 左侧搜索条件 -->
        <a-form
          layout="inline"
          class="flex-nowrap"
          :model="searchForm"
          @finish="handleSearch"
        >
          <a-form-item :label="t('runLog.search.recordId')">
            <a-input
              v-model:value="searchForm.id"
              allow-clear
              :placeholder="t('runLog.search.recordId.placeholder')"
              style="width: 180px"
            />
          </a-form-item>

          <a-form-item :label="t('runLog.search.missionName')">            <a-input
              v-model:value="searchForm.mission_name"
              allow-clear
              :placeholder="t('runLog.search.missionName.placeholder')"
              style="width: 200px"
            />
          </a-form-item>
          <a-form-item :label="t('runLog.search.status')">            <a-select
              v-model:value="searchForm.status"
              style="width: 160px"
              allow-clear
            >
              <a-select-option
                v-for="item in getStatusOptions()"
                :key="item.value"
                :value="item.value"
              >
                {{ item.label }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-form>

        <!-- 右侧按钮组 -->
        <div class="flex gap-2">
          <a-button type="primary" @click="handleSearch" :loading="loading">
            {{ t('runLog.search.query') }}
          </a-button>
          <a-button @click="resetSearch">
            {{ t('runLog.search.reset') }}
          </a-button>        </div>
      </div>

      <a-table
        class="mt-5"
        :columns="getColumns()"
        :data-source="tableData"
        :loading="loading"
        :pagination="pagination"
        :scroll="{ y: 'calc(100vh - 470px)', x: 'max-content' }"
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
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button
                type="primary"
                size="small"
                @click="() => showParamsModal(record)"
              >
                查询参数
              </a-button>
              <a-button
                type="primary"
                danger
                :disabled="record.status !== 0"
                size="small"
                @click="handleCancel(record)"
              >
                中止
              </a-button>
              <a-button
                  type="primary"
                  size="small"
                  @click="showTaskFilesModal(record)"
              >
                任务文件
              </a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <MissionConfigModal
      v-model:open="paramsModal.show"
      :title="paramsModal.title"
      :mode="paramsModal.mode"
      :id="paramsModal.id"
      :data="paramsModal.data"
      :record="paramsModal.record"
    />
  </div>
  <a-modal
      v-model:open="taskFilesModal.visible"
      :title="t('runLog.taskFiles.modal.title')"
      :footer="null"
      width="80%"
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { getTaskRecordList, cancelTaskRecord } from "../api/run_log";
import { message, Modal } from "ant-design-vue";
import type { TablePaginationConfig } from "ant-design-vue";
import MissionConfigModal from "../components/MissionConfigModal.vue";
import { useI18n } from "vue-i18n";
import {deleteFile, getFileListByTaskRecordID} from "../../src/api/file.ts";
import {useUserStore} from "../../src/stores/user.ts"; // 新增引入

const { t } = useI18n(); // 初始化i18n实例
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
});

const loading = ref(false);
const tableData = ref<any[]>([]);
const pagination = reactive<TablePaginationConfig>({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total) => `共 ${total} 条`,
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
    title: t("runLog.table.column.actions"),
    key: "action",
    align: "center",
    fixed: "right",
    width: 170,
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
    });
    if (res && res.code === 0) {
      tableData.value = res.data.list || [];
      pagination.total = res.data.total || 0;
    } else {
      tableData.value = [];
      pagination.total = 0;
    }
  } catch (error) {
    console.error("获取运行记录失败", error);
  } finally {
    loading.value = false;
  }
};

// 查询
const handleSearch = () => {
  pagination.current = 1;
  fetchData();
};

// 重置
const resetSearch = () => {
  searchForm.mission_name = "";
  searchForm.status = -1;
  searchForm.id = "";
  handleSearch();
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

// 显示运行参数弹窗
const showParamsModal = (record: any) => {
  paramsModal.value.id = record.id;
  paramsModal.value.data = record.data || null; 
  paramsModal.value.record = record.task; // 这里传递的是mission
  paramsModal.value.show = true;
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
      title: "File ID",
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
  link.href = import.meta.env.VITE_API_BASE_URL + `/file/${record.path}/${record.id}${record.ex_name}?token=${token}`;
  link.download = record.name;
  link.click();
};

// 删除任务文件
const handleDeleteTaskFile = (record: any) => {
  Modal.confirm({
    title: t('file.delete.confirm.title'),
    content: t('file.delete.confirm.content'),
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
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};
 
onMounted(() => {
  fetchData();
});
</script>
