import {mutateAndRefresh, runExclusive} from './operations.js'

export function validatePortChange(service, port) {
  if (!service?.canChangePort) return '此服务仅支持查看'
  if (!Number.isInteger(port) || port < 1 || port > 65535) return '请输入 1–65535 的整数端口'
  if (service.listenPort > 0 && port === service.listenPort) return '新端口与检测到的端口相同，无需重复提交'
  return null
}

export function createPortState(api) {
  return {
    ports: [], loading: false, loaded: false, stale: true, error: null, requestVersion: 0, saving: '',
    async fetchPorts() {
      const version = ++this.requestVersion
      this.loading = true
      try {
        const result = await api.GetServicePorts()
        if (version !== this.requestVersion) return null
        this.ports = Array.isArray(result) ? result : []
        this.loaded = true
        this.stale = false
        this.error = null
        return null
      } catch (error) {
        if (version !== this.requestVersion) return null
        this.error = String(error)
        this.stale = true
        return this.error
      } finally {
        if (version === this.requestVersion) this.loading = false
      }
    },
    async changePort(service, port) {
      const error = validatePortChange(service, port)
      if (error) throw new Error(error)
      if (this.stale || this.loading) throw new Error('请先刷新端口状态')
      return runExclusive(this, 'saving', service.serviceName, () => mutateAndRefresh(
        async () => {
          try { await api.ChangeServicePort(service.serviceName, port) }
          catch (error) {
            if (String(error).includes('端口配置已写入')) {
              this.stale = true
              this.error = '端口配置可能已改变，请确认服务状态并刷新后再操作: ' + String(error)
            }
            throw error
          }
        }, () => this.fetchPorts()))
    },
  }
}
