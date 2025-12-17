<template>
  <div class="data-source-container">
    <a-card :bordered="false">
      <div class="table-operations">
        <div class="left">
          <a-button type="primary" @click="handleOpenAddDialog">
            <template #icon>
              <PlusOutlined />
            </template>
            {{ $t("datasource.add.title") }}
          </a-button>
        </div>
        <div class="right">
          <a-button shape="circle" @click="fetchDataSourceList">
            <template #icon>
              <ReloadOutlined />
            </template>
          </a-button>
        </div>
      </div>

      <a-table
          :columns="getColumns()"
          :data-source="tableData"
          bordered
          row-key="id"
          :loading="loading"
          :scroll="{ y: 'calc(100vh - 470px)', x: 'max-content' }"
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
        width="600px"
        @ok="handleAddDataSource"
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
            :rules="[{ required: param.required, }]"
        >
          <a-input
              v-model:value="param.value"
              :placeholder="param.description"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import {
  getDataSourceTypeList,
  addDataSource,
  getDataSourceList,
  deleteDataSource,
} from "../api/datasource.ts";
import { useI18n } from "vue-i18n";

import { message, Modal } from "ant-design-vue";
import { PlusOutlined, ReloadOutlined } from "@ant-design/icons-vue";
import type { FormInstance } from "ant-design-vue";
import type { Params } from "@/src/types";
import {SelectValue} from "ant-design-vue/es/select";
import {RuleObject} from "ant-design-vue/es/form";

const { t } = useI18n();

interface DataSourceParam {
  key: string;
  value: string;
  description?: string;
  required: boolean;
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
const formRules:{ [k: string]: RuleObject | RuleObject[]; } = {
  name: [{ required: true, message: t("datasource.name.placeholder"), trigger: "blur" }],
  type: [{ required: true, message: t("datasource.type.placeholder"), trigger: "change" }]
};

const tableData = ref<any[]>([]);
const loading = ref(false);

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
  getDataSourceList()
      .then((res: any) => {
        tableData.value = res.data.list;
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
    // 初始化动态参数
    form.data = selectedType.params.map(param => ({
      key: param.key,
      value: "",
      description: param.description,
      required: param.required || false
    }));
  } else {
    form.data = [];
  }
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
          edit: string;
        } = {
          id: form.id || "",
          name: form.name,
          type: form.type,
          data: form.data.map(item => ({
            key: item.key,
            value: item.value
          })),
          edit: addDataSourceDialog.value.isEdit ? "true" : "false"
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
        }
      })
      .catch((err) => {
        console.error(err);
      });
};

// 编辑数据源
const handleEdit = (row: any) => {
  resetForm();

  addDataSourceDialog.value.show = true;
  addDataSourceDialog.value.title = t("datasource.edit.title");
  addDataSourceDialog.value.isEdit = true;

  // 填充基础信息
  const form = addDataSourceDialog.value.form;
  form.id = row.id;
  form.name = row.name;
  form.type = row.type;

  // 获取对应类型的参数定义
  const selectedType = dataSourceTypeList.value.find(item => item.type === row.type);
  if (selectedType) {
    // 根据类型定义和已有数据构建表单数据
    form.data = selectedType.params.map(param => {
      const existingData = row.data?.find((d: any) => d.key === param.key);
      return {
        key: param.key,
        value: existingData ? existingData.value : "",
        description: param.description,
        required: param.required || false
      };
    });
  }
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
</style>
