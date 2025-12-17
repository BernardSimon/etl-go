<template>
  <div class="p-5">
    <a-card :bordered="false">
      <div class="flex flex-wrap gap-2 items-center justify-between mb-4">
        <h2>{{ t('file.title') }}</h2>

        <!-- 右侧按钮组 -->
        <div class="flex gap-2">
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
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button type="primary" size="small" @click="downloadFile(record)">
                {{ t('file.table.action.download') }}
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
        @ok="handleUpload"
        @cancel="closeUploadModal"
        :confirm-loading="uploadModal.loading"
        :ok-text="t('file.upload.modal.upload')"
        :cancel-text="t('file.upload.modal.cancel')"
    >
      <a-upload-dragger
          v-model:file-list="uploadModal.fileList"
          :before-upload="beforeUpload"
          :max-count="1"
          :show-upload-list="true"
      >
        <p class="ant-upload-drag-icon">
          <inbox-outlined></inbox-outlined>
        </p>
        <p class="ant-upload-text">{{ t('file.upload.select') }}</p>
      </a-upload-dragger>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import {message, Modal, UploadProps} from 'ant-design-vue';
import { InboxOutlined } from '@ant-design/icons-vue';
import { useI18n } from 'vue-i18n';
import { getFileList, uploadFile, deleteFile } from '../api/file';
import {useUserStore} from "../stores/user.ts";

const { t } = useI18n();

const loading = ref(false);
const fileList = ref<any[]>([]);

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`
});

const columns=():any[] =>{return [
  {
    title: "File ID",
    dataIndex: 'id',
    key: 'id',
    align: 'center',
    width:200
  },
  {
    title: t('file.table.column.name'),
    dataIndex: 'name',
    key: 'name',
    align: 'center'
  },
  {
    title:  t('file.table.column.path'),
    dataIndex: 'path',
    key: 'path',
    align: 'center',
    width:200
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
  fileList: [] as any[]
});

// 格式化文件大小
const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

// 获取文件列表
const fetchFileList = async () => {
  loading.value = true;
  try {
    const res = await getFileList({
      page_size: pagination.pageSize,
      page_no: pagination.current
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
};

// 关闭上传模态框
const closeUploadModal = () => {
  uploadModal.visible = false;
  uploadModal.fileList = [];
};

// 上传前检查
const beforeUpload: UploadProps['beforeUpload'] = file => {
  uploadModal.fileList = [file];
  return false;
};

// 处理文件上传
const handleUpload = async () => {
  if (uploadModal.fileList.length === 0) {
    message.warning(t('file.upload.select'));
    return;
  }
  uploadModal.loading = true;
  try {
    const formData = new FormData();
    formData.append('file', uploadModal.fileList[0].originFileObj);
    const res = await uploadFile(formData);

    if (res && res.code === 0) {
      message.success(t('file.upload.success'));
      closeUploadModal();
      fetchFileList();
    }
  } finally {
    uploadModal.loading = false;
  }
};

// 下载文件
const downloadFile = (record: any) => {
  const link = document.createElement('a');
  const token = useUserStore().token
  link.href = import.meta.env.VITE_API_BASE_URL +`/file/${record.path}/${record.id}${record.ex_name}?token=${token}`;
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
