export async function runExclusive(state, field, key, operation) {
  if (state[field]) throw new Error('已有操作正在执行，请稍候')
  state[field] = key
  try {
    return await operation()
  } finally {
    state[field] = ''
  }
}

export async function loadLocalStatus(getOverview, getShares) {
  const [overview, shares] = await Promise.allSettled([
    Promise.resolve().then(getOverview),
    Promise.resolve().then(getShares),
  ])
  return {
    overview: overview.status === 'fulfilled' ? overview.value : null,
    overviewError: overview.status === 'rejected' ? String(overview.reason) : null,
    shares: shares.status === 'fulfilled' && Array.isArray(shares.value) ? shares.value : [],
    sharesError: shares.status === 'rejected' ? String(shares.reason) : null,
  }
}

export async function mutateAndRefresh(operation, refresh) {
  await operation()
  try {
    return {refreshError: await refresh() || null}
  } catch (error) {
    return {refreshError: String(error)}
  }
}
