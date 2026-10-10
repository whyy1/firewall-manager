<script setup>
import {ref, watch, computed} from 'vue'
import {store} from '../stores/rules.js'
import {useMessage} from 'naive-ui'
import {OpenFileDialog} from '../../wailsjs/go/main/App'
import {ruleToForm, supportsPorts, selectProfiles, validateRuleForm} from '../stores/rule_form.js'
import {ruleReadOnlyReason} from '../stores/rule_identity.js'

const props = defineProps({show: Boolean, rule: Object, isAdmin: Boolean})
const emit = defineEmits(['update:show'])
const message = useMessage()
const isEdit = computed(() => !!props.rule)
const submitting = ref(false)
const browsing = ref(false)
const originalRule = ref(null)
const form = ref(ruleToForm(null, store.direction))
const busy = computed(() => submitting.value || browsing.value)
const readOnlyReason = computed(() => isEdit.value ? ruleReadOnlyReason(originalRule.value) : '')
const blocked = computed(() => busy.value || !props.isAdmin || !store.canWrite || !!readOnlyReason.value)
const portsAllowed = computed(() => supportsPorts(form.value.protocol))
const protocolOptions = computed(() => {
  const values = ['any', 'tcp', 'udp', 'icmpv4', 'icmpv6']
  if (!values.includes(form.value.protocol)) values.push(form.value.protocol)
  return values.map(value => ({label: value === 'any' ? '任意' : value.toUpperCase(), value}))
})
const actionOptions = [{label: '允许', value: 'allow'}, {label: '阻止', value: 'block'}]
const directionOptions = [{label: '入站', value: 'in'}, {label: '出站', value: 'out'}]
const profileOptions = [{label: '任意', value: 'any'}, {label: '域', value: 'domain'}, {label: '专用', value: 'private'}, {label: '公用', value: 'public'}]
watch(() => props.show, show => {
  if (!show || busy.value) return
  originalRule.value = props.rule ? {...props.rule} : null
  form.value = ruleToForm(props.rule, store.direction)
})
function updateProfiles(values) { form.value = {...form.value, profile: selectProfiles(values, form.value.profile)} }
function updateProtocol(value) {
  form.value = {...form.value, protocol: value, ...(!supportsPorts(value) ? {localPort: '', remotePort: ''} : {})}
}
async function handleSubmit() {
  if (blocked.value) return
  const error = validateRuleForm(form.value)
  if (error) { message.warning(error); return }
  submitting.value = true
  try {
    const result = await store.saveRule(originalRule.value, {...form.value, profile: form.value.profile.join(',')})
    message.success(isEdit.value ? '规则已更新' : '规则已添加')
    if (result?.refreshError) message.warning('规则操作已成功，但刷新失败，请刷新确认最新状态：' + result.refreshError)
    emit('update:show', false)
  } catch (error) { message.error('操作失败：' + String(error)) }
  finally { submitting.value = false }
}
async function handleBrowseProgram() {
  if (blocked.value) return
  browsing.value = true
  try {
    const path = await OpenFileDialog()
    if (path) form.value = {...form.value, program: path}
  } catch (error) { message.error('选择文件失败：' + String(error)) }
  finally { browsing.value = false }
}
</script>

