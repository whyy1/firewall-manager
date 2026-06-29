<script setup>
import {ref} from 'vue'
import {store} from '../stores/rules.js'
import {useMessage} from 'naive-ui'

const props = defineProps({
  title: String,
  action: String,
})

const visible = ref(false)
const path = ref('')
const message = useMessage()

function open() {
  path.value = ''
  visible.value = true
}

async function handleConfirm() {
  if (!path.value) return
  try {
    if (props.action === 'block') {
      await store.blockApp(path.value)
      message.success('已成功阻止该程序联网')
    } else {
      await store.allowApp(path.value)
      message.success('已成功放行该程序联网')
    }
    visible.value = false
  } catch (e) {
    message.error('操作失败: ' + String(e))
  }
}

defineExpose({open})
</script>

<template>
  <n-modal v-model:show="visible" preset="dialog" :title="title"
           positive-text="确认" negative-text="取消"
           @positive-click="handleConfirm">
    <n-input v-model:value="path" placeholder="输入程序完整路径，如 C:\Program Files\app.exe"
             clearable style="margin-top: 12px"/>
  </n-modal>
</template>
