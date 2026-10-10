<script setup>
import {ref, watch} from 'vue'
import {NConfigProvider, NMessageProvider, NDialogProvider, darkTheme} from 'naive-ui'
import AppLayout from './components/AppLayout.vue'

const isDark = ref(localStorage.getItem('theme') !== 'light')
watch(isDark, value => { document.documentElement.dataset.theme = value ? 'dark' : 'light' }, {immediate: true})
const common = {primaryColorHover: '#5a7bff', primaryColorPressed: '#3451de', borderRadius: '8px'}
const darkOverrides = {common: {...common, primaryColor: '#7a9bff', bodyColor: '#0e131e', cardColor: '#151c2a', modalColor: '#151c2a', popoverColor: '#1a2435', tableColor: '#151c2a', textColorBase: '#e6edf7'}}
const lightOverrides = {common: {...common, primaryColor: '#365fd1', bodyColor: '#f3f6fb', cardColor: '#fff', modalColor: '#fff', popoverColor: '#fff'}}
function toggleTheme() {
  isDark.value = !isDark.value
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}
</script>

<template>
  <NConfigProvider :theme="isDark ? darkTheme : null" :theme-overrides="isDark ? darkOverrides : lightOverrides">
    <NMessageProvider><NDialogProvider><AppLayout :is-dark="isDark" @toggle-theme="toggleTheme"/></NDialogProvider></NMessageProvider>
  </NConfigProvider>
</template>

<style>
.n-config-provider { width: 100%; height: 100%; }
</style>
