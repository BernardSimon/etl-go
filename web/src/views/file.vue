<template>
  <PageContainer>
    <ContentCard>
      <div class="section-stack">
        <div class="table-toolbar">
          <div class="table-toolbar__right">
            <a-button @click="fetchFileList">
              {{ t("file.refresh") }}
            </a-button>
            <a-button type="primary" @click="showUploadModal">
              {{ t("file.upload.button") }}
            </a-button>
          </div>
        </div>

        <FilterBar>
          <div class="filter-field file-search">
            <span class="filter-field__label">{{ t('file.table.column.name') }}</span>
            <a-input
              v-model:value="searchKeyword"
              :placeholder="t('file.search.placeholder')"
              allow-clear
              @pressEnter="handleSearch"
            />
          </div>
          <div class="filter-field filter-field--actions">
            <a-button @click="handleSearch">
              {{ t('common.search') }}
            </a-button>
          </div>
        </FilterBar>

        <a-table
          class="app-shell-table"
          :columns="columns()"
          :data-source="fileList"
          :loading="loading"
          :pagination="pagination"
          :scroll="{ x: 'max-content' }"
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
                <div class="file-meta">{{ record.id }}</div>
              </div>
            </template>
            <template v-else-if="column.key === 'type'">
              <span class="status-pill">{{ record.ex_name || '-' }}</span>
            </template>
            <template v-else-if="column.key === 'action'">
              <a-space wrap>
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
      </div>
    </ContentCard>

    <!-- 上传文件 modal -->
    <a-modal
      v-model:open="uploadModal.visible"
      :footer="null"
      :closable="!uploadModal.loading"
      :mask-closable="!uploadModal.loading"
      :width="480"
      :title="null"
      @cancel="closeUploadModal"
    >
      <div class="upload-modal">
        <div class="upload-modal__header">
          <span class="upload-modal__title">{{ t('file.upload.modal.title') }}</span>
        </div>

        <!-- 拖拽区域（未选文件） -->
        <a-upload-dragger
          v-if="uploadModal.fileList.length === 0"
          v-model:file-list="uploadModal.fileList"
          :before-upload="beforeUpload"
          :max-count="1"
          :show-upload-list="false"
          class="upload-dropzone"
        >
          <div class="upload-dropzone__inner">
            <div class="upload-dropzone__icon">
              <inbox-outlined />
            </div>
            <p class="upload-dropzone__primary">{{ t('file.upload.select') }}</p>
            <p class="upload-dropzone__hint">{{ t('file.upload.hint') }}</p>
          </div>
        </a-upload-dragger>

        <!-- 已选文件卡片 -->
        <div v-else class="upload-file-card">
          <div class="upload-file-card__ext">
            {{ getSelectedFileExt() }}
          </div>
          <div class="upload-file-card__info">
            <div class="upload-file-card__name">{{ getSelectedFileName() }}</div>
            <div class="upload-file-card__size">{{ formatFileSize(getSelectedFileSize()) }}</div>
          </div>
          <button
            v-if="!uploadModal.loading"
            class="upload-file-card__remove"
            type="button"
            @click="uploadModal.fileList = []"
          >✕</button>
        </div>

        <!-- 上传进度 -->
        <div v-if="uploadModal.loading" class="upload-progress">
          <div class="upload-progress__label">
            <span v-if="chunkedState === 'completing'">{{ t('file.upload.assembling') }}</span>
            <span v-else-if="isChunkedUpload">
              {{ t('file.upload.chunkProgress', { current: chunkedCurrentChunk, total: chunkedTotalChunks }) }}
            </span>
            <span v-else>{{ t('file.upload.progress', { progress: uploadModal.progress }) }}</span>
            <span class="upload-progress__pct">{{ uploadModal.progress }}%</span>
          </div>
          <a-progress
            :percent="uploadModal.progress"
            :status="uploadModal.progress >= 100 ? 'success' : 'active'"
            :show-info="false"
          />
          <div class="upload-progress__controls">
            <a-button
              v-if="isChunkedUpload && chunkedState === 'paused'"
              size="small"
              @click="resumeUpload"
            >{{ t('file.upload.resume') }}</a-button>
            <a-button
              v-else-if="isChunkedUpload && chunkedState === 'uploading'"
              size="small"
              @click="pauseUpload"
            >{{ t('file.upload.pause') }}</a-button>
            <a-button size="small" danger @click="cancelUpload">
              {{ t('file.upload.cancelUpload') }}
            </a-button>
          </div>
        </div>

        <!-- 底部操作 -->
        <div class="upload-modal__footer">
          <a-button :disabled="uploadModal.loading" @click="closeUploadModal">
            {{ t('file.upload.modal.cancel') }}
          </a-button>
          <a-button
            type="primary"
            :loading="uploadModal.loading"
            :disabled="uploadModal.fileList.length === 0 || uploadModal.loading"
            @click="handleUpload"
          >
            {{ t('file.upload.modal.upload') }}
          </a-button>
        </div>
      </div>
    </a-modal>
  </PageContainer>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import {message, Modal, UploadProps} from 'ant-design-vue';
