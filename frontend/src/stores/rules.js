import {reactive} from 'vue'
import {GetRules, DeleteRule, ToggleRule, AddRule, GetFirewallStatus, SetFirewallEnabled, ResetFirewall} from '../../wailsjs/go/main/App'

export const store = reactive({
  rules: [],
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
  firewallOn: false,

  get filteredRules() {
    let r = this.rules
    if (this.nameQuery) { const q = this.nameQuery.toLowerCase(); r = r.filter(x => x.name.toLowerCase().includes(q)) }
    if (this.portQuery) { const q = this.portQuery.toLowerCase(); r = r.filter(x => x.localPort.toLowerCase().includes(q) || x.remotePort.toLowerCase().includes(q)) }
    if (this.addrQuery) { const q = this.addrQuery.toLowerCase(); r = r.filter(x => x.localAddr.toLowerCase().includes(q) || x.remoteAddr.toLowerCase().includes(q)) }
    if (this.protocolFilter !== 'all') { r = r.filter(x => x.protocol.toLowerCase() === this.protocolFilter.toLowerCase()) }
    if (this.actionFilter !== 'all') { r = r.filter(x => x.action === this.actionFilter) }
    if (this.statusFilter === 'enabled') { r = r.filter(x => x.enabled) }
    else if (this.statusFilter === 'disabled') { r = r.filter(x => !x.enabled) }
    return r
  },

  get hasAnyFilter() {
    return this.nameQuery || this.portQuery || this.addrQuery ||
      this.protocolFilter !== 'all' || this.actionFilter !== 'all' || this.statusFilter !== 'all'
  },

  get enabledCount() { return this.rules.filter(r => r.enabled).length },

  async fetchRules(limit = 10) {
    this.loading = true; this.error = null
    try { this.rules = await GetRules(this.direction, limit); this.partial = limit > 0 }
    catch (e) { this.error = String(e); this.rules = [] }
    finally { this.loading = false }
  },

  async fetchAllRules() { await this.fetchRules(0) },

  async deleteRule(name) { await DeleteRule(name); await this.fetchRules(this.partial ? 10 : 0) },
  async toggleRule(name, enabled) { await ToggleRule(name, enabled); await this.fetchRules(this.partial ? 10 : 0) },
  async addRule(rule) { await AddRule(rule); await this.fetchRules(this.partial ? 10 : 0) },

  async fetchFirewallStatus() { this.firewallOn = await GetFirewallStatus() },
  async toggleFirewall() { await SetFirewallEnabled(!this.firewallOn); this.firewallOn = !this.firewallOn },
  async resetFirewall() { await ResetFirewall(); await this.fetchRules(10) },

  setDirection(dir) { this.direction = dir; this.fetchRules(10) },
  setNameQuery(q) { this.nameQuery = q },
  setPortQuery(q) { this.portQuery = q },
  setAddrQuery(q) { this.addrQuery = q },
  setProtocolFilter(v) { this.protocolFilter = v },
  setActionFilter(v) { this.actionFilter = v },
  setStatusFilter(v) { this.statusFilter = v },
})
