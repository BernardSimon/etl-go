<template>
  <a-modal
    :open="open"
    :title="title || t('file.library.title')"
    :width="isNarrowScreen ? '96vw' : '920px'"
    :destroy-on-close="false"
    @cancel="handleCancel"
  >
    <div class="file-library">
      <div class="file-library-toolbar">
        <a-input-search
          v-model:value="keyword"
          :placeholder="t('file.library.searchPlaceholder')"
          allow-clear
          @search="handleSearch"
        />
        <a-space>
          <a-button @click="fetchFiles">
            {{ t('file.refresh') }}
          </a-button>
          <a-button type="primary" @click="showUpload = !showUpload">
            {{ showUpload ? t('file.library.hideUpload') : t('file.upload.button') }}
          </a-button>
        </a-space>
      </div>

      <a-card v-if="showUpload" size="small" class="upload-card">
        <a-upload-dragger
          v-model:file-list="uploadState.fileList"
          :before-upload="beforeUpload"
          :show-upload-list="true"
          :max-count="multiple ? 5 : 1"
          :multiple="multiple"
          :disabled="uploadState.loading"
        >
          <p class="ant-upload-drag-icon">
            <inbox-outlined />
          </p>
          <p class="ant-upload-text">{{ t('file.upload.select') }}</p>
          <p class="ant-upload-hint">{{ t('file.upload.hint') }}</p>
        </a-upload-dragger>

        <div v-if="uploadState.fileList.length > 0" class="upload-meta">
          {{ t('file.upload.selectedSize', { size: formatSelectedSize() }) }}
          <div v-if="uploadState.loading" class="upload-progress-text">
            {{ t('file.upload.batchProgress', { current: uploadState.currentIndex, total: uploadState.totalFiles }) }}
          </div>
          <div v-if="uploadState.loading && uploadState.currentFileName" class="upload-progress-text">
            {{ t('file.upload.currentFile', { name: uploadState.currentFileName }) }}
          </div>
        </div>

        <a-progress
          v-if="uploadState.loading"
          :percent="uploadState.progress"
          :status="uploadState.progress >= 100 ? 'success' : 'active'"
          style="margin-top: 12px"
        />

        <a-space class="upload-actions">
          <a-button
            type="primary"
            :loading="uploadState.loading"
            :disabled="uploadState.fileList.length === 0"
            @click="handleUpload"
          >
            {{ t('file.upload.modal.upload') }}
          </a-button>
          <a-button
            v-if="uploadState.loading"
            danger
            @click="cancelUpload"
          >
            {{ t('file.upload.cancelUpload') }}
          </a-button>
        </a-space>
      </a-card>

      <a-table
        row-key="id"
        :columns="columns"
        :data-source="files"
        :loading="loading"
        :pagination="pagination"
        :row-selection="rowSelection"
        :scroll="{ y: 'calc(100vh - 420px)' }"
        @change="handleTableChange"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'name'">
            <div class="file-name-cell">
              <div class="file-name">{{ record.name }}</div>
              <div class="file-secondary">
                {{ record.id }}
              </div>
            </div>
          </template>
          <template v-else-if="column.key === 'size'">
            {{ formatFileSize(record.size) }}
          </template>
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button size="small" @click="downloadFile(record)">
                {{ t('file.table.action.download') }}
              </a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </div>

    <template #footer>
      <a-space>
        <a-button @click="handleCancel">{{ t('common.cancel') }}</a-button>
        <a-button type="primary" @click="handleConfirm">
          {{ t('file.library.confirmSelection') }}
        </a-button>
      </a-space>
    </template>
  </a-modal>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch, onMounted, onUnmounted } from "vue";
import { message } from "ant-design-vue";
import type { UploadProps } from "ant-design-vue";
import { InboxOutlined } from "@ant-design/icons-vue";
import { useI18n } from "vue-i18n";
import { buildFileDownloadUrl, getFileList, uploadFile } from "../api/file";
import type { FileInfo } from "../types";
import { useUserStore } from "../stores/user";

