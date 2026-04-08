<template>
  <div class="data-source-container">
    <a-card :bordered="false">
      <div class="table-operations">
        <div class="left">
          <a-space>
          <a-button type="primary" @click="handleOpenAddDialog">
            <template #icon>
              <PlusOutlined />
            </template>
            {{ $t("datasource.add.title") }}
          </a-button>
          </a-space>
        </div>
        <div class="right">
          <a-button
            shape="circle"
            :title="t('common.refresh')"
            :aria-label="t('common.refresh')"
            @click="fetchDataSourceList"
          >
            <template #icon>
              <ReloadOutlined />
            </template>
          </a-button>
        </div>
      </div>

      <a-space class="filter-bar" style="margin-bottom: 16px" wrap>
        <a-input
          v-model:value="filters.keyword"
          :placeholder="t('datasource.search.keyword')"
          allow-clear
          style="width: 220px"
        />
        <a-select
          v-model:value="filters.type"
          :placeholder="t('datasource.search.type')"
          allow-clear
          style="width: 180px"
        >
          <a-select-option
            v-for="item in dataSourceTypeList"
            :key="item.type"
            :value="item.type"
          >
            {{ item.type }}
          </a-select-option>
        </a-select>
      </a-space>

      <a-table
          :columns="getColumns()"
          :data-source="filteredTableData"
          bordered
          row-key="id"
          :loading="loading"
          :scroll="{ y: 'calc(100vh - 470px)', x: 'max-content' }"
          :locale="{ emptyText: loadError ? t('common.loadFailed') : t('common.empty') }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'action'">
            <a-space>
              <a-button
                  type="primary"
                  class="edit-button"
                  size="small"
                  @click="handleEdit(record)"
              >{{ $t("datasource.action.edit") }}</a-button
              >
              <a-button
                  size="small"
                  @click="handleClone(record)"
              >{{ $t("datasource.action.clone") }}</a-button
              >
              <a-button
                  type="primary"
                  danger
                  class="delete-button"
                  size="small"
                  @click="handleDelete(record)"
              >{{ $t("datasource.action.delete") }}</a-button
              >
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <a-modal
        v-model:open="addDataSourceDialog.show"
        :title="addDataSourceDialog.title"
        width="720px"
        @cancel="handleCancel"
    >
      <a-form
          ref="addDataSourceFormRef"
          :model="addDataSourceDialog.form"
          :rules="formRules"
          :label-col="{ span: 4 }"
          :wrapper-col="{ span: 19 }"
      >
        <a-form-item :label="$t('datasource.name.label')" name="name">
          <a-input
              v-model:value="addDataSourceDialog.form.name"
              :placeholder="$t('datasource.name.placeholder')"
          />
        </a-form-item>
        <a-form-item :label="$t('datasource.type.label')" name="type">
          <a-select
              v-model:value="addDataSourceDialog.form.type"
              :placeholder="$t('datasource.type.placeholder')"
              @change="onDatasourceTypeChange"
          >
            <a-select-option
                v-for="item in dataSourceTypeList"
                :key="item.type"
                :value="item.type"
            >
              {{ item.type }}
            </a-select-option>
          </a-select>
        </a-form-item>

        <!-- 动态参数表单项 -->
        <a-form-item
            v-for="(param, index) in addDataSourceDialog.form.data"
            :key="index"
            :label="param.key"
            :name="['data', index, 'value']"
            :rules="[{ required: param.required, message: isFileParam(param) ? t('datasource.fileSelector.required') : t('datasource.param.required', { param: param.key }) }]"
        >
          <template v-if="isFileParam(param)">
            <div class="file-param-block">
              <div class="file-param-field">
              <div v-if="getSelectedFilesForParam(param).length" class="selected-file-tags">
                <a-tag
                  v-for="item in getSelectedFilesForParam(param)"
                  :key="item.id"
                  closable
                  @close.prevent="removeSelectedFile(param, item.id)"
                >
                  {{ item.name }}
                </a-tag>
              </div>
              <div class="file-param-actions">
                <a-button @click="openFileLibrary(param)">
                  {{ t('datasource.fileSelector.choose') }}
                </a-button>
                <a-button v-if="String(param.value || '').trim()" @click="clearFileParam(param)">
                  {{ t('datasource.fileSelector.clear') }}
                </a-button>
              </div>
              </div>
              <div class="param-help param-help-file">
                <span v-if="param.required" class="required">{{ t('datasource.param.requiredTag') }}</span>
                <span>{{ t('datasource.fileSelector.help') }}</span>
              </div>
            </div>
          </template>
          <a-input-password
              v-else-if="isMaskParam(param)"
              v-model:value="param.value"
              :placeholder="param.placeholder || param.description || t('datasource.param.placeholder', { param: param.key })"
          />
          <a-input
              v-else
              v-model:value="param.value"
              :placeholder="param.placeholder || param.description || t('datasource.param.placeholder', { param: param.key })"
          />
          <div class="param-help">
            <template v-if="!isFileParam(param)">
              <span v-if="param.required" class="required">{{ t('datasource.param.requiredTag') }}</span>
              <span v-if="param.description">{{ param.description }}</span>
              <span v-else>{{ t('datasource.param.noDescription') }}</span>
              <span v-if="param.defaultValue"> · {{ t('datasource.param.defaultValue', { value: param.defaultValue }) }}</span>
              <span v-if="param.example"> · {{ t('datasource.param.example', { value: param.example }) }}</span>
              <span v-if="param.type"> · {{ t('datasource.param.type', { value: param.type }) }}</span>
            </template>
          </div>
        </a-form-item>
      </a-form>
      <template #footer>
        <a-space>
          <a-button @click="handleCancel">{{ t('common.cancel') }}</a-button>
          <a-button
            :loading="testing"
            :disabled="!addDataSourceDialog.form.type"
            @click="handleTestDataSource"
          >
            {{ t('datasource.action.test') }}
          </a-button>
          <a-button type="primary" @click="handleAddDataSource">
            {{ t('common.confirm') }}
          </a-button>
        </a-space>
      </template>
    </a-modal>
    <FileLibraryModal
      v-model:open="fileLibraryOpen"
      :multiple="activeFileParamMultiple"
      :selected-ids="activeFileParamSelectedIds"
      :title="t('datasource.fileSelector.modalTitle')"
      @confirm="handleFileLibraryConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import {
  getDataSourceTypeList,
  addDataSource,
  getDataSourceList,
  getDataSourceById,
  testDataSource,
  deleteDataSource,
} from "../api/datasource.ts";
import { getTypeByComponent } from "../api/mission";
import { useI18n } from "vue-i18n";

