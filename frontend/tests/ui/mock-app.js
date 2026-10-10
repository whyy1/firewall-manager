export async function mockApp(page, options = {}) {
  await page.addInitScript(options => {
    localStorage.setItem('theme', options.theme || 'dark')
    const rule = {
      policyStoreSource: 'PersistentStore', policyStoreSourceType: 'Local', canEdit: true, readOnlyReason: '',
      name: '允许文件共享与团队协作的中文规则', direction: 'in', action: 'allow', enabled: true,
      program: 'Any', localAddr: '任何', remoteAddr: '本地子网', protocol: 'TCP',
      localPort: '445', remotePort: 'Any', profile: 'Domain,Private',
    }
    const rules = Array.from({length: options.ruleCount || 12}, (_, index) => ({
      ...rule, id: 'rule-' + index, name: index === 0 ? rule.name : '测试规则 ' + index,
      protocol: index % 2 ? 'UDP' : 'TCP', localPort: index % 2 ? '53' : '445', enabled: index % 3 !== 2,
    }))
    const state = window.__mock = {calls: [], failures: {}, delays: {}, holds: {}, pending: {}, options, rules}
    const readValues = {
      IsAdmin: options.admin !== false,
      GetServicePorts: [
        {name: 'RDP 远程桌面', serviceName: 'TermService', running: true, listenPort: 3389, defaultPort: 3389, protocol: 'tcp', description: 'Windows 远程桌面连接', canChangePort: true},
        {name: 'WinRM 远程管理', serviceName: 'WinRM', running: true, listenPort: 5985, defaultPort: 5985, protocol: 'tcp', description: 'Windows 远程管理', canChangePort: true},
        {name: 'SMB 文件共享', serviceName: 'LanmanServer', running: true, listenPort: 445, defaultPort: 445, protocol: 'tcp', description: 'Windows SMB TCP 端口固定为 445，不能直接修改', canChangePort: false},
        {name: 'DNS 域名解析', serviceName: 'Dnscache', running: true, listenPort: 0, defaultPort: 53, protocol: 'udp', description: '域名系统解析服务', canChangePort: false},
      ],
      GetSMBOverview: {serviceRunning: true, serviceStatus: 'Running', startupType: 'Automatic', smb1Enabled: true, smb2Enabled: true, requireSecuritySignature: true, encryptData: false},
      GetSMBShares: options.emptyShares ? [] : [
        {name: 'ADMIN$', path: 'C:\\Windows', networkPath: '\\\\DESKTOP\\ADMIN$', special: true, state: 'Online', description: '系统管理共享'},
        {name: '团队项目资料与交付文档', path: 'C:\\团队项目\\长期归档\\研发文档与交付材料', networkPath: '\\\\DESKTOP\\团队项目资料与交付文档', special: false, state: 'Online', description: '用于团队协作的长中文描述'},
      ],
      GetSMBSessions: [{sessionId: '18446744073709551615', clientComputerName: 'CLIENT-01', clientUserName: '团队\\测试用户', numOpens: 3, secondsExists: 3600, secondsIdle: 12, dialect: '3.1.1'}],
      GetSMBFirewallStatus: {available: true, profiles: [{name: 'domain', enabled: true}, {name: 'private', enabled: true}, {name: 'public', enabled: false}]},
      DiscoverSMBNetworkShares: options.networkEmpty ? [] : [{server: 'NAS', name: '团队档案', path: '\\\\NAS\\团队档案', comment: '已连接的归档共享', provider: 'Windows Network'}],
      OpenDirectoryDialog: 'C:\\测试目录',
      OpenFileDialog: '',
    }
    state.values = readValues
    const writes = ['AddRule', 'UpdateRule', 'DeleteRule', 'ToggleRule', 'ChangeServicePort', 'SetSMBServiceRunning', 'DisableSMB1', 'CreateSMBShare', 'RemoveSMBShare', 'CloseSMBSession', 'SetSMBFirewallEnabled', 'OpenSMBLocation', 'OpenExplorer']
    window.go = {main: {App: new Proxy({}, {get: (_, method) => async (...args) => {
      state.calls.push({method, args})
      const delay = method === 'GetRules' ? options.ruleDelays?.[args[0]] : state.delays[method]
      if (delay) await new Promise(resolve => setTimeout(resolve, delay))
      if (state.holds[method]) await new Promise(resolve => { state.pending[method] = resolve })
      const failure = state.failures[method] || options.failures?.[method]
      if (failure) throw new Error(failure)
      if (method === 'GetRules') return options.rulesEmpty ? [] : state.rules.map(item => ({...item, direction: args[0], name: (args[0] === 'out' ? '出站 ' : '') + item.name}))
      if (Object.hasOwn(readValues, method)) return structuredClone(readValues[method])
      if (writes.includes(method)) return undefined
      throw new Error('Unmocked API: ' + String(method))
    }})}}
  }, options)
  await page.goto('/')
}

export const calls = (page, method) => page.evaluate(method => window.__mock.calls.filter(call => call.method === method), method)
