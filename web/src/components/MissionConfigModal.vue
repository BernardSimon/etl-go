<template>
  <a-modal
      :open="open"
      :title="title"
      width="900px"
      :style="{ top: '20px' }"
      @ok="handleOk"
      @cancel="handleCancel"
      :footer="mode === 'read' ? null : undefined"
      :destroyOnClose="true"
  >
    <div class="modal-content" :class="{ 'vertical-layout': isNarrowScreen }">
      <!-- 主表单内容 -->
      <div class="main-content">
        <div v-if="mode === 'read' && !data" class="text-center py-10 text-gray-500">
          {{ t('missionConfig.noData') }}
        </div>

        <a-form
            v-else
            ref="formRef"
            :model="formData"
            :rules="formRules"
            layout="vertical"
            class="workflow-form"
            :disabled="mode === 'read'"
        >
          <!-- 任务名称 -->
          <a-form-item
              :label="t('missionConfig.missionName.label')"
              name="mission_name"
          >
            <a-input
                v-model:value="formData.mission_name"
                :placeholder="t('missionConfig.missionName.placeholder')"
                :disabled="mode === 'read'"
            />
          </a-form-item>

          <!-- 调度规则（仅定时任务显示） -->
          <a-form-item
              v-if="showCronField"
              :label="t('missionConfig.cron.label')"
              name="cron"
          >
            <a-input
                v-model:value="formData.cron"
                :placeholder="t('missionConfig.cron.placeholder')"
            />
          </a-form-item>

          <!-- Before Execute Section -->
          <a-card size="small" :title="t('missionConfig.beforeTask.title')" class="section-card">
            <a-row :gutter="16">
              <a-col :span="12">
                <a-form-item
                    :label="t('missionConfig.type.label')"
                    :name="['before_execute', 'type']"
                    :rules="[]"
                >
                  <a-select
                      v-model:value="formData.before_execute.type"
                      :placeholder="t('missionConfig.type.placeholder')"
                      @change="() => handleTypeChange(formData.before_execute, 'execute')"
                      allowClear
                      :disabled="mode === 'read'"
                  >
                    <a-select-option
                        v-for="item in typesData.execute"
                        :key="item.type"
                        :value="item.type"
                    >{{ item.type }}</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>

              <a-col v-if="showDataSource(formData.before_execute)" :span="12">
                <a-form-item
                    :label="t('missionConfig.datasource.label')"
                    :name="['before_execute', 'data_source']"
                    :rules="isDataSourceRequired(formData.before_execute) ? [{ required: true, message: t('missionConfig.datasource.required') }] : []"
                >
                  <a-select
                      v-model:value="formData.before_execute.data_source as string"
                      :placeholder="t('missionConfig.datasource.placeholder')"
                      :disabled="isDataSourceDisabled(formData.before_execute)"
                      allowClear
                  >
                    <a-select-option
                        v-for="ds in getDataSourceOptions(formData.before_execute)"
                        :key="ds.id"
                        :value="ds.id"
                    >{{ ds.name }}</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
            </a-row>

            <div class="params-section" v-if="formData.before_execute.params && formData.before_execute.params.length > 0">
              <div class="section-label" v-if="formData.before_execute.params.length > 0">
                {{ t('missionConfig.params.label') }}
              </div>
              <div
                  v-for="(param, index) in formData.before_execute.params"
                  :key="param.key + index"
                  class="param-item"
              >
                <a-row :gutter="16">
                  <a-col :span="6">
                    <span class="param-label">
                      {{ formData.before_execute.params[index].key }}:
                      <span v-if="param.required" class="required-asterisk">*</span>
                    </span>
                  </a-col>
                  <a-col :span="18">
                    <a-form-item
                        :name="['before_execute', 'params', index, 'value']"
                        :rules="getValidationRules(param)"
                        style="margin-bottom: 12px;"
                    >
                      <a-textarea
                          v-model:value="formData.before_execute.params[index].value"
                          :auto-size="{ minRows: 1, maxRows: 6 }"
                          :placeholder="getPlaceholder(param)"
                          :disabled="mode === 'read'"
                          :rows="2"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                <div v-if="param.description" class="param-description">
                  {{ param.description }}
                </div>
              </div>
            </div>
          </a-card>

          <!-- Source Section -->
          <a-card size="small" :title="t('missionConfig.source.title')" class="section-card">
            <a-row :gutter="16">
              <a-col :span="12">
                <a-form-item
                    :label="t('missionConfig.type.label')"
                    :name="['source', 'type']"
                    :rules="[{ required: true, message: t('missionConfig.type.required') }]"
                >
                  <a-select
                      v-model:value="formData.source.type"
                      :placeholder="t('missionConfig.type.placeholder')"
                      @change="() => handleTypeChange(formData.source, 'source')"
                      allowClear
                      :disabled="mode === 'read'"
                  >
                    <a-select-option
                        v-for="item in typesData.source"
                        :key="item.type"
                        :value="item.type"
                    >{{ item.type }}</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>

              <a-col v-if="showDataSource(formData.source)" :span="12">
                <a-form-item
                    :label="t('missionConfig.datasource.label')"
                    :name="['source', 'data_source']"
                    :rules="isDataSourceRequired(formData.source) ? [{ required: true, message: t('missionConfig.datasource.required') }] : []"
                >
                  <a-select
                      v-model:value="formData.source.data_source as string"
                      :placeholder="t('missionConfig.datasource.placeholder')"
                      :disabled="isDataSourceDisabled(formData.source)"
                      allowClear
                  >
                    <a-select-option
                        v-for="ds in getDataSourceOptions(formData.source)"
                        :key="ds.id"
                        :value="ds.id"
                    >{{ ds.name }}</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
            </a-row>

            <div class="params-section" v-if="formData.source.params && formData.source.params.length > 0">
              <div class="section-label" v-if="formData.source.params.length > 0">
                {{ t('missionConfig.params.label') }}
              </div>
              <div
                  v-for="(param, index) in formData.source.params"
                  :key="param.key + index"
                  class="param-item"
              >
                <a-row :gutter="16">
                  <a-col :span="6">
                    <span class="param-label">
                      {{ formData.source.params[index].key }}:
                      <span v-if="param.required" class="required-asterisk">*</span>
                    </span>
                  </a-col>
                  <a-col :span="18">
                    <a-form-item
                        :name="['source', 'params', index, 'value']"
                        :rules="getValidationRules(param)"
                        style="margin-bottom: 12px;"
                    >
                      <a-textarea
                          v-model:value="formData.source.params[index].value"
                          :auto-size="{ minRows: 1, maxRows: 6 }"
                          :placeholder="getPlaceholder(param)"
                          :disabled="mode === 'read'"
                          :rows="2"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                <div v-if="param.description" class="param-description">
                  {{ param.description }}
                </div>
              </div>
            </div>
          </a-card>

          <!-- Processor Section -->
          <a-card size="small" :title="t('missionConfig.processor.title')" class="section-card">
            <div
                v-for="(processor, pIndex) in formData.processors"
                :key="pIndex"
                class="processor-item"
            >
              <div class="processor-header">
                <span>{{ t('missionConfig.processor.item') }} {{ pIndex + 1 }}</span>
                <a-button
                    v-if="mode !== 'read'"
                    type="link"
                    danger
                    size="small"
                    @click="removeProcessor(pIndex)"
                >
                  {{ t('missionConfig.processor.remove') }}
                </a-button>
              </div>

              <a-row :gutter="16">
                <a-col :span="24">
                  <a-form-item
                      :label="t('missionConfig.processor.type.label')"
                      :name="['processors', pIndex, 'type']"
                      :rules="[{ required: true, message: t('missionConfig.processor.type.required') }]">
                    <a-select
                        v-model:value="processor.type"
                        :placeholder="t('missionConfig.processor.type.placeholder')"
                        @change="() => handleProcessorTypeChange(processor)"
                        allowClear
                        :disabled="mode === 'read'"
                    >
                      <a-select-option
                          v-for="item in typesData.processor"
                          :key="item.type"
                          :value="item.type"
                      >{{ item.type }}</a-select-option>
                    </a-select>
                  </a-form-item>
                </a-col>
              </a-row>

              <div class="params-section" v-if="processor.params && processor.params.length > 0">
                <div class="section-label" v-if="processor.params.length > 0">
                  {{ t('missionConfig.params.label') }}
                </div>
                <div
                    v-for="(param, index) in processor.params"
                    :key="param.key + index"
                    class="param-item"
                >
                  <a-row :gutter="16">
                    <a-col :span="6">
                      <span class="param-label">
                        {{ processor.params[index].key }}:
                        <span v-if="param.required" class="required-asterisk">*</span>
                      </span>
                    </a-col>
                    <a-col :span="18">
                      <a-form-item
                          :name="['processors', pIndex, 'params', index, 'value']"
                          :rules="getValidationRules(param)"
                          style="margin-bottom: 12px;"
                      >
                        <a-textarea
                            v-model:value="processor.params[index].value"
                            :auto-size="{ minRows: 1, maxRows: 6 }"
                            :placeholder="getPlaceholder(param)"
                            :disabled="mode === 'read'"
                            :rows="2"
                        />
                      </a-form-item>
                    </a-col>
                  </a-row>
                  <div v-if="param.description" class="param-description">
                    {{ param.description }}
                  </div>
                </div>
              </div>
            </div>

            <a-button
                v-if="mode !== 'read'"
                type="dashed"
                block
                @click="addProcessor"
            >
              {{ t('missionConfig.processor.add') }}
            </a-button>
          </a-card>

          <!-- Sink Section -->
          <a-card size="small" :title="t('missionConfig.sink.title')" class="section-card">
            <a-row :gutter="16">
              <a-col :span="12">
                <a-form-item
                    :label="t('missionConfig.type.label')"
                    :name="['sink', 'type']"
                    :rules="[{ required: true, message: t('missionConfig.type.required') }]"
                >
                  <a-select
                      v-model:value="formData.sink.type"
                      :placeholder="t('missionConfig.type.placeholder')"
                      @change="() => handleTypeChange(formData.sink, 'sink')"
                      allowClear
                      :disabled="mode === 'read'"
                  >
                    <a-select-option
                        v-for="item in typesData.sink"
                        :key="item.type"
                        :value="item.type"
                    >{{ item.type }}</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>

              <a-col v-if="showDataSource(formData.sink)" :span="12">
                <a-form-item
                    :label="t('missionConfig.datasource.label')"
                    :name="['sink', 'data_source']"
                    :rules="isDataSourceRequired(formData.sink) ? [{ required: true, message: t('missionConfig.datasource.required') }] : []"
                >
                  <a-select
                      v-model:value="formData.sink.data_source as string"
                      :placeholder="t('missionConfig.datasource.placeholder')"
                      :disabled="isDataSourceDisabled(formData.sink)"
                      allowClear
                  >
                    <a-select-option
                        v-for="ds in getDataSourceOptions(formData.sink)"
                        :key="ds.id"
                        :value="ds.id"
                    >{{ ds.name }}</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
            </a-row>

            <div class="params-section" v-if="formData.sink.params && formData.sink.params.length > 0">
              <div class="section-label" v-if="formData.sink.params.length > 0">
                {{ t('missionConfig.params.label') }}
              </div>
              <div
                  v-for="(param, index) in formData.sink.params"
                  :key="param.key + index"
                  class="param-item"
              >
                <a-row :gutter="16">
                  <a-col :span="6">
                    <span class="param-label">
                      {{ formData.sink.params[index].key }}:
                      <span v-if="param.required" class="required-asterisk">*</span>
                    </span>
                  </a-col>
                  <a-col :span="18">
                    <a-form-item
                        :name="['sink', 'params', index, 'value']"
                        :rules="getValidationRules(param)"
                        style="margin-bottom: 12px;"
                    >
                      <a-textarea
                          v-model:value="formData.sink.params[index].value"
                          :auto-size="{ minRows: 1, maxRows: 6 }"
                          :placeholder="getPlaceholder(param)"
                          :disabled="mode === 'read'"
                          :rows="2"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                <div v-if="param.description" class="param-description">
                  {{ param.description }}
                </div>
              </div>
            </div>
          </a-card>

          <!-- After Execute Section -->
          <a-card size="small" :title="t('missionConfig.afterTask.title')" class="section-card">
            <a-row :gutter="16">
              <a-col :span="12">
                <a-form-item
                    :label="t('missionConfig.type.label')"
                    :name="['after_execute', 'type']"
                    :rules="[]"
                >
                  <a-select
                      v-model:value="formData.after_execute.type"
                      :placeholder="t('missionConfig.type.placeholder')"
                      @change="() => handleTypeChange(formData.after_execute, 'execute')"
                      allowClear
                      :disabled="mode === 'read'"
                  >
                    <a-select-option
                        v-for="item in typesData.execute"
                        :key="item.type"
                        :value="item.type"
                    >{{ item.type }}</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>

              <a-col v-if="showDataSource(formData.after_execute)" :span="12">
                <a-form-item
                    :label="t('missionConfig.datasource.label')"
                    :name="['after_execute', 'data_source']"
                    :rules="isDataSourceRequired(formData.after_execute) ? [{ required: true, message: t('missionConfig.datasource.required') }] : []"
                >
                  <a-select
                      v-model:value="formData.after_execute.data_source as string"
                      :placeholder="t('missionConfig.datasource.placeholder')"
                      :disabled="isDataSourceDisabled(formData.after_execute)"
                      allowClear
                  >
                    <a-select-option
                        v-for="ds in getDataSourceOptions(formData.after_execute)"
                        :key="ds.id"
                        :value="ds.id"
                    >{{ ds.name }}</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
            </a-row>

            <div class="params-section" v-if="formData.after_execute.params && formData.after_execute.params.length > 0">
              <div class="section-label" v-if="formData.after_execute.params.length > 0">
                {{ t('missionConfig.params.label') }}
              </div>
              <div
                  v-for="(param, index) in formData.after_execute.params"
                  :key="param.key + index"
                  class="param-item"
              >
                <a-row :gutter="16">
                  <a-col :span="6">
                    <span class="param-label">
                      {{ formData.after_execute.params[index].key }}:
                      <span v-if="param.required" class="required-asterisk">*</span>
                    </span>
                  </a-col>
                  <a-col :span="18">
                    <a-form-item
                        :name="['after_execute', 'params', index, 'value']"
                        :rules="getValidationRules(param)"
                        style="margin-bottom: 12px;"
                    >
                      <a-textarea
                          v-model:value="formData.after_execute.params[index].value"
                          :auto-size="{ minRows: 1, maxRows: 6 }"
                          :placeholder="getPlaceholder(param)"
                          :disabled="mode === 'read'"
                          :rows="2"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                <div v-if="param.description" class="param-description">
                  {{ param.description }}
                </div>
              </div>
            </div>
          </a-card>
        </a-form>
      </div>

      <!-- 文件上传区域（宽屏时在右侧） -->
      <div
          class="file-upload-section"
          v-if="mode !== 'read' && !isNarrowScreen"
      >
        <a-card size="small" :title="t('missionConfig.fileUpload.title')" class="section-card">
          <div class="file-upload-widget">
            <div class="upload-area">
              <a-upload-dragger
                  v-model:file-list="fileList"
                  :before-upload="beforeUpload"
                  :max-count="5"
                  multiple
                  :show-upload-list="true"
                  :disabled="uploadLoading"
              >
                <p class="ant-upload-drag-icon">
                  <inbox-outlined />
                </p>
                <p class="ant-upload-text">
                  {{ t('missionConfig.fileUpload.dragText') }}
                </p>
                <p class="ant-upload-hint">
                  {{ t('missionConfig.fileUpload.hint') }}
                </p>
              </a-upload-dragger>

              <a-button
                  type="primary"
                  @click="handleUpload"
                  :loading="uploadLoading"
                  :disabled="fileList.length === 0"
                  style="margin-top: 16px; width: 100%"
              >
                {{ t('missionConfig.fileUpload.uploadButton') }}
              </a-button>
            </div>

            <div class="uploaded-files" v-if="uploadedFiles.length > 0">
              <h4>{{ t('missionConfig.fileUpload.uploadedTitle') }}</h4>
              <a-list
                  size="small"
                  :data-source="uploadedFiles"
                  :locale="{ emptyText: t('missionConfig.fileUpload.noFiles') }"
              >
                <template #renderItem="{ item }">
                  <a-list-item class="file-list-item">
                    <div class="file-item">
                      <div class="file-info">
                        <div class="file-name">
                          <file-text-outlined style="margin-right: 8px;" />
                          {{ item.name }}
                        </div>
                        <div class="file-meta">
                          <span class="file-id">ID: {{ item.id }}</span>
                          <span class="file-size" v-if="item.size">
                            ({{ formatFileSize(item.size) }})
                          </span>
                        </div>
                      </div>
                      <div class="file-actions">
                        <a-button
                            type="link"
                            size="small"
                            @click="copyFileId(item.id)"
                            :title="t('missionConfig.fileUpload.copyTooltip')"
                        >
                          <copy-outlined />
                          {{ t('missionConfig.fileUpload.copyButton') }}
                        </a-button>
                        <a-button
                            type="link"
                            size="small"
                            danger
                            @click="removeFile(item.id)"
                            :title="t('missionConfig.fileUpload.removeTooltip')"
                        >
                          <delete-outlined />
                        </a-button>
                      </div>
                    </div>
                  </a-list-item>
                </template>
              </a-list>
            </div>
          </div>
        </a-card>
      </div>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, reactive, watch, computed, onMounted, onUnmounted } from "vue";