import { message, Modal } from "ant-design-vue";
import { PlusOutlined, ReloadOutlined } from "@ant-design/icons-vue";
import type { FormInstance } from "ant-design-vue";
import type { FileInfo, Params } from "@/src/types";
import {SelectValue} from "ant-design-vue/es/select";
import {RuleObject} from "ant-design-vue/es/form";
import { ApiRequestError } from "../utils/request";
import { getFileList } from "../api/file";
import FileLibraryModal from "../components/FileLibraryModal.vue";

const { t } = useI18n();

interface DataSourceParam {
  key: string;
  value: string;
  description?: string;
  required: boolean;
  defaultValue?: string;
  placeholder?: string;
  example?: string;
  type?: string;
  mask?: boolean;
}

// interface DataSourceForm {
//   id?: string;
//   name: string;
//   type: string;
//   data: DataSourceParam[];
// }

const dataSourceTypeList = ref<{ type: string; params: Params[] }[]>([]);

const addDataSourceDialog = ref({
  show: false,
  title: t("datasource.add.title"),
  isEdit: false,
  form: {
    id: undefined as string | undefined,
    name: "",
    type: "",
    data: [] as DataSourceParam[]
  }
});

// 表单引用
const addDataSourceFormRef = ref<FormInstance>();

// 表单验证规则
const formRules = computed<{ [k: string]: RuleObject | RuleObject[] }>(() => ({
  name: [{ required: true, message: t("datasource.name.placeholder"), trigger: "blur" }],
  type: [{ required: true, message: t("datasource.type.placeholder"), trigger: "change" }]
}));

