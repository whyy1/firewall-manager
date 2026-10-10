import {reactive} from 'vue'
import {loadLocalStatus, mutateAndRefresh, runExclusive} from './operations.js'
import {
  CloseSMBSession,
  CreateSMBShare,
  DisableSMB1,
  DiscoverSMBNetworkShares,
  GetSMBFirewallStatus,
  GetSMBOverview,
  GetSMBSessions,
  GetSMBShares,
  OpenSMBLocation,
  RemoveSMBShare,
  SetSMBFirewallEnabled,
  SetSMBServiceRunning,
} from '../../wailsjs/go/main/App'

export const smbStore = reactive({
  overview: null,
  shares: [],
  localLoading: false,
  overviewError: null,
  sharesError: null,
  localLoaded: false,
  localRequestVersion: 0,

  networkShares: [],
  networkLoading: false,
  networkError: null,
  networkLoaded: false,
  networkRequestVersion: 0,

  sessions: [],
  sessionsLoading: false,
  sessionsError: null,
  sessionsLoaded: false,
  sessionsRequestVersion: 0,

  firewall: null,
  firewallLoading: false,
  firewallError: null,
  firewallRequestVersion: 0,

  mutationAction: '',
  openingLocation: '',

  async fetchLocal() {
    const version = ++this.localRequestVersion
    this.localLoading = true
    this.overviewError = null
    this.sharesError = null
    try {
      const result = await loadLocalStatus(GetSMBOverview, GetSMBShares)
      if (version !== this.localRequestVersion) return
      if (!result.overviewError) this.overview = result.overview
      if (!result.sharesError) this.shares = result.shares
      this.localLoaded = true
      this.overviewError = result.overviewError
      this.sharesError = result.sharesError
      return [result.overviewError, result.sharesError].filter(Boolean).join('；') || null
    } finally {
      if (version === this.localRequestVersion) this.localLoading = false
    }
  },

  async scanNetwork() {
    const version = ++this.networkRequestVersion
    this.networkLoading = true
    this.networkError = null
    try {
      const shares = await DiscoverSMBNetworkShares()
      if (version !== this.networkRequestVersion) return
      this.networkShares = Array.isArray(shares) ? shares : []
      this.networkLoaded = true
    } catch (error) {
      if (version !== this.networkRequestVersion) return
      this.networkError = String(error)
    } finally {
      if (version === this.networkRequestVersion) this.networkLoading = false
    }
  },

  async fetchSessions() {
    const version = ++this.sessionsRequestVersion
    this.sessionsLoading = true
    this.sessionsError = null
    try {
      const sessions = await GetSMBSessions()
      if (version !== this.sessionsRequestVersion) return
      this.sessions = Array.isArray(sessions) ? sessions : []
      this.sessionsLoaded = true
      return null
    } catch (error) {
      if (version !== this.sessionsRequestVersion) return
      this.sessionsError = String(error)
      return this.sessionsError
    } finally {
      if (version === this.sessionsRequestVersion) this.sessionsLoading = false
    }
  },

  async fetchFirewall() {
    const version = ++this.firewallRequestVersion
    this.firewallLoading = true
    this.firewallError = null
    try {
      const status = await GetSMBFirewallStatus()
      if (version !== this.firewallRequestVersion) return
      this.firewall = status
      return null
    } catch (error) {
      if (version !== this.firewallRequestVersion) return
      this.firewallError = String(error)
      return this.firewallError
    } finally {
      if (version === this.firewallRequestVersion) this.firewallLoading = false
    }
  },

  async runMutation(key, operation) {
    const localOverview = key === 'service' || key === 'smb1'
    const localShares = key === 'create' || key.startsWith('remove:')
    const sessions = key.startsWith('session:')
    if ((localOverview && (this.localLoading || this.overviewError || !this.overview)) ||
        (localShares && (this.localLoading || this.sharesError || !this.localLoaded)) ||
        (sessions && (this.sessionsLoading || this.sessionsError || !this.sessionsLoaded)) ||
        (key === 'firewall' && (this.firewallLoading || this.firewallError || !this.firewall))) {
      throw new Error('当前数据尚未读取或刷新失败，请刷新确认最新状态后操作')
    }
    return runExclusive(this, 'mutationAction', key, operation)
  },

  async setServiceRunning(running) {
    return this.runMutation('service', () => mutateAndRefresh(
      () => SetSMBServiceRunning(running), () => this.fetchLocal()))
  },

  async disableSMB1() {
    if (this.overview?.configurationError) throw new Error('协议安全状态尚未读取，请启动服务并刷新后重试')
    return this.runMutation('smb1', () => mutateAndRefresh(
      DisableSMB1, () => this.fetchLocal()))
  },

  async createShare(request) {
    return this.runMutation('create', () => mutateAndRefresh(
      () => CreateSMBShare(request), () => this.fetchLocal()))
  },

  async removeShare(name) {
    return this.runMutation(`remove:${name}`, () => mutateAndRefresh(
      () => RemoveSMBShare(name), () => this.fetchLocal()))
  },

  async closeSession(sessionId) {
    return this.runMutation(`session:${sessionId}`, () => mutateAndRefresh(
      () => CloseSMBSession(sessionId), () => this.fetchSessions()))
  },

  async setFirewallEnabled(enabled) {
    return this.runMutation('firewall', () => mutateAndRefresh(
      () => SetSMBFirewallEnabled(enabled), () => this.fetchFirewall()))
  },

  async openLocation(location) {
    if (this.openingLocation) return
    this.openingLocation = location
    try {
      await OpenSMBLocation(location)
    } finally {
      if (this.openingLocation === location) this.openingLocation = ''
    }
  },
})
