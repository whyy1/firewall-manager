<script setup>
import {h, computed, ref, watch, onMounted, onBeforeUnmount} from 'vue'
import {store} from '../stores/rules.js'
import {NTag, NSwitch, NButton, NTooltip, useMessage, useDialog} from 'naive-ui'
import {OpenExplorer} from '../../wailsjs/go/main/App'
import {normalizeAny} from '../stores/rule_form.js'
import {ruleIdentity, ruleReadOnlyReason} from '../stores/rule_identity.js'

const props = defineProps({isDark: Boolean, isAdmin: Boolean})
const emit = defineEmits(['add', 'edit'])
const message = useMessage(), dialog = useDialog()
const tableHost = ref(null), page = ref(1), confirmationOpen = ref(false)
const widths = ref({enabled: 68, name: 220, program: 160, action: 74, protocol: 86, localAddr: 160, remoteAddr: 160, localPort: 110, remotePort: 110, actions: 112})
const keys = Object.keys(widths.value)
let drag = null
const previousStyle = {cursor: '', userSelect: ''}
const blocked = computed(() => !props.isAdmin || !store.canWrite || confirmationOpen.value)
const protocolOptions = computed(() => [{label: '所有协议', value: 'all'}, ...[...new Set(store.rules.map(rule => rule.protocol))].filter(Boolean).map(value => ({label: value, value}))])
const actionOptions = [{label: '所有动作', value: 'all'}, {label: '允许', value: 'allow'}, {label: '阻止', value: 'block'}]
const statusOptions = [{label: '所有状态', value: 'all'}, {label: '已启用', value: 'enabled'}, {label: '已禁用', value: 'disabled'}]
const scrollX = computed(() => Object.values(widths.value).reduce((sum, value) => sum + value, 0))
const pagination = computed(() => ({page: page.value, pageSize: 50, showSizePicker: false, onChange: value => { page.value = value }}))
watch(() => [store.direction, store.nameQuery, store.portQuery, store.addrQuery, store.protocolFilter, store.actionFilter, store.statusFilter], () => { page.value = 1 })
watch(() => store.filteredRules.length, length => { page.value = Math.min(page.value, Math.max(1, Math.ceil(length / 50))) })
const displayAny = value => normalizeAny(value) || '任意'
function actionButton(label, callback, danger = false, reason = '') {
  return h(NTooltip, {}, {
    trigger: () => h(NButton, {text: true, size: 'small', type: danger ? 'error' : 'primary', disabled: blocked.value || !!reason, 'aria-label': label, onClick: callback}, {default: () => label}),
    default: () => !props.isAdmin ? '需要管理员权限' : reason || label,
  })
}
const columns = computed(() => [
  {title: '状态', key: 'enabled', width: widths.value.enabled, fixed: 'left', render: row => h(NSwitch, {value: row.enabled, size: 'small', disabled: blocked.value || !!ruleReadOnlyReason(row), 'aria-disabled': blocked.value || !!ruleReadOnlyReason(row), 'aria-label': `${row.enabled ? '禁用' : '启用'}规则 ${row.name}`, onUpdateValue: () => handleToggle(row)})},
  {title: '规则名称', key: 'name', width: widths.value.name, render: row => h(NTooltip, {width: 360}, {
    trigger: () => h('div', {class: 'rule-name', tabindex: 0}, [h('span', row.name), ...(ruleReadOnlyReason(row) ? [h(NTag, {size: 'small', bordered: false}, {default: () => '只读'})] : [])]),
    default: () => h('div', {style: {overflowWrap: 'anywhere', whiteSpace: 'normal'}}, [h('div', row.name), h('div', '内部标识：' + row.id), h('div', '策略来源：' + (row.policyStoreSource || '未确认')), h('div', '来源类型：' + (row.policyStoreSourceType || '未确认')), ...(ruleReadOnlyReason(row) ? [h('div', '只读原因：' + ruleReadOnlyReason(row))] : [])]),
  })},
  {title: '程序', key: 'program', width: widths.value.program, ellipsis: {tooltip: true}, render: row => normalizeAny(row.program) ? h(NButton, {text: true, type: 'primary', 'aria-label': `打开程序所在目录 ${row.program}`, onClick: () => openProgram(row.program)}, {default: () => row.program.split('\\').pop()}) : '任意程序'},
  {title: '动作', key: 'action', width: widths.value.action, render: row => h(NTag, {size: 'small', bordered: false, type: row.action === 'allow' ? 'success' : 'error'}, {default: () => row.action === 'allow' ? '允许' : row.action === 'block' ? '阻止' : row.action})},
  ...[['protocol', '协议'], ['localAddr', '本地地址'], ['remoteAddr', '远程地址'], ['localPort', '本地端口'], ['remotePort', '远程端口']].map(([key, title]) => ({title, key, width: widths.value[key], ellipsis: {tooltip: true}, render: row => displayAny(row[key])})),
  {title: '管理', key: 'actions', width: widths.value.actions, fixed: 'right', render: row => h('div', {class: 'row-actions'}, [actionButton('编辑', () => emit('edit', row)), actionButton('删除', () => handleDelete(row), true, ruleReadOnlyReason(row))])},
])
function report(result, success) {
  message.success(success)
  if (result?.refreshError) message.warning('操作已成功，但刷新失败，请刷新确认最新状态：' + result.refreshError)
}
async function handleToggle(rule) {
  if (blocked.value || ruleReadOnlyReason(rule)) return
  try { report(await store.toggleRule(rule, !rule.enabled), rule.enabled ? '规则已禁用' : '规则已启用') }
  catch (error) { message.error('操作失败：' + String(error)) }
}
function handleDelete(rule) {
  if (blocked.value || ruleReadOnlyReason(rule)) return
  confirmationOpen.value = true
  let executing = false
  const confirmation = dialog.warning({
    title: '确认删除规则', content: () => h('div', {style: {overflowWrap: 'anywhere', whiteSpace: 'normal'}}, [h('p', `删除${rule.direction === 'in' ? '入站' : '出站'}规则「${rule.name}」后将无法撤销。只删除所选唯一规则。`), h('p', '内部标识：' + rule.id), h('p', '策略来源：' + rule.policyStoreSource + '（' + rule.policyStoreSourceType + '）')]),
    positiveText: '删除', negativeText: '取消',
    onAfterLeave: () => { confirmationOpen.value = false },
    onClose: () => !executing,
    onNegativeClick: () => !executing,
    onPositiveClick: async () => {
      if (executing) return false
      executing = true
      confirmation.loading = true
      confirmation.positiveButtonProps = {disabled: true}
      confirmation.negativeButtonProps = {disabled: true}
      confirmation.closable = false
      confirmation.maskClosable = false
      confirmation.closeOnEsc = false
      try { report(await store.deleteRule(rule), '规则已删除') }
      catch (error) { message.error('删除失败：' + String(error)); return false }
      finally {
        executing = false
        confirmation.loading = false
        confirmation.positiveButtonProps = {disabled: false}
        confirmation.negativeButtonProps = {disabled: false}
        confirmation.closable = true
        confirmation.maskClosable = true
        confirmation.closeOnEsc = true
      }
    },
  })
}
async function openProgram(path) {
  try { await OpenExplorer(path) } catch (error) { message.error('打开失败：' + String(error)) }
}
function edge(event) {
  const headers = [...(tableHost.value?.querySelectorAll('.n-data-table-th') || [])]
  return headers.findIndex(header => {
    const rect = header.getBoundingClientRect()
    return event.clientY >= rect.top && event.clientY <= rect.bottom && event.clientX >= rect.right - 6 && event.clientX <= rect.right + 2
  })
}
function beginResize(event) {
  const index = edge(event)
  if (index < 0 || index >= keys.length) return
  event.preventDefault()
  drag = {key: keys[index], start: event.clientX, width: widths.value[keys[index]]}
  previousStyle.cursor = document.body.style.cursor; previousStyle.userSelect = document.body.style.userSelect
  document.body.style.cursor = 'col-resize'; document.body.style.userSelect = 'none'
  document.addEventListener('mousemove', moveResize); document.addEventListener('mouseup', endResize)
}
function moveResize(event) {
  if (drag) widths.value = {...widths.value, [drag.key]: Math.max(60, drag.width + event.clientX - drag.start)}
}
function endResize() {
  if (drag) { document.body.style.cursor = previousStyle.cursor; document.body.style.userSelect = previousStyle.userSelect }
  drag = null
  document.removeEventListener('mousemove', moveResize); document.removeEventListener('mouseup', endResize)
}
onMounted(() => { if (!store.loaded && !store.loading) store.fetchRules() })
onBeforeUnmount(endResize)
</script>

