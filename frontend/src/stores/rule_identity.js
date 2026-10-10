export const ruleIdentity = rule => JSON.stringify([rule?.policyStoreSourceType || '', rule?.policyStoreSource || '', rule?.id || ''])
export function ruleReadOnlyReason(rule) {
  if (rule?.canEdit === true && rule?.id && rule?.policyStoreSource && rule?.policyStoreSourceType === 'Local') return ''
  return rule?.readOnlyReason || '规则身份、策略来源或可编辑状态未确认，请刷新后查看'
}
export function requireEditableRule(rule) {
  const reason = ruleReadOnlyReason(rule)
  if (reason) throw new Error(reason)
}
