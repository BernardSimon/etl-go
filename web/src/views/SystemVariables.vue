<template>
  <div class="system-variables-container">
    <a-card :bordered="false">
      <div class="table-operations">
        <div class="left">
          <a-button type="primary" @click="handleOpenAddDialog">
            <template #icon>
              <PlusOutlined />
            </template>
            {{ t('systemVariable.add.button') }}
          </a-button>
        </div>
        <div class="right">
          <a-button shape="circle" @click="fetchVariableList">
            <template #icon>
              <ReloadOutlined />
            </template>
          </a-button>
        </div>
      </div>

      <a-table :columns="getColumns()" :data-source="tableData" bordered row-key="id" :loading="loading"
               :scroll="{ y: 'calc(100vh - 470px)', x: 'max-content' }">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'action'">
            <a-space>
              <a-button type="primary" size="small" @click="handleEdit(record)">{{ t('systemVariable.edit.title') }}</a-button>
              <a-button size="small" class="test-button" @click="handleTest(record)">{{ t('runLog.table.action.viewParams') }}</a-button>
              <a-button size="small" type="primary" danger @click="handleDelete(record)">{{ t('systemVariable.delete.title') }}</a-button>
            </a-space>
          </template>
          <template v-if ="column.key === 'datasource'">
            <div>
              {{
                getDatasourceNameById(record.datasource_id, record.type)
              }}
            </div>
          </template>
        </template>
      </a-table>
    </a-card>

    <!-- 新增/编辑系统变量弹窗 -->
    <a-modal v-model:open="variableDialog.show" :title="variableDialog.title" width="600px" @ok="handleSaveVariable"
             @cancel="variableDialog.show = false">
      <a-form ref="variableFormRef" :model="variableDialog.data" :rules="variableDialog.rules" :label-col="{ span: 6 }"
              :wrapper-col="{ span: 17 }">
        <a-form-item :label="t('systemVariable.form.name.label')" name="name">
          <a-input v-model:value="variableDialog.data.name" :placeholder="t('systemVariable.form.name.placeholder')" />
        </a-form-item>

        <a-form-item :label="t('systemVariable.form.type.label')" name="type">
          <a-select
              v-model:value="variableDialog.data.type"
              :placeholder="t('systemVariable.form.type.placeholder')"
              @change="onVariableTypeChange"
          >
            <a-select-option v-for="item in variableTypeList" :key="item.type" :value="item.type">
              {{ item.type }}
            </a-select-option>
          </a-select>
        </a-form-item>

        <a-form-item
            v-if="showDataSourceSelector"
            :label="t('systemVariable.form.datasource.label')"
            name="datasource_id"
            required
        >
          <a-select
              v-model:value="variableDialog.data.datasource_id"
              :placeholder="t('systemVariable.form.datasource.placeholder')"
          >
            <a-select-option v-for="item in currentDatasourceList" :key="item.ID" :value="item.ID">
              {{ item.Name }}
            </a-select-option>
          </a-select>
        </a-form-item>

        <!-- 动态参数表单项 -->
        <a-form-item
            v-for="(param, index) in variableDialog.data.params"
            :key="index"
            :label="param.key"
            :name="['params', index, 'value']"
            :rules="[{ required: param.required, message: `${param.key} is required` }]"
        >
          <a-input
              v-model:value="param.value"
              :placeholder="param.description"
          />
        </a-form-item>

        <a-form-item :label="t('systemVariable.form.description.label')" name="description">
          <a-textarea v-model:value="variableDialog.data.description" :placeholder="t('systemVariable.form.description.placeholder')" :rows="2" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import {
  getVariableList,
  saveVariable,
  deleteVariable,
  testVariable,
  getVariableTypeList
} from "../api/systemVariables";
import { message, Modal } from "ant-design-vue";
import { PlusOutlined, ReloadOutlined } from "@ant-design/icons-vue";
import type { FormInstance } from "ant-design-vue";
import { useI18n } from "vue-i18n";
import type { VariableTypeListItem } from "../api/systemVariables";
import {SelectValue} from "ant-design-vue/es/select";

const { t } = useI18n();

// 变量类型列表
const variableTypeList = ref<VariableTypeListItem[]>([]);

// 当前选中类型的可用数据源列表
const currentDatasourceList = ref<{ID: string; Name: string}[]>([]);

// 是否显示数据源选择器
const showDataSourceSelector = computed(() => {
  const selectedType = variableTypeList.value.find(item => item.type === variableDialog.value.data.type);
  return selectedType && selectedType.datasource_list !== null;
});

// 系统变量表格数据
const tableData = ref<any[]>([]);
const loading = ref(false);

// 表头配置
const getColumns = (): any[] => [
  {
    title: t('systemVariable.table.column.name'),
    dataIndex: "name",
    key: "name",
    align: "center",
  },
  {
    title: t('systemVariable.table.type.name'),
    dataIndex: "type",
    key: "type",
    align: "center",
  },
  {
    title: t('systemVariable.table.column.description'),
    dataIndex: "description",
    key: "description",
    align: "center",
    ellipsis: true,
  },
  {
    title: t('systemVariable.table.column.updatedAt'),
    dataIndex: "updated_at",
    key: "updated_at",
    align: "center",
  },
  {
    title: t('systemVariable.table.column.actions'),
    key: "action",
    align: "center",
    fixed: "right",
    width: 200,
  },
];

