# Playwright MCP 测试数据

本目录包含 Playwright MCP 自动化测试的用户故事和配置数据。

## 目录结构

```
playwright/
├── user-stories/           # 用户故事目录
│   ├── login-logout.md     # 用户故事1：登录和退出功能
│   └── login-error.md      # 用户故事2：登录错误场景
├── config/                 # 配置数据目录
│   ├── users.json          # 用户配置数据
│   └── test-data.json      # 测试数据配置
└── README.md               # 本说明文档
```

## 使用方法

### 1. 读取配置数据

在 Playwright 测试中，可以通过读取 JSON 文件获取测试数据：

```javascript
const { readFile } = require('fs/promises');
const path = require('path');

async function loadTestData() {
  const data = await readFile(path.join(__dirname, 'config/users.json'), 'utf-8');
  return JSON.parse(data);
}
```

### 2. 用户故事驱动测试

每个用户故事文件包含：
- **故事描述**：用户场景说明
- **前置条件**：测试前的准备工作
- **测试步骤**：详细的操作步骤
- **预期结果**：每个步骤的预期输出
- **测试数据引用**：关联的配置数据

## 测试覆盖范围

| 用户故事 | 文件 | 配置数据 |
|---------|------|---------|
| 用户登录和退出功能 | `user-stories/login-logout.md` | `config/users.json` |
| 用户登录错误场景 | `user-stories/login-error.md` | `config/users.json` |

## 注意事项

1. **敏感信息**：密码等敏感信息仅用于测试环境，请勿在生产环境使用
2. **数据隔离**：测试数据与代码分离，便于维护和扩展
3. **版本控制**：配置数据文件应纳入版本控制
