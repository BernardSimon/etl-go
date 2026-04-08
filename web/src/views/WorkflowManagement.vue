<template>
  <div class="workflow-container">
    <a-card :bordered="false" class="main-card">
      <div class="table-operations">
        <div class="left">
          <a-space>
          <a-button type="primary" @click="handleAdd">
            <template #icon>
              <PlusOutlined />
            </template>
            {{ t('workflow.add.button') }}
          </a-button>
          <a-button type="default" @click="handleAddManual">
            <template #icon>
              <PlusOutlined />
            </template>
            {{ t('workflow.addManual.button') }}
          </a-button>
          <a-button type="default" @click="templateModal.visible = true">
            {{ t('workflow.template.useButton') }}
          </a-button>
          </a-space>
        </div>

        <div class="right">
          <a-button
            class="refresh-button"
            shape="circle"
            :title="t('common.refresh')"
            :aria-label="t('common.refresh')"
            @click="fetchData"
          >
            <template #icon>
              <ReloadOutlined />
            </template>
          </a-button>
        </div>
      </div>

      <div class="workflow-toolbar">
        <div class="filter-bar">
          <a-input
            v-model:value="filters.keyword"
            :placeholder="t('workflow.search.keyword')"
            allow-clear
            @pressEnter="handleFilterChange"
          />
          <a-select
            v-model:value="filters.status"
            :placeholder="t('workflow.search.status')"
            allow-clear
          >
            <a-select-option :value="0">{{ t('workflow.status.paused') }}</a-select-option>
            <a-select-option :value="1">{{ t('workflow.status.scheduling') }}</a-select-option>
            <a-select-option :value="2">{{ t('workflow.status.error') }}</a-select-option>
          </a-select>
          <a-select
            v-model:value="filters.taskType"
            :placeholder="t('workflow.search.taskType')"
            allow-clear
          >
            <a-select-option value="scheduled">{{ t('workflow.taskType.scheduled') }}</a-select-option>
            <a-select-option value="manual">{{ t('workflow.taskType.manual') }}</a-select-option>
          </a-select>
          <a-button type="primary" class="search-button" @click="handleFilterChange">
            {{ t('common.search') }}
          </a-button>
        </div>

        <div class="batch-bar">
          <div class="batch-actions">
            <a-button :disabled="selectedRowKeys.length === 0" @click="handleBatchRun">
              {{ t('workflow.batch.start') }}
            </a-button>
            <a-button :disabled="selectedRowKeys.length === 0" @click="handleBatchStop">
              {{ t('workflow.batch.stop') }}
            </a-button>
            <a-button danger :disabled="selectedRowKeys.length === 0" @click="handleBatchDelete">
              {{ t('workflow.batch.delete') }}
            </a-button>
          </div>
          <span class="selection-hint">
            {{ t('workflow.batch.selected', { count: selectedRowKeys.length }) }}
          </span>
        </div>
      </div>

      <a-table
          :columns="getColumns()"
          :data-source="tableData"
          :row-selection="rowSelection"
          row-key="id"
          :loading="loading"
          :pagination="pagination"
          :scroll="{ y: 'calc(100vh - 470px)', x: 'max-content' }"
          :locale="{ emptyText: loadError ? t('common.loadFailed') : t('common.empty') }"
          @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'action'">
            <a-space>
              <a-button
                  type="primary"
                  size="small"
                  :disabled="record.status === 1"
                  @click="handleEdit(record.data, record.id, record)"
              >{{ t('workflow.action.edit') }}</a-button>
              <a-button
                  v-if="record.status !== 1"
                  type="default"
                  size="small"
                  class="success-button"
                  :disabled="record.cron === 'manual'"
                  @click="handleRun(record.id)"
              >{{ t('workflow.action.start') }}</a-button>
              <a-button
                  v-else
                  type="default"
                  size="small"
                  class="test-button"
                  @click="handleStop(record.id)"
              >{{ t('workflow.action.stop') }}</a-button>
              <a-button
                  type="default"
                  size="small"
                  class="success-button"
                  @click="handleRunOnce(record.id)"
              >{{ t('workflow.action.runOnce') }}</a-button>
              <a-dropdown>
                <a-button size="small">{{ t('workflow.action.more') }} ▾</a-button>
                <template #overlay>
                  <a-menu>
                    <a-menu-item @click="handleCopy(record)">{{ t('workflow.action.copy') }}</a-menu-item>
                    <a-menu-item @click="handlePreviewTask(record.id)">{{ t('missionConfig.preview.btn') }}</a-menu-item>
                    <a-menu-item @click="handleViewRecords(record)">{{ t('workflow.action.records') }}</a-menu-item>
                    <a-menu-divider />
                    <a-menu-item :disabled="record.status === 1" @click="record.status !== 1 && handleDelete(record.id)">
                      <span :class="record.status !== 1 ? 'text-danger' : ''">{{ t('workflow.action.delete') }}</span>
                    </a-menu-item>
                  </a-menu>
                </template>
              </a-dropdown>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <!-- 新增或编辑弹窗 -->
    <MissionConfigModal
        v-model:open="addOrEditDialog.show"
        :title="addOrEditDialog.title"
        :mode="addOrEditDialog.mode"
        :id="addOrEditDialog.id"
        :data="addOrEditDialog.data"
        :record="addOrEditDialog.record"
        :taskType = "addOrEditDialog.taskType"
        @success="fetchData"
        @templateSaved="fetchTemplates"
    />
    <!-- 数据预览弹窗 -->
    <a-modal
        v-model:open="previewVisible"
        :title="t('missionConfig.preview.modalTitle')"
        width="80vw"
        :footer="null"
        :destroy-on-close="true"
    >
      <a-spin :spinning="previewLoading">
        <div v-if="!previewLoading && previewColumns.length === 0" style="text-align:center;padding:40px;color:#939393;">
          {{ t('missionConfig.preview.empty') }}
        </div>
        <a-table
            v-else
            :columns="previewColumns"
            :data-source="previewRows"
            :row-key="'_key'"
            :scroll="{ x: 'max-content', y: 400 }"
            size="small"
            :pagination="false"
        />
      </a-spin>
    </a-modal>

    <a-modal
        v-model:open="templateModal.visible"
        :title="t('workflow.template.modalTitle')"
        :footer="null"
        :width="isNarrowScreen ? '96vw' : '720px'"
    >
      <a-table
          :columns="templateColumns"
          :data-source="taskTemplates"
          row-key="id"
          :pagination="false"
          :locale="{ emptyText: t('common.empty') }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'tasktypes'">
            {{ record.tasktypes === 'manual' ? t('workflow.taskType.manual') : t('workflow.taskType.scheduled') }}
          </template>
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button size="small" type="primary" @click="handleUseTemplate(record)">
                {{ t('workflow.template.useAction') }}
              </a-button>
              <a-button size="small" danger @click="handleDeleteTemplate(record.id)">
                {{ t('common.delete') }}
              </a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from "vue";
