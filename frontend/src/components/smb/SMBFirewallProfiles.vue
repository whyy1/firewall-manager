<script setup>
import {computed} from 'vue'

const props = defineProps({
  status: {type: Object, default: null},
  loading: {type: Boolean, default: false},
  error: {type: String, default: ''},
  isAdmin: {type: Boolean, default: false},
  busy: {type: Boolean, default: false},
})
const emit = defineEmits(['refresh', 'toggle'])

const profileLabels = {domain: '域网络', private: '专用网络', public: '公用网络'}
const allEnabled = computed(() => Boolean(props.status?.profiles?.length) && props.status.profiles.every(profile => profile.enabled))
</script>

<template>
  <section class="firewall-panel">
    <div class="firewall-title">
      <div><h3>文件和打印机共享防火墙</h3><p>显示各网络配置文件的有效状态；切换时统一修改整个内置规则组</p></div>
      <div class="firewall-actions">
        <n-button size="tiny" :loading="loading" :disabled="loading || busy" @click="emit('refresh')">刷新</n-button>
        <n-button v-if="status?.available" size="tiny" :type="allEnabled ? 'error' : 'primary'" :loading="busy"
          :disabled="!isAdmin || busy" @click="emit('toggle', !allEnabled)">
          {{ allEnabled ? '全部禁用' : '全部启用' }}
        </n-button>
      </div>
    </div>
    <n-alert v-if="error" type="error" :bordered="false">{{ error }}</n-alert>
    <div v-else-if="status?.available" class="profile-list">
      <div v-for="profile in status.profiles" :key="profile.name" class="profile-row">
        <div><span class="profile-name">{{ profileLabels[profile.name] || profile.name }}</span><span v-if="profile.name === 'public'" class="risk">暴露风险较高</span></div>
        <n-tag :type="profile.enabled ? 'success' : 'default'" :bordered="false" size="small">
          {{ profile.enabled ? '已启用' : '已禁用' }}
        </n-tag>
      </div>
    </div>
    <n-empty v-else-if="!loading" description="未找到文件和打印机共享规则组" size="small"/>
  </section>
</template>

<style scoped>
.firewall-panel { display: flex; flex-direction: column; gap: 10px; padding: 14px; border: 1px solid rgba(128,128,128,0.18); border-radius: 8px; background: rgba(128,128,128,0.05); }
.firewall-title { display: flex; align-items: center; justify-content: space-between; }
.firewall-actions { display: flex; align-items: center; gap: 8px; }
h3 { margin: 0; color: inherit; font-size: 14px; }
p { margin: 3px 0 0; color: #888; font-size: 12px; }
.profile-list { display: grid; grid-template-columns: repeat(3, minmax(130px, 1fr)); gap: 10px; }
.profile-row { display: flex; align-items: center; justify-content: space-between; gap: 10px; padding: 10px; border-radius: 6px; background: rgba(128,128,128,0.08); }
.profile-name { font-size: 13px; }
.risk { display: block; margin-top: 2px; color: #e88080; font-size: 10px; }
@media (max-width: 900px) { .profile-list { grid-template-columns: 1fr; } }
</style>