import { message } from "ant-design-vue";
import type { FormInstance } from "ant-design-vue";
import { addTask, updateTask, getTypeByComponent } from "../api/mission";
import type { ConfigItem, TaskType } from "../types/mission";
import { useI18n } from "vue-i18n";
import { uploadFile } from "../api/file.ts";
import {
  InboxOutlined,
  FileTextOutlined,
  CopyOutlined,
  DeleteOutlined
} from '@ant-design/icons-vue';
import {RuleObject} from "ant-design-vue/es/form";

const { t } = useI18n();

// Props 定义
interface Props {
  open: boolean;
  title?: string;
  mode?: "add" | "edit" | "read";
  id?: string;
  data?: any;
  record?: any;
  taskType?: TaskType; // 'scheduled' | 'manual'
}

const props = withDefaults(defineProps<Props>(), {
  mode: "add",
  taskType: "scheduled"
});

const emit = defineEmits<{
  (e: "update:open", value: boolean): void;
  (e: "success"): void;
}>();

// 响应式数据
const formRef = ref<FormInstance>();
const screenWidth = ref(window.innerWidth);

// 屏幕宽度监听
const isNarrowScreen = computed(() => screenWidth.value < 768);

const handleResize = () => {
  screenWidth.value = window.innerWidth;
};

