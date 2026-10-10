import {test, expect} from '@playwright/test'
import {mockApp, calls} from './mock-app.js'

const nav = (page, name) => page.getByRole('button', {name, exact: true})
async function confirm(page, positiveText) {
  const dialog = page.locator('.n-dialog')
  await expect(dialog).toBeVisible()
  await dialog.getByRole('button', {name: positiveText, exact: true}).click()
  await expect(dialog).toBeHidden()
}

test('new rule, program folder and WinRM actions send exact mock parameters', async ({page}) => {
  await mockApp(page)
  await expect(page.getByText('共 12 条', {exact: true})).toBeVisible()
  await page.getByRole('button', {name: '添加规则', exact: true}).click()
  const drawer = page.locator('.n-drawer')
  await drawer.getByPlaceholder('输入便于识别的名称').fill('交付验收临时规则')
  await drawer.getByRole('button', {name: '添加', exact: true}).click()
  await expect(drawer).toBeHidden()
  const added = await calls(page, 'AddRule')
  expect(added).toHaveLength(1)
  expect(added[0].args[0]).toMatchObject({name: '交付验收临时规则', direction: 'in', action: 'allow', protocol: 'any', profile: 'any', enabled: true})

  await page.evaluate(() => { window.__mock.rules[0].program = 'C:\\Program Files\\验收工具\\app.exe' })
  await page.getByRole('button', {name: '刷新', exact: true}).click()
  await page.getByRole('button', {name: '打开程序所在目录 C:\\Program Files\\验收工具\\app.exe', exact: true}).click()
  await expect.poll(async () => (await calls(page, 'OpenExplorer')).length).toBe(1)
  expect((await calls(page, 'OpenExplorer'))[0].args).toEqual(['C:\\Program Files\\验收工具\\app.exe'])

  await nav(page, '常用端口').click()
  await page.getByRole('button', {name: '修改WinRM 远程管理端口', exact: true}).click()
  const dialog = page.locator('.n-dialog')
  await expect(dialog).toContainText('WinRM 远程管理')
  await expect(dialog).toContainText('5985')
  await dialog.locator('input').fill('5987')
  await confirm(page, '确认修改')
  expect((await calls(page, 'ChangeServicePort')).map(call => call.args)).toEqual([['WinRM', 5987]])
})

test('SMB successful confirmations preserve share, session and firewall targets', async ({page}) => {
  await mockApp(page)
  await nav(page, 'SMB 文件共享').click()
  const shareName = '团队项目资料与交付文档'
  const row = page.locator('tr').filter({hasText: shareName})
  await expect(row).toBeVisible()
  await row.getByRole('button', {name: '打开目录', exact: true}).click()
  await expect.poll(async () => (await calls(page, 'OpenSMBLocation')).length).toBe(1)
  await row.getByRole('button', {name: '打开共享', exact: true}).click()
  await expect.poll(async () => (await calls(page, 'OpenSMBLocation')).length).toBe(2)
  expect((await calls(page, 'OpenSMBLocation')).map(call => call.args[0])).toEqual([
    'C:\\团队项目\\长期归档\\研发文档与交付材料',
    '\\\\DESKTOP\\团队项目资料与交付文档',
  ])
  await row.getByRole('button', {name: '删除', exact: true}).click()
  await expect(page.locator('.n-dialog')).toContainText(shareName)
  await expect(page.locator('.n-dialog')).toContainText('不会删除目录')
  await confirm(page, '确认删除')
  expect((await calls(page, 'RemoveSMBShare')).map(call => call.args)).toEqual([[shareName]])

  await page.getByRole('button', {name: '关闭 SMB1', exact: true}).click()
  await expect(page.locator('.n-dialog')).toContainText('旧设备')
  await page.evaluate(() => { window.__mock.values.GetSMBOverview.smb1Enabled = false })
  await confirm(page, '确认关闭')
  expect(await calls(page, 'DisableSMB1')).toHaveLength(1)
  await expect(page.getByRole('button', {name: '关闭 SMB1', exact: true})).toHaveCount(0)

  await page.getByRole('button', {name: '全部启用', exact: true}).click()
  await expect(page.locator('.n-dialog')).toContainText('公用网络')
  await page.evaluate(() => { window.__mock.values.GetSMBFirewallStatus.profiles = window.__mock.values.GetSMBFirewallStatus.profiles.map(profile => ({...profile, enabled: true})) })
  await confirm(page, '确认启用')
  await expect(page.getByRole('button', {name: '全部禁用', exact: true})).toBeEnabled()
  expect((await calls(page, 'SetSMBFirewallEnabled')).map(call => call.args)).toEqual([[true]])

  await page.getByText('活动会话', {exact: true}).click()
  await page.getByRole('button', {name: '断开 CLIENT-01 的会话', exact: true}).click()
  await expect(page.locator('.n-dialog')).toContainText('3 个文件')
  await confirm(page, '确认断开')
  expect((await calls(page, 'CloseSMBSession')).map(call => call.args)).toEqual([['18446744073709551615']])

  await page.getByText('已连接共享', {exact: true}).click()
  await page.getByRole('button', {name: '检测已连接共享', exact: true}).click()
  await page.locator('tr').filter({hasText: 'NAS'}).getByRole('button', {name: '打开', exact: true}).click()
  await expect.poll(async () => (await calls(page, 'OpenSMBLocation')).length).toBe(3)
  expect((await calls(page, 'OpenSMBLocation'))[2].args).toEqual(['\\\\NAS\\团队档案'])
})
