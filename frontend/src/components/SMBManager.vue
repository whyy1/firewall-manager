<script setup>
import {computed, onMounted, reactive, ref} from 'vue'
import {NIcon, NTabPane, NTabs, NTooltip, useDialog, useMessage} from 'naive-ui'
import {
  AddOutline,
  CloseCircleOutline,
  InformationCircleOutline,
  LockClosedOutline,
  ServerOutline,
} from '@vicons/ionicons5'
import {OpenDirectoryDialog} from '../../wailsjs/go/main/App'
import {smbStore} from '../stores/smb.js'
import SMBFirewallProfiles from './smb/SMBFirewallProfiles.vue'
import SMBLocalShares from './smb/SMBLocalShares.vue'
import SMBNetworkShares from './smb/SMBNetworkShares.vue'
import SMBSessions from './smb/SMBSessions.vue'

const props = defineProps({isAdmin: {type: Boolean, default: false}})
const dialog = useDialog()
const message = useMessage()
const activeTab = ref('local')
const confirmationOpen = ref(false)
const showCreateDialog = ref(false)
const createForm = reactive({name: '', path: '', description: '', accessMode: 'read'})

const accessOptions = [
  {label: '只读（推荐）', value: 'read'},
  {label: '更改', value: 'change'},
  {label: '完全控制', value: 'full'},
]
const startupTypeLabels = {Auto: '自动', Automatic: '自动', Manual: '手动', Disabled: '已禁用'}
const action = computed(() => smbStore.mutationAction)
const busy = computed(() => Boolean(smbStore.mutationAction))
const currentTabLoading = computed(() => activeTab.value === 'local'
  ? smbStore.localLoading || smbStore.firewallLoading
  : smbStore.sessionsLoading)

onMounted(() => {
  smbStore.fetchLocal()
  smbStore.fetchFirewall()
})

async function runAction(operation, successText) {
  try {
    await operation()
    message.success(successText)
    return true
  } catch (error) {
    message.error(String(error))
    return false
  }
}

function confirmMutation({title, content, positiveText, operation, successText}) {
  if (busy.value || confirmationOpen.value) return
  confirmationOpen.value = true
  let handle
  handle = dialog.warning({
    title, content, positiveText, negativeText: '取消',
    closable: false, maskClosable: false, closeOnEsc: false,
    onNegativeClick: () => { confirmationOpen.value = false },
    onPositiveClick: async () => {
      if (busy.value) return false
      handle.positiveButtonProps = {loading: true, disabled: true}
      handle.negativeButtonProps = {disabled: true}
      const succeeded = await runAction(operation, successText)
      if (succeeded) {
        confirmationOpen.value = false
        return undefined
      }
      handle.positiveButtonProps = {loading: false, disabled: false}
      handle.negativeButtonProps = {disabled: false}
      return false
    },
  })
}

function handleTabChange(tab) {
  activeTab.value = tab
  if (tab === 'sessions' && !smbStore.sessions.length && !smbStore.sessionsLoading) {
    smbStore.fetchSessions()
  }
}

function refreshCurrentTab() {
  if (activeTab.value === 'local') {
    smbStore.fetchLocal()
    smbStore.fetchFirewall()
  } else if (activeTab.value === 'sessions') {
    smbStore.fetchSessions()
  }
}

function confirmServiceChange() {
  const running = !smbStore.overview?.serviceRunning
  confirmMutation({
    title: running ? '启动 SMB 服务' : '停止 SMB 服务',
    content: running ? '将启动 Windows Server 服务，使本机共享可被访问。' : '停止服务会立即中断现有连接，并使所有本机共享不可访问。',
    positiveText: running ? '确认启动' : '确认停止',
    operation: () => smbStore.setServiceRunning(running),
    successText: running ? 'SMB 服务已启动' : 'SMB 服务已停止',
  })
}

function confirmDisableSMB1() {
  confirmMutation({
    title: '关闭 SMB1',
    content: '关闭后，只支持 SMB1 的旧设备将无法连接。SMB2/3 不受影响。',
    positiveText: '确认关闭',
    operation: () => smbStore.disableSMB1(),
    successText: 'SMB1 已关闭',
  })
}

function confirmRemove(row) {
  confirmMutation({
    title: `删除共享“${row.name}”`,
    content: '删除会强制断开现有连接，但不会删除目录和其中的文件。',
    positiveText: '确认删除',
    operation: () => smbStore.removeShare(row.name),
    successText: 'SMB 共享已删除',
  })
}

function confirmDisconnect(session) {
  confirmMutation({
    title: `断开“${session.clientComputerName || session.clientUserName}”`,
    content: `该会话当前打开 ${session.numOpens} 个文件。强制断开可能导致客户端尚未写回的数据丢失。`,
    positiveText: '确认断开',
    operation: () => smbStore.closeSession(session.sessionId),
    successText: 'SMB 会话已断开',
  })
}