onMounted(() => {
  window.addEventListener('resize', handleResize);
});

onUnmounted(() => {
  window.removeEventListener('resize', handleResize);
});

// 是否显示 cron 字段
const showCronField = computed(() => {
  return props.taskType === 'scheduled';
});

// 类型数据
const typesData = ref<Record<string, any[]>>({
  execute: [],
  source: [],
  sink: [],
  processor: [],
});

// 创建空配置
const createEmptyConfig = (): ConfigItem => ({
  type: undefined,
  data_source: undefined,
  params: [],
});

// 表单数据
const formData = reactive({
  id: "",
  mission_name: "",
  cron: props.taskType === 'manual' ? 'manual' : "",
  before_execute: createEmptyConfig(),
  source: createEmptyConfig(),
  processors: [] as ConfigItem[],
  sink: createEmptyConfig(),
  after_execute: createEmptyConfig(),
});

// 表单验证规则
const formRules = computed((): Record<string, RuleObject[]> => ({
  "mission_name": [
    { required: true, message: t('missionConfig.form.missionName.required'), trigger: "blur", type: 'string' as const },
  ],
  "cron": showCronField.value
      ? [{ required: true, message: t('missionConfig.form.cron.required'), trigger: "blur", type: 'string' as const }]
      : []
}));

