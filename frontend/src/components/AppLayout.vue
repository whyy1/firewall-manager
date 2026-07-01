<script setup>
import {onMounted, ref, watch} from 'vue'
import {store} from '../stores/rules.js'
import {NIcon, useMessage, useDialog} from 'naive-ui'
import {
  ShieldCheckmarkOutline, SwapHorizontalOutline, SearchOutline, AddOutline,
  PowerOutline, RefreshOutline, SunnyOutline, MoonOutline,
} from '@vicons/ionicons5'
import RuleList from './RuleList.vue'
import RuleEditor from './RuleEditor.vue'
import {IsAdmin} from '../../wailsjs/go/main/App'

const props = defineProps({ isDark: Boolean })
const emit = defineEmits(['toggleTheme'])

const isAdmin = ref(false)
const showEditor = ref(false)
const editingRule = ref(null)
const message = useMessage()
const dialog = useDialog()

// 搜索防抖
const nameInput = ref('')
const portInput = ref('')
const addrInput = ref('')
let t1, t2, t3
watch(nameInput, v => { clearTimeout(t1); t1 = setTimeout(() => store.setNameQuery(v), 300) })
watch(portInput, v => { clearTimeout(t2); t2 = setTimeout(() => store.setPortQuery(v), 300) })
watch(addrInput, v => { clearTimeout(t3); t3 = setTimeout(() => store.setAddrQuery(v), 300) })

// 协议选项
const protocolOpts = [
  { label: '全部协议', value: 'all' },
  { label: 'TCP', value: 'tcp' },
  { label: 'UDP', value: 'udp' },
  { label: 'ICMPv4', value: 'icmpv4' },
  { label: 'ICMPv6', value: 'icmpv6' },
  { label: '任意', value: 'any' },
]

onMounted(async () => {
  isAdmin.value = await IsAdmin()
  store.fetchRules()
  store.fetchFirewallStatus()
})

function handleAdd() { editingRule.value = null; showEditor.value = true }
function handleEdit(rule) { editingRule.value = rule; showEditor.value = true }
function handleDirectionChange(dir) { store.setDirection(dir) }

function handleToggleFirewall() {
  dialog.warning({
    title: store.firewallOn ? '关闭防火墙' : '开启防火墙',
    content: store.firewallOn ? '确定要关闭防火墙吗？这会降低系统安全性。' : '确定要开启防火墙吗？',
    positiveText: '确定', negativeText: '取消',
    onPositiveClick: async () => {
      try { await store.toggleFirewall(); message.success(store.firewallOn ? '防火墙已开启' : '防火墙已关闭') }
      catch (e) { message.error('操作失败: ' + String(e)) }
    }
  })
}

function handleResetFirewall() {
  dialog.error({
    title: '⚠️ 重置防火墙',
    content: '这将恢复 Windows 防火墙的默认规则，所有自定义规则将被清除！确定要继续吗？',
    positiveText: '确认重置', negativeText: '取消',
    onPositiveClick: async () => {
      try { await store.resetFirewall(); message.success('防火墙已重置为默认规则') }
      catch (e) { message.error('重置失败: ' + String(e)) }
    }
  })
}
</script>

