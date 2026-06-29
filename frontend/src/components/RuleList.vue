<script setup>
import {store} from '../stores/rules.js'
import {NIcon, useMessage, useDialog} from 'naive-ui'
import {
  TrashOutline,
  CreateOutline,
  ToggleOutline,
  FolderOpenOutline,
  ShieldCheckmarkOutline,
  CloseCircleOutline,
} from '@vicons/ionicons5'

const emit = defineEmits(['edit'])
const message = useMessage()
const dialog = useDialog()

async function handleToggle(rule) {
  try {
    await store.toggleRule(rule.name, !rule.enabled)
    message.success(rule.enabled ? '已禁用' : '已启用')
  } catch (e) {
    message.error('操作失败: ' + String(e))
  }
}

function handleDelete(rule) {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除规则「${rule.name}」吗？此操作不可撤销。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await store.deleteRule(rule.name)
        message.success('规则已删除')
      } catch (e) {
        message.error('删除失败: ' + String(e))
      }
    }
  })
}

function truncatePath(path, maxLen) {
  if (!path || path === 'Any') return '-'
  if (path.length <= maxLen) return path
  return '...' + path.slice(-(maxLen - 3))
}
</script>

<template>
  <div class="rule-list">
    <!-- 表头 -->
    <div class="list-header">
      <div class="col col-status">状态</div>
      <div class="col col-name">规则名称</div>
      <div class="col col-action">操作</div>
      <div class="col col-program">程序</div>
      <div class="col col-protocol">协议</div>
      <div class="col col-port">端口</div>
      <div class="col col-actions">管理</div>
    </div>

    <!-- 加载中 -->
    <div v-if="store.loading" class="loading-state">
      <n-spin size="medium"/>
      <span>加载规则中...</span>
    </div>

    <!-- 错误提示 -->
    <div v-else-if="store.error" class="error-state">
      <NIcon :size="40" color="#e88080">
        <CloseCircleOutline/>
      </NIcon>
      <span>{{ store.error }}</span>
      <n-button size="small" @click="store.fetchRules">重试</n-button>
    </div>

    <!-- 空状态 -->
    <div v-else-if="store.filteredRules.length === 0" class="empty-state">
      <NIcon :size="48" color="#444">
        <ShieldCheckmarkOutline/>
      </NIcon>
      <span>{{ store.searchQuery ? '没有匹配的规则' : '暂无规则' }}</span>
    </div>

    <!-- 规则列表 -->
    <div v-else class="list-body">
      <div
        v-for="(rule, index) in store.filteredRules"
        :key="rule.name + index"
        class="rule-row"
        :class="{disabled: !rule.enabled}"
      >
        <div class="col col-status">
          <n-switch
            :value="rule.enabled"
            size="small"
            @update:value="handleToggle(rule)"
          />
        </div>
        <div class="col col-name" :title="rule.name">
          <span class="rule-name">{{ rule.name }}</span>
        </div>
        <div class="col col-action">
          <n-tag
            :type="rule.action === 'allow' ? 'success' : 'error'"
            size="tiny"
            :bordered="false"
          >
            {{ rule.action === 'allow' ? '允许' : '阻止' }}
          </n-tag>
        </div>
        <div class="col col-program" :title="rule.program">
          <span class="path-text">{{ truncatePath(rule.program, 36) }}</span>
        </div>
        <div class="col col-protocol">
          <span>{{ rule.protocol === 'any' ? '任意' : rule.protocol }}</span>
        </div>
        <div class="col col-port">
          <span>{{ rule.localPort || '-' }}</span>
        </div>
        <div class="col col-actions">
          <n-button text size="small" @click="$emit('edit', rule)" title="编辑">
            <NIcon :size="16" color="#4361ee">
              <CreateOutline/>
            </NIcon>
          </n-button>
          <n-button text size="small" @click="handleDelete(rule)" title="删除">
            <NIcon :size="16" color="#e88080">
              <TrashOutline/>
            </NIcon>
          </n-button>
        </div>
      </div>
    </div>

    <!-- 底部统计 -->
    <div v-if="!store.loading && !store.error" class="list-footer">
      共 {{ store.filteredRules.length }} 条规则
      <span v-if="store.searchQuery">（已过滤）</span>
    </div>
  </div>
</template>

<style scoped>
.rule-list {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.list-header {
  display: flex;
  align-items: center;
  padding: 10px 16px;
  background: #1a1a22;
  border-radius: 8px 8px 0 0;
  font-size: 12px;
  font-weight: 600;
  color: #888;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.list-body {
  flex: 1;
  overflow-y: auto;
}

.rule-row {
  display: flex;
  align-items: center;
  padding: 10px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.03);
  transition: background 0.12s;
  font-size: 13px;
}

.rule-row:hover {
  background: rgba(255, 255, 255, 0.03);
}

.rule-row.disabled {
  opacity: 0.45;
}

.col {
  display: flex;
  align-items: center;
}

.col-status {
  width: 50px;
  flex-shrink: 0;
}

.col-name {
  flex: 2;
  min-width: 0;
  overflow: hidden;
}

.col-action {
  width: 60px;
  flex-shrink: 0;
}

.col-program {
  flex: 2.5;
  min-width: 0;
  overflow: hidden;
}

.col-protocol {
  width: 70px;
  flex-shrink: 0;
  color: #888;
  font-size: 12px;
}

.col-port {
  width: 70px;
  flex-shrink: 0;
  color: #888;
  font-size: 12px;
}

.col-actions {
  width: 70px;
  flex-shrink: 0;
  display: flex;
  gap: 6px;
  justify-content: flex-end;
}

.rule-name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: #ddd;
}

.path-text {
  font-family: "Cascadia Code", "Consolas", monospace;
  font-size: 12px;
  color: #888;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.loading-state,
.error-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 60px 0;
  color: #666;
  font-size: 14px;
}

.list-footer {
  padding: 10px 16px;
  font-size: 12px;
  color: #555;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  background: #1a1a22;
  border-radius: 0 0 8px 8px;
}
</style>