// 获取类型参数
const getParamsForType = (category: string, type: string | undefined) => {
  if (!type) return [];
  const list = typesData.value[category] || [];
  const item = list.find((i: any) => i.type === type);
  return item ? item.params || [] : [];
};

// 同步参数
const syncParams = (configItem: ConfigItem, category: string) => {
  if (props.mode === "read") return;

  if (!configItem.type) {
    configItem.params = [];
    return;
  }

  const requiredParams = getParamsForType(category, configItem.type);
  const existingMap = new Map();

  if (configItem.params) {
    configItem.params.forEach((p) => existingMap.set(p.key, p.value));
  }

  configItem.params = requiredParams.map((param: any) => ({
    key: param.key,
    value: existingMap.get(param.key) || param.defaultValue || "",
    required: param.required,
    description: param.description
  }));
};

// 类型改变处理
const handleTypeChange = (configItem: ConfigItem, category: string) => {
  if (props.mode === "read") return;

  if ("data_source" in configItem) {
    (configItem as ConfigItem).data_source = undefined;
  }
  syncParams(configItem, category);
};

const handleProcessorTypeChange = (processor: ConfigItem) => {
  handleTypeChange(processor, "processor");

  // 重置data_source
  if (processor.data_source !== undefined) {
    processor.data_source = undefined;
  }
};

