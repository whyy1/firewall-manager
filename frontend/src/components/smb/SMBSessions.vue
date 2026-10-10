<script setup>
import {computed, h} from 'vue'
import {NButton, NTag} from 'naive-ui'

const props = defineProps({
  sessions: {type: Array, default: () => []},
  loading: {type: Boolean, default: false},
  loaded: {type: Boolean, default: false},
  error: {type: String, default: ''},
  isAdmin: {type: Boolean, default: false},
  busy: {type: Boolean, default: false},
})
const emit = defineEmits(['refresh', 'disconnect'])

function duration(seconds) {
  const value = Number(seconds) || 0
  if (value < 60) return `${value} 秒`
  if (value < 3600) return `${Math.floor(value / 60)} 分钟`
  return `${Math.floor(value / 3600)} 小时`
}

const columns = computed(() => [
  {title: '客户端', key: 'clientComputerName', minWidth: 160, ellipsis: {tooltip: true}},
  {title: '用户', key: 'clientUserName', minWidth: 190, ellipsis: {tooltip: true}},
  {title: '协议', key: 'dialect', width: 90, render: (row) => h(NTag, {size: 'small', bordered: false, type: 'info'}, {default: () => `SMB ${row.dialect || '未知'}`})},
  {title: '打开文件', key: 'numOpens', width: 90, align: 'center'},
  {title: '已连接', key: 'secondsExists', width: 100, render: (row) => duration(row.secondsExists)},
  {title: '空闲', key: 'secondsIdle', width: 100, render: (row) => duration(row.secondsIdle)},
  {
    title: '操作', key: 'disconnect', fixed: 'right', width: 90, align: 'center',
    render: (row) => h(NButton, {
      text: true, type: 'error', size: 'small',
      'aria-label': `断开 ${row.clientComputerName || row.clientUserName} 的会话`,
      disabled: !props.isAdmin || props.busy || props.loading || Boolean(props.error),
      onClick: () => emit('disconnect', row),
    }, {default: () => '断开'}),
  },
])
</script>

<template>
  <section class="sessions-panel">
    <div class="sessions-header">
      <div><h3>活动 SMB 会话</h3><p>这些设备当前正在访问本机共享</p></div>
      <n-button size="small" :loading="loading" :disabled="loading || busy" @click="emit('refresh')">刷新</n-button>
    </div>
    <n-alert v-if="error" type="error" :bordered="false">{{ error }}{{ sessions.length ? '；下方为最后成功读取的数据，断开操作已暂停。' : '' }}</n-alert>
    <div v-if="loading && sessions.length === 0" class="center-state"><n-spin/><span>正在读取活动会话...</span></div>
    <n-empty v-else-if="!error && loaded && sessions.length === 0" description="当前没有设备连接本机共享"/>
    <n-data-table v-if="sessions.length" :columns="columns" :data="sessions" :bordered="false"
      :scroll-x="850"  size="small" striped/>
  </section>
</template>

<style scoped>
.sessions-panel { display: flex; flex-direction: column; gap: 14px; }
.sessions-header { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 10px; }
h3 { margin: 0; color: var(--color-text); font-size: 15px; }
p { margin: 3px 0 0; color: var(--color-muted); font-size: 12px; }
.center-state { min-height: 220px; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; color: var(--color-muted); }
</style>
