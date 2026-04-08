<template>
  <a-modal
      :open="open"
      :title="title"
      :width="isNarrowScreen ? '96vw' : '900px'"
      :style="{ top: '20px' }"
      @cancel="handleCancel"
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
          <a-collapse v-model:activeKey="activeSections" class="config-collapse" ghost>
            <a-collapse-panel key="basic" :header="t('missionConfig.basic.title')">
              <div class="section-intro">
                {{ t('missionConfig.basic.description') }}
              </div>
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
              <div v-if="showCronField" class="section-example">
                {{ t('missionConfig.cron.example') }}
              </div>
            </a-collapse-panel>

            <a-collapse-panel key="before" :header="t('missionConfig.beforeTask.title')">
              <div class="section-intro">
                {{ t('missionConfig.beforeTask.description') }}
              </div>
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
                      <template v-if="isFileParam(param)">
                        <div class="file-param-field">
                          <div v-if="getSelectedFilesForParam(param).length" class="selected-file-tags">
                            <a-tag
                              v-for="item in getSelectedFilesForParam(param)"
                              :key="item.id"
                              :closable="mode !== 'read'"
                              @close.prevent="removeSelectedFile(param, item.id)"
                            >
                              {{ item.name }}
                            </a-tag>
                          </div>
                          <div v-if="mode !== 'read'" class="file-param-actions">
                            <a-button @click="openFileLibrary(param)">
                              {{ t('missionConfig.fileSelector.choose') }}
                            </a-button>
                            <a-button
                              v-if="String(formData.before_execute.params[index].value || '').trim()"
                              @click="clearFileParam(param)"
                            >
                              {{ t('missionConfig.fileSelector.clear') }}
                            </a-button>
                          </div>
                        </div>
                      </template>
                      <a-auto-complete
                          v-else-if="isTableParam(param)"
                          v-model:value="formData.before_execute.params[index].value"
                          :options="getTableOptions(formData.before_execute)"
                          :placeholder="getPlaceholder(param)"
                          :disabled="mode === 'read'"
                          style="width: 100%"
                          allow-clear
                          filter-option
                      />
                      <template v-else-if="isQueryParam(param) && getTableOptions(formData.before_execute).length">
                        <a-textarea
                            v-model:value="formData.before_execute.params[index].value"
                            :auto-size="{ minRows: 2, maxRows: 6 }"
                            :placeholder="getPlaceholder(param)"
                            :disabled="mode === 'read'"
                        />
                        <div v-if="mode !== 'read'" class="table-quick-pick">
                          <a-select
                              size="small"
                              :placeholder="t('missionConfig.schema.quickPick')"
                              style="width: 200px"
                              @select="(v: any) => fillQuery(param, String(v))"
                              :value="undefined"
                              :options="getTableOptions(formData.before_execute)"
                          />
                        </div>
                      </template>
                      <a-textarea
                          v-else
                          v-model:value="formData.before_execute.params[index].value"
                          :auto-size="{ minRows: 1, maxRows: 6 }"
                          :placeholder="getPlaceholder(param)"
                          :disabled="mode === 'read'"
                          :rows="2"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                <div v-if="getParamDescription(param)" class="param-description">
                  {{ getParamDescription(param) }}
                </div>
              </div>
            </div>
            </a-collapse-panel>

            <a-collapse-panel key="source" :header="t('missionConfig.source.title')">
              <div class="section-intro">
                {{ t('missionConfig.source.description') }}
              </div>
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
                      <template v-if="isFileParam(param)">
                        <div class="file-param-field">
                          <div v-if="getSelectedFilesForParam(param).length" class="selected-file-tags">
                            <a-tag
                              v-for="item in getSelectedFilesForParam(param)"
                              :key="item.id"
                              :closable="mode !== 'read'"
                              @close.prevent="removeSelectedFile(param, item.id)"
                            >
                              {{ item.name }}
                            </a-tag>
                          </div>
                          <div v-if="mode !== 'read'" class="file-param-actions">
                            <a-button @click="openFileLibrary(param)">
                              {{ t('missionConfig.fileSelector.choose') }}
                            </a-button>
                            <a-button
                              v-if="String(formData.source.params[index].value || '').trim()"
                              @click="clearFileParam(param)"
                            >
                              {{ t('missionConfig.fileSelector.clear') }}
                            </a-button>
                          </div>
                        </div>
                      </template>
                      <a-auto-complete
                          v-else-if="isTableParam(param)"
                          v-model:value="formData.source.params[index].value"
                          :options="getTableOptions(formData.source)"
                          :placeholder="getPlaceholder(param)"
                          :disabled="mode === 'read'"
                          style="width: 100%"
                          allow-clear
                          filter-option
                      />
                      <template v-else-if="isQueryParam(param) && getTableOptions(formData.source).length">
                        <a-textarea
                            v-model:value="formData.source.params[index].value"
                            :auto-size="{ minRows: 2, maxRows: 6 }"
                            :placeholder="getPlaceholder(param)"
                            :disabled="mode === 'read'"
                        />
                        <div v-if="mode !== 'read'" class="table-quick-pick">
                          <a-select
                              size="small"
                              :placeholder="t('missionConfig.schema.quickPick')"
                              style="width: 200px"
                              @select="(v: any) => fillQuery(param, String(v))"
                              :value="undefined"
                              :options="getTableOptions(formData.source)"
                          />
                        </div>
                      </template>
                      <a-textarea
                          v-else
                          v-model:value="formData.source.params[index].value"
                          :auto-size="{ minRows: 1, maxRows: 6 }"
                          :placeholder="getPlaceholder(param)"
                          :disabled="mode === 'read'"
                          :rows="2"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                <div v-if="getParamDescription(param)" class="param-description">
                  {{ getParamDescription(param) }}
                </div>
              </div>
            </div>
            </a-collapse-panel>

            <a-collapse-panel key="processor" :header="t('missionConfig.processor.title')">
              <div class="section-intro">
                {{ t('missionConfig.processor.description') }}
              </div>
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
                        <template v-if="isFileParam(param)">
                          <div class="file-param-field">
                            <div v-if="getSelectedFilesForParam(param).length" class="selected-file-tags">
                              <a-tag
                                v-for="item in getSelectedFilesForParam(param)"
                                :key="item.id"
                                :closable="mode !== 'read'"
                                @close.prevent="removeSelectedFile(param, item.id)"
                              >
                                {{ item.name }}
                              </a-tag>
                            </div>
                            <div v-if="mode !== 'read'" class="file-param-actions">
                              <a-button @click="openFileLibrary(param)">
                                {{ t('missionConfig.fileSelector.choose') }}
                              </a-button>
                              <a-button
                                v-if="String(processor.params[index].value || '').trim()"
                                @click="clearFileParam(param)"
                              >
                                {{ t('missionConfig.fileSelector.clear') }}
                              </a-button>
                            </div>
                          </div>
                        </template>
                        <a-textarea
                            v-else
                            v-model:value="processor.params[index].value"
                            :auto-size="{ minRows: 1, maxRows: 6 }"
                            :placeholder="getPlaceholder(param)"
                            :disabled="mode === 'read'"
                            :rows="2"
                        />
                      </a-form-item>
                    </a-col>
                  </a-row>
                  <div v-if="getParamDescription(param)" class="param-description">
                    {{ getParamDescription(param) }}
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
            </a-collapse-panel>

            <a-collapse-panel key="sink" :header="t('missionConfig.sink.title')">
              <div class="section-intro">
                {{ t('missionConfig.sink.description') }}
              </div>
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
                      <template v-if="isFileParam(param)">
                        <div class="file-param-field">
                          <div v-if="getSelectedFilesForParam(param).length" class="selected-file-tags">
                            <a-tag
                              v-for="item in getSelectedFilesForParam(param)"
                              :key="item.id"
                              :closable="mode !== 'read'"
                              @close.prevent="removeSelectedFile(param, item.id)"
                            >
                              {{ item.name }}
                            </a-tag>
                          </div>
                          <div v-if="mode !== 'read'" class="file-param-actions">
                            <a-button @click="openFileLibrary(param)">
                              {{ t('missionConfig.fileSelector.choose') }}
                            </a-button>
                            <a-button
                              v-if="String(formData.sink.params[index].value || '').trim()"
                              @click="clearFileParam(param)"
                            >
                              {{ t('missionConfig.fileSelector.clear') }}
                            </a-button>
                          </div>
                        </div>
                      </template>
                      <a-auto-complete
                          v-else-if="isTableParam(param)"
                          v-model:value="formData.sink.params[index].value"
                          :options="getTableOptions(formData.sink)"
                          :placeholder="getPlaceholder(param)"
                          :disabled="mode === 'read'"
                          style="width: 100%"
                          allow-clear
                          filter-option
                      />
                      <template v-else-if="isQueryParam(param) && getTableOptions(formData.sink).length">
                        <a-textarea
                            v-model:value="formData.sink.params[index].value"
                            :auto-size="{ minRows: 2, maxRows: 6 }"
                            :placeholder="getPlaceholder(param)"
                            :disabled="mode === 'read'"
                        />
                        <div v-if="mode !== 'read'" class="table-quick-pick">
                          <a-select
                              size="small"
                              :placeholder="t('missionConfig.schema.quickPick')"
                              style="width: 200px"
                              @select="(v: any) => fillQuery(param, String(v))"
                              :value="undefined"
                              :options="getTableOptions(formData.sink)"
                          />
                        </div>
                      </template>
                      <a-textarea
                          v-else
                          v-model:value="formData.sink.params[index].value"
                          :auto-size="{ minRows: 1, maxRows: 6 }"
                          :placeholder="getPlaceholder(param)"
                          :disabled="mode === 'read'"
                          :rows="2"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                <div v-if="getParamDescription(param)" class="param-description">
                  {{ getParamDescription(param) }}
                </div>
              </div>
            </div>
            </a-collapse-panel>

            <a-collapse-panel key="after" :header="t('missionConfig.afterTask.title')">
              <div class="section-intro">
                {{ t('missionConfig.afterTask.description') }}
              </div>
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
                      <template v-if="isFileParam(param)">
                        <div class="file-param-field">
                          <div v-if="getSelectedFilesForParam(param).length" class="selected-file-tags">
                            <a-tag
                              v-for="item in getSelectedFilesForParam(param)"
                              :key="item.id"
                              :closable="mode !== 'read'"
                              @close.prevent="removeSelectedFile(param, item.id)"
                            >
                              {{ item.name }}
                            </a-tag>
                          </div>
                          <div v-if="mode !== 'read'" class="file-param-actions">
                            <a-button @click="openFileLibrary(param)">
                              {{ t('missionConfig.fileSelector.choose') }}
                            </a-button>
                            <a-button
                              v-if="String(formData.after_execute.params[index].value || '').trim()"
                              @click="clearFileParam(param)"
                            >
                              {{ t('missionConfig.fileSelector.clear') }}
                            </a-button>
                          </div>
                        </div>
                      </template>
                      <a-auto-complete
                          v-else-if="isTableParam(param)"
                          v-model:value="formData.after_execute.params[index].value"
                          :options="getTableOptions(formData.after_execute)"
                          :placeholder="getPlaceholder(param)"
                          :disabled="mode === 'read'"
                          style="width: 100%"
                          allow-clear
                          filter-option
                      />
                      <template v-else-if="isQueryParam(param) && getTableOptions(formData.after_execute).length">
                        <a-textarea
                            v-model:value="formData.after_execute.params[index].value"
                            :auto-size="{ minRows: 2, maxRows: 6 }"
                            :placeholder="getPlaceholder(param)"
                            :disabled="mode === 'read'"
                        />
                        <div v-if="mode !== 'read'" class="table-quick-pick">
                          <a-select
                              size="small"
                              :placeholder="t('missionConfig.schema.quickPick')"
                              style="width: 200px"
                              @select="(v: any) => fillQuery(param, String(v))"
                              :value="undefined"
                              :options="getTableOptions(formData.after_execute)"
                          />
                        </div>
                      </template>
                      <a-textarea
                          v-else
                          v-model:value="formData.after_execute.params[index].value"
                          :auto-size="{ minRows: 1, maxRows: 6 }"
                          :placeholder="getPlaceholder(param)"
                          :disabled="mode === 'read'"
                          :rows="2"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                <div v-if="getParamDescription(param)" class="param-description">
                  {{ getParamDescription(param) }}
                </div>
              </div>
            </div>
            </a-collapse-panel>
          </a-collapse>

          <a-card size="small" :title="t('missionConfig.preview.title')" class="section-card preview-card">
            <div class="preview-description">
              {{ t('missionConfig.preview.description') }}
            </div>
            <pre class="config-preview">{{ configPreview }}</pre>
          </a-card>
        </a-form>
      </div>

    </div>
    <FileLibraryModal
      v-model:open="fileLibraryOpen"
      :multiple="activeFileParamMultiple"
      :selected-ids="activeFileParamSelectedIds"
      :title="t('missionConfig.fileSelector.modalTitle')"
      @confirm="handleFileLibraryConfirm"
    />

    <template #footer>
      <a-space v-if="mode !== 'read'">
        <a-button @click="handleCancel">{{ t('common.cancel') }}</a-button>
        <a-button :loading="templateSaving" @click="handleSaveTemplate">
          {{ t('missionConfig.template.save') }}
        </a-button>
        <a-button type="primary" @click="handleOk">{{ t('common.confirm') }}</a-button>
      </a-space>
      <a-space v-else>
        <a-button @click="handleCancel">{{ t('common.close') }}</a-button>
      </a-space>
    </template>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, reactive, watch, computed, onMounted, onUnmounted } from "vue";
