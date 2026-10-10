import {reactive} from 'vue'
import * as api from '../../wailsjs/go/main/App'
import {createRulesState} from './rules_state.js'

export const store = reactive(createRulesState(api))
