import {reactive} from 'vue'
import {GetRules, DeleteRule, ToggleRule, AddRule, BlockApp, AllowApp} from '../../wailsjs/go/main/App'

export const store = reactive({
  rules: [],
  loading: false,
  direction: 'in',
  searchQuery: '',
  error: null,

  get filteredRules() {
    if (!this.searchQuery) return this.rules
    const q = this.searchQuery.toLowerCase()
    return this.rules.filter(r =>
      r.name.toLowerCase().includes(q) ||
      r.program.toLowerCase().includes(q)
    )
  },

  get enabledCount() {
    return this.rules.filter(r => r.enabled).length
  },

  async fetchRules() {
    this.loading = true
    this.error = null
    try {
      this.rules = await GetRules(this.direction)
    } catch (e) {
      this.error = String(e)
      this.rules = []
    } finally {
      this.loading = false
    }
  },

  async deleteRule(name) {
    await DeleteRule(name)
    await this.fetchRules()
  },

  async toggleRule(name, enabled) {
    await ToggleRule(name, enabled)
    await this.fetchRules()
  },

  async addRule(rule) {
    await AddRule(rule)
    await this.fetchRules()
  },

  async blockApp(programPath) {
    await BlockApp(programPath)
    await this.fetchRules()
  },

  async allowApp(programPath) {
    await AllowApp(programPath)
    await this.fetchRules()
  },

  setDirection(dir) {
    this.direction = dir
    this.fetchRules()
  },

  setSearch(query) {
    this.searchQuery = query
  }
})