function confirmFirewallToggle(enabled) {
  confirmMutation({
    title: `${enabled ? '全部启用' : '全部禁用'}共享规则`,
    content: enabled
      ? '将为域、专用和公用网络统一启用“文件和打印机共享”规则组。公用网络暴露风险较高，请确认当前网络可信。'
      : '将为域、专用和公用网络统一禁用“文件和打印机共享”规则组，现有远程访问可能中断。',
    positiveText: enabled ? '确认启用' : '确认禁用',
    operation: () => smbStore.setFirewallEnabled(enabled),
    successText: 'SMB 防火墙规则已更新',
  })
}

function openCreateDialog() {
  if (busy.value || confirmationOpen.value) return
  Object.assign(createForm, {name: '', path: '', description: '', accessMode: 'read'})
  showCreateDialog.value = true
}

async function chooseDirectory() {
  try {
    const selected = await OpenDirectoryDialog()
    if (selected) createForm.path = selected
  } catch (error) {
    message.error('选择目录失败: ' + String(error))
  }
}

async function createShare() {
  if (!createForm.name.trim() || !createForm.path.trim()) {
    message.error('请输入共享名称并选择本机目录')
    return false
  }
  const succeeded = await runAction(() => smbStore.createShare({...createForm}), 'SMB 共享已创建')
  if (succeeded) showCreateDialog.value = false
  return false
}

async function openLocation(location) {
  try {
    await smbStore.openLocation(location)
  } catch (error) {
    message.error('打开失败: ' + String(error))
  }
}
</script>

<template>
  <div class="smb-manager">
    <div class="page-header">
      <div class="page-heading">
        <NIcon :size="24" color="#4361ee"><ServerOutline/></NIcon>
        <div><h2>SMB 文件共享</h2><p>管理本机共享，并打开或检测当前已连接的网络共享</p></div>
      </div>
      <div class="header-actions">
        <NTooltip v-if="activeTab === 'local'">
          <template #trigger>
            <n-button type="primary" size="small" :disabled="!isAdmin || busy || confirmationOpen" @click="openCreateDialog">
              <template #icon><NIcon><AddOutline/></NIcon></template>新建共享
            </n-button>
          </template>
          {{ isAdmin ? '把一个本机目录发布为网络共享' : '需要管理员权限' }}
        </NTooltip>
        <n-button v-if="activeTab !== 'network'" size="small" :loading="currentTabLoading" :disabled="busy || confirmationOpen || currentTabLoading" @click="refreshCurrentTab">刷新当前页</n-button>
      </div>
    </div>

    <NTabs :value="activeTab" type="line" animated @update:value="handleTabChange">
      <NTabPane name="local" tab="本机共享">
        <div v-if="smbStore.localLoading && !smbStore.overview" class="center-state"><n-spin/><span>正在读取 SMB 配置...</span></div>
        <div v-else-if="smbStore.localError" class="center-state">
          <NIcon :size="44" color="#e88080"><CloseCircleOutline/></NIcon><span>{{ smbStore.localError }}</span>
          <n-button size="small" type="primary" @click="smbStore.fetchLocal()">重试</n-button>
        </div>
        <div v-else-if="smbStore.overview" class="local-content">
          <section class="status-grid">
            <div class="status-card">
              <span class="status-label">SMB 服务</span>
              <n-tag :type="smbStore.overview.serviceRunning ? 'success' : 'default'" :bordered="false" size="small">{{ smbStore.overview.serviceRunning ? '运行中' : '已停止' }}</n-tag>
              <span class="detail">启动类型：{{ startupTypeLabels[smbStore.overview.startupType] || smbStore.overview.startupType }}</span>
              <n-button size="tiny" :type="smbStore.overview.serviceRunning ? 'error' : 'primary'" :disabled="!isAdmin || busy || confirmationOpen" @click="confirmServiceChange">{{ smbStore.overview.serviceRunning ? '停止服务' : '启动服务' }}</n-button>
            </div>
            <div class="status-card" :class="{'unsafe-card': smbStore.overview.smb1Enabled}">
              <span class="status-label">SMB1</span>
              <n-tag :type="smbStore.overview.smb1Enabled ? 'error' : 'success'" :bordered="false" size="small">{{ smbStore.overview.smb1Enabled ? '已启用，不安全' : '已关闭' }}</n-tag>
              <n-button v-if="smbStore.overview.smb1Enabled" size="tiny" type="error" :disabled="!isAdmin || busy || confirmationOpen" @click="confirmDisableSMB1">关闭 SMB1</n-button>
            </div>
            <div class="status-card"><span class="status-label">SMB2/3</span><n-tag :type="smbStore.overview.smb2Enabled ? 'success' : 'error'" :bordered="false" size="small">{{ smbStore.overview.smb2Enabled ? '已启用' : '已关闭' }}</n-tag></div>
            <div class="status-card"><span class="status-label">安全策略</span><span class="detail">签名：{{ smbStore.overview.requireSecuritySignature ? '必需' : '非必需' }}</span><span class="detail">全局加密：{{ smbStore.overview.encryptData ? '开启' : '关闭' }}</span></div>
          </section>
          <SMBFirewallProfiles :status="smbStore.firewall" :loading="smbStore.firewallLoading" :error="smbStore.firewallError || ''" :is-admin="isAdmin" :busy="busy || confirmationOpen" @refresh="smbStore.fetchFirewall()" @toggle="confirmFirewallToggle"/>
          <SMBLocalShares :shares="smbStore.shares" :is-admin="isAdmin" :busy="busy || confirmationOpen" :opening-location="smbStore.openingLocation" @open="openLocation" @remove="confirmRemove"/>
        </div>
      </NTabPane>

      <NTabPane name="network" tab="已连接共享">
        <SMBNetworkShares :shares="smbStore.networkShares" :loading="smbStore.networkLoading" :error="smbStore.networkError || ''" :opening-location="smbStore.openingLocation" @scan="smbStore.scanNetwork()" @open="openLocation"/>
      </NTabPane>

      <NTabPane name="sessions" tab="活动会话">
        <SMBSessions :sessions="smbStore.sessions" :loading="smbStore.sessionsLoading" :error="smbStore.sessionsError || ''" :is-admin="isAdmin" :busy="busy || confirmationOpen" @refresh="smbStore.fetchSessions()" @disconnect="confirmDisconnect"/>
      </NTabPane>
    </NTabs>

    <n-modal v-model:show="showCreateDialog" preset="dialog" title="新建 SMB 共享" positive-text="创建共享" negative-text="取消"
      :closable="action !== 'create'" :mask-closable="action !== 'create'" :close-on-esc="action !== 'create'"
      :positive-button-props="{loading: action === 'create', disabled: busy}" :negative-button-props="{disabled: action === 'create'}" @positive-click="createShare">
      <div class="create-form">
        <label>共享名称</label><n-input v-model:value="createForm.name" maxlength="80" placeholder="例如 TeamFiles"/>
        <label>本机目录</label><n-input-group><n-input v-model:value="createForm.path" placeholder="选择一个已存在的目录"/><n-button @click="chooseDirectory">浏览</n-button></n-input-group>
        <label>共享权限</label><n-select v-model:value="createForm.accessMode" :options="accessOptions"/>
        <label>说明</label><n-input v-model:value="createForm.description" type="textarea" maxlength="256" show-count placeholder="可选"/>
        <div class="notice"><NIcon :size="18" color="#f0a020"><InformationCircleOutline/></NIcon><span>默认向 Everyone 授予只读共享权限，但不会修改目录现有的 NTFS 权限。</span></div>
        <div class="notice"><NIcon :size="18" color="#4361ee"><LockClosedOutline/></NIcon><span>创建后可在列表中点击“打开目录”或“打开共享”。</span></div>
      </div>
    </n-modal>
  </div>