// 新增/编辑弹窗配置
const variableDialog = ref<any>({
  show: false,
  title: t('systemVariable.add.title'),
  isEdit: false,
  data: {
    id: "",
    name: "",
    type: "",
    datasource_id: "",
    params: [],
    description: ""
  },
  rules: {
    name: [{ required: true, message: t('systemVariable.form.name.required'), trigger: "blur" }],
    type: [{ required: true, message: t('systemVariable.form.type.required'), trigger: "blur" }],
    datasource_id: [{ required: false, message: t('systemVariable.form.datasource.required'), trigger: "blur" }],
    description: [{ required: true, message: t('systemVariable.form.description.required'), trigger: "blur" }],
  },
});

const variableFormRef = ref<FormInstance>();

// 获取系统变量列表
const fetchVariableList = () => {
  loading.value = true;
  getVariableList().then((res: any) => {
    tableData.value = res.data.list || [];
    loading.value = false;
  });
};

// 获取变量类型列表
const fetchVariableTypeList = () => {
  getVariableTypeList().then((res: any) => {
    variableTypeList.value = res.data.list || [];
  });
};

// 添加一个方法来根据ID获取数据源名称
const getDatasourceNameById = (id: string, type: string) => {
  if (!id) return '-';

  const variableType = variableTypeList.value.find(item => item.type === type);
  if (variableType && variableType.datasource_list) {
    const datasource = variableType.datasource_list.find(item => item.ID === id);
    return datasource ? datasource.Name : '-';
  }
  return '-';
};

// 变量类型变化处理
const onVariableTypeChange = (value: SelectValue) => {
  const selectedType = variableTypeList.value.find(item => item.type === value);

  if (selectedType) {
    // 初始化动态参数
    variableDialog.value.data.params = selectedType.params.map(param => ({
      key: param.key,
      value: "",
      description: param.description,
      required: param.required || false
    }));

    // 设置数据源列表
    if (selectedType.datasource_list) {
      currentDatasourceList.value = selectedType.datasource_list;
      // 清空之前选择的数据源
      variableDialog.value.data.datasource_id = "";
    } else {
      currentDatasourceList.value = [];
      variableDialog.value.data.datasource_id = "";
    }
  } else {
    variableDialog.value.data.params = [];
    currentDatasourceList.value = [];
    variableDialog.value.data.datasource_id = "";
  }
};

// 打开新增弹窗
const handleOpenAddDialog = () => {
  variableDialog.value.show = true;
  variableDialog.value.title = t('systemVariable.add.title');
  variableDialog.value.isEdit = false;
  variableDialog.value.data = {
    id: "",
    name: "",
    type: "",
    datasource_id: "",
    params: [],
    description: ""
  };
};

// 编辑系统变量
const handleEdit = (row: any) => {
  variableDialog.value.show = true;
  variableDialog.value.title = t('systemVariable.edit.title');
  variableDialog.value.isEdit = true;

  // 构建参数值映射
  const paramValues: Record<string, string> = {};
  if (Array.isArray(row.value)) {
    row.value.forEach((kv: any) => {
      paramValues[kv.key] = kv.value;
    });
  }

  // 查找变量类型信息
  const variableType = variableTypeList.value.find(type => type.type === row.type);

  // 构建参数列表
  let params: any[] = [];
  if (variableType) {
    params = variableType.params.map(param => ({
      key: param.key,
      value: paramValues[param.key] || "",
      description: param.description,
      required: param.required || false
    }));
  }
  // 触发类型变更以设置数据源列表
  onVariableTypeChange(row.type);

  variableDialog.value.data = {
    id: row.id,
    name: row.name,
    type: row.type,
    datasource_id: row.datasource_id || "",
    params: params,
    description: row.description
  };
  console.log(variableDialog.value.data)

};

// 新增/编辑提交
const handleSaveVariable = () => {
  variableFormRef.value
      ?.validate()
      .then(() => {
        // 构造参数值数组
        const value = variableDialog.value.data.params.map((param: any) => ({
          key: param.key,
          value: param.value
        }));

        // 如果需要数据源，则添加到value中

        const payload = {
          id: variableDialog.value.isEdit ? variableDialog.value.data.id : undefined,
          name: variableDialog.value.data.name,
          type: variableDialog.value.data.type,
          datasource_id: variableDialog.value.data.datasource_id || null,
          description: variableDialog.value.data.description,
          value: value,
          edit: variableDialog.value.isEdit ? "true" : "false"
        };

        saveVariable(payload).then((res: any) => {
          if (res.code === 0) {
            message.success(
                variableDialog.value.isEdit ? t('systemVariable.edit.success') : t('systemVariable.add.success')
            );
            variableDialog.value.show = false;
            fetchVariableList();
          }
        });
      })
      .catch((err: any) => {
        console.log("Validation failed", err);
      });
};

// 删除系统变量
const handleDelete = (row: any) => {
  Modal.confirm({
    title: t('systemVariable.delete.confirm.title'),
    content: t('systemVariable.delete.confirm.content'),
    onOk: () => {
      if (!row.id) return;
      deleteVariable({ id: row.id }).then((res: any) => {
        if (res.code === 0) {
          message.success(t('systemVariable.delete.success'));
          fetchVariableList();
        }
      });
    },
  });
};

// 测试系统变量
const handleTest = (row: any) => {
  if (!row.id) return;
  testVariable({ id: row.id })
      .then((res: any) => {
        if (res.code === 0) {
          message.success(res.data?.result || t('systemVariable.test.success'));
        }
      })
      .catch((err: any) => {
        console.error("测试变量失败：", err);
      });
};

onMounted(() => {
  fetchVariableList();
  fetchVariableTypeList();
});
</script>

<style scoped lang="scss">
.system-variables-container {
  padding: 20px;
}

.table-operations {
  margin-bottom: 16px;
  display: flex;
  justify-content: space-between;
}
</style>