// 处理器管理
const addProcessor = () => {
  formData.processors.push(createEmptyConfig());
};

const removeProcessor = (index: number) => {
  formData.processors.splice(index, 1);
};

// 重置配置项
const resetConfigItem = (config: ConfigItem, recordConfig?: any, category?: string) => {
  if (recordConfig) {
    config.type = recordConfig.type;
    config.data_source = recordConfig.data_source || undefined;
    config.params = recordConfig.params || [];
  } else {
    config.type = undefined;
    config.data_source = undefined;
    config.params = [];
  }

  if (category && props.mode !== "read") {
    syncParams(config, category);
  }
};

// 初始化表单
const initForm = async () => {
  try {
    // 加载类型数据
    const res = await getTypeByComponent();
    if (res.code === 0) {
      const transformData = (items: any[]) => {
        return items.map(item => ({
          type: item.type,
          data_source: item.data_source,
          params: item.params || []
        }));
      };

      typesData.value = {
        execute: transformData(res.data.executor || []),
        source: transformData(res.data.source || []),
        sink: transformData(res.data.sink || []),
        processor: transformData(res.data.processor || [])
      };
    }

    // 根据模式初始化数据
    if (props.mode === "add") {
      // 新增模式
      Object.assign(formData, {
        id: "",
        mission_name: "",
        cron: props.taskType === 'manual' ? 'manual' : "",
        before_execute: createEmptyConfig(),
        source: createEmptyConfig(),
        processors: [],
        sink: createEmptyConfig(),
        after_execute: createEmptyConfig(),
      });
    } else if (props.data) {
      // 编辑/只读模式
      const data = props.data;
      const record = props.record;

      formData.id = props.id || "";
      formData.mission_name = record.mission_name;

      // 处理 cron 字段：如果是 manual 则设置为手动任务
      const isManualTask = record.cron === 'manual';
      formData.cron = isManualTask ? 'manual' : record.cron;

      resetConfigItem(formData.before_execute, data.before_execute, "execute");
      resetConfigItem(formData.source, data.source, "source");

      // 处理器数组处理
      if (Array.isArray(data.processors)) {
        formData.processors = data.processors.map((item: any) => {
          const cfg: ConfigItem = {
            type: item.type,
            data_source: item.data_source || undefined,
            params: item.params || [],
          };
          if (props.mode !== 'read') syncParams(cfg, "processor");
          return cfg;
        });
      } else {
        formData.processors = [];
      }

      resetConfigItem(formData.sink, data.sink, "sink");
      resetConfigItem(formData.after_execute, data.after_execute, "execute");
    }
  } catch (error) {
    console.error("初始化表单失败:", error);
    message.error("表单初始化失败");
  }
};

