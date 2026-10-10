import test from 'node:test'
import assert from 'node:assert/strict'
import {readFile} from 'node:fs/promises'
import {loadLocalStatus, mutateAndRefresh, runExclusive} from './operations.js'

// Exercise the actual store with isolated APIs; never call Windows bindings.
const source = await readFile(new URL('./smb.js', import.meta.url), 'utf8')
const isolatedSource = source
  .replace(/import \{reactive\} from 'vue'\s*/, '')
  .replace(/import \{loadLocalStatus, mutateAndRefresh, runExclusive\} from '.\/operations.js'\s*/, '')
  .replace(/import \{([\s\S]*?)\} from '\.\.\/\.\.\/wailsjs\/go\/main\/App'/, 'const {$1} = api')
  .replace('export const smbStore', 'const smbStore')
const createStore = new Function('reactive', 'loadLocalStatus', 'mutateAndRefresh', 'runExclusive', 'api', isolatedSource + '\nreturn smbStore')
const mockStore = api => createStore(value => value, loadLocalStatus, mutateAndRefresh, runExclusive, api)

test('SMB partial refresh preserves prior data independently', async () => {
  const store = mockStore({
    GetSMBOverview: async () => ({serviceRunning: true}),
    GetSMBShares: async () => { throw new Error('列表失败') },
  })
  store.shares = [{name: '资料'}]
  const error = await store.fetchLocal()
  assert.equal(store.overview.serviceRunning, true)
  assert.deepEqual(store.shares, [{name: '资料'}])
  assert.match(error, /列表失败/)
  assert.equal(store.localLoading, false)
  await assert.rejects(store.createShare({name: '新共享'}), /刷新失败/)
})

test('SMB successful empty network scan differs from initial and failure states', async () => {
  let fail = false
  const store = mockStore({DiscoverSMBNetworkShares: async () => {
    if (fail) throw new Error('检测失败')
    return []
  }})
  assert.equal(store.networkLoaded, false)
  await store.scanNetwork()
  assert.equal(store.networkLoaded, true)
  assert.equal(store.networkError, null)
  fail = true
  await store.scanNetwork()
  assert.match(store.networkError, /检测失败/)
  assert.equal(store.networkLoaded, true)
})

test('SMB stale session data remains visible and cannot be disconnected', async () => {
  let writes = 0
  const store = mockStore({
    GetSMBSessions: async () => { throw new Error('会话读取失败') },
    CloseSMBSession: async () => { writes++ },
  })
  store.sessions = [{sessionId: '12'}]
  store.sessionsLoaded = true
  await store.fetchSessions()
  assert.deepEqual(store.sessions, [{sessionId: '12'}])
  await assert.rejects(store.closeSession('12'), /刷新失败/)
  assert.equal(writes, 0)
})

test('SMB request versions reject obsolete responses', async () => {
  const pending = []
  const store = mockStore({DiscoverSMBNetworkShares: () => new Promise(resolve => pending.push(resolve))})
  const first = store.scanNetwork()
  const second = store.scanNetwork()
  pending[1]([{name: '最新'}])
  await second
  pending[0]([{name: '旧结果'}])
  await first
  assert.deepEqual(store.networkShares, [{name: '最新'}])
  assert.equal(store.networkLoading, false)
})

test('SMB successful write with failed refresh is not retried', async () => {
  let writes = 0
  const store = mockStore({
    SetSMBFirewallEnabled: async () => { writes++ },
    GetSMBFirewallStatus: async () => { throw new Error('刷新失败') },
  })
  store.firewall = {available: true, profiles: []}
  const result = await store.setFirewallEnabled(true)
  assert.match(result.refreshError, /刷新失败/)
  assert.equal(writes, 1)
  assert.equal(store.firewall.available, true)
  assert.equal(store.mutationAction, '')
  await assert.rejects(store.setFirewallEnabled(true), /刷新失败/)
  assert.equal(writes, 1)
})

test('SMB duplicate share submissions execute once and preserve failure state', async () => {
  let failWrite
  let writes = 0
  const store = mockStore({CreateSMBShare: () => {
    writes++
    return new Promise((resolve, reject) => { failWrite = reject })
  }})
  store.localLoaded = true
  const first = store.createShare({name: '资料'})
  await assert.rejects(store.createShare({name: '资料'}), /已有操作/)
  assert.equal(writes, 1)
  failWrite(new Error('创建失败'))
  await assert.rejects(first, /创建失败/)
  assert.equal(store.mutationAction, '')
  assert.equal(store.localLoading, false)
})
