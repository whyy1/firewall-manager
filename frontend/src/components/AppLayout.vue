<script setup>
import {onMounted, ref, watch} from 'vue'
import {store} from '../stores/rules.js'
import {NIcon} from 'naive-ui'
import {
  ShieldCheckmarkOutline,
  SwapHorizontalOutline,
  SearchOutline,
  AddOutline,
  BanOutline,
  CheckmarkCircleOutline,
} from '@vicons/ionicons5'
import RuleList from './RuleList.vue'
import RuleEditor from './RuleEditor.vue'
import QuickActionDialog from './QuickActionDialog.vue'
import {IsAdmin} from '../../wailsjs/go/main/App'

const isAdmin = ref(false)
const showEditor = ref(false)
const editingRule = ref(null)
const searchInput = ref('')
let searchTimer = null

// 防抖：300ms 后才更新 store
watch(searchInput, (val) => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => store.setSearch(val), 300)
})
const blockDialog = ref(null)
const allowDialog = ref(null)

onMounted(async () => {
  isAdmin.value = await IsAdmin()
  store.fetchRules()
})

function handleAdd() {
  editingRule.value = null
  showEditor.value = true
}

function handleEdit(rule) {
  editingRule.value = rule
  showEditor.value = true
}

function handleDirectionChange(dir) {
  store.setDirection(dir)
}
</script>

<template>
  <div class="layout">
    <!-- 顶部标题栏 -->
    <header class="header">
      <div class="header-left">
        <NIcon :size="22" color="#4361ee">
          <ShieldCheckmarkOutline/>
        </NIcon>
        <span class="title">防火墙管理器</span>
        <n-tag v-if="isAdmin" type="success" size="small" :bordered="false">管理员</n-tag>
        <n-tag v-else type="warning" size="small" :bordered="false">需要管理员权限</n-tag>
      </div>
      <div class="header-right">
        <n-input
          v-model:value="searchInput"
          placeholder="搜索规则名称、端口、地址..."
          clearable
          size="small"
          style="width: 260px"
          @update:value="(v) => { if (!v) store.setSearch('') }"
        >
          <template #prefix>
            <NIcon :size="16" color="#999">
              <SearchOutline/>
            </NIcon>
          </template>
        </n-input>
        <n-select
          v-model:value="store.actionFilter"
          :options="[{label:'全部动作',value:'all'},{label:'允许',value:'allow'},{label:'阻止',value:'block'}]"
          size="small"
          style="width: 100px"
          @update:value="store.setActionFilter"
        />
        <n-select
          v-model:value="store.statusFilter"
          :options="[{label:'全部状态',value:'all'},{label:'已启用',value:'enabled'},{label:'已禁用',value:'disabled'}]"
          size="small"
          style="width: 100px"
          @update:value="store.setStatusFilter"
        />
        <n-button type="primary" size="small" @click="handleAdd">
          <template #icon>
            <NIcon>
              <AddOutline/>
            </NIcon>
          </template>
          添加规则
        </n-button>
      </div>
    </header>

    <div class="body">
      <!-- 侧边栏 -->
      <aside class="sidebar">
        <div class="nav-group">
          <div
            class="nav-item"
            :class="{active: store.direction === 'in'}"
            @click="handleDirectionChange('in')"
          >
            <NIcon :size="18">
              <SwapHorizontalOutline/>
            </NIcon>
            <span>入站规则</span>
          </div>
          <div
            class="nav-item"
            :class="{active: store.direction === 'out'}"
            @click="handleDirectionChange('out')"
          >
            <NIcon :size="18">
              <SwapHorizontalOutline style="transform: rotate(180deg)"/>
            </NIcon>
            <span>出站规则</span>
          </div>
        </div>

        <div class="nav-group">
          <div class="nav-group-title">快捷操作</div>
          <div class="nav-item" @click="blockDialog?.open()">
            <NIcon :size="18">
              <BanOutline/>
            </NIcon>
            <span>阻止程序联网</span>
          </div>
          <div class="nav-item" @click="allowDialog?.open()">
            <NIcon :size="18">
              <CheckmarkCircleOutline/>
            </NIcon>
            <span>放行程序联网</span>
          </div>
        </div>

        <div class="sidebar-footer">
          <div class="stat">
            <span class="stat-label">总规则数</span>
            <span class="stat-value">{{ store.rules.length }}</span>
          </div>
          <div class="stat">
            <span class="stat-label">已启用</span>
            <span class="stat-value enabled">{{ store.enabledCount }}</span>
          </div>
        </div>
      </aside>

      <!-- 主内容区 -->
      <main class="content">
        <RuleList @edit="handleEdit"/>
      </main>
    </div>

    <!-- 规则编辑器抽屉 -->
    <RuleEditor
      v-model:show="showEditor"
      :rule="editingRule"
    />

    <!-- 快捷操作对话框 -->
    <QuickActionDialog ref="blockDialog" title="阻止程序联网" action="block"/>
    <QuickActionDialog ref="allowDialog" title="放行程序联网" action="allow"/>
  </div>
</template>

<style scoped>
.layout {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100vh;
  background: #121218;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  height: 52px;
  background: #18181f;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  flex-shrink: 0;
  -webkit-app-region: drag;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  -webkit-app-region: no-drag;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
  -webkit-app-region: no-drag;
}

.title {
  font-size: 15px;
  font-weight: 600;
  color: #e8e8ed;
  letter-spacing: 0.5px;
}

.body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.sidebar {
  width: 200px;
  background: #16161d;
  border-right: 1px solid rgba(255, 255, 255, 0.06);
  display: flex;
  flex-direction: column;
  padding: 12px 0;
  flex-shrink: 0;
}

.nav-group {
  padding: 0 8px;
  margin-bottom: 8px;
}

.nav-group-title {
  font-size: 11px;
  color: #666;
  padding: 8px 12px 4px;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 14px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: #aaa;
  transition: all 0.15s;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.04);
  color: #ddd;
}

.nav-item.active {
  background: rgba(67, 97, 238, 0.15);
  color: #4361ee;
}

.sidebar-footer {
  margin-top: auto;
  padding: 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.stat {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
}

.stat-label {
  color: #666;
}

.stat-value {
  color: #999;
  font-weight: 600;
}

.stat-value.enabled {
  color: #63e2b7;
}

.content {
  flex: 1;
  overflow: auto;
  padding: 16px 20px;
}
</style>
