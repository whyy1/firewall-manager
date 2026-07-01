<script setup>
import {h, computed, ref, onMounted, onBeforeUnmount} from 'vue'
import {store} from '../stores/rules.js'
import {NIcon, NTag, NSwitch, NButton, useMessage, useDialog} from 'naive-ui'
import {
  TrashOutline,
  CreateOutline,
  ShieldCheckmarkOutline,
  CloseCircleOutline,
} from '@vicons/ionicons5'

const emit = defineEmits(['edit'])
const message = useMessage()
const dialog = useDialog()

// ============ 列宽拖拽 ============
const colWidths = ref({
  enabled: 60,
  name: 240,
  action: 70,
  protocol: 70,
  localAddr: 150,
  remoteAddr: 150,
  localPort: 90,
  remotePort: 90,
  actions: 80,
})

let dragState = null

function onResizeStart(e, colKey) {
  e.preventDefault()
  const startX = e.clientX
  const startWidth = colWidths.value[colKey]
  dragState = {colKey, startX, startWidth}
  document.addEventListener('mousemove', onResizeMove)
  document.addEventListener('mouseup', onResizeEnd)
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
}

function onResizeMove(e) {
  if (!dragState) return
  const diff = e.clientX - dragState.startX
  const newWidth = Math.max(50, dragState.startWidth + diff)
  colWidths.value[dragState.colKey] = newWidth
}

function onResizeEnd() {
  dragState = null
  document.removeEventListener('mousemove', onResizeMove)
  document.removeEventListener('mouseup', onResizeEnd)
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
}

onBeforeUnmount(() => {
  document.removeEventListener('mousemove', onResizeMove)
  document.removeEventListener('mouseup', onResizeEnd)
})

// ============ 列定义 ============
function makeResizeHandle(colKey) {
  return h('div', {
    class: 'resize-handle',
    onMousedown: (e) => onResizeStart(e, colKey),
  })
}

const columns = computed(() => {
  const w = colWidths.value
  return [
    {
      title: () => ['状态', makeResizeHandle('enabled')],
      key: 'enabled',
      width: w.enabled,
      minWidth: 50,
      maxWidth: 80,
      align: 'center',
      render(row) {
        return h(NSwitch, {
          value: row.enabled,
          size: 'small',
          onUpdateValue: () => handleToggle(row),
        })
      },
    },
    {
      title: () => ['规则名称', makeResizeHandle('name')],
      key: 'name',
      width: w.name,
      minWidth: 100,
      ellipsis: {tooltip: true},
      render(row) {
        return h('span', {style: {color: '#ddd'}}, row.name)
      },
    },
    {
      title: () => ['动作', makeResizeHandle('action')],
      key: 'action',
      width: w.action,
      minWidth: 60,
      maxWidth: 90,
      align: 'center',
      render(row) {
        return h(NTag, {
          type: row.action === 'allow' ? 'success' : 'error',
          size: 'tiny',
          bordered: false,
        }, {default: () => row.action === 'allow' ? '允许' : '阻止'})
      },
    },
    {
      title: () => ['协议', makeResizeHandle('protocol')],
      key: 'protocol',
      width: w.protocol,
      minWidth: 50,
      maxWidth: 100,
      align: 'center',
      render(row) {
        return h('span', {style: {color: '#888', fontSize: '12px'}},
          row.protocol === 'any' ? '任意' : row.protocol)
      },
    },
    {
      title: () => ['本地地址', makeResizeHandle('localAddr')],
      key: 'localAddr',
      width: w.localAddr,
      minWidth: 80,
      ellipsis: {tooltip: true},
      render(row) {
        const val = row.localAddr || 'Any'
        return h('span', {style: {color: '#888', fontSize: '12px'}}, val === 'Any' ? '任意' : val)
      },
    },
    {
      title: () => ['远程地址', makeResizeHandle('remoteAddr')],
      key: 'remoteAddr',
      width: w.remoteAddr,
      minWidth: 80,
      ellipsis: {tooltip: true},
      render(row) {
        const val = row.remoteAddr || 'Any'
        return h('span', {style: {color: '#888', fontSize: '12px'}},
          val === 'Any' ? '任意' : val === 'LocalSubnet' ? '本地子网' : val)
      },
    },
    {
      title: () => ['本地端口', makeResizeHandle('localPort')],
      key: 'localPort',
      width: w.localPort,
      minWidth: 60,
      maxWidth: 140,
      ellipsis: {tooltip: true},
      render(row) {
        return h('span', {style: {color: '#888', fontSize: '12px'}}, row.localPort || '-')
      },
    },
    {
      title: () => ['远程端口', makeResizeHandle('remotePort')],
      key: 'remotePort',
      width: w.remotePort,
      minWidth: 60,
      maxWidth: 140,
      ellipsis: {tooltip: true},
      render(row) {
        return h('span', {style: {color: '#888', fontSize: '12px'}}, row.remotePort || '-')
      },
    },
    {
      title: () => ['管理', makeResizeHandle('actions')],
      key: 'actions',
      width: w.actions,
      minWidth: 70,
      maxWidth: 100,
      align: 'center',
      render(row) {
        return h('div', {style: {display: 'flex', gap: '6px', justifyContent: 'center'}}, [
          h(NButton, {text: true, size: 'small', onClick: () => emit('edit', row)}, {
            icon: () => h(NIcon, {size: 16, color: '#4361ee'}, {default: () => h(CreateOutline)}),
          }),
          h(NButton, {text: true, size: 'small', onClick: () => handleDelete(row)}, {
            icon: () => h(NIcon, {size: 16, color: '#e88080'}, {default: () => h(TrashOutline)}),
          }),
        ])
      },
    },
  ]
})

