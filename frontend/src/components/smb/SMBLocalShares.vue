<script setup>
import {computed, h} from 'vue'
import {NButton, NIcon, NTag, NTooltip} from 'naive-ui'
import {FolderOpenOutline, OpenOutline, TrashOutline} from '@vicons/ionicons5'

const props = defineProps({
  shares: {type: Array, default: () => []},
  isAdmin: {type: Boolean, default: false},
  busy: {type: Boolean, default: false},
  openingLocation: {type: String, default: ''},
})
const emit = defineEmits(['open', 'remove'])

function openButton(label, location) {
  return h(NButton, {
    text: true,
    size: 'small',
    type: 'primary',
    disabled: !location || (Boolean(props.openingLocation) && props.openingLocation !== location),
    loading: props.openingLocation === location,
    onClick: () => emit('open', location),
  }, {default: () => label, icon: () => h(NIcon, null, {default: () => h(OpenOutline)})})
}

const columns = computed(() => [
  {
    title: '共享名称', key: 'name', width: 145,
    render: (row) => h('div', {class: 'share-name'}, [
      h(NIcon, {size: 16, color: row.special ? '#888' : '#4361ee'}, {default: () => h(FolderOpenOutline)}),
      h('span', row.name),
      row.special ? h(NTag, {size: 'tiny', bordered: false}, {default: () => '系统'}) : null,
    ]),
  },
  {title: '本机目录', key: 'path', minWidth: 180, ellipsis: {tooltip: true}, render: (row) => row.path || '—'},
  {title: '网络访问地址', key: 'networkPath', minWidth: 190, ellipsis: {tooltip: true}},
  {
    title: '状态', key: 'state', width: 80, align: 'center',
    render: (row) => h(NTag, {size: 'small', bordered: false, type: String(row.state).toLowerCase() === 'online' ? 'success' : 'default'}, {default: () => row.state || '未知'}),
  },
  {title: '说明', key: 'description', minWidth: 130, ellipsis: {tooltip: true}, render: (row) => row.description || '—'},
  {
    title: '操作', key: 'actions', width: 245, align: 'center',
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
    <n-empty v-if="shares.length === 0" description="暂无 SMB 共享"/>
    <n-data-table v-else :columns="columns" :data="shares" :bordered="false"
      :scroll-x="1100" :max-height="350" size="small" striped/>
  </section>
</template>

<style scoped>
.panel { display: flex; flex-direction: column; gap: 12px; min-height: 230px; }
.section-title { display: flex; align-items: center; justify-content: space-between; }
h3 { margin: 0; color: inherit; font-size: 15px; }
p { margin: 3px 0 0; color: #888; font-size: 12px; }
.share-name, .row-actions { display: flex; align-items: center; gap: 7px; }
.share-name { font-weight: 500; }
.row-actions { justify-content: center; gap: 10px; }
</style>
