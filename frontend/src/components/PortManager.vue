<script setup>
import {h, onMounted, computed, ref} from 'vue'
import {portStore} from '../stores/ports.js'
import {validatePortChange} from '../stores/port_state.js'
import {NTag, NButton, useMessage} from 'naive-ui'

const props = defineProps({isAdmin: {type: Boolean, default: false}})
const message = useMessage()
const showPortDialog = ref(false)
const editingService = ref(null)
const newPort = ref(null)
const formError = ref('')
const changing = computed(() => Boolean(portStore.saving))
const unavailable = computed(() => !props.isAdmin || portStore.stale || portStore.loading || changing.value)
onMounted(() => portStore.fetchPorts())

function openPortDialog(row) {
  if (unavailable.value || !row.canChangePort) return
  editingService.value = {...row}
  newPort.value = row.listenPort || row.defaultPort
  formError.value = ''
  showPortDialog.value = true
}
async function handlePortChange() {
  if (changing.value || unavailable.value) return false
  formError.value = validatePortChange(editingService.value, newPort.value) || ''
  if (formError.value) return false
  const service = {...editingService.value}
  const port = newPort.value
  try {
    const result = await portStore.changePort(service, port)
    message.success('端口配置操作已完成，请确认服务和连接状态')
    if (result?.refreshError) message.warning('配置操作已成功，但刷新失败。请刷新确认状态，不要重复修改: ' + result.refreshError)
    showPortDialog.value = false
  } catch (error) {
    formError.value = '修改失败: ' + String(error)
  }
  return false
}
const columns = computed(() => [
  {title: '服务名称', key: 'name', width: 185, fixed: 'left', ellipsis: {tooltip: true}},
  {title: '运行状态', key: 'running', width: 95, render: row => h(NTag, {size: 'small', bordered: false, type: row.running ? 'success' : 'default'}, {default: () => row.running ? '运行中' : '未运行'})},
  {title: '检测到的端口', key: 'listenPort', width: 125, render: row => row.listenPort ? h(NTag, {size: 'small', bordered: false, type: 'info'}, {default: () => String(row.listenPort)}) : h('span', {class: 'muted'}, '未检测到')},
  {title: '默认端口', key: 'defaultPort', width: 95},
  {title: '协议', key: 'protocol', width: 80, render: row => String(row.protocol).toUpperCase()},
  {title: '说明', key: 'description', minWidth: 200, ellipsis: {tooltip: true}},
  {title: '操作', key: 'operation', width: 120, fixed: 'right', render: row => row.canChangePort
    ? h(NButton, {size: 'small', text: true, type: 'primary', disabled: unavailable.value, title: !props.isAdmin ? '需要管理员权限' : portStore.stale ? '请先刷新状态' : '修改此服务的端口', 'aria-label': '修改' + row.name + '端口', onClick: () => openPortDialog(row)}, {default: () => '修改端口'})
    : h('span', {class: 'muted', title: row.serviceName === 'LanmanServer' ? 'Windows SMB TCP 监听端口固定为 445，不支持直接修改' : '此服务暂不支持端口修改'}, '仅查看')},
])
</script>

<template>
  <div class="console-page">
    <div class="page-heading">
      <div><h2>常用端口</h2><p>查看本机服务状态与端口参考信息，支持的服务可修改配置。</p></div>
      <div class="page-actions"><n-button size="small" :loading="portStore.loading" :disabled="portStore.loading || changing" @click="portStore.fetchPorts()">刷新端口</n-button></div>
    </div>
    <n-alert type="info" :bordered="false">检测到的端口仅供参考，不能替代实际连接验证。“未检测到”不代表服务没有监听端口。Windows SMB TCP 端口固定为 445，仅供查看。</n-alert>
    <n-alert v-if="portStore.error" type="error" :title="portStore.loaded ? '刷新失败，以下为上次结果' : '端口读取失败'"><div class="diagnostic">{{ portStore.error }}</div><n-button size="small" :disabled="portStore.loading || changing" @click="portStore.fetchPorts()">重试</n-button></n-alert>
    <section class="console-panel">
      <div v-if="portStore.loading && !portStore.loaded" class="state-panel"><n-spin/><span>正在读取端口信息…</span></div>
      <n-empty v-else-if="portStore.loaded && !portStore.ports.length && !portStore.error" description="没有可显示的服务"/>
      <n-data-table v-else-if="portStore.ports.length" :columns="columns" :data="portStore.ports" :loading="portStore.loading" :row-key="row => row.serviceName" :scroll-x="900" :bordered="false" size="small" striped/>
      <div v-else-if="portStore.error" class="state-panel">请重试读取服务状态</div>
    </section>
    <n-modal v-model:show="showPortDialog" preset="dialog" title="修改服务端口" positive-text="确认修改" negative-text="取消"
      :positive-button-props="{loading: changing, disabled: unavailable}"
      :negative-button-props="{disabled: changing}" :closable="!changing" :mask-closable="!changing" :close-on-esc="!changing" @positive-click="handlePortChange">
      <div v-if="editingService" class="port-dialog">
        <dl class="port-details"><dt>服务</dt><dd>{{ editingService.name }}</dd><dt>检测到的端口</dt><dd>{{ editingService.listenPort || '未检测到' }}</dd><dt>默认端口</dt><dd>{{ editingService.defaultPort }}</dd></dl>
        <n-form-item label="新端口" :validation-status="formError ? 'error' : undefined" :feedback="formError">
          <n-input-number v-model:value="newPort" :min="1" :max="65535" :disabled="changing" :input-props="{'aria-label': '新端口'}" placeholder="1–65535 的整数"/>
        </n-form-item>
        <n-alert type="warning" :bordered="false">确认后将写入服务配置并尝试重启服务，当前连接可能中断。请准备好使用新端口重新连接。</n-alert>
      </div>
    </n-modal>
  </div>
</template>

<style scoped>
.port-dialog { display: flex; flex-direction: column; gap: 14px; padding-top: 12px; }
.port-details { display: grid; grid-template-columns: 115px 1fr; gap: 8px; font-size: 13px; color: var(--color-text); }
dt { color: var(--color-muted); }
dd { overflow-wrap: anywhere; }
</style>
