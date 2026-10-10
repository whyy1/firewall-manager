export function normalizeAny(value) {
  const text = String(value || '').trim()
  return /^(any|任意|任何|所有|全部|\*)$/i.test(text) ? '' : text
}
export function normalizeProtocol(value) { return normalizeAny(value).toLowerCase() || 'any' }
export function normalizeProfiles(value) {
  const aliases = {域: 'domain', 专用: 'private', 公用: 'public', 域配置文件: 'domain', 专用配置文件: 'private', 公用配置文件: 'public'}
  const values = String(value || 'any').split(/[,，]/).map(part => {
    const text = part.trim().toLowerCase()
    return aliases[text] || (normalizeAny(text) || 'any')
  })
  return values.includes('any') ? ['any'] : [...new Set(values)]
}
export function selectProfiles(next, previous) {
  if (!next.length) return ['any']
  if (!next.includes('any')) return next
  return previous.includes('any') ? next.filter(value => value !== 'any') : ['any']
}
export function ruleToForm(rule, direction) {
  return {
    id: rule?.id || '', policyStoreSource: rule?.policyStoreSource || '', policyStoreSourceType: rule?.policyStoreSourceType || '',
    canEdit: rule?.canEdit === true, readOnlyReason: rule?.readOnlyReason || '',
    icmpType: rule?.icmpType || '', name: rule?.name || '', direction: rule?.direction || direction, action: rule?.action || 'allow',
    program: normalizeAny(rule?.program), localAddr: normalizeAny(rule?.localAddr), remoteAddr: normalizeAny(rule?.remoteAddr),
    protocol: normalizeProtocol(rule?.protocol), localPort: normalizeAny(rule?.localPort), remotePort: normalizeAny(rule?.remotePort),
    profile: normalizeProfiles(rule?.profile), enabled: rule?.enabled ?? true,
  }
}
export function supportsPorts(protocol) { return ['tcp', 'udp', '6', '17'].includes(normalizeProtocol(protocol)) }
export function validateRuleForm(form) {
  if (!form.name.trim()) return '请输入规则名称'
  if (!supportsPorts(form.protocol) && (normalizeAny(form.localPort) || normalizeAny(form.remotePort))) return '只有 TCP 或 UDP 协议支持端口，请清空端口或选择对应协议'
  for (const [field, port] of [['localPort', form.localPort], ['remotePort', form.remotePort]]) {
    const value = normalizeAny(port)
    if (!value) continue
    for (const part of value.replaceAll('，', ',').split(',').map(item => item.trim())) {
      if (/^(rpc|rpc-epmap|iphttps)$/i.test(part)) {
        if (field !== 'localPort' || !['tcp', '6'].includes(normalizeProtocol(form.protocol))) return 'RPC、RPC-EPMap 和 IPHTTPS 仅支持 TCP 本地端口'
        continue
      }
      const match = /^(\d{1,5})(?:-(\d{1,5}))?$/.exec(part)
      if (!match || Number(match[1]) > 65535 || match[2] && (Number(match[2]) < Number(match[1]) || Number(match[2]) > 65535)) return '端口必须是 0–65535 的整数、范围或逗号分隔列表'
    }
  }
  return null
}