<template>
  <n-drawer :show="show" :width="440" :mask-closable="!busy" :close-on-esc="!busy" @update:show="!busy && emit('update:show', $event)">
    <n-drawer-content :title="isEdit ? '编辑规则' : '添加规则'">
      <n-alert v-if="!isAdmin" type="warning" :bordered="false">修改规则需要管理员权限。</n-alert>
      <section v-if="isEdit" class="rule-identity">
        <n-alert v-if="readOnlyReason" type="warning" :bordered="false">只读规则：{{ readOnlyReason }}</n-alert>
        <label>内部标识<n-input :value="form.id" readonly aria-label="规则内部标识" /></label>
        <label>策略来源<n-input :value="form.policyStoreSource || '未确认'" readonly aria-label="规则策略来源" /></label>
        <label>来源类型<n-input :value="form.policyStoreSourceType || '未确认'" readonly aria-label="规则来源类型" /></label>
        <label v-if="originalRule?.icmpType">原 ICMP 类型<n-input :value="originalRule.icmpType" readonly aria-label="原 ICMP 类型" /></label>
        <p v-if="originalRule?.icmpType" class="form-hint">协议未修改时保留原 ICMP 类型与代码。</p>
        <p class="form-hint">可选择并复制身份信息；同名规则按内部标识和来源区分。</p>
      </section>
      <n-form :disabled="blocked" label-placement="top" :model="form" class="rule-form">
        <section class="form-section">
          <h3>基本信息</h3>
          <n-form-item label="规则名称" required><n-input v-model:value="form.name" :disabled="isEdit || blocked" placeholder="输入便于识别的名称" clearable /></n-form-item>
          <n-grid :cols="2" :x-gap="12">
            <n-gi><n-form-item label="方向"><n-select v-model:value="form.direction" :options="directionOptions" /></n-form-item></n-gi>
            <n-gi><n-form-item label="动作"><n-select v-model:value="form.action" :options="actionOptions" /></n-form-item></n-gi>
          </n-grid>
        </section>
        <section class="form-section">
          <h3>程序与地址</h3>
          <n-form-item label="程序路径">
            <n-input-group><n-input v-model:value="form.program" placeholder="留空表示任意程序" clearable style="flex: 1" /><n-button :disabled="blocked" :loading="browsing" @click="handleBrowseProgram">浏览</n-button></n-input-group>
          </n-form-item>
          <n-form-item label="本地地址"><n-input v-model:value="form.localAddr" placeholder="留空表示任意；支持 IP、子网、范围或列表" clearable /></n-form-item>
          <n-form-item label="远程地址"><n-input v-model:value="form.remoteAddr" placeholder="留空表示任意；如 10.0.0.0/24 或 LocalSubnet" clearable /></n-form-item>
        </section>
        <section class="form-section">
          <h3>协议与端口</h3>
          <n-form-item label="协议"><n-select :value="form.protocol" :options="protocolOptions" @update:value="updateProtocol" /></n-form-item>
          <n-form-item label="适用配置文件"><n-select :value="form.profile" multiple :options="profileOptions" @update:value="updateProfiles" /></n-form-item>
          <n-grid :cols="2" :x-gap="12">
            <n-gi><n-form-item label="本地端口"><n-input v-model:value="form.localPort" :disabled="blocked || !portsAllowed" placeholder="如 80,443 或 1000-2000" clearable /></n-form-item></n-gi>
            <n-gi><n-form-item label="远程端口"><n-input v-model:value="form.remotePort" :disabled="blocked || !portsAllowed" placeholder="留空表示任意" clearable /></n-form-item></n-gi>
          </n-grid>
          <p class="form-hint">仅 TCP、UDP 支持端口限制；切换到其他协议会清空端口。</p>
        </section>
        <n-form-item label="启用规则"><n-switch v-model:value="form.enabled" aria-label="启用规则" /></n-form-item>
      </n-form>
      <template #footer><div class="form-footer"><n-button :disabled="busy" @click="emit('update:show', false)">取消</n-button><n-button type="primary" :loading="submitting" :disabled="blocked" :aria-label="isEdit ? '保存' : '添加'" @click="handleSubmit">{{ isEdit ? '保存' : '添加' }}</n-button></div></template>
    </n-drawer-content>
  </n-drawer>
</template>

<style scoped>
.rule-identity { display: grid; gap: 10px; margin-bottom: 16px; }
.rule-identity label { display: grid; gap: 4px; font-size: 12px; color: var(--color-muted); }
.rule-form { display: flex; flex-direction: column; gap: 16px; }
.form-section { padding-bottom: 12px; border-bottom: 1px solid var(--color-border); }
h3 { font-size: 14px; color: var(--color-text); margin: 12px 0; }
.form-hint { font-size: 12px; color: var(--color-muted); margin: 0; }
.form-footer { display: flex; gap: 8px; justify-content: flex-end; }
</style>
