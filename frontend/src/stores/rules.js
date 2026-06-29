import {reactive} from 'vue'
import {GetRules, DeleteRule, ToggleRule, AddRule, BlockApp, AllowApp} from '../../wailsjs/go/main/App'

export const store = reactive({
  rules: [],
  loading: false,
  direction: 'in',
  searchQuery: '',
  actionFilter: 'all',
  statusFilter: 'all',
  error: null,
  partial: true, // 是否只加载了部分规则
  totalCount: 0,

  get filteredRules() {
    let result = this.rules

    // 搜索过滤（名称、程序、端口、地址）
    if (this.searchQuery) {
      const q = this.searchQuery.toLowerCase()
      result = result.filter(r =>
        r.name.toLowerCase().includes(q) ||
        r.localAddr.toLowerCase().includes(q) ||
        r.remoteAddr.toLowerCase().includes(q) ||
        r.localPort.toLowerCase().includes(q) ||
        r.remotePort.toLowerCase().includes(q) ||
        r.protocol.toLowerCase().includes(q)
      )
    }

    // 动作过滤
    if (this.actionFilter !== 'all') {
      result = result.filter(r => r.action === this.actionFilter)
    }

    // 状态过滤
    if (this.statusFilter === 'enabled') {
      result = result.filter(r => r.enabled)
    } else if (this.statusFilter === 'disabled') {
      result = result.filter(r => !r.enabled)
    }

    return result
  },

  get enabledCount() {
    return this.rules.filter(r => r.enabled).length
  },

  async fetchRules(limit = 10) {
    this.loading = true
    this.error = null
    try {
      this.rules = await GetRules(this.direction, limit)
      this.partial = limit > 0
    } catch (e) {
      this.error = String(e)
      this.rules = []
    } finally {
      this.loading = false
    }
  },

  async fetchAllRules() {
    await this.fetchRules(0)
  },

  async deleteRule(name) {
    await DeleteRule(name)
    await this.fetchRules(this.partial ? 10 : 0)
  },

  async toggleRule(name, enabled) {
    await ToggleRule(name, enabled)
    await this.fetchRules(this.partial ? 10 : 0)
  },

  async addRule(rule) {
    await AddRule(rule)
    await this.fetchRules(this.partial ? 10 : 0)
  },

  async blockApp(programPath) {
    await BlockApp(programPath)
    await this.fetchRules(this.partial ? 10 : 0)
  },

  async allowApp(programPath) {
    await AllowApp(programPath)
    await this.fetchRules(this.partial ? 10 : 0)
  },

  setDirection(dir) {
    this.direction = dir
    this.fetchRules(10)
  },

  setSearch(query) {
    this.searchQuery = query
  },

  setActionFilter(val) {
    this.actionFilter = val
  },

  setStatusFilter(val) {
    this.statusFilter = val
  }
})
