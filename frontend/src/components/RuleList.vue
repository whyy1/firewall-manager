<script setup>
import {h, computed, ref, onMounted, onBeforeUnmount} from 'vue'
import {store} from '../stores/rules.js'
import {NIcon, NTag, NSwitch, NButton, NTooltip, useMessage, useDialog} from 'naive-ui'
import {
  TrashOutline,
  CreateOutline,
  ShieldCheckmarkOutline,
  CloseCircleOutline,
} from '@vicons/ionicons5'
import {OpenExplorer} from '../../wailsjs/go/main/App'

const props = defineProps({ isDark: { type: Boolean, default: true } })
const emit = defineEmits(['edit'])
const message = useMessage()
const dialog = useDialog()

// ============ 列宽拖拽（DOM 直接操作 + rAF 节流，避免拖拽期间 Vue 重渲染） ============
const colWidths = ref({
  enabled: 60, name: 200, program: 150, action: 70, protocol: 80,
  localAddr: 140, remoteAddr: 140, localPort: 90, remotePort: 90, actions: 80,
})
const colKeys = ['enabled','name','program','action','protocol','localAddr','remoteAddr','localPort','remotePort','actions']
let dragState = null
let rafId = 0

function getColEdgeIndex(x) {
  const table = document.querySelector('.n-data-table')
  if (!table) return -1
  const ths = table.querySelectorAll('.n-data-table-th')
  for (let i = 0; i < ths.length && i < colKeys.length; i++) {
    const rect = ths[i].getBoundingClientRect()
    if (x >= rect.right - 8 && x <= rect.right + 4) return i
  }
  return -1
}

function onGlobalMouseDown(e) {
  const idx = getColEdgeIndex(e.clientX)
  if (idx < 0) return
  e.preventDefault()
  e.stopPropagation()
  // 拖拽开始时一次性缓存 DOM 引用，避免每帧重复查询
  const table = document.querySelector('.n-data-table')
  const th = table ? table.querySelectorAll('.n-data-table-th')[idx] : null
  const tds = table
    ? Array.from(table.querySelectorAll('.n-data-table-tr'))
        .map(tr => tr.querySelectorAll('td')[idx])
        .filter(Boolean)
    : []
  dragState = { idx, startX: e.clientX, startW: colWidths.value[colKeys[idx]], pendingW: 0, th, tds }
  document.addEventListener('mousemove', onGlobalMouseMove)
  document.addEventListener('mouseup', onGlobalMouseUp)
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
}

function onGlobalMouseMove(e) {
  if (!dragState) return
  const newW = Math.max(50, dragState.startW + e.clientX - dragState.startX)
  dragState.pendingW = newW
  // 用 rAF 节流：拖拽期间只改 DOM 宽度，不触发 Vue 响应式
  if (!rafId) {
    rafId = requestAnimationFrame(() => {
      rafId = 0
      if (!dragState) return
      const w = dragState.pendingW + 'px'
      // 直接使用缓存的 DOM 引用，避免每帧重复查询
      if (dragState.th) {
        dragState.th.style.width = w
        dragState.th.style.minWidth = w
        dragState.th.style.maxWidth = w
      }
      dragState.tds.forEach(td => {
        td.style.width = w
        td.style.minWidth = w
        td.style.maxWidth = w
      })
    })
  }
}

function onGlobalMouseUp() {
  if (dragState) {
    // mouseup 时才同步回 Vue 状态（触发一次重渲染）
    colWidths.value[colKeys[dragState.idx]] = dragState.pendingW || colWidths.value[colKeys[dragState.idx]]
  }
  dragState = null
  if (rafId) { cancelAnimationFrame(rafId); rafId = 0 }
  document.removeEventListener('mousemove', onGlobalMouseMove)
  document.removeEventListener('mouseup', onGlobalMouseUp)
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
}

function onGlobalMouseMoveHover(e) {
  if (dragState) return
  document.body.style.cursor = getColEdgeIndex(e.clientX) >= 0 ? 'col-resize' : ''
}