<template>
  <section class="rule-list">
    <div class="page-heading"><div><h1>防火墙规则</h1><p>查看并管理当前方向的全部规则</p></div><div class="heading-actions"><n-button :loading="store.loading" :disabled="!!store.saving" @click="store.fetchRules()">刷新</n-button><n-button type="primary" :disabled="blocked" @click="emit('add')">添加规则</n-button></div></div>
    <n-tabs :value="store.direction" type="segment" :disabled="!!store.saving" @update:value="store.setDirection($event)"><n-tab name="in">入站规则</n-tab><n-tab name="out">出站规则</n-tab></n-tabs>
    <div class="rule-filters">
      <n-input v-model:value="store.nameQuery" aria-label="搜索规则名称" placeholder="搜索规则名称" clearable />
      <n-input v-model:value="store.portQuery" aria-label="搜索端口" placeholder="搜索本地或远程端口" clearable />
      <n-input v-model:value="store.addrQuery" aria-label="搜索地址" placeholder="搜索本地或远程地址" clearable />
      <n-select v-model:value="store.protocolFilter" aria-label="协议筛选" :options="protocolOptions" />
      <n-select v-model:value="store.actionFilter" aria-label="动作筛选" :options="actionOptions" />
      <n-select v-model:value="store.statusFilter" aria-label="状态筛选" :options="statusOptions" />
      <n-button :disabled="!store.hasAnyFilter" @click="store.clearFilters()">清空筛选</n-button>
    </div>

    <n-alert v-if="store.error" type="error" :bordered="false">{{ store.loaded ? '刷新失败，显示上次成功数据；已暂停修改。' : '加载规则失败。' }} {{ store.error }}</n-alert>
    <div class="rule-stats"><template v-if="store.loaded"><span>共 {{ store.rules.length }} 条</span><span>已启用 {{ store.enabledCount }} 条</span><span>匹配 {{ store.filteredRules.length }} 条</span></template><span v-else>尚未读取规则</span><span v-if="store.loading">正在{{ store.loaded ? '刷新' : '加载' }}…</span><span v-if="store.saving">正在执行规则操作…</span></div>
    <div v-if="!store.loaded && store.loading" class="list-state"><n-spin /><span>加载规则中…</span></div>
    <div v-else ref="tableHost" class="table-host" @mousedown="beginResize">
      <n-data-table :row-key="ruleIdentity" :columns="columns" :data="store.filteredRules" :scroll-x="scrollX" :pagination="pagination" :bordered="false" size="small" striped :row-class-name="row => row.enabled ? '' : 'row-disabled'">
        <template #empty>{{ store.error ? '读取失败，请点击刷新重试' : store.hasAnyFilter ? '没有匹配的规则，请调整或清空筛选' : '当前方向暂无规则' }}</template>
      </n-data-table>
    </div>
  </section>