const tableData = ref<any[]>([]);
const loading = ref(false);
const loadError = ref(false);
const testing = ref(false);
const fileLibraryOpen = ref(false);
const activeFileParam = ref<DataSourceParam | null>(null);
const selectedFileMap = ref<Record<string, FileInfo>>({});
const filters = ref({
  keyword: "",
  type: undefined as string | undefined,
});

const filteredTableData = computed(() => {
  return tableData.value.filter((item) => {
    const matchKeyword = !filters.value.keyword ||
      item.name?.toLowerCase().includes(filters.value.keyword.toLowerCase());
    const matchType = !filters.value.type || item.type === filters.value.type;
    return matchKeyword && matchType;
  });
});

const isMaskParam = (param: DataSourceParam) => !!param.mask;
const isFileParam = (param: DataSourceParam) => String(param.key || "").includes("file_id");
const isMultiFileParam = (param: DataSourceParam) => String(param.key || "").includes("file_ids");
const parseFileIds = (value: string) =>
  String(value || "")
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);

const mergeSelectedFiles = (items: FileInfo[]) => {
  const nextMap = { ...selectedFileMap.value };
  items.forEach((item) => {
    nextMap[item.id] = item;
  });
  selectedFileMap.value = nextMap;
};

const refreshSelectedFilesFromForm = async () => {
  const ids = Array.from(
    new Set(
      addDataSourceDialog.value.form.data
        .filter((param) => isFileParam(param))
        .flatMap((param) => parseFileIds(param.value))
    )
  );

  if (!ids.length) {
    selectedFileMap.value = {};
    return;
  }

  const res = await getFileList({
    ids: ids.join(","),
    page_no: 1,
    page_size: Math.max(ids.length, 10),
  });
  if (res.code === 0) {
    mergeSelectedFiles(res.data.list || []);
  }
};

const getSelectedFilesForParam = (param: DataSourceParam) =>
  parseFileIds(param.value).map((id) => selectedFileMap.value[id] || ({
    id,
    name: id,
    size: 0,
    created_at: "",
    path: "",
    ex_name: "",
  }));

const openFileLibrary = (param: DataSourceParam) => {
  activeFileParam.value = param;
  fileLibraryOpen.value = true;
};

const clearFileParam = (param: DataSourceParam) => {
  param.value = "";
};

const removeSelectedFile = (param: DataSourceParam, id: string) => {
  const nextIds = parseFileIds(param.value).filter((item) => item !== id);
  param.value = isMultiFileParam(param) ? nextIds.join(",") : "";
};

const activeFileParamMultiple = computed(() => activeFileParam.value ? isMultiFileParam(activeFileParam.value) : false);
const activeFileParamSelectedIds = computed(() => activeFileParam.value ? parseFileIds(activeFileParam.value.value) : []);

const handleFileLibraryConfirm = (files: FileInfo[]) => {
  if (!activeFileParam.value) {
    return;
  }
  mergeSelectedFiles(files);
  activeFileParam.value.value = isMultiFileParam(activeFileParam.value)
    ? files.map((item) => item.id).join(",")
    : (files[0]?.id || "");
};

const getColumns = (): any[] => {
  return [
    {
      title: t("datasource.name.label"),
      dataIndex: "name",
      key: "name",
      align: "center",
    },
    {
      title: t("datasource.type.label"),
      dataIndex: "type",
      key: "type",
      align: "center",
    },
    {
      title: t("datasource.updated_at.label"),
      dataIndex: "updated_at",
      key: "updated_at",
      align: "center",
    },
    {
      title: t("datasource.action.label"),
      key: "action",
      align: "center",
      width: 150,
      fixed: "right",
    },
  ];
};