onMounted(() => {
  document.addEventListener('mousedown', onGlobalMouseDown, true)
  document.addEventListener('mousemove', onGlobalMouseMoveHover, true)
})
onBeforeUnmount(() => {
  document.removeEventListener('mousedown', onGlobalMouseDown, true)
  document.removeEventListener('mousemove', onGlobalMouseMoveHover, true)
  document.removeEventListener('mousemove', onGlobalMouseMove)
  document.removeEventListener('mouseup', onGlobalMouseUp)
})

// ============ 列定义（纯 title 字符串，不注入 div） ============
const columns = computed(() => {
  const w = colWidths.value
  return [
    { title: '状态', key: 'enabled', width: w.enabled, minWidth: 50, maxWidth: 80, align: 'center',
      render: (row) => h(NSwitch, { value: row.enabled, size: 'small', onUpdateValue: () => handleToggle(row) }) },
    { title: '规则名称', key: 'name', width: w.name, minWidth: 100, ellipsis: { tooltip: true },
      render: (row) => h('span', { style: { color: '#ddd' } }, row.name) },
    { title: '程序', key: 'program', width: w.program, minWidth: 100, ellipsis: { tooltip: true },
      render: (row) => {
        if (!row.program) return h('span', { style: { color: '#555' } }, '-')
        const fileName = row.program.split('\\').pop()
        return h(NTooltip, { trigger: 'hover' }, {
          trigger: () => h('span', {
            style: { color: '#4361ee', cursor: 'pointer', fontSize: '12px' },
            onClick: () => handleOpenExplorer(row.program),
          }, fileName),
          default: () => h('div', [
            h('p', { style: { margin: '0 0 4px' } }, row.program),
            h('p', { style: { margin: '0', color: '#999', fontSize: '11px' } }, '点击打开文件所在目录'),
          ])
        })
      }
    },
    { title: '动作', key: 'action', width: w.action, minWidth: 60, maxWidth: 90, align: 'center',
      render: (row) => h(NTag, { type: row.action === 'allow' ? 'success' : 'error', size: 'tiny', bordered: false }, { default: () => row.action === 'allow' ? '允许' : '阻止' }) },
    { title: '协议', key: 'protocol', width: w.protocol, minWidth: 50, maxWidth: 100, align: 'center',
      render: (row) => h('span', { style: { color: '#888', fontSize: '12px' } }, row.protocol === 'any' ? '任意' : row.protocol) },
    { title: '本地地址', key: 'localAddr', width: w.localAddr, minWidth: 80, ellipsis: { tooltip: true },
      render: (row) => { const v = row.localAddr || 'Any'; return h('span', { style: { color: '#888', fontSize: '12px' } }, v === 'Any' ? '任意' : v) } },
    { title: '远程地址', key: 'remoteAddr', width: w.remoteAddr, minWidth: 80, ellipsis: { tooltip: true },
      render: (row) => { const v = row.remoteAddr || 'Any'; return h('span', { style: { color: '#888', fontSize: '12px' } }, v === 'Any' ? '任意' : v === 'LocalSubnet' ? '本地子网' : v) } },
    { title: '本地端口', key: 'localPort', width: w.localPort, minWidth: 60, maxWidth: 140, ellipsis: { tooltip: true },
      render: (row) => h('span', { style: { color: '#888', fontSize: '12px' } }, row.localPort || '-') },
    { title: '远程端口', key: 'remotePort', width: w.remotePort, minWidth: 60, maxWidth: 140, ellipsis: { tooltip: true },
      render: (row) => h('span', { style: { color: '#888', fontSize: '12px' } }, row.remotePort || '-') },
    { title: '管理', key: 'actions', width: w.actions, minWidth: 70, maxWidth: 100, align: 'center',
      render: (row) => h('div', { style: { display: 'flex', gap: '6px', justifyContent: 'center' } }, [
        h(NButton, { text: true, size: 'small', onClick: () => emit('edit', row) }, { icon: () => h(NIcon, { size: 16, color: '#4361ee' }, { default: () => h(CreateOutline) }) }),
        h(NButton, { text: true, size: 'small', onClick: () => handleDelete(row) }, { icon: () => h(NIcon, { size: 16, color: '#e88080' }, { default: () => h(TrashOutline) }) }),
      ]) },
  ]
})