</template>

<style scoped>
.smb-manager { display: flex; flex-direction: column; min-height: 100%; gap: 12px; color: inherit; }
.page-header { display: flex; align-items: center; justify-content: space-between; padding-bottom: 12px; border-bottom: 1px solid rgba(128,128,128,0.18); }
.page-heading, .header-actions { display: flex; align-items: center; gap: 12px; }
h2 { margin: 0; color: inherit; font-size: 18px; }
.page-heading p { margin: 3px 0 0; color: #888; font-size: 12px; }
.center-state { min-height: 280px; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 14px; color: #888; }
.local-content { display: flex; flex-direction: column; gap: 14px; }
.status-grid { display: grid; grid-template-columns: repeat(4, minmax(145px, 1fr)); gap: 10px; }
.status-card { min-height: 100px; padding: 12px; border: 1px solid rgba(128,128,128,0.18); border-radius: 8px; background: rgba(128,128,128,0.06); display: flex; flex-direction: column; align-items: flex-start; gap: 8px; }
.unsafe-card { border-color: rgba(232,128,128,0.5); background: rgba(232,128,128,0.08); }
.status-label, .detail, .create-form label { color: #888; font-size: 12px; }
.create-form { display: flex; flex-direction: column; gap: 9px; margin-bottom: 8px; }
.notice { display: flex; align-items: flex-start; gap: 8px; padding: 10px; border-radius: 6px; background: rgba(128,128,128,0.08); color: #888; font-size: 12px; line-height: 1.5; }
@media (max-width: 980px) { .status-grid { grid-template-columns: repeat(2, minmax(145px, 1fr)); } }
</style>