// 刷新数据源列表
const fetchDataSourceList = () => {
  loading.value = true;
  loadError.value = false;
  getDataSourceList()
      .then((res: any) => {
        tableData.value = res.data.list || [];
      })
      .catch(() => {
        loadError.value = true;
        tableData.value = [];
      })
      .finally(() => {
        loading.value = false;
      });
};

// 打开新增弹窗
const handleOpenAddDialog = () => {
  resetForm();
  addDataSourceDialog.value.show = true;
  addDataSourceDialog.value.title = t("datasource.add.title");
  addDataSourceDialog.value.isEdit = false;
};

// 取消操作
const handleCancel = () => {
  addDataSourceDialog.value.show = false;
  resetForm();
};

// 重置表单
const resetForm = () => {
  const form = addDataSourceDialog.value.form;
  form.id = undefined;
  form.name = "";
  form.type = "";
  form.data = [];
  activeFileParam.value = null;
  fileLibraryOpen.value = false;
  selectedFileMap.value = {};

  // 清除表单验证状态
  if (addDataSourceFormRef.value) {
    addDataSourceFormRef.value.clearValidate();
  }
};

// 数据源类型变化处理
const onDatasourceTypeChange = (value: SelectValue) => {
  const selectedType = dataSourceTypeList.value.find((item) => item.type === value);
  const form = addDataSourceDialog.value.form;

  if (selectedType) {
    const existingMap = new Map(form.data.map((item) => [item.key, item.value]));
    // 初始化动态参数
    form.data = selectedType.params.map(param => ({
      key: param.key,
      value: String(existingMap.get(param.key) || param.defaultValue || ""),
      description: param.description,
      required: param.required || false,
      defaultValue: param.defaultValue,
      placeholder: param.placeholder,
      example: param.example,
      type: param.type,
      mask: param.mask || false,
    }));
  } else {
    form.data = [];
  }
  refreshSelectedFilesFromForm();
};

// 新增/编辑提交
const handleAddDataSource = () => {
  addDataSourceFormRef.value?.validate()
      .then(() => {
        // 构造符合后端接口要求的数据格式
        const form = addDataSourceDialog.value.form;
        const payload:{
          id?: string | undefined;
          name: string;
          type: string;
          data: { key: string; value: string }[];
          edit: boolean;
        } = {
          id: form.id || "",
          name: form.name,
          type: form.type,
          data: form.data.map(item => ({
            key: item.key,
        value: item.value
          })),
          edit: addDataSourceDialog.value.isEdit
        };

        return addDataSource(payload);
      })
      .then((res) => {
        if (res.code === 0) {
          message.success(
              addDataSourceDialog.value.isEdit
                  ? t("datasource.edit.success")
                  : t("datasource.add.success")
          );
          addDataSourceDialog.value.show = false;
          fetchDataSourceList();
          getTypeByComponent(true);
        }
      })
      .catch((err) => {
        if (err instanceof ApiRequestError && err.details?.errors?.length) {
          (addDataSourceFormRef.value as any)?.setFields?.(
            err.details.errors.map((item) => ({
              name: item.field,
              errors: [item.message],
            }))
          );
        }
        console.error(err);
      });
};

const handleTestDataSource = () => {
  addDataSourceFormRef.value?.validate()
      .then(() => {
        testing.value = true;
        const form = addDataSourceDialog.value.form;
        return testDataSource({
          id: form.id,
          type: form.type,
          data: form.data.map((item) => ({
            key: item.key,
            value: item.value,
          })),
        });
      })
      .then((res) => {
        if (res?.code === 0) {
          message.success(res.message || t("datasource.test.success"));
        }
      })
      .finally(() => {
        testing.value = false;
      });
};

