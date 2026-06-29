<script setup>
import {NConfigProvider, NMessageProvider, NDialogProvider, darkTheme} from 'naive-ui'
import AppLayout from './components/AppLayout.vue'
</script>

<template>
  <NConfigProvider :theme="darkTheme" :theme-overrides="themeOverrides">
    <NMessageProvider>
      <NDialogProvider>
        <AppLayout/>
      </NDialogProvider>
    </NMessageProvider>
  </NConfigProvider>
</template>

<script>
export default {
  data() {
    return {
      themeOverrides: {
        common: {
          primaryColor: '#4361ee',
          primaryColorHover: '#5a7bff',
          primaryColorPressed: '#3451de',
          bodyColor: '#121218',
          cardColor: '#1a1a24',
          modalColor: '#1e1e2a',
          popoverColor: '#1e1e2a',
          borderRadius: '8px',
        },
      }
    }
  }
}
</script>

<style>
/* 覆盖 Naive UI 默认字体 */
.n-config-provider {
  width: 100%;
  height: 100%;
}
</style>
