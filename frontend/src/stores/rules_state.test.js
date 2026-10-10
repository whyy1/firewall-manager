import test from 'node:test'
import assert from 'node:assert/strict'
import {createRulesState} from './rules_state.js'
import {ruleToForm, normalizeProfiles, selectProfiles, supportsPorts, validateRuleForm} from './rule_form.js'
import {ruleIdentity} from './rule_identity.js'
const rule = {id: 'id-local', policyStoreSource: 'PersistentStore', policyStoreSourceType: 'Local', canEdit: true, name: '规则', direction: 'in', protocol: 'tcp', action: 'allow', enabled: true, localPort: '445'}
const deferred = () => { let resolve; const promise = new Promise(yes => { resolve = yes }); return {promise, resolve} }
test('full dataset filters and accurate counts', async () => {
  const state = createRulesState({GetRules: async (direction, limit) => { assert.equal(limit, 0); return Array.from({length: 80}, (_, i) => ({...rule, name: `规则 ${i}`, enabled: i % 2 === 0})) }})
  await state.fetchRules(); assert.equal(state.enabledCount, 40)
  state.nameQuery = '79'; assert.equal(state.filteredRules.length, 1)
  state.clearFilters(); state.statusFilter = 'disabled'; assert.equal(state.filteredRules.length, 40)
})
test('old response cannot replace new direction', async () => {
  const first = deferred(), second = deferred()
  const state = createRulesState({GetRules: direction => direction === 'in' ? first.promise : second.promise})
  const pending = state.fetchRules(), switched = state.setDirection('out')
  second.resolve([{...rule, name: 'new'}]); await switched
  first.resolve([rule]); await pending
  assert.equal(state.rules[0].name, 'new'); assert.equal(state.loading, false)
})
test('refresh failure preserves data and pauses writes', async () => {
  let failed = false
  const state = createRulesState({GetRules: async () => { if (failed) throw new Error('query failed'); return [rule] }})
  await state.fetchRules(); failed = true; await state.fetchRules()
  assert.equal(state.rules.length, 1); assert.equal(state.canWrite, false)
  await assert.rejects(state.deleteRule(rule), /尚未就绪/)
})
test('selected object passed intact and duplicate writes blocked', async () => {
  const write = deferred(); let received
  const state = createRulesState({GetRules: async () => [rule], ToggleRule: selected => { received = selected; return write.promise }})
  await state.fetchRules(); const pending = state.toggleRule(rule, false)
  await assert.rejects(state.deleteRule(rule), /尚未就绪/)
  assert.deepEqual(received, rule); assert.notEqual(received, rule)
  write.resolve(); await pending; assert.equal(state.saving, '')
})
test('write success reports refresh failure separately', async () => {
  let reads = 0
  const state = createRulesState({GetRules: async () => { if (++reads > 1) throw new Error('refresh failed'); return [rule] }, AddRule: async () => {}})
  await state.fetchRules(); const result = await state.addRule(rule)
  assert.match(result.refreshError, /refresh failed/); assert.equal(state.rules.length, 1)
})
test('failed write keeps original data and releases lock', async () => {
  const state = createRulesState({GetRules: async () => [rule], UpdateRule: async () => { throw new Error('write failed') }})
  await state.fetchRules(); await assert.rejects(state.updateRule(rule, {...rule, name: 'updated'}), /write failed/)
  assert.deepEqual(state.rules, [rule]); assert.equal(state.saving, '')
})
test('localized values normalize and custom protocol retained', () => {
  const form = ruleToForm({...rule, program: '任意', localPort: 'Any', protocol: 'ICMPv4:8,any', profile: '域,专用'}, 'in')
  assert.equal(form.program, ''); assert.equal(form.localPort, ''); assert.equal(form.protocol, 'icmpv4:8,any')
  assert.deepEqual(form.profile, ['domain', 'private']); assert.deepEqual(normalizeProfiles('Any'), ['any'])
  assert.deepEqual(selectProfiles(['any', 'public'], ['any']), ['public'])
  assert.deepEqual(selectProfiles(['private', 'any'], ['private']), ['any'])
})
test('invalid ports and protocol combinations rejected', () => {
  const form = ruleToForm(rule, 'in')
  assert.equal(supportsPorts('6'), true); assert.equal(validateRuleForm(form), null)
  assert.match(validateRuleForm({...form, localPort: '65536'}), /端口必须/)
  assert.match(validateRuleForm({...form, localPort: '200-100'}), /端口必须/)
  assert.match(validateRuleForm({...form, protocol: 'any'}), /只有 TCP/)
})

test('RPC family requires local TCP port including mixed lists', () => {
  const form = ruleToForm(rule, 'in')
  for (const value of ['RPC', 'RPC-EPMap', 'IPHTTPS', '445,RPC']) {
    assert.equal(validateRuleForm({...form, localPort: value}), null)
    assert.match(validateRuleForm({...form, protocol: 'udp', localPort: value}), /仅支持 TCP 本地端口/)
    assert.match(validateRuleForm({...form, remotePort: value}), /仅支持 TCP 本地端口/)
  }
  assert.equal(validateRuleForm({...form, protocol: '6', localPort: 'RPC'}), null)
  assert.equal(validateRuleForm({...form, localPort: '80，443'}), null)
})

test('same internal ID in different sources remains distinct and form preserves identity', () => {
  const policyRule = {...rule, policyStoreSource: 'example.test/policy', policyStoreSourceType: 'GroupPolicy', canEdit: false, readOnlyReason: '组策略规则仅查看'}
  assert.notEqual(ruleIdentity(rule), ruleIdentity(policyRule))
  const form = ruleToForm(policyRule, 'in')
  for (const field of ['id', 'policyStoreSource', 'policyStoreSourceType', 'canEdit', 'readOnlyReason']) assert.equal(form[field], policyRule[field])
})

test('readonly unknown identity and changed provenance cannot bypass state write guards', async () => {
  const received = []
  const state = createRulesState({GetRules: async () => [rule], DeleteRule: async value => received.push(value), ToggleRule: async value => received.push(value), UpdateRule: async value => received.push(value)})
  await state.fetchRules()
  for (const selected of [{...rule, canEdit: false, readOnlyReason: '策略规则仅查看'}, {...rule, id: ''}, {...rule, policyStoreSource: ''}, {...rule, policyStoreSourceType: 'GroupPolicy'}]) {
    await assert.rejects(state.deleteRule(selected))
    await assert.rejects(state.toggleRule(selected, false))
    await assert.rejects(state.updateRule(selected, selected))
  }
  await assert.rejects(state.updateRule(rule, {...rule, policyStoreSource: 'different'}), /身份不可修改/)
  assert.deepEqual(received, [])
  assert.equal(state.saving, '')
})

test('adding duplicate display name still uses add API', async () => {
  let received
  const state = createRulesState({GetRules: async () => [rule], AddRule: async value => { received = value }})
  await state.fetchRules()
  await state.addRule(ruleToForm(rule, 'in'))
  assert.equal(received.name, rule.name)
})