// 根据 id 加载数据源配置并填充表单（敏感字段后端已脱敏为 ****）
const loadDataSourceIntoForm = async (id: string, type: string) => {
  const res = await getDataSourceById(id);
  if (res.code !== 0) return;
  const remoteData: {key: string; value: string}[] = res.data.data || [];
  const remoteMap = new Map(remoteData.map(d => [d.key, d.value]));

  const selectedType = dataSourceTypeList.value.find(item => item.type === type);
  if (selectedType) {
    addDataSourceDialog.value.form.data = selectedType.params.map(param => ({
      key: param.key,
      value: remoteMap.has(param.key) ? remoteMap.get(param.key)! : (param.defaultValue || ""),
      description: param.description,
      required: param.required || false,
      defaultValue: param.defaultValue,
      placeholder: param.placeholder,
      example: param.example,
      type: param.type,
      mask: param.mask || false,
    }));
  }
  refreshSelectedFilesFromForm();
};

// 编辑数据源
const handleEdit = async (row: any) => {
  resetForm();
  addDataSourceDialog.value.show = true;
  addDataSourceDialog.value.title = t("datasource.edit.title");
  addDataSourceDialog.value.isEdit = true;
  const form = addDataSourceDialog.value.form;
  form.id = row.id;
  form.name = row.name;
  form.type = row.type;
  await loadDataSourceIntoForm(row.id, row.type);
};

const handleClone = async (row: any) => {
  resetForm();
  addDataSourceDialog.value.show = true;
  addDataSourceDialog.value.title = t("datasource.clone.title");
  addDataSourceDialog.value.isEdit = false;
  const form = addDataSourceDialog.value.form;
  form.id = undefined;
  form.name = `${row.name}-${t("datasource.clone.suffix")}`;
  form.type = row.type;
  await loadDataSourceIntoForm(row.id, row.type);
  // 克隆时清空敏感字段，要求重新填写
  form.data.forEach(param => {
    if (param.mask) param.value = "";
  });
};

// 删除数据源
const handleDelete = (row: any) => {
  Modal.confirm({
    title: t("datasource.delete.confirm.title"),
    content: t("datasource.delete.confirm.content"),
    onOk: () => {
      if (!row.id) return;
      deleteDataSource({ id: row.id })
          .then((res: any) => {
            if (res.code === 0) {
              message.success(t("datasource.delete.success"));
              fetchDataSourceList();
              getTypeByComponent(true);
            }
          })
          .catch((err: any) => {
            console.error(err);
          });
    },
  });
};

onMounted(() => {
  // 获取数据源类型列表
  getDataSourceTypeList().then((res) => {
    dataSourceTypeList.value = res.data.list;
  });
  fetchDataSourceList();
});
</script>

<style scoped lang="scss">
.data-source-container {
  padding: 20px;
}

.table-operations {
  margin-bottom: 16px;
  display: flex;
  justify-content: space-between;
}

.file-param-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: flex-start;
}

.file-param-block {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
}

.file-param-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}

.file-param-actions :deep(.ant-btn) {
  min-width: 140px;
}

.selected-file-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.param-help {
  color: #666;
  font-size: 13px;
  line-height: 1.6;
  word-break: break-word;
}

.param-help-file {
  margin-top: 0;
}

.required {
  margin-right: 6px;
  color: #ff4d4f;
}

@media (max-width: 768px) {
  .data-source-container {
    padding: 12px;
  }

  .table-operations {
    flex-direction: column;
    gap: 12px;
  }

  .data-source-container :deep(.filter-bar),
  .data-source-container .left,
  .data-source-container .right {
    width: 100%;
  }

  .data-source-container :deep(.filter-bar .ant-space-item),
  .data-source-container :deep(.table-operations .ant-space-item) {
    flex: 1 1 100%;
  }

  .data-source-container :deep(.filter-bar .ant-input),
  .data-source-container :deep(.filter-bar .ant-select),
  .data-source-container :deep(.table-operations .ant-btn) {
    width: 100% !important;
  }
}
</style>