<template>
  <div class="layout" :class="{ 'light-theme': !isDark }">
    <header class="header">
      <div class="header-left">
        <NIcon :size="22" color="#4361ee"><ShieldCheckmarkOutline/></NIcon>
        <span class="title">防火墙管理器</span>
        <n-tag v-if="isAdmin" type="success" size="small" :bordered="false">管理员</n-tag>
        <n-tag v-else type="warning" size="small" :bordered="false">需要管理员权限</n-tag>
      </div>
      <div class="header-right">
        <n-input v-model:value="nameInput" placeholder="规则名称" clearable size="small" style="width:140px">
          <template #prefix><NIcon :size="14" color="#999"><SearchOutline/></NIcon></template>
        </n-input>
        <n-input v-model:value="portInput" placeholder="端口" clearable size="small" style="width:70px"/>
        <n-input v-model:value="addrInput" placeholder="地址" clearable size="small" style="width:100px"/>
        <n-select v-model:value="store.protocolFilter" :options="protocolOpts" size="small" style="width:100px" @update:value="store.setProtocolFilter"/>
        <n-select v-model:value="store.actionFilter" :options="[{label:'全部动作',value:'all'},{label:'允许',value:'allow'},{label:'阻止',value:'block'}]" size="small" style="width:90px" @update:value="store.setActionFilter"/>
        <n-select v-model:value="store.statusFilter" :options="[{label:'全部状态',value:'all'},{label:'已启用',value:'enabled'},{label:'已禁用',value:'disabled'}]" size="small" style="width:90px" @update:value="store.setStatusFilter"/>
        <n-button text size="small" @click="emit('toggleTheme')" :title="isDark ? '切换亮色' : '切换暗色'">
          <NIcon :size="20" :color="isDark ? '#f0c040' : '#666'"><MoonOutline v-if="isDark"/><SunnyOutline v-else/></NIcon>
        </n-button>
        <n-button type="primary" size="small" @click="handleAdd">
          <template #icon><NIcon><AddOutline/></NIcon></template>添加规则
        </n-button>
      </div>
    </header>

    <div class="body">
      <aside class="sidebar">
        <div class="nav-group">
          <div class="nav-item" :class="{active: store.direction === 'in'}" @click="handleDirectionChange('in')">
            <NIcon :size="18"><SwapHorizontalOutline/></NIcon><span>入站规则</span>
          </div>
          <div class="nav-item" :class="{active: store.direction === 'out'}" @click="handleDirectionChange('out')">
            <NIcon :size="18"><SwapHorizontalOutline style="transform:rotate(180deg)"/></NIcon><span>出站规则</span>
          </div>
        </div>

        <div class="nav-group">
          <div class="nav-group-title">快捷操作</div>
          <div class="nav-item" @click="handleToggleFirewall">
            <NIcon :size="18" :color="store.firewallOn ? '#63e2b7' : '#e88080'"><PowerOutline/></NIcon>
            <span>防火墙 {{ store.firewallOn ? '已开启' : '已关闭' }}</span>
          </div>
          <div class="nav-item" @click="handleResetFirewall">
            <NIcon :size="18" color="#f0c040"><RefreshOutline/></NIcon><span>重置防火墙</span>
          </div>
        </div>

        <div class="sidebar-footer">
          <div class="stat"><span class="stat-label">总规则数</span><span class="stat-value">{{ store.rules.length }}</span></div>
          <div class="stat"><span class="stat-label">已启用</span><span class="stat-value enabled">{{ store.enabledCount }}</span></div>
        </div>
      </aside>

      <main class="content"><RuleList @edit="handleEdit"/></main>
    </div>

    <RuleEditor v-model:show="showEditor" :rule="editingRule"/>
  </div>
</template>

<style scoped>
.layout { display: flex; flex-direction: column; width: 100%; height: 100vh; background: #121218; color: #e0e0e6; }
.light-theme { background: #f5f5f5; color: #333; }
.light-theme .header { background: #fff; border-bottom-color: rgba(0,0,0,0.08); }
.light-theme .sidebar { background: #fafafa; border-right-color: rgba(0,0,0,0.08); }
.light-theme .nav-item { color: #555; }
.light-theme .nav-item:hover { background: rgba(0,0,0,0.04); color: #333; }
.light-theme .nav-item.active { background: rgba(67,97,238,0.1); color: #4361ee; }
.light-theme .nav-group-title { color: #999; }
.light-theme .stat-label { color: #999; }
.light-theme .stat-value { color: #666; }
.light-theme .title { color: #333; }
.header { display: flex; align-items: center; justify-content: space-between; padding: 0 20px; height: 52px; background: #18181f; border-bottom: 1px solid rgba(255,255,255,0.06); flex-shrink: 0; -webkit-app-region: drag; }
.header-left { display: flex; align-items: center; gap: 10px; -webkit-app-region: no-drag; }
.header-right { display: flex; align-items: center; gap: 8px; -webkit-app-region: no-drag; }
.title { font-size: 15px; font-weight: 600; color: #e8e8ed; letter-spacing: 0.5px; }
.body { display: flex; flex: 1; overflow: hidden; }
.sidebar { width: 200px; background: #16161d; border-right: 1px solid rgba(255,255,255,0.06); display: flex; flex-direction: column; padding: 12px 0; flex-shrink: 0; }
.nav-group { padding: 0 8px; margin-bottom: 8px; }
.nav-group-title { font-size: 11px; color: #666; padding: 8px 12px 4px; text-transform: uppercase; letter-spacing: 1px; }
.nav-item { display: flex; align-items: center; gap: 10px; padding: 9px 14px; border-radius: 6px; cursor: pointer; font-size: 13px; color: #aaa; transition: all 0.15s; }
.nav-item:hover { background: rgba(255,255,255,0.04); color: #ddd; }
.nav-item.active { background: rgba(67,97,238,0.15); color: #4361ee; }
.sidebar-footer { margin-top: auto; padding: 16px; border-top: 1px solid rgba(255,255,255,0.06); display: flex; flex-direction: column; gap: 8px; }
.stat { display: flex; justify-content: space-between; font-size: 12px; }
.stat-label { color: #666; }
.stat-value { color: #999; font-weight: 600; }
.stat-value.enabled { color: #63e2b7; }
.content { flex: 1; overflow: auto; padding: 16px 20px; }
</style>
