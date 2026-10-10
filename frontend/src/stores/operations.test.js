import test from 'node:test'
import assert from 'node:assert/strict'
import {loadLocalStatus, mutateAndRefresh, runExclusive} from './operations.js'

test('rejects duplicate mutations without executing the second command', async () => {
  const state = {saving: ''}
  let finish
  let calls = 0
  const first = runExclusive(state, 'saving', 'update', async () => {
    calls++
    await new Promise(resolve => { finish = resolve })
  })
  await assert.rejects(runExclusive(state, 'saving', 'update', async () => { calls++ }), /已有操作/)
  assert.equal(calls, 1)
  finish()
  await first
  assert.equal(state.saving, '')
})

test('command failure releases the mutation lock and does not refresh', async () => {
  const state = {saving: ''}
  let refreshed = false
  await assert.rejects(runExclusive(state, 'saving', 'update', () => mutateAndRefresh(
    async () => { throw new Error('写入失败') },
    async () => { refreshed = true })), /写入失败/)
  assert.equal(refreshed, false)
  assert.equal(state.saving, '')
})

test('share query failure preserves service overview', async () => {
  const overview = {serviceRunning: false}
  const result = await loadLocalStatus(async () => overview, async () => { throw new Error('共享读取失败') })
  assert.deepEqual(result.overview, overview)
  assert.equal(result.overviewError, null)
  assert.match(result.sharesError, /共享读取失败/)
})

test('overview query failure preserves a successful share list', async () => {
  const shares = [{name: 'TeamFiles'}]
  const result = await loadLocalStatus(() => { throw new Error('概览读取失败') }, async () => shares)
  assert.equal(result.overview, null)
  assert.match(result.overviewError, /概览读取失败/)
  assert.deepEqual(result.shares, shares)
  assert.equal(result.sharesError, null)
})

test('successful mutation reports partial refresh failure without retrying the command', async () => {
  let calls = 0
  const result = await mutateAndRefresh(async () => { calls++ }, async () => {
    const status = await loadLocalStatus(async () => ({serviceRunning: true}), async () => { throw new Error('列表读取失败') })
    return status.sharesError
  })
  assert.equal(calls, 1)
  assert.match(result.refreshError, /列表读取失败/)
})

test('thrown refresh failures remain successful mutation results', async () => {
  const result = await mutateAndRefresh(async () => {}, async () => { throw new Error('刷新失败') })
  assert.match(result.refreshError, /刷新失败/)
  assert.deepEqual(await mutateAndRefresh(async () => {}, async () => null), {refreshError: null})
})
