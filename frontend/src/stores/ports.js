import {reactive} from 'vue'
import {GetServicePorts, ChangeServicePort} from '../../wailsjs/go/main/App'
import {createPortState} from './port_state.js'

export const portStore = reactive(createPortState({GetServicePorts, ChangeServicePort}))