interface Props {
  open: boolean;
  multiple?: boolean;
  selectedIds?: string[];
  title?: string;
}

const props = withDefaults(defineProps<Props>(), {
  multiple: false,
  selectedIds: () => [],
  title: "",
});

const emit = defineEmits<{
  (e: "update:open", value: boolean): void;
  (e: "confirm", files: FileInfo[]): void;
}>();

const { t } = useI18n();
const screenWidth = ref(window.innerWidth);
const isNarrowScreen = computed(() => screenWidth.value < 768);
const loading = ref(false);
const files = ref<FileInfo[]>([]);
const keyword = ref("");
const showUpload = ref(false);
const selectedRowKeys = ref<string[]>([]);
const selectedMap = ref<Record<string, FileInfo>>({});

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => t("common.pagination.total", { total }),
});

const uploadState = reactive({
  loading: false,
  progress: 0,
  fileList: [] as any[],
  controller: null as AbortController | null,
  currentFileName: "",
  currentIndex: 0,
  totalFiles: 0,
});

const columns = computed(() => [
  {
    title: t("file.table.column.name"),
    dataIndex: "name",
    key: "name",
  },
  {
    title: t("file.table.column.type"),
    dataIndex: "ex_name",
    key: "ex_name",
    width: 120,
    customRender: ({ text }: { text: string }) => text || "-",
  },
  {
    title: t("file.table.column.size"),
    dataIndex: "size",
    key: "size",
    width: 120,
  },
  {
    title: t("file.table.column.createdAt"),
    dataIndex: "created_at",
    key: "created_at",
    width: 200,
  },
  {
    title: t("file.table.column.actions"),
    key: "action",
    width: 120,
  },
]);

const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys: (string | number)[], rows: FileInfo[]) => {
    selectedRowKeys.value = keys.map((key) => String(key));
    const nextMap = { ...selectedMap.value };
    rows.forEach((row) => {
      nextMap[row.id] = row;
    });
    selectedMap.value = nextMap;
  },
  type: props.multiple ? ("checkbox" as const) : ("radio" as const),
}));

const handleResize = () => {
  screenWidth.value = window.innerWidth;
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
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};

const formatSelectedSize = () => {
  const total = uploadState.fileList.reduce((sum, item) => {
    const rawFile = item.originFileObj || item;
    return sum + (rawFile?.size || 0);
  }, 0);
  return formatFileSize(total);
};

const mergeSelectedFiles = (items: FileInfo[]) => {
  const nextMap = { ...selectedMap.value };
  items.forEach((item) => {
    nextMap[item.id] = item;
  });
  selectedMap.value = nextMap;
};

const loadSelectedFiles = async () => {
  if (!props.selectedIds.length) {
    selectedRowKeys.value = [];
    selectedMap.value = {};
    return;
  }

  selectedRowKeys.value = [...props.selectedIds];
  const res = await getFileList({
    ids: props.selectedIds.join(","),
    page_no: 1,
    page_size: Math.max(props.selectedIds.length, 10),
  });
  if (res.code === 0) {
    mergeSelectedFiles(res.data.list || []);
  }
};

const fetchFiles = async () => {
  loading.value = true;
  try {
    const res = await getFileList({
      page_no: pagination.current,
      page_size: pagination.pageSize,
      keyword: keyword.value.trim(),
    });
    if (res.code === 0) {
      files.value = res.data.list || [];
      pagination.total = res.data.total || 0;
      mergeSelectedFiles(files.value);
    }
  } finally {
    loading.value = false;
  }
};

const handleSearch = () => {
  pagination.current = 1;
  fetchFiles();
};

const handleTableChange = (pag: any) => {
  pagination.current = pag.current;
  pagination.pageSize = pag.pageSize;
  fetchFiles();
};

