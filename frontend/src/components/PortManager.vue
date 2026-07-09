<script setup>
import {h, onMounted, computed, ref} from 'vue'
import {portStore} from '../stores/ports.js'
import {NIcon, NTag, NButton, useMessage, NInputNumber} from 'naive-ui'
import {CloseCircleOutline, HardwareChipOutline, CreateOutline} from '@vicons/ionicons5'

const props = defineProps({ isDark: { type: Boolean, default: true } })
const message = useMessage()

// 修改端口对话框状态
const showPortDialog = ref(false)
const editingService = ref(null)
const newPort = ref(0)
const changing = ref(false)

onMounted(() => {
  portStore.fetchPorts()
})

function openPortDialog(row) {
  editingService.value = row
  newPort.value = row.listenPort || row.defaultPort
  showPortDialog.value = true
}

async function handlePortChange() {
  if (!editingService.value || newPort.value <= 0 || newPort.value > 65535) {
    message.error('端口范围: 1-65535')
    return
  }
  changing.value = true
  try {
    await portStore.changePort(editingService.value.serviceName, newPort.value)
    message.success(`端口已修改为 ${newPort.value}，服务已重启`)
    showPortDialog.value = false
  } catch (e) {
    message.error('修改失败: ' + String(e))
  } finally {
    changing.value = false
  }
}

const columns = computed(() => [
  {
    title: '服务名称',
    key: 'name',
    width: 150,
    minWidth: 120,
    render: (row) => h('span', { style: { fontWeight: 500, color: '#e0e0e6' } }, row.name),
  },
  {
    title: '服务状态',
    key: 'running',
    width: 90,
    minWidth: 80,
    align: 'center',
    render: (row) => row.running
      ? h(NTag, { type: 'success', size: 'small', bordered: false }, { default: () => '运行中' })
      : h(NTag, { type: 'default', size: 'small', bordered: false }, { default: () => '未运行' }),
  },
  {
    title: '监听端口',
    key: 'listenPort',
    width: 120,
    minWidth: 90,
    align: 'center',
    render: (row) => {
      if (row.listenPort === 0) return h('span', { style: { color: '#666' } }, '未监听')
      const isDefault = row.listenPort === row.defaultPort
      return h('span', [
        h(NTag, { type: 'info', size: 'small', bordered: false }, { default: () => row.listenPort }),
        !isDefault ? h('span', { style: { color: '#e88080', fontSize: '11px', marginLeft: '4px' } }, '(非默认)') : null,
      ])
    },
  },
  {
    title: '默认端口',
    key: 'defaultPort',
    width: 90,
    minWidth: 70,
    align: 'center',
    render: (row) => h('span', { style: { color: '#888', fontSize: '12px' } }, row.defaultPort),
  },
  {
    title: '协议',
    key: 'protocol',
    width: 60,
    minWidth: 50,
    align: 'center',
    render: (row) => h('span', { style: { color: '#888', fontSize: '12px', textTransform: 'uppercase' } }, row.protocol),
  },
  {
    title: '说明',
    key: 'description',
    minWidth: 200,
    ellipsis: { tooltip: true },
    render: (row) => h('span', { style: { color: '#888', fontSize: '12px' } }, row.description),
  },
  {
    title: '修改端口',
    key: 'editPort',
    width: 80,
    minWidth: 70,
    align: 'center',
    render: (row) => h(NButton, {
      text: true,
      size: 'small',
      onClick: () => openPortDialog(row),
    }, {
      icon: () => h(NIcon, { size: 16, color: '#4361ee' }, { default: () => h(CreateOutline) }),
    }),
  },
])

const scrollX = computed(() => columns.value.reduce((s, c) => s + (c.width || c.minWidth || 100), 0))
</script>

<template>
  <div class="port-manager">
    <div class="port-header">
      <NIcon :size="20" color="#4361ee"><HardwareChipOutline/></NIcon>
      <span class="port-title">常用端口管理</span>
      <span class="port-desc">查看服务运行状态、实际监听端口</span>
    </div>

    <div v-if="portStore.loading" class="loading-state">
      <n-spin size="medium"/><span>加载端口信息...</span>
    </div>
    <div v-else-if="portStore.error" class="error-state">
      <NIcon :size="40" color="#e88080"><CloseCircleOutline/></NIcon>
      <span>{{ portStore.error }}</span>
      <n-button size="small" @click="portStore.fetchPorts()">重试</n-button>
    </div>
    <n-data-table
      v-else
      :columns="columns"
      :data="portStore.ports"
      :scroll-x="scrollX"
      :max-height="500"
      :scrollbar-props="{size: 6}"
      :bordered="false"
      size="small"
      striped
    />

    <!-- 修改端口对话框 -->
    <n-modal v-model:show="showPortDialog" preset="dialog" title="修改服务端口"
      :positive-text="changing ? '修改中...' : '确认修改'" negative-text="取消"
      :positive-button-props="{ loading: changing, disabled: changing }"
      :negative-button-props="{ disabled: changing }"
      :closable="!changing" :mask-closable="!changing"
      @positive-click="handlePortChange">
      <div v-if="editingService" style="margin-bottom: 16px;">
        <p style="margin-bottom: 8px; color: #e0e0e6;">
          服务: <strong>{{ editingService.name }}</strong>
        </p>
        <p style="margin-bottom: 8px; color: #888; font-size: 12px;">
          当前端口: {{ editingService.listenPort || editingService.defaultPort }}
          (默认: {{ editingService.defaultPort }})
        </p>
        <div style="display: flex; align-items: center; gap: 8px;">
          <span style="color: #e0e0e6;">新端口:</span>
          <n-input-number v-model:value="newPort" :min="1" :max="65535" size="small" style="width: 120px;" :disabled="changing"/>
        </div>
        <p style="margin-top: 12px; color: #e88080; font-size: 12px;">
          ⚠️ 修改端口后服务将自动重启，当前连接可能会断开
        </p>
      </div>
    </n-modal>
  </div>
</template>

<style scoped>
.port-manager {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}
.port-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}
.port-title {
  font-size: 15px;
  font-weight: 600;
  color: #e0e0e6;
}
.port-desc {
  font-size: 12px;
  color: #666;
  margin-left: 8px;
}
.loading-state, .error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 60px 0;
  color: #666;
  font-size: 14px;
}
</style>
