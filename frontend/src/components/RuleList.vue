<script setup>
import {h, computed} from 'vue'
import {store} from '../stores/rules.js'
import {NIcon, NTag, NSwitch, NButton, NTooltip, useMessage, useDialog} from 'naive-ui'
import {
  TrashOutline,
  CreateOutline,
  ShieldCheckmarkOutline,
  CloseCircleOutline,
} from '@vicons/ionicons5'

const emit = defineEmits(['edit'])
const message = useMessage()
const dialog = useDialog()

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

const columns = computed(() => [
  {
    title: '状态',
    key: 'enabled',
    width: 60,
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
    title: '规则名称',
    key: 'name',
    width: 200,
    minWidth: 100,
    ellipsis: {tooltip: true},
    render(row) {
      return h('span', {style: {color: '#ddd'}}, row.name)
    },
  },
  {
    title: '动作',
    key: 'action',
    width: 70,
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
    title: '程序',
    key: 'program',
    width: 280,
    minWidth: 100,
    ellipsis: {tooltip: true},
    render(row) {
      const text = !row.program || row.program === 'Any' ? '-' : row.program
      return h('span', {
        style: {fontFamily: '"Cascadia Code","Consolas",monospace', fontSize: '12px', color: '#999'},
      }, text)
    },
  },
  {
    title: '协议',
    key: 'protocol',
    width: 70,
    minWidth: 50,
    maxWidth: 100,
    align: 'center',
    render(row) {
      return h('span', {style: {color: '#888', fontSize: '12px'}},
        row.protocol === 'any' ? '任意' : row.protocol)
    },
  },
  {
    title: '本地地址',
    key: 'localAddr',
    width: 140,
    minWidth: 80,
    ellipsis: {tooltip: true},
    render(row) {
      const val = row.localAddr || 'Any'
      return h('span', {style: {color: '#888', fontSize: '12px'}}, val === 'Any' ? '任意' : val)
    },
  },
  {
    title: '远程地址',
    key: 'remoteAddr',
    width: 140,
    minWidth: 80,
    ellipsis: {tooltip: true},
    render(row) {
      const val = row.remoteAddr || 'Any'
      return h('span', {style: {color: '#888', fontSize: '12px'}}, val === 'Any' ? '任意' : val === 'LocalSubnet' ? '本地子网' : val)
    },
  },
  {
    title: '本地端口',
    key: 'localPort',
    width: 90,
    minWidth: 60,
    maxWidth: 140,
    ellipsis: {tooltip: true},
    render(row) {
      return h('span', {style: {color: '#888', fontSize: '12px'}}, row.localPort || '-')
    },
  },
  {
    title: '远程端口',
    key: 'remotePort',
    width: 90,
    minWidth: 60,
    maxWidth: 140,
    ellipsis: {tooltip: true},
    render(row) {
      return h('span', {style: {color: '#888', fontSize: '12px'}}, row.remotePort || '-')
    },
  },
  {
    title: '管理',
    key: 'actions',
    width: 80,
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
])

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
      <n-button size="small" @click="store.fetchRules">重试</n-button>
    </div>

    <!-- 空状态 -->
    <div v-else-if="store.filteredRules.length === 0" class="empty-state">
      <NIcon :size="48" color="#444">
        <ShieldCheckmarkOutline/>
      </NIcon>
      <span>{{ store.searchQuery ? '没有匹配的规则' : '暂无规则' }}</span>
    </div>

    <!-- 数据表格 -->
    <n-data-table
      v-else
      :columns="columns"
      :data="store.filteredRules"
      :row-class-name="rowClassName"
      :max-height="600"
      :scrollbar-props="{size: 6}"
      :bordered="false"
      size="small"
      striped
    />

    <!-- 底部统计 -->
    <div v-if="!store.loading && !store.error" class="list-footer">
      共 {{ store.filteredRules.length }} 条规则
      <span v-if="store.searchQuery || store.actionFilter !== 'all' || store.statusFilter !== 'all'">（已过滤）</span>
      <span style="margin-left: 12px; color: #555;">← 拖拽列边框调整宽度</span>
    </div>
  </div>
</template>

<style scoped>
.rule-list {
  display: flex;
  flex-direction: column;
  height: 100%;
}

/* 禁用行半透明 */
:deep(.row-disabled td) {
  opacity: 0.45;
}

/* 表格深色主题适配 */
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

/* 列拖拽手柄 */
:deep(.n-data-table .n-data-table-th__resize-handle) {
  background: rgba(255, 255, 255, 0.15);
}

:deep(.n-data-table .n-data-table-th__resize-handle:hover) {
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