const beforeUpload: UploadProps["beforeUpload"] = (file) => {
  uploadState.fileList = props.multiple ? [...uploadState.fileList, file] : [file];
  return false;
};

const cancelUpload = () => {
  uploadState.controller?.abort();
};

const handleUpload = async () => {
  if (uploadState.fileList.length === 0) {
    message.warning(t("file.upload.select"));
    return;
  }

  uploadState.loading = true;
  uploadState.progress = 0;
  uploadState.controller = new AbortController();
  uploadState.totalFiles = uploadState.fileList.length;
  uploadState.currentIndex = 0;
  uploadState.currentFileName = "";

  try {
    const totalFiles = uploadState.fileList.length;
    const results: FileInfo[] = [];

    for (const [index, item] of uploadState.fileList.entries()) {
      uploadState.currentIndex = index + 1;
      uploadState.currentFileName = item?.name || item?.originFileObj?.name || "";

      const formData = new FormData();
      formData.append("file", item.originFileObj || item);
      const res = await uploadFile(formData, {
        signal: uploadState.controller?.signal,
        onUploadProgress: (event) => {
          if (!event.total) {
            return;
          }
          const currentFileProgress = event.loaded / event.total;
          uploadState.progress = Math.min(
            99,
            Math.round(((index + currentFileProgress) / totalFiles) * 100)
          );
        },
      });
      results.push(res.data);
    }

    uploadState.progress = 100;
    uploadState.fileList = [];
    await fetchFiles();
    mergeSelectedFiles(results);
    selectedRowKeys.value = props.multiple
      ? Array.from(new Set([...selectedRowKeys.value, ...results.map((item) => item.id)]))
      : results.slice(-1).map((item) => item.id);
    message.success(t("file.upload.success"));
  } catch (error: any) {
    if (error?.name === "CanceledError" || error?.code === "ERR_CANCELED") {
      message.warning(t("file.upload.cancelled"));
    } else {
      message.error(error?.message || t("request.failed"));
    }
  } finally {
    uploadState.loading = false;
    uploadState.controller = null;
    uploadState.currentFileName = "";
    uploadState.currentIndex = 0;
    uploadState.totalFiles = 0;
  }
};

const downloadFile = (record: any) => {
  const link = document.createElement("a");
  link.href = buildFileDownloadUrl(record, useUserStore().token);
  link.download = record.name;
  link.click();
};

const handleConfirm = () => {
  const result = selectedRowKeys.value
    .map((id) => selectedMap.value[id])
    .filter(Boolean);
  emit("confirm", result);
  emit("update:open", false);
};

const handleCancel = () => {
  emit("update:open", false);
};

watch(
  () => props.open,
  async (value) => {
    if (!value) {
      return;
    }
    keyword.value = "";
    showUpload.value = false;
    pagination.current = 1;
    await Promise.all([loadSelectedFiles(), fetchFiles()]);
  }
);

watch(
  () => props.selectedIds,
  (value) => {
    selectedRowKeys.value = [...value];
  }
);

onMounted(() => {
  window.addEventListener("resize", handleResize);
});

onUnmounted(() => {
  window.removeEventListener("resize", handleResize);
});
</script>

<style scoped lang="scss">
.file-library {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.file-library-toolbar {
  display: flex;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
}

.upload-card {
  border-radius: 12px;
}

.upload-actions {
  margin-top: 12px;
}

.upload-meta {
  margin-top: 8px;
  color: #666;
  font-size: 13px;
}

.upload-progress-text {
  margin-top: 4px;
}

.file-name-cell {
  display: flex;
  flex-direction: column;
}

.file-name {
  font-weight: 500;
  word-break: break-all;
}

.file-secondary {
  font-size: 12px;
  color: #8c8c8c;
  word-break: break-all;
}

@media (max-width: 768px) {
  .file-library-toolbar {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