// 监听打开状态
watch(() => props.open, (val) => {
  if (val) {
    initForm();
    // 重置文件列表
    fileList.value = [];
    uploadedFiles.value = [];
  }
});

// 任务类型变化监听
watch(() => props.taskType, (newType) => {
  if (newType === 'manual') {
    formData.cron = 'manual';
  } else if (formData.cron === 'manual') {
    formData.cron = '';
  }
});

// 确认提交
const handleOk = async () => {
  if (props.mode === "read") {
    emit("update:open", false);
    return;
  }

  try {
    await formRef.value?.validate();

    const payload = {
      id: props.id,
      mission_name: formData.mission_name,
      cron: formData.cron,
      params: {
        before_execute: formData.before_execute.type ? formData.before_execute : null,
        source: formData.source.type ? formData.source : null,
        processors: formData.processors.filter(p => p.type),
        sink: formData.sink.type ? formData.sink : null,
        after_execute: formData.after_execute.type ? formData.after_execute : null,
      },
    };

    const apiCall = props.mode === "edit" ? updateTask : addTask;
    const res = await apiCall(payload);

    if (res.code === 0) {
      message.success(props.mode === "edit"
          ? t('missionConfig.save.success.edit')
          : t('missionConfig.save.success.add')
      );
      emit("update:open", false);
      emit("success");
    }
  } catch (error) {
    console.error("保存任务失败：", error);
  }
};

