<script setup>
import {ref} from 'vue'
import {NConfigProvider, NMessageProvider, NDialogProvider, darkTheme} from 'naive-ui'
import AppLayout from './components/AppLayout.vue'

const isDark = ref(localStorage.getItem('theme') !== 'light')

const darkOverrides = {
  common: {
    primaryColor: '#4361ee',
    primaryColorHover: '#5a7bff',
    primaryColorPressed: '#3451de',
    borderRadius: '8px',
  },
}

const lightOverrides = {
  common: {
    primaryColor: '#4361ee',
    primaryColorHover: '#5a7bff',
    primaryColorPressed: '#3451de',
    borderRadius: '8px',
    bodyColor: '#f5f5f5',
    cardColor: '#fff',
    modalColor: '#fff',
    popoverColor: '#fff',
  },
}

function toggleTheme() {
  isDark.value = !isDark.value
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}
</script>

<template>
  <NConfigProvider :theme="isDark ? darkTheme : null" :theme-overrides="isDark ? darkOverrides : lightOverrides">
    <NMessageProvider>
      <NDialogProvider>
        <AppLayout :is-dark="isDark" @toggle-theme="toggleTheme"/>
      </NDialogProvider>
    </NMessageProvider>
  </NConfigProvider>
</template>

<style>
html, body, #app {
  margin: 0;
  padding: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
}
.n-config-provider {
  width: 100%;
  height: 100%;
}
</style>