import { useRouter } from "vue-router";
import { PlusOutlined, ReloadOutlined } from "@ant-design/icons-vue";
import {
  getTaskAll,
  deleteTask,
  runTask,
  stopTask,
  runTaskOnce,
  getTaskTemplates,
  deleteTaskTemplate,
  previewTask,
} from "../api/mission";
import { message, Modal } from "ant-design-vue";
import type { TablePaginationConfig } from "ant-design-vue";
import MissionConfigModal from "../components/MissionConfigModal.vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();
const router = useRouter();
const screenWidth = ref(window.innerWidth);
const isNarrowScreen = computed(() => screenWidth.value < 768);

const handleResize = () => {
  screenWidth.value = window.innerWidth;
};

const loading = ref(false);
const loadError = ref(false);
const filters = ref({
  keyword: "",
  status: undefined as number | undefined,
  taskType: undefined as "scheduled" | "manual" | undefined,
});
const pagination = ref<TablePaginationConfig>({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total) => t("common.pagination.total", { total }),
});

const getColumns = (): any[] => [
  {
    title: t('workflow.table.column.missionName'),
    dataIndex: "mission_name",
    key: "mission_name",
    align: "center",
    width: 230,
  },
  {
    title: t('workflow.table.column.isRunning'),
    dataIndex: "is_running",
    key: "is_running",
    align: "center",
    width: 140,
    customRender: ({ text }: { text: boolean }) => (text ? t('common.yes') : t('common.no')),
  },
  {
    title: t('workflow.table.column.status'),
    dataIndex: "status",
    key: "status",
    align: "center",
    customRender: ({ text }: { text: number }) => {
      const statusMap: Record<number, string> = {
        0: t('workflow.status.paused'),
        1: t('workflow.status.scheduling'),
        2: t('workflow.status.error'),
      };
      return statusMap[text] || '';
    },
  },
  {
    title: t('workflow.table.column.cron'),
    dataIndex: "cron",
    key: "cron",
    align: "center",
  },
  {
    title: t('workflow.table.column.errorMessage'),
    dataIndex: "err_msg",
    key: "err_msg",
    align: "center",
  },
  {
    title: t('workflow.table.column.lastEndTime'),
    dataIndex: "last_end_time",
    key: "last_end_time",
    align: "center",
  },
  {
    title: t('workflow.table.column.lastRunTime'),
    dataIndex: "last_run_time",
    key: "last_run_time",
    align: "center",
  },
  {
    title: t('workflow.table.column.lastSuccessTime'),
    dataIndex: "last_success_time",
    key: "last_success_time",
    align: "center",
  },
  {
    title: t('workflow.table.column.updatedAt'),
    dataIndex: "updated_at",
    key: "updated_at",
    align: "center",
  },
  {
    title: t('workflow.table.column.actions'),
    key: "action",
    align: "center",
    fixed: "right",
    width: 300,
  },
];

