<script setup>
import {onMounted, ref} from 'vue'
import {store} from '../stores/rules.js'
import {NIcon} from 'naive-ui'
import {
  ShieldCheckmarkOutline,
  SwapHorizontalOutline,
  SearchOutline,
  AddOutline,
  FolderOpenOutline,
  BanOutline,
  CheckmarkCircleOutline,
} from '@vicons/ionicons5'
import RuleList from './RuleList.vue'
import RuleEditor from './RuleEditor.vue'
import {IsAdmin} from '../../wailsjs/go/main/App'

const isAdmin = ref(false)
const showEditor = ref(false)
const editingRule = ref(null)

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
          v-model:value="store.searchQuery"
          placeholder="搜索规则名称或程序..."
          clearable
          size="small"
          style="width: 260px"
          @update:value="store.setSearch"
        >
          <template #prefix>
            <NIcon :size="16" color="#999">
              <SearchOutline/>
            </NIcon>
          </template>
        </n-input>
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
          <div class="nav-item" @click="$refs.blockDialog?.show()">
            <NIcon :size="18">
              <BanOutline/>
            </NIcon>
            <span>阻止程序联网</span>
          </div>
          <div class="nav-item" @click="$refs.allowDialog?.show()">
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

<script>
import {defineComponent, ref} from 'vue'
import {store} from '../stores/rules.js'
import {useMessage} from 'naive-ui'
import {FolderOpenOutline} from '@vicons/ionicons5'

const QuickActionDialog = defineComponent({
  props: {
    title: String,
    action: String,
  },
  setup(props) {
    const show = ref(false)
    const path = ref('')
    const message = useMessage()

    return {show, path, message}
  },
  methods: {
    show() {
      this.show = true
      this.path = ''
    },
    async handleConfirm() {
      if (!this.path) return
      try {
        if (this.action === 'block') {
          await store.blockApp(this.path)
          this.message.success('已成功阻止该程序联网')
        } else {
          await store.allowApp(this.path)
          this.message.success('已成功放行该程序联网')
        }
        this.show = false
      } catch (e) {
        this.message.error('操作失败: ' + String(e))
      }
    }
  },
  template: `
    <n-modal v-model:show="show" preset="dialog" :title="title"
             positive-text="确认" negative-text="取消"
             @positive-click="handleConfirm">
      <n-input v-model:value="path" placeholder="输入程序完整路径，如 C:\Program Files\app.exe"
               clearable style="margin-top: 12px"/>
    </n-modal>
  `
})

export default {
  components: {QuickActionDialog}
}
</script>

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
