# 用户故事1：用户登录和退出功能

## 故事描述

**作为** 一个系统用户  
**我想要** 能够使用正确的用户名和密码登录系统  
**以便于** 访问系统的各项功能，并在使用完毕后安全退出

## 前置条件

1. 系统已部署并正常运行
2. 用户账号已存在于系统中
3. 浏览器已打开系统登录页面

## 测试数据

引用配置文件：`config/users.json` 中的 `valid_user`

```json
{
  "username": "Yabo.sui",
  "password": "eisoo.com123"
}
```

## 测试步骤

### 场景1：用户成功登录

| 步骤 | 操作 | 预期结果 |
|-----|------|---------|
| 1 | 打开登录页面 | 显示登录表单，包含用户名和密码输入框 |
| 2 | 输入用户名 `Yabo.sui` | 用户名输入框显示输入内容 |
| 3 | 输入密码 `eisoo.com123` | 密码输入框显示为掩码（****） |
| 4 | 点击登录按钮 | 系统验证通过，跳转到首页 |
| 5 | 验证登录状态 | 页面右上角显示用户名或用户头像 |

### 场景2：用户成功退出

| 步骤 | 操作 | 预期结果 |
|-----|------|---------|
| 1 | 用户已登录状态 | 页面显示用户已登录状态 |
| 2 | 点击用户头像/用户名 | 显示下拉菜单 |
| 3 | 点击"退出"按钮 | 系统清除登录状态 |
| 4 | 验证退出结果 | 页面跳转到登录页面 |

## 验收标准

- [ ] 使用正确凭据可以成功登录
- [ ] 登录后页面显示用户信息
- [ ] 登录后可以访问需要认证的功能
- [ ] 退出后跳转到登录页面
- [ ] 退出后无法访问需要认证的功能

## 异常处理

| 异常场景 | 处理方式 |
|---------|---------|
| 登录按钮无响应 | 检查网络请求，显示错误提示 |
| 服务器返回错误 | 显示友好的错误提示信息 |
| 会话超时 | 自动跳转到登录页面 |

## 自动化测试脚本参考

```javascript
// tests/login-logout.spec.js
const { test, expect } = require('@playwright/test');
const testData = require('../config/users.json');

test('用户成功登录', async ({ page }) => {
  // 打开登录页面
  await page.goto('/login');
  
  // 输入用户名和密码
  await page.fill('input[name="username"]', testData.valid_user.username);
  await page.fill('input[name="password"]', testData.valid_user.password);
  
  // 点击登录按钮
  await page.click('button[type="submit"]');
  
  // 验证登录成功
  await expect(page).toHaveURL(/.*home|.*dashboard/);
  await expect(page.locator('.user-info')).toContainText(testData.valid_user.username);
});

test('用户成功退出', async ({ page }) => {
  // 先登录
  await page.goto('/login');
  await page.fill('input[name="username"]', testData.valid_user.username);
  await page.fill('input[name="password"]', testData.valid_user.password);
  await page.click('button[type="submit"]');
  
  // 点击退出
  await page.click('.user-avatar');
  await page.click('text=退出');
  
  // 验证退出成功
  await expect(page).toHaveURL(/.*login/);
});
```

## 优先级

**P0** - 核心功能，必须通过

## 相关文档

- 系统登录页面：`/login`
- 用户认证 API：`POST /api/auth/login`
- 用户退出 API：`POST /api/auth/logout`
