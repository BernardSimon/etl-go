<template>
  <div class="p-5">
    <a-card :bordered="false">
      <div class="file-toolbar flex flex-wrap gap-2 items-center justify-between mb-4">
        <h2>{{ t('file.title') }}</h2>

        <div class="file-toolbar-actions">
          <a-input-search
            v-model:value="searchKeyword"
            :placeholder="t('file.search.placeholder')"
            allow-clear
            @search="handleSearch"
          />
          <a-button @click="fetchFileList">
            {{ t('file.refresh') }}
          </a-button>
          <a-button type="primary" @click="showUploadModal">
            {{ t('file.upload.button') }}
          </a-button>
        </div>
      </div>

      <a-table
          class="mt-5"
          :columns="columns()"
          :data-source="fileList"
          :loading="loading"
          :pagination="pagination"
          :scroll="{ y: 'calc(100vh - 350px)' }"
          @change="handleTableChange"
          row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'size'">
            {{ formatFileSize(record.size) }}
          </template>
          <template v-else-if="column.key === 'name'">
            <div class="file-name-cell">
              <div class="file-name">{{ record.name }}</div>
              <div class="file-meta">
                {{ record.id }}
              </div>
            </div>
          </template>
          <template v-else-if="column.key === 'type'">
            {{ record.ex_name || '-' }}
          </template>
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button type="primary" size="small" @click="downloadFile(record)">
                {{ t('file.table.action.download') }}
              </a-button>
              <a-button size="small" @click="copyFileId(record.id)">
                {{ t('file.table.action.copyId') }}
              </a-button>
              <a-button type="primary" danger size="small" @click="handleDeleteFile(record)">
                {{ t('file.table.action.delete') }}
              </a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <!-- 上传文件模态框 -->
    <a-modal
        v-model:open="uploadModal.visible"
        :title="t('file.upload.modal.title')"
        :width="isNarrowScreen ? '96vw' : '520px'"
        @ok="handleUpload"
        @cancel="closeUploadModal"
        :confirm-loading="uploadModal.loading"
        :ok-text="t('file.upload.modal.upload')"
        :cancel-text="t('file.upload.modal.cancel')"
        :mask-closable="!uploadModal.loading"
        :closable="!uploadModal.loading"
        :ok-button-props="{ disabled: uploadModal.fileList.length === 0 || uploadModal.loading }"
    >
      <a-upload-dragger
          v-model:file-list="uploadModal.fileList"
          :before-upload="beforeUpload"
          :max-count="1"
          :show-upload-list="true"
          :disabled="uploadModal.loading"
      >
        <p class="ant-upload-drag-icon">
          <inbox-outlined></inbox-outlined>
        </p>
        <p class="ant-upload-text">{{ t('file.upload.select') }}</p>
        <p class="ant-upload-hint">{{ t('file.upload.hint') }}</p>
      </a-upload-dragger>
      <div v-if="uploadModal.fileList.length > 0" class="upload-meta">
        <div>{{ t('file.upload.selectedSize', { size: formatFileSize(getSelectedFileSize()) }) }}</div>
        <div>{{ t('file.upload.currentFile', { name: getSelectedFileName() }) }}</div>
        <div v-if="uploadModal.loading">{{ t('file.upload.progress', { progress: uploadModal.progress }) }}</div>
      </div>
      <a-progress
          v-if="uploadModal.loading"
          :percent="uploadModal.progress"
          :status="uploadModal.progress >= 100 ? 'success' : 'active'"
          style="margin-top: 12px"
      />
      <a-button
          v-if="uploadModal.loading"
          danger
          block
          style="margin-top: 12px"
          @click="cancelUpload"
      >
        {{ t('file.upload.cancelUpload') }}
      </a-button>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, computed } from 'vue';
import {message, Modal, UploadProps} from 'ant-design-vue';
import { InboxOutlined } from '@ant-design/icons-vue';
import { useI18n } from 'vue-i18n';
import { buildFileDownloadUrl, getFileList, uploadFile, deleteFile } from '../api/file';
import {useUserStore} from "../stores/user.ts";

const { t } = useI18n();
const screenWidth = ref(window.innerWidth);
const isNarrowScreen = computed(() => screenWidth.value < 768);
const handleResize = () => {
  screenWidth.value = window.innerWidth;
};

const loading = ref(false);
const fileList = ref<any[]>([]);
const searchKeyword = ref("");

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => t("common.pagination.total", { total })
});

