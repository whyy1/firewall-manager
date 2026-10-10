import {mutateAndRefresh, runExclusive} from './operations.js'
import {ruleIdentity, requireEditableRule} from './rule_identity.js'
const text = value => String(value || '').toLowerCase()
export function createRulesState(api) {
  let requestVersion = 0
  return {
    rules: [], loading: false, saving: '', loaded: false,
    direction: 'in', nameQuery: '', portQuery: '', addrQuery: '',
    protocolFilter: 'all', actionFilter: 'all', statusFilter: 'all', error: null,
    get enabledCount() { return this.rules.filter(rule => rule.enabled).length },
    get canWrite() { return this.loaded && !this.loading && !this.error && !this.saving },
    get hasAnyFilter() {
      return !!(this.nameQuery || this.portQuery || this.addrQuery || this.protocolFilter !== 'all' || this.actionFilter !== 'all' || this.statusFilter !== 'all')
    },
    get filteredRules() {
      return this.rules.filter(rule => {
        const name = text(this.nameQuery.trim()), port = text(this.portQuery.trim()), address = text(this.addrQuery.trim())
        return (!name || text(rule.name).includes(name)) &&
          (!port || text(rule.localPort).includes(port) || text(rule.remotePort).includes(port)) &&
          (!address || text(rule.localAddr).includes(address) || text(rule.remoteAddr).includes(address)) &&
          (this.protocolFilter === 'all' || text(rule.protocol) === text(this.protocolFilter)) &&
          (this.actionFilter === 'all' || rule.action === this.actionFilter) &&
          (this.statusFilter === 'all' || rule.enabled === (this.statusFilter === 'enabled'))
      })
    },
    async fetchRules() {
      const version = ++requestVersion, direction = this.direction
      this.loading = true; this.error = null
      try {
        const rules = await api.GetRules(direction, 0)
        if (version !== requestVersion) return null
        this.rules = Array.isArray(rules) ? rules : []; this.loaded = true
        return null
      } catch (error) {
        if (version !== requestVersion) return null
        this.error = String(error); return this.error
      } finally { if (version === requestVersion) this.loading = false }
    },
    async setDirection(direction) {
      if (this.saving || direction === this.direction) return
      this.direction = direction; this.rules = []; this.loaded = false
      return this.fetchRules()
    },
    clearFilters() {
      this.nameQuery = ''; this.portQuery = ''; this.addrQuery = ''
      this.protocolFilter = 'all'; this.actionFilter = 'all'; this.statusFilter = 'all'
    },
    setNameQuery(value) { this.nameQuery = value }, setPortQuery(value) { this.portQuery = value },
    setAddrQuery(value) { this.addrQuery = value }, setProtocolFilter(value) { this.protocolFilter = value },
    setActionFilter(value) { this.actionFilter = value }, setStatusFilter(value) { this.statusFilter = value },
    async mutate(key, operation) {
      if (!this.canWrite) throw new Error('规则状态尚未就绪，请刷新后重试')
      return runExclusive(this, 'saving', key, () => mutateAndRefresh(operation, () => this.fetchRules()))
    },
    async deleteRule(rule) { requireEditableRule(rule); return this.mutate('delete', () => api.DeleteRule({...rule})) },
    async toggleRule(rule, enabled) { requireEditableRule(rule); return this.mutate('toggle', () => api.ToggleRule({...rule}, enabled)) },
    async saveRule(original, rule) {
      if (original) {
        requireEditableRule(original)
        if (ruleIdentity(original) !== ruleIdentity(rule)) throw new Error('规则身份不可修改，请刷新后重试')
      }
      return this.mutate(original ? 'update' : 'add', () => original ? api.UpdateRule({...original}, {...rule}) : api.AddRule({...rule}))
    },
    addRule(rule) { return this.saveRule(null, rule) }, updateRule(original, rule) { return this.saveRule(original, rule) },
  }
}
