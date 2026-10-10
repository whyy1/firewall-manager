<script setup>
import {computed, nextTick, onMounted, ref, watch} from 'vue'
import {NIcon} from 'naive-ui'
import {ShieldCheckmarkOutline, SunnyOutline, MoonOutline, HardwareChipOutline, FolderOpenOutline} from '@vicons/ionicons5'
import RuleList from './RuleList.vue'
import RuleEditor from './RuleEditor.vue'
import PortManager from './PortManager.vue'
import SMBManager from './SMBManager.vue'
import {IsAdmin} from '../../wailsjs/go/main/App'

defineProps({isDark: Boolean})
const emit = defineEmits(['toggleTheme'])
const isAdmin = ref(false)
const privilegeState = ref('checking')
const privilegeError = ref('')
const currentView = ref('rules')
const showEditor = ref(false)
const editingRule = ref(null)
const contentHost = ref(null)
watch(currentView, async () => { await nextTick(); if (contentHost.value) contentHost.value.scrollTop = 0 })
const navigation = [
  {key: 'rules', label: '防火墙规则', icon: ShieldCheckmarkOutline},
  {key: 'ports', label: '常用端口', icon: HardwareChipOutline},
  {key: 'smb', label: 'SMB 文件共享', icon: FolderOpenOutline},
]
const privilegeLabel = computed(() => privilegeState.value === 'checking' ? '检查权限中' : isAdmin.value ? '管理员模式' : '只读模式')
onMounted(async () => {
  try { isAdmin.value = await IsAdmin(); privilegeState.value = 'ready' }
  catch (error) { privilegeState.value = 'error'; privilegeError.value = '权限检查失败: ' + String(error) }
})
function openEditor(rule = null) {
  if (!isAdmin.value || showEditor.value) return
  editingRule.value = rule
  showEditor.value = true
}
</script>

<template>
  <div class="app-shell">
    <header class="app-header">
      <div class="brand"><div class="brand-mark"><NIcon :size="20"><ShieldCheckmarkOutline/></NIcon></div><span>防火墙管理器</span><span class="brand-caption">本机管理控制台</span></div>
      <div class="header-actions">
        <n-tag :type="isAdmin ? 'success' : 'default'" size="small" :bordered="false">{{ privilegeLabel }}</n-tag>
        <n-button quaternary circle :aria-label="isDark ? '切换亮色主题' : '切换暗色主题'" :title="isDark ? '切换亮色主题' : '切换暗色主题'" @click="emit('toggleTheme')"><template #icon><NIcon :size="19"><SunnyOutline v-if="isDark"/><MoonOutline v-else/></NIcon></template></n-button>
      </div>
    </header>
    <div class="app-body">
      <aside class="app-sidebar">
        <span class="nav-caption">管理功能</span>
        <nav aria-label="功能导航"><button v-for="item in navigation" :key="item.key" class="nav-button" :class="{active: currentView === item.key}" :aria-label="item.label" :aria-current="currentView === item.key ? 'page' : undefined" :title="item.label" @click="currentView = item.key"><NIcon :size="20"><component :is="item.icon"/></NIcon><span class="nav-label">{{ item.label }}</span></button></nav>
        <div class="sidebar-note"><span class="status-dot" :class="{admin: isAdmin}"/><span class="nav-label">{{ isAdmin ? '本机管理员操作' : '查看配置与状态' }}</span></div>
      </aside>
      <main ref="contentHost" class="app-content" aria-label="管理内容">
        <n-alert v-if="privilegeError" type="warning" :bordered="false">{{ privilegeError }}。写操作暂不可用。</n-alert>
        <div v-else-if="privilegeState === 'ready' && !isAdmin" class="readonly-banner">当前为只读模式。需要修改配置时，请以管理员身份重新启动应用。</div>
        <RuleList v-if="currentView === 'rules'" :is-dark="isDark" :is-admin="isAdmin" @add="openEditor()" @edit="openEditor"/>
        <PortManager v-else-if="currentView === 'ports'" :is-admin="isAdmin"/>
        <SMBManager v-else :is-admin="isAdmin"/>
      </main>
    </div>
    <RuleEditor v-model:show="showEditor" :rule="editingRule" :is-admin="isAdmin"/>
  </div>
</template>

<style scoped>
.app-shell { height: 100vh; display: flex; flex-direction: column; color: var(--color-text); background: var(--color-bg); }
.app-header { height: 56px; padding: 0 22px; display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid var(--color-border); background: var(--color-surface); flex-shrink: 0; }
.brand, .header-actions { display: flex; align-items: center; gap: 12px; }
.brand { font-size: 15px; font-weight: 650; }
.brand-mark { display: grid; place-items: center; width: 32px; height: 32px; border-radius: 9px; color: var(--color-primary); background: var(--color-primary-soft); }
.brand-caption { font-size: 11px; color: var(--color-muted); font-weight: 400; border-left: 1px solid var(--color-border); padding-left: 12px; }
.app-body { display: flex; flex: 1; min-height: 0; }
.app-sidebar { width: 184px; padding: 22px 12px 14px; background: var(--color-surface); border-right: 1px solid var(--color-border); display: flex; flex-direction: column; flex-shrink: 0; }
.nav-caption { padding: 0 12px 12px; color: var(--color-muted); font-size: 11px; }
nav { display: flex; flex-direction: column; gap: 6px; }
.nav-button { display: flex; align-items: center; gap: 10px; width: 100%; padding: 12px; border: 0; border-radius: 8px; color: var(--color-muted); background: transparent; cursor: pointer; text-align: left; font: inherit; font-size: 13px; }
.nav-button:hover { background: var(--color-hover); color: var(--color-text); }
.nav-button.active { color: var(--color-primary); background: var(--color-primary-soft); font-weight: 600; }
.sidebar-note { margin-top: auto; padding: 14px 10px 0; display: flex; align-items: center; gap: 8px; font-size: 11px; color: var(--color-muted); border-top: 1px solid var(--color-border); }
.status-dot { width: 6px; height: 6px; border-radius: 50%; background: var(--color-muted); flex-shrink: 0; }
.status-dot.admin { background: var(--color-success); }
.app-content { flex: 1; min-width: 0; overflow: auto; padding: 22px 24px; }
.readonly-banner { margin-bottom: 16px; padding: 10px 14px; font-size: 12px; background: var(--color-primary-soft); color: var(--color-text); border: 1px solid var(--color-border); border-radius: 8px; }
@media (max-width: 1000px) { .app-sidebar { width: 64px; padding: 18px 8px; } .nav-label, .nav-caption { display: none; } .nav-button { justify-content: center; padding: 12px 0; } .sidebar-note { justify-content: center; padding: 14px 0 0; } .app-content { padding: 16px; } }
@media (max-width: 850px) { .brand-caption { display: none; } .app-header { padding: 0 16px; } }
</style>