</template>

<style scoped>
.rule-list { display: flex; flex-direction: column; gap: 14px; min-width: 0; color: var(--color-text); }
.page-heading { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
h1 { font-size: 22px; margin: 0 0 4px; }
p { margin: 0; color: var(--color-muted); font-size: 13px; }
.heading-actions, .rule-stats { display: flex; align-items: center; gap: 12px; }
.rule-filters { display: grid; grid-template-columns: repeat(3, minmax(150px, 1fr)); gap: 10px; }
.rule-filters > :last-child { justify-self: start; }
.rule-stats { color: var(--color-muted); font-size: 12px; flex-wrap: wrap; }
.table-host { min-width: 0; background: var(--color-surface); border: 1px solid var(--color-border); border-radius: 8px; overflow: hidden; }
.list-state { display: flex; align-items: center; justify-content: center; gap: 12px; min-height: 180px; }
:deep(.row-actions) { display: flex; gap: 12px; }
:deep(.rule-name) { display: flex; align-items: center; gap: 6px; min-width: 0; }
:deep(.rule-name > span:first-child) { overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
:deep(.row-disabled td) { color: var(--color-muted); }
:deep(.n-data-table-th) { position: relative; }
:deep(.n-data-table-th)::after { content: ''; position: absolute; right: 0; top: 4px; bottom: 4px; width: 5px; cursor: col-resize; }
:deep(.n-data-table-th):hover::after { background: var(--color-primary); }
@media (max-width: 900px) { .rule-filters { grid-template-columns: repeat(2, minmax(120px, 1fr)); } .page-heading { flex-wrap: wrap; } }
</style>
