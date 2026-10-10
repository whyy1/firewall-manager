<script setup>
import {computed, h} from 'vue'
import {NButton, NIcon, NTag, NTooltip} from 'naive-ui'
import {FolderOpenOutline, OpenOutline, TrashOutline} from '@vicons/ionicons5'

const props = defineProps({
  shares: {type: Array, default: () => []},
  loading: {type: Boolean, default: false},
  loaded: {type: Boolean, default: false},
  error: {type: String, default: ''},
  isAdmin: {type: Boolean, default: false},
  busy: {type: Boolean, default: false},
  openingLocation: {type: String, default: ''},
})
const emit = defineEmits(['open', 'remove'])

function openButton(label, location) {
  return h(NButton, {
    text: true,
    size: 'small',
    'aria-label': label,
    type: 'primary',
    disabled: !location || props.loading || Boolean(props.error) || (Boolean(props.openingLocation) && props.openingLocation !== location),
    loading: props.openingLocation === location,
    onClick: () => emit('open', location),
  }, {default: () => label, icon: () => h(NIcon, null, {default: () => h(OpenOutline)})})
}

const columns = computed(() => [
  {
    title: '共享名称', key: 'name', width: 160, ellipsis: {tooltip: true},
    render: (row) => h('div', {class: 'share-name'}, [
      h(NIcon, {size: 16, color: row.special ? 'var(--color-muted)' : 'var(--color-primary)'}, {default: () => h(FolderOpenOutline)}),
      h('span', {title: row.name}, row.name),
      row.special ? h(NTag, {size: 'tiny', bordered: false}, {default: () => '系统'}) : null,
    ]),
  },
  {title: '本机目录', key: 'path', minWidth: 180, ellipsis: {tooltip: true}, render: (row) => row.path || '—'},
  {title: '网络访问地址', key: 'networkPath', minWidth: 190, ellipsis: {tooltip: true}},
  {
    title: '状态', key: 'state', width: 80, align: 'center',
    render: (row) => h(NTag, {size: 'small', bordered: false, type: String(row.state).toLowerCase() === 'online' ? 'success' : 'default'}, {default: () => ({online: '在线', offline: '离线'})[String(row.state).toLowerCase()] || row.state || '未知'}),
  },
  {title: '说明', key: 'description', minWidth: 130, ellipsis: {tooltip: true}, render: (row) => row.description || '—'},
  {
    title: '操作', key: 'actions', fixed: 'right', width: 245, align: 'center',
    render: (row) => h('div', {class: 'row-actions'}, [
      openButton('打开目录', row.path),
      openButton('打开共享', row.networkPath),
      h(NTooltip, {trigger: 'hover'}, {
        trigger: () => h(NButton, {
          text: true, size: 'small', type: 'error',
          disabled: row.special || !props.isAdmin || props.busy,
          onClick: () => emit('remove', row),
        }, {default: () => '删除', icon: () => h(NIcon, null, {default: () => h(TrashOutline)})}),
        default: () => row.special ? '系统特殊共享不可删除' : props.isAdmin ? '删除共享' : '需要管理员权限',
      }),
    ]),
  },
])
</script>

<template>
  <section class="panel">
    <div class="section-title">
      <div>
        <h3>本机共享目录</h3>
        <p>“本机目录”是磁盘位置；“网络访问地址”供局域网内其他设备访问</p>
      </div>
    </div>
    <n-alert v-if="error" type="error" title="共享列表读取失败">{{ error }}{{ shares.length ? '；下方为最后成功读取的数据，操作已暂停。' : '' }}</n-alert>
    <div v-if="loading && !shares.length" class="loading-state"><n-spin size="small"/>正在读取共享列表...</div>
    <n-empty v-else-if="!error && loaded && shares.length === 0" description="暂无 SMB 共享"/>
    <n-data-table v-if="shares.length" :columns="columns" :data="shares" :bordered="false"
      :scroll-x="1100"  size="small" striped/>
  </section>
</template>

<style scoped>
.panel { display: flex; flex-direction: column; gap: 12px; min-height: 230px; }
.section-title { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 10px; }
h3 { margin: 0; color: var(--color-text); font-size: 15px; }
p { margin: 3px 0 0; color: var(--color-muted); font-size: 12px; }
.share-name, .row-actions { display: flex; align-items: center; gap: 7px; }
.loading-state { display: flex; align-items: center; gap: 8px; color: var(--color-muted); }
.share-name { font-weight: 500; min-width: 0; }
.share-name span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.row-actions { justify-content: center; gap: 10px; }
</style>