// 取消操作
const handleCancel = () => {
  emit("update:open", false);
};

// FormSection 相关计算属性和方法
const showDataSource = (config: ConfigItem) => {
  if (!config.type) return false;
  const category = getCategoryByConfig(config);
  const list = typesData.value[category] || [];
  const item = list.find(i => i.type === config.type);
  return item && item.data_source !== null;
};

const isDataSourceRequired = (config: ConfigItem) => {
  if (!config.type) return false;
  const category = getCategoryByConfig(config);
  const list = typesData.value[category] || [];
  const item = list.find(i => i.type === config.type);
  return item && item.data_source !== null && item.data_source.length > 0;
};

const isDataSourceDisabled = (config: ConfigItem) => {
  if (!config.type) return true;
  const category = getCategoryByConfig(config);
  const list = typesData.value[category] || [];
  const item = list.find(i => i.type === config.type);
  return !item || !item.data_source || item.data_source.length === 0 || props.mode === 'read';
};

const getDataSourceOptions = (config: ConfigItem) => {
  if (!config.type) return [];
  const category = getCategoryByConfig(config);
  const list = typesData.value[category] || [];
  const item = list.find(i => i.type === config.type);
  return item?.data_source || [];
};

const getCategoryByConfig = (config: ConfigItem) => {
  // 这里需要根据具体的业务逻辑判断config属于哪个类别
  // 简化实现，实际应用中可能需要更复杂的判断逻辑
  if (config === formData.source) return 'source';
  if (config === formData.sink) return 'sink';
  return 'execute';
};

// ParamFields 相关方法
const getValidationRules = (param: any) : RuleObject[]=> {
  if (!param.required) return [];

  return [
    {
      required: true,
      message: t('missionConfig.param.required', { param: param.key }),
      trigger: 'blur'
    },
  ];
};

const getPlaceholder = (param: any) => {
  if (param.defaultValue) {
    return `${param.description || ''} (默认值: ${param.defaultValue})`;
  }
  return param.description || '';
};

// FileUploadWidget 相关功能
const fileList = ref<any[]>([]);
const uploadedFiles = ref<any[]>([]);
const uploadLoading = ref(false);

const beforeUpload = (file: any) => {
  // 文件类型和大小验证
  const isLt10M = file.size / 1024 / 1024 < 10;
  if (!isLt10M) {
    message.error(t('missionConfig.fileUpload.sizeError'));
    return false;
  }
  return false; // 阻止自动上传
};

