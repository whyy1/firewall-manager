<script setup>
import {ref, watch, computed} from 'vue'
import {store} from '../stores/rules.js'
import {useMessage} from 'naive-ui'
import {OpenFileDialog} from '../../wailsjs/go/main/App'

const props = defineProps({
  show: Boolean,
  rule: Object,
})
const emit = defineEmits(['update:show'])
const message = useMessage()

const isEdit = computed(() => !!props.rule)

const form = ref({
  name: '',
  direction: 'in',
  action: 'allow',
  program: '',
  localAddr: '',
  remoteAddr: '',
  protocol: 'any',
  localPort: '',
  remotePort: '',
  profile: 'any',
  enabled: true,
})

watch(() => props.show, (val) => {
  if (val && props.rule) {
    form.value = {
      name: props.rule.name,
      direction: props.rule.direction,
      action: props.rule.action,
      program: props.rule.program === 'Any' ? '' : props.rule.program,
      localAddr: props.rule.localAddr === 'Any' ? '' : props.rule.localAddr || '',
      remoteAddr: props.rule.remoteAddr === 'Any' ? '' : props.rule.remoteAddr || '',
      protocol: props.rule.protocol || 'any',
      localPort: props.rule.localPort || '',
      remotePort: props.rule.remotePort || '',
      profile: props.rule.profile || 'any',
      enabled: props.rule.enabled,
    }
  } else if (val) {
    form.value = {
      name: '',
      direction: store.direction,
      action: 'allow',
      program: '',
      localAddr: '',
      remoteAddr: '',
      protocol: 'any',
      localPort: '',
      remotePort: '',
      profile: 'any',
      enabled: true,
    }
  }
})

const protocolOptions = [
  {label: '任意', value: 'any'},
  {label: 'TCP', value: 'tcp'},
  {label: 'UDP', value: 'udp'},
]

const actionOptions = [
  {label: '允许', value: 'allow'},
  {label: '阻止', value: 'block'},
]

const directionOptions = [
  {label: '入站', value: 'in'},
  {label: '出站', value: 'out'},
]

const profileOptions = [
  {label: '任意', value: 'any'},
  {label: '域', value: 'domain'},
  {label: '专用', value: 'private'},
  {label: '公用', value: 'public'},
]

async function handleSubmit() {
  if (!form.value.name.trim()) {
    message.warning('请输入规则名称')
    return
  }

  try {
    await store.addRule({
      name: form.value.name,
      direction: form.value.direction,
      action: form.value.action,
      program: form.value.program || '',
      localAddr: form.value.localAddr || '',
      remoteAddr: form.value.remoteAddr || '',
      protocol: form.value.protocol,
      localPort: form.value.localPort || '',
      remotePort: form.value.remotePort || '',
      profile: form.value.profile,
      enabled: form.value.enabled,
    })
    message.success(isEdit.value ? '规则已更新' : '规则已添加')
    emit('update:show', false)
  } catch (e) {
    message.error('操作失败: ' + String(e))
  }
}

async function handleBrowseProgram() {
  try {
    const path = await OpenFileDialog()
    if (path) {
      form.value.program = path
    }
  } catch (e) {
    message.error('选择文件失败: ' + String(e))
  }
}
</script>

<template>
  <n-drawer
    :show="show"
    :width="440"
    @update:show="emit('update:show', $event)"
  >
    <n-drawer-content :title="isEdit ? '编辑规则' : '添加规则'">
      <div class="rule-editor">
        <n-form label-placement="top" :model="form" class="rule-form">
          <!-- 基本信息 -->
          <div class="form-section">
            <h3 class="form-section-title">基本信息</h3>
            <n-form-item label="规则名称" required>
              <n-input
                v-model:value="form.name"
                placeholder="如 Allow Chrome"
                :disabled="isEdit"
                clearable
              />
            </n-form-item>

            <n-grid :cols="2" :x-gap="12">
              <n-gi>
                <n-form-item label="方向">
                  <n-select v-model:value="form.direction" :options="directionOptions"/>
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="动作">
                  <n-select v-model:value="form.action" :options="actionOptions"/>
                </n-form-item>
              </n-gi>
            </n-grid>
          </div>

          <!-- 程序与地址 -->
          <div class="form-section">
            <h3 class="form-section-title">程序与地址</h3>
            <n-form-item label="程序路径">
              <n-input-group>
                <n-input
                  v-model:value="form.program"
                  placeholder="留空表示任意程序"
                  clearable
                  style="flex: 1;"
                />
                <n-button type="primary" @click="handleBrowseProgram">
                  浏览
                </n-button>
              </n-input-group>
              <template #feedback>
                <span class="form-hint">如 C:\Program Files\app.exe</span>
              </template>
            </n-form-item>

            <n-grid :cols="2" :x-gap="12">
              <n-gi>
                <n-form-item label="本地地址">
                  <n-input v-model:value="form.localAddr" placeholder="留空表示任意" clearable/>
                  <template #feedback>
                    <span class="form-hint">如 192.168.1.100</span>
                  </template>
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="远程地址">
                  <n-input v-model:value="form.remoteAddr" placeholder="留空表示任意" clearable/>
                  <template #feedback>
                    <span class="form-hint">如 10.0.0.0/24</span>
                  </template>
                </n-form-item>
              </n-gi>
            </n-grid>
          </div>

          <!-- 协议与端口 -->
          <div class="form-section">
            <h3 class="form-section-title">协议与端口</h3>
            <n-grid :cols="2" :x-gap="12">
              <n-gi>
                <n-form-item label="协议">
                  <n-select v-model:value="form.protocol" :options="protocolOptions"/>
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="配置文件">
                  <n-select v-model:value="form.profile" :options="profileOptions"/>
                </n-form-item>
              </n-gi>
            </n-grid>

            <n-grid :cols="2" :x-gap="12">
              <n-gi>
                <n-form-item label="本地端口">
                  <n-input v-model:value="form.localPort" placeholder="如 80,443" clearable/>
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="远程端口">
                  <n-input v-model:value="form.remotePort" placeholder="如 8080" clearable/>
                </n-form-item>
              </n-gi>
            </n-grid>
          </div>

          <!-- 状态 -->
          <div class="form-section">
            <n-form-item label="启用规则">
              <n-switch v-model:value="form.enabled"/>
            </n-form-item>
          </div>
        </n-form>
      </div>

      <template #footer>
        <div class="form-footer">
          <n-button @click="emit('update:show', false)">取消</n-button>
          <n-button type="primary" @click="handleSubmit">
            {{ isEdit ? '保存' : '添加' }}
          </n-button>
        </div>
      </template>
    </n-drawer-content>
  </n-drawer>
</template>

<style scoped>
.rule-editor {
  padding: 0 4px;
}

.rule-form {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-bottom: 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.form-section:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.form-section-title {
  font-size: 14px;
  font-weight: 600;
  color: #e0e0e6;
  margin: 0;
  line-height: 1.4;
}

.form-hint {
  font-size: 11px;
  color: #666;
  line-height: 1.4;
}

.form-footer {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}
</style>
