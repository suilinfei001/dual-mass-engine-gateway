# 用户故事2：用户登录用户名或密码错误

## 故事描述

**作为** 一个系统用户  
**我想要** 在输入错误的用户名或密码时收到明确的错误提示  
**以便于** 了解登录失败的原因并采取正确的行动

## 前置条件

1. 系统已部署并正常运行
2. 浏览器已打开系统登录页面

## 测试数据

引用配置文件：`config/users.json` 中的 `invalid_users`

| 场景 | 用户名 | 密码 | 说明 |
|-----|-------|------|------|
| 场景1 | testuser | eisoo.com123 | 用户名不存在 |
| 场景2 | Yabo.sui | eisoo.com1234 | 密码错误 |

## 测试步骤

### 场景1：用户名不存在

| 步骤 | 操作 | 预期结果 |
|-----|------|---------|
| 1 | 打开登录页面 | 显示登录表单 |
| 2 | 输入用户名 `testuser` | 用户名输入框显示输入内容 |
| 3 | 输入密码 `eisoo.com123` | 密码输入框显示为掩码 |
| 4 | 点击登录按钮 | 系统验证失败 |
| 5 | 验证错误提示 | 显示错误信息："用户名或密码错误" |
| 6 | 验证页面状态 | 停留在登录页面，输入框内容保留 |

### 场景2：密码错误

| 步骤 | 操作 | 预期结果 |
|-----|------|---------|
| 1 | 打开登录页面 | 显示登录表单 |
| 2 | 输入用户名 `Yabo.sui` | 用户名输入框显示输入内容 |
| 3 | 输入密码 `eisoo.com1234` | 密码输入框显示为掩码 |
| 4 | 点击登录按钮 | 系统验证失败 |
| 5 | 验证错误提示 | 显示错误信息："用户名或密码错误" |
| 6 | 验证页面状态 | 停留在登录页面，输入框内容保留 |

## 验收标准

- [ ] 用户名不存在时显示错误提示
- [ ] 密码错误时显示错误提示
- [ ] 错误提示信息不暴露具体是用户名错误还是密码错误（安全考虑）
- [ ] 登录失败后停留在登录页面
- [ ] 连续失败多次后可能触发验证码或账户锁定（可选）

## 安全考虑

1. **错误信息统一**：不区分"用户名不存在"和"密码错误"，统一显示"用户名或密码错误"
2. **防暴力破解**：连续失败多次后可能触发验证码或临时锁定
3. **日志记录**：记录失败的登录尝试，便于安全审计

## 异常处理

| 异常场景 | 处理方式 |
|---------|---------|
| 网络错误 | 显示"网络连接失败，请稍后重试" |
| 服务器错误 | 显示"系统繁忙，请稍后重试" |
| 账户锁定 | 显示"账户已锁定，请联系管理员" |

## 自动化测试脚本参考

```javascript
// tests/login-error.spec.js
const { test, expect } = require('@playwright/test');
const testData = require('../config/users.json');

test.describe('登录错误场景', () => {
  
  test('用户名不存在 - 显示错误提示', async ({ page }) => {
    const invalidUser = testData.invalid_users.find(u => u.scenario === 'username_not_exist');
    
    // 打开登录页面
    await page.goto('/login');
    
    // 输入错误的用户名
    await page.fill('input[name="username"]', invalidUser.username);
    await page.fill('input[name="password"]', invalidUser.password);
    
    // 点击登录按钮
    await page.click('button[type="submit"]');
    
    // 验证错误提示
    await expect(page.locator('.error-message')).toBeVisible();
    await expect(page.locator('.error-message')).toContainText(/用户名或密码错误/i);
    
    // 验证停留在登录页面
    await expect(page).toHaveURL(/.*login/);
  });
  
  test('密码错误 - 显示错误提示', async ({ page }) => {
    const invalidUser = testData.invalid_users.find(u => u.scenario === 'password_wrong');
    
    // 打开登录页面
    await page.goto('/login');
    
    // 输入正确的用户名和错误的密码
    await page.fill('input[name="username"]', invalidUser.username);
    await page.fill('input[name="password"]', invalidUser.password);
    
    // 点击登录按钮
    await page.click('button[type="submit"]');
    
    // 验证错误提示
    await expect(page.locator('.error-message')).toBeVisible();
    await expect(page.locator('.error-message')).toContainText(/用户名或密码错误/i);
    
    // 验证停留在登录页面
    await expect(page).toHaveURL(/.*login/);
  });
  
  test('连续失败登录 - 检查错误提示一致性', async ({ page }) => {
    await page.goto('/login');
    
    // 第一次尝试：用户名不存在
    await page.fill('input[name="username"]', 'nonexist_user');
    await page.fill('input[name="password"]', 'any_password');
    await page.click('button[type="submit"]');
    const errorMsg1 = await page.locator('.error-message').textContent();
    
    // 第二次尝试：密码错误
    await page.fill('input[name="username"]', 'Yabo.sui');
    await page.fill('input[name="password"]', 'wrong_password');
    await page.click('button[type="submit"]');
    const errorMsg2 = await page.locator('.error-message').textContent();
    
    // 验证两种情况的错误信息相同（安全考虑）
    expect(errorMsg1).toBe(errorMsg2);
  });
});
```

## 优先级

**P0** - 核心功能，必须通过

## 相关文档

- 系统登录页面：`/login`
- 用户认证 API：`POST /api/auth/login`
- 错误码定义：
  - `AUTH_FAILED` - 认证失败
  - `USER_NOT_FOUND` - 用户不存在（内部使用，不返回给前端）
  - `PASSWORD_WRONG` - 密码错误（内部使用，不返回给前端）