const tableData = ref<any[]>([]);
const taskTemplates = ref<any[]>([]);
const selectedRowKeys = ref<string[]>([]);
const templateModal = ref({
  visible: false,
});
const templateColumns: any[] = [
  {
    title: t('workflow.template.column.name'),
    dataIndex: "name",
    key: "name",
    align: "center",
  },
  {
    title: t('workflow.template.column.taskType'),
    dataIndex: "tasktypes",
    key: "tasktypes",
    align: "center",
    width: 140,
  },
  {
    title: t('workflow.template.column.updatedAt'),
    dataIndex: "updated_at",
    key: "updated_at",
    align: "center",
    width: 180,
  },
  {
    title: t('workflow.table.column.actions'),
    key: "action",
    align: "center",
    width: 180,
  },
];

// 获取任务列表
const fetchData = () => {
  loading.value = true;
  loadError.value = false;
  getTaskAll({
    page_no: pagination.value.current as number,
    page_size: pagination.value.pageSize as number,
    mission_name: filters.value.keyword || undefined,
    search: filters.value.keyword || undefined,
    status: filters.value.status,
    tasktypes: filters.value.taskType,
  }).then((res: any) => {
    tableData.value = res.data?.list || [];
    pagination.value.total = res.data?.total || 0;
    pagination.value.current = res.data?.page_no || pagination.value.current;
    pagination.value.pageSize = res.data?.page_size || pagination.value.pageSize;
  }).catch(() => {
    loadError.value = true;
    tableData.value = [];
    pagination.value.total = 0;
    selectedRowKeys.value = [];
  }).finally(() => {
    loading.value = false;
  });
};

const rowSelection = ref({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys: (string | number)[]) => {
    selectedRowKeys.value = keys.map((item) => String(item));
    rowSelection.value.selectedRowKeys = selectedRowKeys.value;
  },
});

const handleFilterChange = () => {
  pagination.value.current = 1;
  fetchData();
};

const handleTableChange = (pag: TablePaginationConfig) => {
  pagination.value.current = pag.current || 1;
  pagination.value.pageSize = pag.pageSize || 10;
  fetchData();
};

const addOrEditDialog = ref<any>({
  show: false,
  title: t('workflow.add.title'),
  mode: "add",
  id: undefined,
  data: undefined,
  record: undefined,
  taskType: "scheduled",
});

// 新增任务
const handleAdd = async () => {
  addOrEditDialog.value.title = t('workflow.add.title');
  addOrEditDialog.value.mode = "add";
  addOrEditDialog.value.id = undefined;
  addOrEditDialog.value.data = undefined;
  addOrEditDialog.value.record = undefined;
  addOrEditDialog.value.show = true;
  addOrEditDialog.value.taskType = "scheduled";
};
const handleAddManual = async () => {
  addOrEditDialog.value.title = t('workflow.add.title');
  addOrEditDialog.value.mode = "add";
  addOrEditDialog.value.id = undefined;
  addOrEditDialog.value.data = undefined;
  addOrEditDialog.value.record = undefined;
  addOrEditDialog.value.show = true;
  addOrEditDialog.value.taskType = "manual";
};