import { InboxOutlined } from '@ant-design/icons-vue';
import { useI18n } from 'vue-i18n';
import { buildFileDownloadUrl, getFileList, uploadFile, deleteFile, LARGE_FILE_THRESHOLD } from '../api/file';
import { useChunkedUpload, type ChunkedUploadState } from '../composables/useChunkedUpload';
import {useUserStore} from "../stores/user.ts";

const { t } = useI18n();

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

// Chunked upload instance (set while a large-file upload is in progress)
const chunkedUpload = ref<ReturnType<typeof useChunkedUpload> | null>(null);
const isChunkedUpload = computed(() => chunkedUpload.value !== null);
const chunkedState = computed<ChunkedUploadState>(() => chunkedUpload.value?.state ?? 'idle');
const chunkedCurrentChunk = computed(() => chunkedUpload.value?.currentChunk ?? 0);
const chunkedTotalChunks = computed(() => chunkedUpload.value?.totalChunks ?? 0);

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

const getSelectedFileExt = () => {
  const name = getSelectedFileName();
  const parts = name.split(".");
  return parts.length > 1 ? parts[parts.length - 1].toUpperCase() : "FILE";
};

const cancelUpload = async () => {
  if (chunkedUpload.value) {
    await chunkedUpload.value.cancel();
  } else {
    uploadModal.controller?.abort();
  }
};

const pauseUpload = () => chunkedUpload.value?.pause();
const resumeUpload = () => chunkedUpload.value?.resume();

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

  const rawFile = uploadModal.fileList[0].originFileObj || uploadModal.fileList[0];
  uploadModal.loading = true;
  uploadModal.progress = 0;

  if (rawFile.size >= LARGE_FILE_THRESHOLD) {
    // ── Chunked path ──────────────────────────────────────────────────────
    const upload = useChunkedUpload();
    chunkedUpload.value = upload;
    try {
      await upload.start(rawFile, (percent) => {
        uploadModal.progress = percent;
      });
      uploadModal.progress = 100;
      uploadModal.loading = false;
      chunkedUpload.value = null;
      message.success(t('file.upload.success'));
      closeUploadModal();
      fetchFileList();
    } catch (err: any) {
      if (err?.name === 'CanceledError' || err?.code === 'ERR_CANCELED' || upload.state.value === 'cancelled') {
        message.warning(t('file.upload.cancelled'));
      } else {
        message.error(err?.message || t('request.failed'));
      }
    } finally {
      uploadModal.loading = false;
      chunkedUpload.value = null;
    }
  } else {
    // ── Small-file path ───────────────────────────────────────────────────
    uploadModal.controller = new AbortController();
    try {
      const formData = new FormData();
      formData.append('file', rawFile);
      const res = await uploadFile(formData, {
        signal: uploadModal.controller.signal,
        onUploadProgress: (event) => {
          if (!event.total) return;
          uploadModal.progress = Math.min(99, Math.round((event.loaded / event.total) * 100));
        },
      });
      if (res && res.code === 0) {
        uploadModal.progress = 100;
        uploadModal.loading = false;
        uploadModal.controller = null;
        message.success(t('file.upload.success'));
        closeUploadModal();
        fetchFileList();
      } else {
        message.error(res?.message || t('request.failed'));
      }
    } catch (error: any) {
      if (error?.name === 'CanceledError' || error?.code === 'ERR_CANCELED') {
        message.warning(t('file.upload.cancelled'));
      } else {
        message.error(error?.message || t('request.failed'));
      }
    } finally {
      uploadModal.loading = false;
      uploadModal.controller = null;
    }
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
  fetchFileList();
});
</script>

