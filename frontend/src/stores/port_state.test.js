import test from 'node:test'
import assert from 'node:assert/strict'
import {createPortState, validatePortChange} from './port_state.js'

const service = {name: 'RDP', serviceName: 'TermService', listenPort: 3389, defaultPort: 3389, canChangePort: true}
test('port validation rejects unsupported, noninteger, same and out of range values', () => {
  for (const port of [null, 0, 65536, 1.5, 3389]) assert.ok(validatePortChange(service, port))
  assert.ok(validatePortChange({...service, canChangePort: false}, 4000))
  assert.equal(validatePortChange(service, 4000), null)
})
test('port read failure preserves the last result and blocks edits', async () => {
  let fail = false
  const state = createPortState({GetServicePorts: async () => { if (fail) throw new Error('读取失败'); return [service] }, ChangeServicePort: async () => assert.fail('unexpected write')})
  await state.fetchPorts()
  fail = true
  assert.match(await state.fetchPorts(), /读取失败/)
  assert.deepEqual(state.ports, [service])
  assert.equal(state.stale, true)
  await assert.rejects(state.changePort(service, 4000), /刷新/)
})
test('older port requests cannot overwrite the latest response', async () => {
  const resolve = []
  const state = createPortState({GetServicePorts: () => new Promise(done => resolve.push(done))})
  const first = state.fetchPorts()
  const second = state.fetchPorts()
  resolve[1]([{...service, listenPort: 4000}])
  await second
  resolve[0]([service])
  await first
  assert.equal(state.ports[0].listenPort, 4000)
})
test('port write failure retains data and releases the lock', async () => {
  const state = createPortState({GetServicePorts: async () => [service], ChangeServicePort: async () => { throw new Error('写入失败') }})
  await state.fetchPorts()
  await assert.rejects(state.changePort(service, 4000), /写入失败/)
  assert.equal(state.saving, '')
  assert.deepEqual(state.ports, [service])
})
test('port write success with refresh failure is not retried', async () => {
  let calls = 0
  const state = createPortState({GetServicePorts: async () => { if (calls) throw new Error('刷新失败'); return [service] }, ChangeServicePort: async () => { calls++ }})
  await state.fetchPorts()
  const result = await state.changePort(service, 4000)
  assert.match(result.refreshError, /刷新失败/)
  assert.equal(calls, 1)
})


test('partial configuration write blocks retry until refreshed', async () => {
  const state = createPortState({GetServicePorts: async () => [service], ChangeServicePort: async () => {throw new Error('端口配置已写入，但服务重启失败')}})
  await state.fetchPorts()
  await assert.rejects(state.changePort(service, 3390), /端口配置已写入/)
  assert.equal(state.stale, true)
  assert.match(state.error, /确认服务状态/)
  await assert.rejects(state.changePort(service, 3390), /请先刷新端口状态/)
  await state.fetchPorts()
  assert.equal(state.stale, false)
})