import { message } from "ant-design-vue";
import type { FormInstance } from "ant-design-vue";
import { addTask, updateTask, getTypeByComponent, saveTaskTemplate } from "../api/mission";
import type { ConfigItem, TaskType, ParamItem } from "../types/mission";
import { useI18n } from "vue-i18n";
import { getFileList } from "../api/file.ts";
import {RuleObject} from "ant-design-vue/es/form";
import { ApiRequestError } from "../utils/request";
import type { FileInfo } from "../types";
import FileLibraryModal from "./FileLibraryModal.vue";
import { getDataSourceSchema } from "../api/datasource";

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
  (e: "templateSaved"): void;
}>();

// 响应式数据
const formRef = ref<FormInstance>();
const screenWidth = ref(window.innerWidth);
const activeSections = ref<string[]>(["basic", "source", "sink"]);
const templateSaving = ref(false);

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

const buildTaskPayload = (includeId = props.mode === "edit") => {
  const payload: Record<string, any> = {
    mission_name: formData.mission_name,
    cron: formData.cron,
    params: {
      before_execute: formData.before_execute.type ? formData.before_execute : null,
      source: formData.source.type ? formData.source : null,
      processors: formData.processors.filter((p) => p.type),
      sink: formData.sink.type ? formData.sink : null,
      after_execute: formData.after_execute.type ? formData.after_execute : null,
    },
  };

  if (includeId && props.id) {
    payload.id = props.id;
  }

  return payload;
};

