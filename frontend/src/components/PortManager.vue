<script setup>
import {h, onMounted, computed, ref} from 'vue'
import {portStore} from '../stores/ports.js'
import {NIcon, NTag, NButton, NTooltip, useMessage, NInputNumber} from 'naive-ui'
import {CloseCircleOutline, HardwareChipOutline, CreateOutline, InformationCircleOutline} from '@vicons/ionicons5'

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
    width: 160,
    minWidth: 120,
    render: (row) => h('div', { style: { display: 'flex', alignItems: 'center', gap: '8px' } }, [
      h(NIcon, {
        size: 16,
        color: row.running ? '#63e2b7' : '#666',
      }, { default: () => h(HardwareChipOutline) }),
      h('span', {
        style: {
          fontWeight: 500,
          color: '#e0e0e6',
          fontSize: '13px',
        }
      }, row.name),
    ]),
  },
  {
    title: '服务状态',
    key: 'running',
    width: 100,
    minWidth: 80,
    align: 'center',
    render: (row) => row.running
      ? h(NTag, {
          type: 'success',
          size: 'small',
          bordered: false,
          round: true,
        }, { default: () => '运行中' })
      : h(NTag, {
          type: 'default',
          size: 'small',
          bordered: false,
          round: true,
        }, { default: () => '未运行' }),
  },
  {
    title: '监听端口',
    key: 'listenPort',
    width: 130,
    minWidth: 100,
    align: 'center',
    render: (row) => {
      if (row.listenPort === 0) return h('span', { style: { color: '#666', fontSize: '12px' } }, '未监听')
      const isDefault = row.listenPort === row.defaultPort
      return h('div', { style: { display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '4px' } }, [
        h(NTag, {
          type: 'info',
          size: 'small',
          bordered: false,
        }, { default: () => row.listenPort }),
        !isDefault ? h(NTooltip, { trigger: 'hover' }, {
          trigger: () => h(NIcon, {
            size: 14,
            color: '#e88080',
            style: { cursor: 'help' },
          }, { default: () => h(InformationCircleOutline) }),
          default: () => h('span', { style: { fontSize: '12px' } }, `默认端口: ${row.defaultPort}`),
        }) : null,
      ])
    },
  },
  {
    title: '默认端口',
    key: 'defaultPort',
    width: 90,
    minWidth: 70,
    align: 'center',
    render: (row) => h('span', {
      style: {
        color: '#888',
        fontSize: '12px',
        fontFamily: 'monospace',
      }
    }, row.defaultPort),
  },
  {
    title: '协议',
    key: 'protocol',
    width: 70,
    minWidth: 60,
    align: 'center',
    render: (row) => h(NTag, {
      size: 'small',
      bordered: false,
      type: row.protocol === 'tcp' ? 'warning' : 'info',
    }, { default: () => row.protocol.toUpperCase() }),
  },
  {
    title: '说明',
    key: 'description',
    minWidth: 200,
    ellipsis: { tooltip: true },
    render: (row) => h('span', {
      style: {
        color: '#888',
        fontSize: '12px',
        lineHeight: '1.5',
      }
    }, row.description),
  },
  {
    title: '操作',
    key: 'editPort',
    width: 80,
    minWidth: 70,
    align: 'center',
    render: (row) => h(NButton, {
      text: true,
      size: 'small',
      type: 'primary',
      onClick: () => openPortDialog(row),
      style: { minWidth: '44px', minHeight: '44px' },
    }, {
      icon: () => h(NIcon, { size: 16 }, { default: () => h(CreateOutline) }),
    }),
  },
])

const scrollX = computed(() => columns.value.reduce((s, c) => s + (c.width || c.minWidth || 100), 0))
</script>

<template>
  <div class="port-manager">
    <div class="port-header">
      <div class="port-header-left">
        <NIcon :size="24" color="#4361ee"><HardwareChipOutline/></NIcon>
        <div class="port-header-text">
          <h2 class="port-title">常用端口管理</h2>
          <p class="port-desc">查看服务运行状态、实际监听端口</p>
        </div>
      </div>
      <n-button size="small" @click="portStore.fetchPorts()" :loading="portStore.loading">
        刷新
      </n-button>
    </div>

    <div v-if="portStore.loading && portStore.ports.length === 0" class="loading-state">
      <n-spin size="large"/>
      <span class="loading-text">加载端口信息...</span>
    </div>
    <div v-else-if="portStore.error" class="error-state">
      <NIcon :size="48" color="#e88080"><CloseCircleOutline/></NIcon>
      <span class="error-text">{{ portStore.error }}</span>
      <n-button type="primary" size="small" @click="portStore.fetchPorts()">重试</n-button>
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
      :row-class-name="(row) => !row.running ? 'row-stopped' : ''"
    />

    <!-- 修改端口对话框 -->
    <n-modal v-model:show="showPortDialog" preset="dialog" title="修改服务端口"
      :positive-text="changing ? '修改中...' : '确认修改'" negative-text="取消"
      :positive-button-props="{ loading: changing, disabled: changing }"
      :negative-button-props="{ disabled: changing }"
      :closable="!changing" :mask-closable="!changing"
      @positive-click="handlePortChange">
      <div v-if="editingService" class="port-dialog-content">
        <div class="port-dialog-info">
          <p class="port-dialog-label">服务</p>
          <p class="port-dialog-value">{{ editingService.name }}</p>
        </div>
        <div class="port-dialog-info">
          <p class="port-dialog-label">当前端口</p>
          <p class="port-dialog-value">
            {{ editingService.listenPort || editingService.defaultPort }}
            <span v-if="editingService.listenPort !== editingService.defaultPort" class="port-dialog-default">
              (默认: {{ editingService.defaultPort }})
            </span>
          </p>
        </div>
        <div class="port-dialog-input">
          <span class="port-dialog-label">新端口</span>
          <n-input-number
            v-model:value="newPort"
            :min="1"
            :max="65535"
            size="small"
            style="width: 120px;"
            :disabled="changing"
            placeholder="1-65535"
          />
        </div>
        <div class="port-dialog-warning">
          <NIcon :size="16" color="#e88080"><InformationCircleOutline/></NIcon>
          <span>修改端口后服务将自动重启，当前连接可能会断开</span>
        </div>
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
  gap: 20px;
}

.port-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.port-header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.port-header-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.port-title {
  font-size: 18px;
  font-weight: 600;
  color: #e0e0e6;
  margin: 0;
  line-height: 1.3;
}

.port-desc {
  font-size: 13px;
  color: #888;
  margin: 0;
  line-height: 1.4;
}

.loading-state, .error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 80px 0;
  color: #888;
}

.loading-text, .error-text {
  font-size: 14px;
  line-height: 1.5;
}

:deep(.row-stopped td) {
  opacity: 0.6;
}

.port-dialog-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 8px;
}

.port-dialog-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.port-dialog-label {
  font-size: 12px;
  color: #888;
  margin: 0;
  line-height: 1.4;
}

.port-dialog-value {
  font-size: 14px;
  color: #e0e0e6;
  margin: 0;
  font-weight: 500;
  line-height: 1.5;
}

.port-dialog-default {
  font-size: 12px;
  color: #888;
  font-weight: normal;
}

.port-dialog-input {
  display: flex;
  align-items: center;
  gap: 12px;
}

.port-dialog-warning {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 12px;
  background: rgba(232, 128, 128, 0.1);
  border-radius: 6px;
  font-size: 12px;
  color: #e88080;
  line-height: 1.5;
}
</style>