const handleEdit = async (data: any, id: string, record: any) => {
  addOrEditDialog.value.title = t('workflow.edit.title');
  addOrEditDialog.value.mode = "edit";
  addOrEditDialog.value.id = id;
  addOrEditDialog.value.data = data;
  addOrEditDialog.value.record = record;
  addOrEditDialog.value.taskType = record?.cron === "manual" ? "manual" : "scheduled";
  addOrEditDialog.value.show = true;
};

const handleCopy = (record: any) => {
  addOrEditDialog.value.title = t('workflow.copy.title');
  addOrEditDialog.value.mode = "add";
  addOrEditDialog.value.id = undefined;
  addOrEditDialog.value.data = record.data;
  addOrEditDialog.value.record = record;
  addOrEditDialog.value.taskType = record?.cron === "manual" ? "manual" : "scheduled";
  addOrEditDialog.value.show = true;
};

const handleViewRecords = (record: any) => {
  router.push({
    path: "/run-logs",
    query: {
      task_id: record.id,
      mission_name: record.mission_name,
    },
  });
};

const fetchTemplates = () => {
  getTaskTemplates().then((res: any) => {
    taskTemplates.value = res.data?.list || [];
  });
};

const handleUseTemplate = (record: any) => {
  addOrEditDialog.value.title = t('workflow.template.createTitle');
  addOrEditDialog.value.mode = "add";
  addOrEditDialog.value.id = undefined;
  addOrEditDialog.value.data = record.data;
  addOrEditDialog.value.record = {
    mission_name: record.name,
    cron: record.cron,
  };
  addOrEditDialog.value.taskType = record.tasktypes === "manual" ? "manual" : "scheduled";
  addOrEditDialog.value.show = true;
  templateModal.value.visible = false;
};

const handleDeleteTemplate = (id: string) => {
  Modal.confirm({
    title: t('workflow.template.deleteConfirmTitle'),
    content: t('workflow.template.deleteConfirmContent'),
    onOk: () => {
      deleteTaskTemplate(id).then((res: any) => {
        if (res.code === 0) {
          message.success(t('workflow.template.deleteSuccess'));
          fetchTemplates();
        }
      });
    },
  });
};

// 删除任务
const handleDelete = (id: any) => {
  Modal.confirm({
    title: t('workflow.delete.confirm.title'),
    content: t('workflow.delete.confirm.content'),
    onOk: () => {
      if (!id) return;
      deleteTask({ id }).then((res: any) => {
        if (res.code === 0) {
          message.success(t('workflow.delete.success'));
          fetchData();
        }
      });
    },
  });
};

const patchTaskRow = (id: string, patch: Record<string, any>) => {
  const target = tableData.value.find((item) => item.id === id);
  if (target) {
    Object.assign(target, patch);
  }
};

const getSelectedRows = () => tableData.value.filter((item) => selectedRowKeys.value.includes(item.id));

const handleBatchAction = (
  title: string,
  content: string,
  action: (row: any) => Promise<any>,
  successMessage: string
) => {
  const rows = getSelectedRows();
  if (rows.length === 0) return;
  Modal.confirm({
    title,
    content,
    async onOk() {
      const results = await Promise.allSettled(rows.map((row) => action(row)));
      const successCount = results.filter((item) => item.status === "fulfilled").length;
      if (successCount > 0) {
        message.success(`${successMessage} (${successCount}/${rows.length})`);
      }
      selectedRowKeys.value = [];
      rowSelection.value.selectedRowKeys = [];
      fetchData();
    },
  });
};

const handleBatchRun = () => {
  handleBatchAction(
    t('workflow.batch.startTitle'),
    t('workflow.batch.startContent'),
    (row) => runTask({ id: row.id }),
    t('workflow.batch.startSuccess')
  );
};

const handleBatchStop = () => {
  handleBatchAction(
    t('workflow.batch.stopTitle'),
    t('workflow.batch.stopContent'),
    (row) => stopTask({ id: row.id }),
    t('workflow.batch.stopSuccess')
  );
};

const handleBatchDelete = () => {
  handleBatchAction(
    t('workflow.batch.deleteTitle'),
    t('workflow.batch.deleteContent'),
    (row) => deleteTask({ id: row.id }),
    t('workflow.batch.deleteSuccess')
  );
};