const scrollX = computed(() => Object.values(colWidths.value).reduce((s, v) => s + v, 0))
const rowClassName = (row) => row.enabled ? '' : 'row-disabled'

async function handleToggle(rule) {
  try { await store.toggleRule(rule.name, !rule.enabled); message.success(rule.enabled ? '已禁用' : '已启用') }
  catch (e) { message.error('操作失败: ' + String(e)) }
}

function handleDelete(rule) {
  dialog.warning({
    title: '确认删除', content: `确定要删除规则「${rule.name}」吗？`, positiveText: '删除', negativeText: '取消',
    onPositiveClick: async () => {
      try { await store.deleteRule(rule.name); message.success('规则已删除') }
      catch (e) { message.error('删除失败: ' + String(e)) }
    }
  })
}

async function handleOpenExplorer(path) {
  try { await OpenExplorer(path) }
  catch (e) { message.error('打开失败: ' + String(e)) }
}
</script>

<template>
  <div class="rule-list">
    <div v-if="store.loading" class="loading-state"><n-spin size="medium"/><span>加载规则中...</span></div>
    <div v-else-if="store.error" class="error-state">
      <NIcon :size="40" color="#e88080"><CloseCircleOutline/></NIcon>
      <span>{{ store.error }}</span>
      <n-button size="small" @click="() => store.fetchRules()">重试</n-button>
    </div>
    <div v-else-if="store.filteredRules.length === 0" class="empty-state">
      <NIcon :size="48" color="#444"><ShieldCheckmarkOutline/></NIcon>
      <span>{{ store.hasAnyFilter ? '没有匹配的规则' : '暂无规则' }}</span>
    </div>
    <n-data-table v-else :columns="columns" :data="store.filteredRules" :row-class-name="rowClassName"
      :scroll-x="scrollX" :max-height="500" :scrollbar-props="{size: 6}" :bordered="false" size="small" striped />
    <div v-if="!store.loading && !store.error" class="list-footer" :class="{ 'footer-light': !isDark }">
      <span>共 {{ store.filteredRules.length }} 条规则</span>
      <span v-if="store.partial" style="color:#999">（仅前 10 条）</span>
      <span v-if="store.hasAnyFilter" style="color:#999">（已过滤）</span>
      <n-button v-if="store.partial" text type="primary" size="tiny" style="margin-left:auto" :loading="store.loading" @click="store.fetchAllRules()">加载全部</n-button>
      <span v-else style="margin-left:auto; color:#555; font-size:11px">已加载全部</span>
    </div>
  </div>
</template>

<style scoped>
.rule-list { display: flex; flex-direction: column; height: 100%; overflow: hidden; }
:deep(.row-disabled td) { opacity: 0.45; }
:deep(.n-data-table .n-data-table-th) { position: relative; overflow: visible !important; }
:deep(.n-data-table .n-data-table-th)::after {
  content: ''; position: absolute; right: 0; top: 4px; bottom: 4px; width: 4px;
  cursor: col-resize; border-radius: 2px; transition: background 0.15s;
}
:deep(.n-data-table .n-data-table-th):hover::after { background: rgba(67,97,238,0.5); }
.loading-state, .error-state, .empty-state {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 12px; padding: 60px 0; color: #666; font-size: 14px;
}
.list-footer {
  padding: 10px 16px; font-size: 12px; color: #555;
  border-top: 1px solid rgba(255,255,255,0.06); background: #1a1a22;
  border-radius: 0 0 8px 8px; display: flex; align-items: center; gap: 8px;
}
.list-footer.footer-light {
  background: #fff; color: #666; border-top-color: rgba(0,0,0,0.06);
}
</style>
