<script setup>
import {computed, h} from 'vue'
import {NButton, NIcon, NTag} from 'naive-ui'
import {OpenOutline} from '@vicons/ionicons5'

const props = defineProps({
  shares: {type: Array, default: () => []},
  loading: {type: Boolean, default: false},
  loaded: {type: Boolean, default: false},
  error: {type: String, default: ''},
  openingLocation: {type: String, default: ''},
})
const emit = defineEmits(['scan', 'open'])

const columns = computed(() => [
  {title: '设备', key: 'server', width: 150, render: (row) => h(NTag, {size: 'small', bordered: false, type: 'info'}, {default: () => row.server})},
  {title: '共享名称', key: 'name', width: 160, ellipsis: {tooltip: true}},
  {title: '网络地址', key: 'path', minWidth: 260, ellipsis: {tooltip: true}},
  {title: '说明', key: 'comment', minWidth: 180, ellipsis: {tooltip: true}, render: (row) => row.comment || '—'},
  {
    title: '操作', key: 'open', fixed: 'right', width: 100, align: 'center',
    render: (row) => h(NButton, {
      text: true, type: 'primary', size: 'small',
      disabled: !row.path || props.loading || Boolean(props.error) || Boolean(props.openingLocation) && props.openingLocation !== row.path,
      loading: props.openingLocation === row.path,
      onClick: () => emit('open', row.path),
    }, {default: () => '打开', icon: () => h(NIcon, null, {default: () => h(OpenOutline)})}),
  },
])
</script>

<template>
  <section class="network-panel">
    <div class="network-header">
      <div>
        <h3>已连接的局域网共享</h3>
        <p>检测 Windows 当前已经连接或映射的网络磁盘共享，不会读取共享内的文件</p>
      </div>
      <n-button type="primary" size="small" :loading="loading" :disabled="loading" @click="emit('scan')">
        {{ loaded ? '重新检测' : '检测已连接共享' }}
      </n-button>
    </div>
    <n-alert type="info" :bordered="false">
      这里只显示本机当前已连接的共享。未连接的 NAS 或其他电脑不会被主动探测，因此空结果不代表局域网没有共享。
    </n-alert>
    <n-alert v-if="error" type="error" :bordered="false">{{ error }}{{ shares.length ? '；下方为最后成功读取的数据，打开操作已暂停。' : '' }}</n-alert>
    <div v-if="loading && shares.length === 0" class="center-state"><n-spin/><span>正在检测已连接共享...</span></div>
    <n-empty v-else-if="!error && shares.length === 0" :description="loaded ? '检测完成，当前没有已连接的网络共享' : '点击“检测已连接共享”开始检查'"/>
    <n-data-table v-if="shares.length" :columns="columns" :data="shares" :bordered="false"
      :scroll-x="850"  size="small" striped/>
  </section>
</template>

<style scoped>
.network-panel { display: flex; flex-direction: column; gap: 14px; }
.network-header { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 10px; }
h3 { margin: 0; color: var(--color-text); font-size: 15px; }
p { margin: 3px 0 0; color: var(--color-muted); font-size: 12px; }
.center-state { min-height: 220px; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; color: var(--color-muted); }
</style>