const columns=():any[] =>{return [
  {
    title: t('file.table.column.name'),
    dataIndex: 'name',
    key: 'name',
    align: 'left'
  },
  {
    title:  t('file.table.column.type'),
    dataIndex: 'ex_name',
    key: 'type',
    align: 'center',
    width:120
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
]};

const uploadModal = reactive({
  visible: false,
  loading: false,
  progress: 0,
  fileList: [] as any[],
  controller: null as AbortController | null,
});

// 格式化文件大小
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

// 获取文件列表
const fetchFileList = async () => {
  loading.value = true;
  try {
    const res = await getFileList({
      page_size: pagination.pageSize,
      page_no: pagination.current,
      keyword: searchKeyword.value.trim(),
    });

    if (res && res.code === 0) {
      fileList.value = res.data.list || [];
      pagination.total = res.data.total || 0;
    } else {
      fileList.value = [];
      pagination.total = 0;
    }
  }  finally {
    loading.value = false;
  }
};

const handleSearch = () => {
  pagination.current = 1;
  fetchFileList();
};

// 处理表格分页变化
const handleTableChange = (pag: any) => {
  pagination.current = pag.current;
  pagination.pageSize = pag.pageSize;
  fetchFileList();
};

// 显示上传模态框
const showUploadModal = () => {
  uploadModal.visible = true;
  uploadModal.fileList = [];
  uploadModal.progress = 0;
  uploadModal.controller = null;
};

// 关闭上传模态框
const closeUploadModal = () => {
  if (uploadModal.loading) {
    return;
  }
  uploadModal.visible = false;
  uploadModal.fileList = [];
  uploadModal.progress = 0;
  uploadModal.controller = null;
};

// 上传前检查
const beforeUpload: UploadProps['beforeUpload'] = file => {
  uploadModal.fileList = [file];
  return false;
};

const getSelectedFileSize = () => {
  const selectedFile = uploadModal.fileList[0];
  const rawFile = selectedFile?.originFileObj || selectedFile;
  return rawFile?.size || 0;
};

const getSelectedFileName = () => {
  const selectedFile = uploadModal.fileList[0];
  return selectedFile?.name || selectedFile?.originFileObj?.name || "-";
};

const cancelUpload = () => {
  uploadModal.controller?.abort();
};

const copyFileId = async (id: string) => {
  try {
    await navigator.clipboard.writeText(id);
    message.success(t("file.copyId.success"));
  } catch {
    message.error(t("file.copyId.failed"));
  }
};

// 处理文件上传
const handleUpload = async () => {
  if (uploadModal.fileList.length === 0) {
    message.warning(t('file.upload.select'));
    return;
  }
  uploadModal.loading = true;
  uploadModal.progress = 0;
  uploadModal.controller = new AbortController();
  try {
    const formData = new FormData();
    formData.append('file', uploadModal.fileList[0].originFileObj || uploadModal.fileList[0]);
    const res = await uploadFile(formData, {
      signal: uploadModal.controller.signal,
      onUploadProgress: (event) => {
        if (!event.total) {
          return;
        }
        uploadModal.progress = Math.min(99, Math.round((event.loaded / event.total) * 100));
      },
    });

    if (res && res.code === 0) {
      uploadModal.progress = 100;
      message.success(t('file.upload.success'));
      uploadModal.loading = false;
      closeUploadModal();
      fetchFileList();
    }
  } catch (error: any) {
    if (error?.name === "CanceledError" || error?.code === "ERR_CANCELED") {
      message.warning(t('file.upload.cancelled'));
      return;
    }
    message.error(error?.message || t('request.failed'));
  } finally {
    uploadModal.loading = false;
    uploadModal.controller = null;
  }
};

// 下载文件
const downloadFile = (record: any) => {
  const link = document.createElement('a');
  const token = useUserStore().token
  link.href = buildFileDownloadUrl(record, token);
  link.download = record.name;
  link.click();
};

// 删除文件
const handleDeleteFile = (record: any) => {
  Modal.confirm({
    title: t('file.delete.confirm.title'),
    content: t('file.delete.confirm.content'),
    onOk: async () => {
      try {
        const res = await deleteFile({ id: record.id });
        if (res && res.code === 0) {
          message.success(t('file.delete.success'));
          fetchFileList();
        }
      }finally {
        return ;
      }
    }
  });
};

onMounted(() => {
  window.addEventListener("resize", handleResize);
  fetchFileList();
});

onUnmounted(() => {
  window.removeEventListener("resize", handleResize);
});
</script>

<style scoped lang="scss">
.file-toolbar {
  gap: 12px;
}

.file-toolbar-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.file-toolbar-actions :deep(.ant-input-search) {
  width: 260px;
}

.file-name-cell {
  display: flex;
  flex-direction: column;
}

.file-name {
  font-weight: 500;
  word-break: break-all;
}

.file-meta {
  font-size: 12px;
  color: #8c8c8c;
  word-break: break-all;
}

@media (max-width: 768px) {
  .p-5 {
    padding: 12px;
  }

  .file-toolbar,
  .file-toolbar > div,
  .file-toolbar-actions {
    width: 100%;
  }

  .file-toolbar > div:last-child {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .file-toolbar-actions {
    flex-direction: column;
    align-items: stretch;
  }

  .file-toolbar-actions :deep(.ant-input-search) {
    width: 100%;
  }

  .file-toolbar :deep(.ant-btn) {
    width: 100%;
  }
}
</style>
