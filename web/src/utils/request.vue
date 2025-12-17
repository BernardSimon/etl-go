<template>
  <div class="workflow-container">
    <a-card :bordered="false" class="main-card">
      <div class="table-operations">
        <div class="left">
          <a-button type="primary" @click="handleAdd">
            <template #icon>
              <PlusOutlined />
            </template>
            {{ t('workflow.add.button') }}
          </a-button>
        </div>
        <div class="right">
          <a-button shape="circle" @click="fetchData">
            <template #icon>
              <ReloadOutlined />
            </template>
          </a-button>
        </div>
      </div>

      <a-table
        :columns="getColumns()"
        :data-source="tableData"
        row-key="id"
        :loading="loading"
        :scroll="{ y: 'calc(100vh - 470px)', x: 'max-content' }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'action'">
            <a-space wrap>
              <a-button
                type="primary"
                size="small"
                @click="handleEdit(record.data, record.id, record)"
                >{{ t('workflow.action.edit') }}</a-button
              >
              <a-button
                type="default"
                size="small"
                class="success-button"
                @click="handleRun(record.id)"
                >{{ t('workflow.action.start') }}</a-button
              >
              <a-button
                type="default"
                size="small"
                class="test-button"
                @click="handleStop(record.id)"
                >{{ t('workflow.action.stop') }}</a-button
              >
              <a-button
                type="default"
                size="small"
                class="success-button"
                @click="handleRunOnce(record.id)"
                >{{ t('workflow.action.runOnce') }}</a-button
              >
              <a-button
                type="default"
                class="error-button"
                size="small"
                @click="handleDelete(record.id)"
                >{{ t('workflow.action.delete') }}</a-button
              >
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
      @success="fetchData"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { PlusOutlined, ReloadOutlined } from "@ant-design/icons-vue";
import {
  getTaskAll,
  deleteTask,
  runTask,
  stopTask,
  runTaskOnce,
} from "../api/mission";
import { message, Modal } from "ant-design-vue";
import MissionConfigModal from "../components/MissionConfigModal.vue";
import { useI18n } from "vue-i18n";
import {ColumnsType} from "ant-design-vue/es/table";

const { t } = useI18n();

const loading = ref(false);

const getColumns = (): ColumnsType => [
  {
    title: t('workflow.table.column.missionName'),
    dataIndex: "mission_name",
    key: "mission_name",
    align: "center",
  },
  {
    title: t('workflow.table.column.isRunning'),
    dataIndex: "isRunning",
    key: "isRunning",
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
    width: 350,
  },
];

const tableData = ref<any[]>([]);

// 获取任务列表
const fetchData = () => {
  loading.value = true;
  getTaskAll().then((res: any) => {
    tableData.value = res.data;
    loading.value = false;
  });
};

const addOrEditDialog = ref<any>({
  show: false,
  title: t('workflow.add.title'),
  mode: "add",
  id: undefined,
  data: undefined,
  record: undefined,
});

// 新增任务
const handleAdd = async () => {
  addOrEditDialog.value.title = t('workflow.add.title');
  addOrEditDialog.value.mode = "add";
  addOrEditDialog.value.id = undefined;
  addOrEditDialog.value.data = undefined;
  addOrEditDialog.value.record = undefined;
  addOrEditDialog.value.show = true;
};

const handleEdit = async (data: any, id: string, record: any) => {
  addOrEditDialog.value.title = t('workflow.edit.title');
  addOrEditDialog.value.mode = "edit";
  addOrEditDialog.value.id = id;
  addOrEditDialog.value.data = data; 
  addOrEditDialog.value.record = record; 
  addOrEditDialog.value.show = true;
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

// 启动任务
const handleRun = (id: string) => {
  if (!id) return;
  runTask({ id })
    .then((res: any) => {
      if (res.code === 0) {
        message.success(t('workflow.start.success'));
        fetchData();
      }
    })
    .catch((err: any) => { 
      console.error("启动任务失败：", err);
    });
};

// 停止任务
const handleStop = (id: string) => {
  if (!id) return;
  stopTask({ id })
    .then((res: any) => {
      if (res.code === 0) {
        message.success(t('workflow.stop.success'));
        fetchData();
      }
    })
    .catch((err: any) => {
      console.error("停止任务失败：", err);
    });
};

// 手动执行任务一次
const handleRunOnce = (id: string) => {
  if (!id) return;
  runTaskOnce({ id })
    .then((res: any) => {
      if (res.code === 0) {
        message.success(t('workflow.runOnce.success'));
        fetchData();
      }
    })
    .catch((err: any) => {
      console.error("执行任务失败：", err);
    });
};

// 挂载
onMounted(() => {
  fetchData();
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
  }
}
</style>