<style scoped lang="scss">
.file-search {
  width: 320px;
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
  color: var(--app-text-faint);
  word-break: break-all;
}

// ── Upload modal ──────────────────────────────────────────────────────────────

.upload-modal {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 4px 0 0;
}

.upload-modal__header {
  display: flex;
  align-items: center;
}

.upload-modal__title {
  font-size: 16px;
  font-weight: 600;
  color: var(--app-text);
}

.upload-dropzone {
  :deep(.ant-upload-drag) {
    border-radius: var(--app-radius-md) !important;
    border-color: var(--app-border) !important;
    background: var(--app-surface-muted) !important;
    transition: border-color 0.2s, background 0.2s;
  }

  :deep(.ant-upload-drag:hover) {
    border-color: var(--app-primary) !important;
    background: var(--app-primary-soft) !important;
  }
}

.upload-dropzone__inner {
  padding: 32px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.upload-dropzone__icon {
  font-size: 40px;
  color: var(--app-primary);
  line-height: 1;
}

.upload-dropzone__primary {
  font-size: 14px;
  font-weight: 600;
  color: var(--app-text);
  margin: 0;
}

.upload-dropzone__hint {
  font-size: 12px;
  color: var(--app-text-faint);
  margin: 0;
}

.upload-file-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 16px;
  border-radius: var(--app-radius-md);
  border: 1px solid var(--app-border);
  background: var(--app-surface-muted);
}

.upload-file-card__ext {
  flex: 0 0 auto;
  width: 44px;
  height: 44px;
  border-radius: var(--app-radius-sm);
  background: var(--app-primary-soft);
  color: var(--app-primary);
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.04em;
  display: grid;
  place-items: center;
}

.upload-file-card__info {
  flex: 1;
  min-width: 0;
}

.upload-file-card__name {
  font-size: 14px;
  font-weight: 500;
  color: var(--app-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.upload-file-card__size {
  font-size: 12px;
  color: var(--app-text-faint);
  margin-top: 2px;
}

.upload-file-card__remove {
  flex: 0 0 auto;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 1px solid var(--app-border);
  background: transparent;
  color: var(--app-text-soft);
  font-size: 14px;
  cursor: pointer;
  display: grid;
  place-items: center;
  transition: all 0.15s;

  &:hover {
    background: var(--app-danger);
    border-color: var(--app-danger);
    color: white;
  }

  &:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
}

.upload-progress {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 14px 16px;
  border-radius: var(--app-radius-md);
  border: 1px solid var(--app-border);
  background: var(--app-surface-muted);
}

.upload-progress__label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  color: var(--app-text-soft);
}

.upload-progress__pct {
  font-weight: 600;
  color: var(--app-primary);
}

.upload-progress__controls {
  display: flex;
  gap: 8px;
  margin-top: 4px;
}

.upload-modal__footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 4px;
  border-top: 1px solid var(--app-border);
}

@media (max-width: 768px) {
  .file-search {
    width: 240px;
  }
}
</style>