// 启动任务
const handleRun = (id: string) => {
  if (!id) return;
  Modal.confirm({
    title: t('workflow.start.confirm.title'),
    content: t('workflow.start.confirm.content'),
    onOk: () => runTask({ id })
        .then((res: any) => {
          if (res.code === 0) {
            message.success(t('workflow.start.success'));
            patchTaskRow(id, { status: 1, is_running: true });
          }
        })
        .catch((err: any) => {
          console.error("启动任务失败：", err);
        }),
  });
};

// 数据预览
const previewVisible = ref(false)
const previewLoading = ref(false)
const previewColumns = ref<{ title: string; dataIndex: string; key: string; ellipsis: boolean }[]>([])
const previewRows = ref<Record<string, any>[]>([])

const handlePreviewTask = async (id: string) => {
  if (!id) return
  previewLoading.value = true
  previewVisible.value = true
  previewColumns.value = []
  previewRows.value = []
  try {
    const res = await previewTask(id)
    if (res.code === 0) {
      previewColumns.value = (res.data.columns || []).map((col: string) => ({
        title: col,
        dataIndex: col,
        key: col,
        ellipsis: true,
      }))
      previewRows.value = (res.data.rows || []).map((row: Record<string, any>, i: number) => ({ ...row, _key: i }))
    }
  } finally {
    previewLoading.value = false
  }
}

// 停止任务
const handleStop = (id: string) => {
  if (!id) return;
  Modal.confirm({
    title: t('workflow.stop.confirm.title'),
    content: t('workflow.stop.confirm.content'),
    onOk: () => stopTask({ id })
        .then((res: any) => {
          if (res.code === 0) {
            message.success(t('workflow.stop.success'));
            patchTaskRow(id, { status: 0, is_running: false });
          }
        })
        .catch((err: any) => {
          console.error("停止任务失败：", err);
        }),
  });
};

// 手动执行任务一次
const handleRunOnce = (id: string) => {
  if (!id) return;
  Modal.confirm({
    title: t('workflow.runOnce.confirm.title'),
    content: t('workflow.runOnce.confirm.content'),
    onOk: () => runTaskOnce({ id })
        .then((res: any) => {
          if (res.code === 0) {
            message.success(res.message || t('workflow.runOnce.success'));
          }
        })
        .catch((err: any) => {
          console.error("执行任务失败：", err);
        }),
  });
};

// 挂载
onMounted(() => {
  window.addEventListener("resize", handleResize);
  fetchData();
  fetchTemplates();
});

onUnmounted(() => {
  window.removeEventListener("resize", handleResize);
});
</script>

<style scoped lang="scss">
.workflow-container {
  padding: 20px;

  .main-card {
    border-radius: 4px;
  }

  .table-operations {
    margin-bottom: 16px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
  }
}

.workflow-toolbar {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 16px;
}

.filter-bar {
  display: grid;
  grid-template-columns: minmax(220px, 1.5fr) repeat(2, minmax(160px, 1fr)) auto;
  gap: 12px;
  align-items: center;
}

.refresh-button {
  flex: 0 0 auto;
}

.search-button {
  min-width: 96px;
  white-space: nowrap;
}

.batch-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.batch-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.selection-hint {
  color: #6b7280;
  font-size: 14px;
  white-space: nowrap;
}

@media (max-width: 768px) {
  .workflow-container {
    padding: 12px;
  }

  .workflow-container .table-operations {
    flex-direction: column;
    gap: 12px;
  }

  .workflow-container .left,
  .workflow-container .right {
    width: 100%;
  }

  .filter-bar {
    grid-template-columns: 1fr;
  }

  .batch-bar,
  .batch-actions {
    flex-direction: column;
    align-items: stretch;
  }

  .workflow-container :deep(.filter-bar .ant-input),
  .workflow-container :deep(.filter-bar .ant-select),
  .workflow-container :deep(.table-operations .left .ant-btn),
  .batch-actions :deep(.ant-btn) {
    width: 100% !important;
  }

  .workflow-container .right {
    display: flex;
    justify-content: flex-end;
  }

  .workflow-container .refresh-button {
    width: 32px !important;
    min-width: 32px;
    height: 32px;
    padding: 0;
  }

  .workflow-container .search-button {
    width: 100%;
  }

  .selection-hint {
    white-space: normal;
  }
}
</style>
