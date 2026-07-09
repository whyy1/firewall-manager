import {reactive} from 'vue'
import {GetRules, DeleteRule, ToggleRule, AddRule} from '../../wailsjs/go/main/App'

export const store = reactive({
  rules: [],           // 当前显示的规则（可能是全部或前10条）
  allRules: null,      // 缓存的全部规则（搜索时用）
  loading: false,
  direction: 'in',
  nameQuery: '',
  portQuery: '',
  addrQuery: '',
  protocolFilter: 'all',
  actionFilter: 'all',
  statusFilter: 'all',
  error: null,
  partial: true,

  get filteredRules() {
    // 如果有搜索条件，从缓存的全部规则中过滤
    const source = this.hasAnyQuery && this.allRules ? this.allRules : this.rules
    let r = source

    if (this.nameQuery) { const q = this.nameQuery.toLowerCase(); r = r.filter(x => x.name.toLowerCase().includes(q)) }
    if (this.portQuery) { const q = this.portQuery.toLowerCase(); r = r.filter(x => x.localPort.toLowerCase().includes(q) || x.remotePort.toLowerCase().includes(q)) }
    if (this.addrQuery) { const q = this.addrQuery.toLowerCase(); r = r.filter(x => x.localAddr.toLowerCase().includes(q) || x.remoteAddr.toLowerCase().includes(q)) }
    if (this.protocolFilter !== 'all') { r = r.filter(x => x.protocol.toLowerCase() === this.protocolFilter.toLowerCase()) }
    if (this.actionFilter !== 'all') { r = r.filter(x => x.action === this.actionFilter) }
    if (this.statusFilter === 'enabled') { r = r.filter(x => x.enabled) }
    else if (this.statusFilter === 'disabled') { r = r.filter(x => !x.enabled) }
    return r
  },

  get hasAnyQuery() {
    return this.nameQuery || this.portQuery || this.addrQuery
  },

  get hasAnyFilter() {
    return this.nameQuery || this.portQuery || this.addrQuery ||
      this.protocolFilter !== 'all' || this.actionFilter !== 'all' || this.statusFilter !== 'all'
  },

  async fetchRules(limit = 10) {
    this.loading = true; this.error = null
    try { this.rules = await GetRules(this.direction, limit); this.partial = limit > 0 }
    catch (e) { this.error = String(e); this.rules = [] }
    finally { this.loading = false }
  },

  async fetchAllRules() {
    this.loading = true; this.error = null
    try {
      const all = await GetRules(this.direction, 0)
      this.allRules = all
      this.rules = all
      this.partial = false
    } catch (e) { this.error = String(e); this.rules = [] }
    finally { this.loading = false }
  },

  // 加载全部规则到缓存（搜索时用，不显示 loading）
  async ensureAllRules() {
    if (this.allRules) return
    try { this.allRules = await GetRules(this.direction, 0) }
    catch (e) { this.allRules = null }
  },

  async deleteRule(name) { await DeleteRule(name); this.allRules = null; await this.fetchRules(this.partial ? 10 : 0) },
  async toggleRule(name, enabled) { await ToggleRule(name, enabled); this.allRules = null; await this.fetchRules(this.partial ? 10 : 0) },
  async addRule(rule) { await AddRule(rule); this.allRules = null; await this.fetchRules(this.partial ? 10 : 0) },

  setDirection(dir) {
    this.direction = dir
    this.allRules = null  // 方向变化时清空缓存
    this.fetchRules(10)
  },

  async setNameQuery(q) {
    this.nameQuery = q
    if (q && this.partial) await this.ensureAllRules()
  },
  async setPortQuery(q) {
    this.portQuery = q
    if (q && this.partial) await this.ensureAllRules()
  },
  async setAddrQuery(q) {
    this.addrQuery = q
    if (q && this.partial) await this.ensureAllRules()
  },
  setProtocolFilter(v) { this.protocolFilter = v },
  setActionFilter(v) { this.actionFilter = v },
  setStatusFilter(v) { this.statusFilter = v },
})