const configPreview = computed(() => JSON.stringify(buildTaskPayload(false), null, 2));

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
    description: param.description,
    defaultValue: param.defaultValue,
    placeholder: param.placeholder,
    example: param.example,
    type: param.type,
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
    if (!props.data) {
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
      await refreshSelectedFilesFromParams();
      return;
    }

    const data = props.data;
    const record = props.record || {};

    formData.id = props.mode === "edit" ? props.id || "" : "";
    formData.mission_name = props.mode === "add"
        ? `${record.mission_name || ""}${t("missionConfig.copy.suffix")}`
        : record.mission_name || "";

    const isManualTask = props.taskType === "manual" || record.cron === 'manual';
    formData.cron = isManualTask ? 'manual' : (record.cron || "");

    resetConfigItem(formData.before_execute, data.before_execute, "execute");
    resetConfigItem(formData.source, data.source, "source");

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
    await refreshSelectedFilesFromParams();
  } catch (error) {
    console.error("初始化表单失败:", error);
    message.error(t("missionConfig.form.initFailed"));
  }
};

// 监听打开状态
watch(() => props.open, (val) => {
  if (val) {
    activeFileParam.value = null;
    initForm();
  } else {
    activeFileParam.value = null;
    fileLibraryOpen.value = false;
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
    const payload = buildTaskPayload();

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
    const apiError = error as ApiRequestError;
    if (apiError?.details?.errors?.length) {
      (formRef.value as any)?.setFields?.(
        apiError.details.errors.map((item) => ({
          name: item.field,
          errors: [item.message],
        }))
      );
    }
    console.error("保存任务失败：", error);
  }
};

const handleSaveTemplate = async () => {
  try {
    await formRef.value?.validate();
    templateSaving.value = true;
    const payload = buildTaskPayload(false);
    const taskType = formData.cron === "manual" ? "manual" : "scheduled";
    const res = await saveTaskTemplate({
      name: formData.mission_name,
      cron: formData.cron,
      tasktypes: taskType,
      params: payload.params,
    });
    if (res.code === 0) {
      message.success(t("missionConfig.template.saveSuccess"));
      emit("templateSaved");
    }
  } catch (error) {
    const apiError = error as ApiRequestError;
    if (apiError?.details?.errors?.length) {
      (formRef.value as any)?.setFields?.(
        apiError.details.errors.map((item) => ({
          name: item.field,
          errors: [item.message],
        }))
      );
    }
  } finally {
    templateSaving.value = false;
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
  if (isFileParam(param)) {
    return t("missionConfig.fileSelector.help");
  }
  if (param.placeholder) {
    return param.placeholder;
  }
  if (param.defaultValue) {
    return `${param.description || ''} (默认值: ${param.defaultValue})`;
  }
  return param.description || '';
};

const getParamDescription = (param: any) => {
  if (isFileParam(param)) {
    return t("missionConfig.fileSelector.help");
  }
  const chunks = [];
  if (param.description) {
    chunks.push(param.description);
  }
  if (param.defaultValue) {
    chunks.push(`Default: ${param.defaultValue}`);
  }
  if (param.example) {
    chunks.push(`Example: ${param.example}`);
  }
  if (param.type) {
    chunks.push(`Type: ${param.type}`);
  }
  return chunks.join(" | ");
};

// File selector 相关功能
const fileLibraryOpen = ref(false);
const activeFileParam = ref<ParamItem | null>(null);
const selectedFileMap = ref<Record<string, FileInfo>>({});

const getAllConfigItems = () => {
  return [
    formData.before_execute,
    formData.source,
    ...formData.processors,
    formData.sink,
    formData.after_execute,
  ].filter(Boolean) as ConfigItem[];
};

const isFileParam = (param: ParamItem) => {
  return String(param.key || "").includes("file_id");
};

const isMultiFileParam = (param: ParamItem) => {
  return String(param.key || "").includes("file_ids");
};

const parseFileIds = (value: any) => {
  if (!value) {
    return [];
  }
  return String(value)
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
};

const mergeSelectedFiles = (items: FileInfo[]) => {
  const nextMap = { ...selectedFileMap.value };
  items.forEach((item) => {
    nextMap[item.id] = item;
  });
  selectedFileMap.value = nextMap;
};

const collectSelectedIdsFromParams = () => {
  const ids = new Set<string>();
  getAllConfigItems().forEach((config) => {
    (config.params || []).forEach((param) => {
      if (!isFileParam(param)) {
        return;
      }
      parseFileIds(param.value).forEach((id) => ids.add(id));
    });
  });
  return Array.from(ids);
};

const refreshSelectedFilesFromParams = async () => {
  const ids = collectSelectedIdsFromParams();
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

const getSelectedFilesForParam = (param: ParamItem) => {
  return parseFileIds(param.value).map((id) => selectedFileMap.value[id] || {
    id,
    name: id,
    size: 0,
    created_at: "",
    path: "",
    ex_name: "",
  });
};

const activeFileParamMultiple = computed(() => {
  return activeFileParam.value ? isMultiFileParam(activeFileParam.value) : false;
});

const activeFileParamSelectedIds = computed(() => {
  return activeFileParam.value ? parseFileIds(activeFileParam.value.value) : [];
});

const openFileLibrary = (param: ParamItem) => {
  activeFileParam.value = param;
  fileLibraryOpen.value = true;
};

const handleFileLibraryConfirm = (files: FileInfo[]) => {
  if (!activeFileParam.value) {
    return;
  }
  mergeSelectedFiles(files);
  activeFileParam.value.value = isMultiFileParam(activeFileParam.value)
    ? files.map((item) => item.id).join(",")
    : (files[0]?.id || "");
};

const clearFileParam = (param: ParamItem) => {
  param.value = "";
};

const removeSelectedFile = (param: ParamItem, id: string) => {
  const nextIds = parseFileIds(param.value).filter((item) => item !== id);
  param.value = isMultiFileParam(param) ? nextIds.join(",") : "";
};

// ── Schema 发现 ───────────────────────────────────────────────────────────────

interface TableInfo {
  name: string
  columns: { name: string; type: string; nullable: boolean }[]
}

// 按 datasource ID 缓存 schema，避免重复请求
const schemaCache = ref<Record<string, TableInfo[]>>({})
const schemaLoading = ref<Set<string>>(new Set())

const loadSchema = async (datasourceId: string | null | undefined) => {
  if (!datasourceId) return
  if (schemaCache.value[datasourceId] !== undefined) return
  if (schemaLoading.value.has(datasourceId)) return
  schemaLoading.value = new Set([...schemaLoading.value, datasourceId])
  try {
    const res = await getDataSourceSchema(datasourceId)
    if (res.code === 0) {
      schemaCache.value = { ...schemaCache.value, [datasourceId]: res.data.tables || [] }
    }
  } catch {
    schemaCache.value = { ...schemaCache.value, [datasourceId]: [] }
  } finally {
    const next = new Set(schemaLoading.value)
    next.delete(datasourceId)
    schemaLoading.value = next
  }
}

// datasource 变更时自动拉取 schema
watch(() => formData.source.data_source, loadSchema)
watch(() => formData.sink.data_source, loadSchema)
watch(() => formData.before_execute.data_source, loadSchema)
watch(() => formData.after_execute.data_source, loadSchema)

const getTableOptions = (config: ConfigItem) => {
  if (!config.data_source) return []
  return (schemaCache.value[config.data_source] || []).map(t => ({ value: t.name }))
}

const isTableParam = (param: ParamItem) => param.key === 'table'
const isQueryParam = (param: ParamItem) => param.key === 'query'

// 点击表名：将 query 填入 SELECT * FROM <table>
const fillQuery = (param: ParamItem, tableName: string) => {
  param.value = `SELECT * FROM ${tableName}`
}

// const formatFileSize = (bytes: number): string => {
//   if (bytes === 0) return '0 Bytes';
//   const k = 1024;
//   const sizes = ['Bytes', 'KB', 'MB', 'GB'];
//   const i = Math.floor(Math.log(bytes) / Math.log(k));
//   return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
// };
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

// 响应式设计
@media (max-width: 768px) {
  .modal-content {
    flex-direction: column;
    gap: 16px;
  }

  .main-content {
    width: 100%;
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

.config-collapse {
  margin-bottom: 16px;

  :deep(.ant-collapse-item) {
    margin-bottom: 12px;
    border: 1px solid #f0f0f0;
    border-radius: 8px;
    background: #fff;
    overflow: hidden;
  }

  :deep(.ant-collapse-header) {
    font-weight: 600;
    background: #fafafa;
  }

  :deep(.ant-collapse-content-box) {
    padding-top: 4px;
  }
}

.section-intro {
  margin-bottom: 12px;
  color: #6b7280;
  font-size: 13px;
  line-height: 1.6;
}

.section-example {
  margin-top: -4px;
  margin-bottom: 4px;
  color: #8b5e3c;
  font-size: 12px;
  line-height: 1.6;
}

.preview-card {
  margin-bottom: 0;
}

.preview-description {
  margin-bottom: 12px;
  color: #6b7280;
  font-size: 13px;
  line-height: 1.6;
}

.config-preview {
  margin: 0;
  padding: 16px;
  max-height: 280px;
  overflow: auto;
  border-radius: 8px;
  background: #0f172a;
  color: #e2e8f0;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
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

.file-param-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: flex-start;
}

.selected-file-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
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

.table-quick-pick {
  margin-top: 6px;
}

.preview-empty {
  text-align: center;
  color: #999;
  padding: 40px 0;
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