const handleUpload = async () => {
  if (fileList.value.length === 0) {
    message.warning(t('missionConfig.fileUpload.noFilesSelected'));
    return;
  }

  uploadLoading.value = true;

  try {
    const uploadPromises = fileList.value.map(async (fileItem) => {
      const formData = new FormData();
      formData.append('file', fileItem.originFileObj || fileItem);
      const res = await uploadFile(formData);

      if (res?.code === 0) {
        return {
          id: res.data.id,
          name: fileItem.name,
          size: res.data.size,
          path: res.data.path,
          ex_name: res.data.ex_name
        };
      }
      throw new Error(t('missionConfig.fileUpload.uploadError', { file: fileItem.name }));
    });

    const results = await Promise.all(uploadPromises);
    uploadedFiles.value = [...uploadedFiles.value, ...results];
    fileList.value = [];
    message.success(t('missionConfig.fileUpload.uploadSuccess'));
  } catch (error: any) {
    message.error(error.message || t('missionConfig.fileUpload.uploadFailed'));
  } finally {
    uploadLoading.value = false;
  }
};

const copyFileId = async (id: string) => {
  try {
    await navigator.clipboard.writeText(id);
    message.success(t('missionConfig.fileUpload.copySuccess'));
  } catch {
    message.error(t('missionConfig.fileUpload.copyFailed'));
  }
};

const removeFile = (id: string) => {
  uploadedFiles.value = uploadedFiles.value.filter(file => file.id !== id);
  message.success(t('missionConfig.fileUpload.removeSuccess'));
};

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};
</script>

<style scoped lang="scss">
.modal-content {
  display: flex;
  gap: 20px;
  min-height: 500px;

  &.vertical-layout {
    flex-direction: column;
  }
}

.main-content {
  flex: 1;
  min-width: 0; // 防止flex布局溢出

  .workflow-form {
    max-height: 600px;
    overflow-y: auto;
    padding-right: 10px;
  }
}

.file-upload-section {
  width: 280px;
  flex-shrink: 0;

  &.upload-first {
    order: -1;
    width: 100%;
    margin-bottom: 20px;
  }

  .section-card {
    height: fit-content;
    background-color: #fff;
  }
}

// 响应式设计
@media (max-width: 768px) {
  .modal-content {
    flex-direction: column;
    gap: 16px;
  }

  .main-content,
  .file-upload-section {
    width: 100%;
  }

  .file-upload-section.upload-first {
    order: -1;
  }
}

@media (max-width: 480px) {
  .modal-content {
    gap: 12px;
  }
}

.section-card {
  margin-bottom: 16px;

  :deep(.ant-card-head) {
    min-height: 40px;
    padding: 0 12px;

    .ant-card-head-title {
      font-size: 14px;
      font-weight: 600;
    }
  }

  :deep(.ant-card-body) {
    padding: 12px;
  }
}

.text-center {
  text-align: center;
}

.py-10 {
  padding-top: 40px;
  padding-bottom: 40px;
}

.text-gray-500 {
  color: #6b7280;
}

/* ParamFields styles */
.params-section {
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px dashed #f0f0f0;
}

.section-label {
  margin-bottom: 12px;
  color: #666;
  font-size: 14px;
  font-weight: 500;
}

.param-item {
  margin-bottom: 5px;
  padding: 10px;
  background: #fff;
  border-radius: 4px;
  border: 1px solid #f0f0f0;
}

.param-label {
  display: block;
  text-align: right;
  margin-top: 8px;
  font-weight: 500;
  color: #333;
}

.required-asterisk {
  color: #ff4d4f;
  margin-left: 2px;
}

.param-description {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
  font-style: italic;
}

/* FileUploadWidget styles */
.file-upload-widget {
  width: 100%;
}

.upload-area {
  margin-bottom: 20px;
}

.uploaded-files {
  margin-top: 20px;
}

.uploaded-files h4 {
  margin-bottom: 12px;
  color: #333;
  font-weight: 600;
}

.file-list-item {
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;
}

.file-list-item:last-child {
  border-bottom: none;
}

.file-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.file-info {
  flex: 1;
  min-width: 0;
}

.file-name {
  font-weight: 500;
  margin-bottom: 4px;
  display: flex;
  align-items: center;
  color: #1890ff;
}

.file-meta {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: #666;
}

.file-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

@media (max-width: 480px) {
  .file-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .file-actions {
    align-self: flex-end;
  }
}

/* ProcessorSection styles */
.processor-item {
  margin-bottom: 16px;
  padding: 12px;
  background: #fafafa;
  border-radius: 4px;
  border: 1px solid #f0f0f0;
}

.processor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #f0f0f0;
}
</style>
