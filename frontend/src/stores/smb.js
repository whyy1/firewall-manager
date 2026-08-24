import {reactive} from 'vue'
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
  localError: null,
  localRequestVersion: 0,

  networkShares: [],
  networkLoading: false,
  networkError: null,
  networkRequestVersion: 0,

  sessions: [],
  sessionsLoading: false,
  sessionsError: null,
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
    this.localError = null
    try {
      const [overview, shares] = await Promise.all([
        GetSMBOverview(),
        GetSMBShares(),
      ])
      if (version !== this.localRequestVersion) return
      this.overview = overview
      this.shares = Array.isArray(shares) ? shares : []
    } catch (error) {
      if (version !== this.localRequestVersion) return
      this.localError = String(error)
      this.overview = null
      this.shares = []
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
    } catch (error) {
      if (version !== this.networkRequestVersion) return
      this.networkError = String(error)
      this.networkShares = []
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
    } catch (error) {
      if (version !== this.sessionsRequestVersion) return
      this.sessionsError = String(error)
      this.sessions = []
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
    } catch (error) {
      if (version !== this.firewallRequestVersion) return
      this.firewallError = String(error)
      this.firewall = null
    } finally {
      if (version === this.firewallRequestVersion) this.firewallLoading = false
    }
  },

  async runMutation(key, operation) {
    if (this.mutationAction) throw new Error('已有 SMB 操作正在执行，请稍候')
    this.mutationAction = key
    try {
      return await operation()
    } finally {
      if (this.mutationAction === key) this.mutationAction = ''
    }
  },

  async setServiceRunning(running) {
    return this.runMutation('service', async () => {
      await SetSMBServiceRunning(running)
      await this.fetchLocal()
    })
  },

  async disableSMB1() {
    return this.runMutation('smb1', async () => {
      await DisableSMB1()
      await this.fetchLocal()
    })
  },

  async createShare(request) {
    return this.runMutation('create', async () => {
      await CreateSMBShare(request)
      await this.fetchLocal()
    })
  },

  async removeShare(name) {
    return this.runMutation(`remove:${name}`, async () => {
      await RemoveSMBShare(name)
      await this.fetchLocal()
    })
  },

  async closeSession(sessionId) {
    return this.runMutation(`session:${sessionId}`, async () => {
      await CloseSMBSession(sessionId)
      await this.fetchSessions()
    })
  },

  async setFirewallEnabled(enabled) {
	return this.runMutation('firewall', async () => {
	  await SetSMBFirewallEnabled(enabled)
      await this.fetchFirewall()
    })
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
