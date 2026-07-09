import {reactive} from 'vue'
import {GetServicePorts, ChangeServicePort} from '../../wailsjs/go/main/App'

export const portStore = reactive({
  ports: [],
  loading: false,
  error: null,

  async fetchPorts() {
    this.loading = true
    this.error = null
    try {
      this.ports = await GetServicePorts()
    } catch (e) {
      this.error = String(e)
      this.ports = []
    } finally {
      this.loading = false
    }
  },

  async changePort(serviceName, newPort) {
    await ChangeServicePort(serviceName, newPort)
    await this.fetchPorts()
  },
})