// 计算总宽度用于水平滚动
const scrollX = computed(() => {
  return Object.values(colWidths.value).reduce((sum, w) => sum + w, 0)
})

// ============ 操作 ============
async function handleToggle(rule) {
  try {
    await store.toggleRule(rule.name, !rule.enabled)
    message.success(rule.enabled ? '已禁用' : '已启用')
  } catch (e) {
    message.error('操作失败: ' + String(e))
  }
}

function handleDelete(rule) {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除规则「${rule.name}」吗？此操作不可撤销。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await store.deleteRule(rule.name)
        message.success('规则已删除')
      } catch (e) {
        message.error('删除失败: ' + String(e))
      }
    }
  })
}

const rowClassName = (row) => row.enabled ? '' : 'row-disabled'
</script>

<template>
  <div class="rule-list">
    <!-- 加载中 -->
    <div v-if="store.loading" class="loading-state">
      <n-spin size="medium"/>
      <span>加载规则中...</span>
    </div>

    <!-- 错误提示 -->
    <div v-else-if="store.error" class="error-state">
      <NIcon :size="40" color="#e88080">
        <CloseCircleOutline/>
      </NIcon>
      <span>{{ store.error }}</span>
      <n-button size="small" @click="() => store.fetchRules()">重试</n-button>
    </div>

    <!-- 空状态 -->
    <div v-else-if="store.filteredRules.length === 0" class="empty-state">
      <NIcon :size="48" color="#444">
        <ShieldCheckmarkOutline/>
      </NIcon>
      <span>{{ store.searchQuery ? '没有匹配的规则' : '暂无规则' }}</span>
    </div>

    <!-- 数据表格 -->
    <div v-else class="table-wrapper">
      <n-data-table
        :columns="columns"
        :data="store.filteredRules"
        :row-class-name="rowClassName"
        :scroll-x="scrollX"
        :max-height="500"
        :scrollbar-props="{size: 6}"
        :bordered="false"
        size="small"
        striped
      />
    </div>

    <!-- 底部统计 -->
    <div v-if="!store.loading && !store.error" class="list-footer">
      <span>共 {{ store.filteredRules.length }} 条规则</span>
      <span v-if="store.partial" class="partial-hint">（仅显示前 10 条）</span>
      <span v-if="store.searchQuery || store.actionFilter !== 'all' || store.statusFilter !== 'all'">（已过滤）</span>
      <n-button
        v-if="store.partial"
        text
        type="primary"
        size="tiny"
        style="margin-left: auto"
        :loading="store.loading"
        @click="store.fetchAllRules()"
      >
        加载全部规则
      </n-button>
      <span v-else style="margin-left: auto; color: #555; font-size: 11px;">已加载全部</span>
    </div>
  </div>
</template>

<style scoped>
.rule-list {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.table-wrapper {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

/* 禁用行半透明 */
:deep(.row-disabled td) {
  opacity: 0.45;
}

/* 表格深色主题 */
:deep(.n-data-table) {
  --n-th-color: #1a1a22;
  --n-td-color: transparent;
  --n-border-color: rgba(255, 255, 255, 0.06);
  --n-th-text-color: #888;
  --n-td-text-color: #ccc;
}

:deep(.n-data-table .n-data-table-th) {
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.5px;
  text-transform: uppercase;
  position: relative;
  padding-right: 8px !important;
}

:deep(.n-data-table .n-data-table-td) {
  font-size: 13px;
}

:deep(.n-data-table .n-data-table-tr:hover .n-data-table-td) {
  background: rgba(255, 255, 255, 0.03) !important;
}

:deep(.n-data-table .n-data-table-tr--striped .n-data-table-td) {
  background: rgba(255, 255, 255, 0.015);
}

/* 列宽拖拽手柄 */
.resize-handle {
  position: absolute;
  right: 0;
  top: 0;
  bottom: 0;
  width: 6px;
  cursor: col-resize;
  background: transparent;
  transition: background 0.15s;
  z-index: 1;
}

.resize-handle:hover,
.resize-handle:active {
  background: #4361ee;
}

.loading-state,
.error-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 60px 0;
  color: #666;
  font-size: 14px;
}

.list-footer {
  padding: 10px 16px;
  font-size: 12px;
  color: #555;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  background: #1a1a22;
  border-radius: 0 0 8px 8px;
  display: flex;
  align-items: center;
}
</style>
