<template>
  <div class="api-reference-page">
    <div class="api-header">
      <h1>API 参考</h1>
      <p class="api-desc">双引擎质量网关 RESTful API 文档</p>
    </div>

    <div class="api-content">
      <nav class="api-nav">
        <div class="nav-section">
          <h4>Event Processor</h4>
          <ul>
            <li><a href="#ep-auth" :class="{ active: activeSection === 'ep-auth' }" @click.prevent="scrollTo('ep-auth')">认证接口</a></li>
            <li><a href="#ep-events" :class="{ active: activeSection === 'ep-events' }" @click.prevent="scrollTo('ep-events')">事件接口</a></li>
            <li><a href="#ep-tasks" :class="{ active: activeSection === 'ep-tasks' }" @click.prevent="scrollTo('ep-tasks')">任务接口</a></li>
            <li><a href="#ep-resources" :class="{ active: activeSection === 'ep-resources' }" @click.prevent="scrollTo('ep-resources')">资源接口</a></li>
          </ul>
        </div>
        <div class="nav-section">
          <h4>Resource Pool</h4>
          <ul>
            <li><a href="#rp-internal" :class="{ active: activeSection === 'rp-internal' }" @click.prevent="scrollTo('rp-internal')">内部接口</a></li>
            <li><a href="#rp-external" :class="{ active: activeSection === 'rp-external' }" @click.prevent="scrollTo('rp-external')">外部接口</a></li>
            <li><a href="#rp-admin" :class="{ active: activeSection === 'rp-admin' }" @click.prevent="scrollTo('rp-admin')">管理接口</a></li>
          </ul>
        </div>
      </nav>

      <main class="api-main">
        <section id="ep-auth" class="api-section">
          <h2>Event Processor - 认证接口</h2>
          <p class="section-desc">用户认证和会话管理接口</p>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/api/auth/register</span>
            </div>
            <div class="endpoint-body">
              <h4>用户注册</h4>
              <p>注册新用户账号</p>
              <div class="request-example">
                <h5>请求体</h5>
                <pre><code>{
  "username": "string",
  "password": "string",
  "email": "string"
}</code></pre>
              </div>
              <div class="response-example">
                <h5>响应</h5>
                <pre><code>{
  "success": true,
  "message": "注册成功",
  "data": {
    "user_id": 1,
    "username": "testuser"
  }
}</code></pre>
              </div>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/api/auth/login</span>
            </div>
            <div class="endpoint-body">
              <h4>用户登录</h4>
              <p>用户登录获取会话</p>
              <div class="request-example">
                <h5>请求体</h5>
                <pre><code>{
  "username": "string",
  "password": "string"
}</code></pre>
              </div>
              <div class="response-example">
                <h5>响应</h5>
                <pre><code>{
  "success": true,
  "message": "登录成功",
  "data": {
    "user_id": 1,
    "username": "testuser",
    "role": "admin"
  }
}</code></pre>
              </div>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/api/auth/logout</span>
            </div>
            <div class="endpoint-body">
              <h4>用户登出</h4>
              <p>退出当前会话</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/api/auth/status</span>
            </div>
            <div class="endpoint-body">
              <h4>获取认证状态</h4>
              <p>检查当前用户的登录状态</p>
              <div class="response-example">
                <h5>响应</h5>
                <pre><code>{
  "logged_in": true,
  "username": "testuser",
  "role": "admin"
}</code></pre>
              </div>
            </div>
          </div>
        </section>

        <section id="ep-events" class="api-section">
          <h2>Event Processor - 事件接口</h2>
          <p class="section-desc">事件查询和管理接口</p>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/api/events</span>
            </div>
            <div class="endpoint-body">
              <h4>获取事件列表</h4>
              <p>获取所有事件或按条件筛选</p>
              <div class="params-table">
                <h5>查询参数</h5>
                <table>
                  <thead>
                    <tr>
                      <th>参数</th>
                      <th>类型</th>
                      <th>说明</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                      <td>status</td>
                      <td>string</td>
                      <td title="事件状态: pending, processing, completed, failed, cancelled">事件状态: pending, processing, completed, failed, cancelled</td>
                    </tr>
                    <tr>
                      <td>event_type</td>
                      <td>string</td>
                      <td title="事件类型: push, pull_request">事件类型: push, pull_request</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <div class="response-example">
                <h5>响应</h5>
                <pre><code>{
  "success": true,
  "data": [
    {
      "id": 1,
      "event_id": "evt-123",
      "event_type": "push",
      "event_status": "completed",
      "repository": "org/repo",
      "branch": "main",
      "commit_sha": "abc123",
      "created_at": "2026-03-19T10:00:00Z"
    }
  ]
}</code></pre>
              </div>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/api/events/{id}</span>
            </div>
            <div class="endpoint-body">
              <h4>获取事件详情</h4>
              <p>根据 ID 获取单个事件的详细信息</p>
              <div class="params-table">
                <h5>路径参数</h5>
                <table>
                  <thead>
                    <tr>
                      <th>参数</th>
                      <th>类型</th>
                      <th>说明</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                      <td>id</td>
                      <td>integer</td>
                      <td title="事件 ID">事件 ID</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/api/events/{id}/tasks</span>
            </div>
            <div class="endpoint-body">
              <h4>获取事件关联的任务</h4>
              <p>获取指定事件下的所有任务</p>
            </div>
          </div>
        </section>

        <section id="ep-tasks" class="api-section">
          <h2>Event Processor - 任务接口</h2>
          <p class="section-desc">任务查询接口</p>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/api/tasks</span>
            </div>
            <div class="endpoint-body">
              <h4>获取任务列表</h4>
              <p>获取所有任务或按条件筛选</p>
              <div class="params-table">
                <h5>查询参数</h5>
                <table>
                  <thead>
                    <tr>
                      <th>参数</th>
                      <th>类型</th>
                      <th>说明</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                      <td>status</td>
                      <td>string</td>
                      <td title="任务状态: pending, running, passed, failed, cancelled, skipped">任务状态: pending, running, passed, failed, cancelled, skipped</td>
                    </tr>
                    <tr>
                      <td>event_id</td>
                      <td>integer</td>
                      <td title="关联的事件 ID">关联的事件 ID</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <div class="response-example">
                <h5>响应</h5>
                <pre><code>{
  "success": true,
  "data": [
    {
      "id": 1,
      "event_id": 1,
      "task_name": "basic_ci_all",
      "status": "passed",
      "execute_order": 1,
      "created_at": "2026-03-19T10:00:00Z",
      "completed_at": "2026-03-19T10:30:00Z"
    }
  ]
}</code></pre>
              </div>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/api/tasks/{id}</span>
            </div>
            <div class="endpoint-body">
              <h4>获取任务详情</h4>
              <p>根据 ID 获取单个任务的详细信息，包含执行结果</p>
            </div>
          </div>
        </section>

        <section id="ep-resources" class="api-section">
          <h2>Event Processor - 资源接口</h2>
          <p class="section-desc">可执行资源配置接口</p>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/api/resources</span>
            </div>
            <div class="endpoint-body">
              <h4>获取所有可执行资源</h4>
              <p>获取配置的所有 Azure DevOps 管道资源</p>
              <div class="response-example">
                <h5>响应</h5>
                <pre><code>{
  "success": true,
  "data": [
    {
      "id": 1,
      "resource_name": "basic_ci_all",
      "resource_type": "basic_ci",
      "pipeline_url": "azure://devops.aishu.cn/org/project/pipeline/123",
      "enabled": true
    }
  ]
}</code></pre>
              </div>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/api/resources/{id}</span>
            </div>
            <div class="endpoint-body">
              <h4>获取资源详情</h4>
              <p>根据 ID 获取单个资源的详细信息</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/api/config/event-receiver</span>
            </div>
            <div class="endpoint-body">
              <h4>获取 Event Receiver 配置</h4>
              <p>获取 Event Receiver 服务的连接配置</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/api/health</span>
            </div>
            <div class="endpoint-body">
              <h4>健康检查</h4>
              <p>检查服务运行状态</p>
            </div>
          </div>
        </section>

        <section id="rp-internal" class="api-section">
          <h2>Resource Pool - 内部接口</h2>
          <p class="section-desc">供 Event Processor 内部调用的接口，无需认证</p>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/internal/testbeds/acquire</span>
            </div>
            <div class="endpoint-body">
              <h4>申请测试床</h4>
              <p>根据类别申请一个可用的测试床</p>
              <div class="request-example">
                <h5>请求体</h5>
                <pre><code>{
  "category_uuid": "string",
  "requester": "string"
}</code></pre>
              </div>
              <div class="response-example">
                <h5>响应</h5>
                <pre><code>{
  "uuid": "allocation-uuid",
  "testbed": {
    "uuid": "testbed-uuid",
    "ip_address": "10.4.111.100",
    "ssh_user": "root",
    "ssh_password": "password"
  }
}</code></pre>
              </div>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/internal/testbeds/acquire-robot</span>
            </div>
            <div class="endpoint-body">
              <h4>申请 Robot 测试床</h4>
              <p>自动申请一个 robot 类别的测试床（用于部署任务）</p>
              <div class="response-example">
                <h5>响应</h5>
                <pre><code>{
  "success": true,
  "allocation": {
    "uuid": "allocation-uuid"
  },
  "testbed": {
    "uuid": "testbed-uuid",
    "ip_address": "10.4.111.137",
    "ssh_user": "root",
    "ssh_password": "123qweASD"
  }
}</code></pre>
              </div>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/internal/testbeds/{uuid}/release</span>
            </div>
            <div class="endpoint-body">
              <h4>释放测试床</h4>
              <p>释放指定的测试床，使其可被重新分配</p>
              <div class="params-table">
                <h5>路径参数</h5>
                <table>
                  <thead>
                    <tr>
                      <th>参数</th>
                      <th>类型</th>
                      <th>说明</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                      <td>uuid</td>
                      <td>string</td>
                      <td title="分配 UUID (allocation_uuid)">分配 UUID (allocation_uuid)</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/internal/testbeds</span>
            </div>
            <div class="endpoint-body">
              <h4>获取测试床列表</h4>
              <p>获取所有测试床</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/internal/testbeds/{uuid}</span>
            </div>
            <div class="endpoint-body">
              <h4>获取测试床详情</h4>
              <p>根据 UUID 获取测试床详细信息</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/internal/health</span>
            </div>
            <div class="endpoint-body">
              <h4>健康检查</h4>
              <p>检查服务运行状态</p>
            </div>
          </div>
        </section>

        <section id="rp-external" class="api-section">
          <h2>Resource Pool - 外部接口</h2>
          <p class="section-desc">供普通用户使用的接口，需要认证</p>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/external/categories</span>
            </div>
            <div class="endpoint-body">
              <h4>获取资源类别列表</h4>
              <p>获取所有可用的资源类别</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/external/categories/{uuid}</span>
            </div>
            <div class="endpoint-body">
              <h4>获取类别详情</h4>
              <p>获取指定类别的详细信息</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/external/categories/{uuid}/quota</span>
            </div>
            <div class="endpoint-body">
              <h4>获取类别配额</h4>
              <p>获取指定类别的配额策略</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/external/resource-instances/my</span>
            </div>
            <div class="endpoint-body">
              <h4>获取我的资源实例</h4>
              <p>获取当前用户创建的所有资源实例</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/external/resource-instances/public</span>
            </div>
            <div class="endpoint-body">
              <h4>获取公开资源实例</h4>
              <p>获取所有公开的资源实例</p>
              <div class="params-table">
                <h5>查询参数</h5>
                <table>
                  <thead>
                    <tr>
                      <th>参数</th>
                      <th>类型</th>
                      <th>说明</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                      <td>resource_type</td>
                      <td>string</td>
                      <td title="资源类型: virtual_machine, physical_machine">资源类型: virtual_machine, physical_machine</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <div class="response-example">
                <h5>响应</h5>
                <pre><code>{
  "success": true,
  "instances": [
    {
      "id": 1,
      "uuid": "instance-uuid",
      "instance_type": "VirtualMachine",
      "ip_address": "10.4.111.100",
      "port": 22,
      "ssh_user": "root",
      "status": "active"
    }
  ],
  "total": 1
}</code></pre>
              </div>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/external/allocations</span>
            </div>
            <div class="endpoint-body">
              <h4>获取我的分配列表</h4>
              <p>获取当前用户的所有资源分配</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/external/allocations/{uuid}</span>
            </div>
            <div class="endpoint-body">
              <h4>获取分配详情</h4>
              <p>获取指定分配的详细信息</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/external/allocations/{uuid}/extend</span>
            </div>
            <div class="endpoint-body">
              <h4>延长分配时间</h4>
              <p>延长资源分配的使用时间</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method delete">DELETE</span>
              <span class="path">/external/allocations/{uuid}</span>
            </div>
            <div class="endpoint-body">
              <h4>释放分配</h4>
              <p>主动释放资源分配</p>
            </div>
          </div>
        </section>

        <section id="rp-admin" class="api-section">
          <h2>Resource Pool - 管理接口</h2>
          <p class="section-desc">管理员专用接口，需要管理员权限</p>

          <h3>测试床管理</h3>
          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/admin/testbeds</span>
            </div>
            <div class="endpoint-body">
              <h4>获取所有测试床</h4>
              <p>管理员查看所有测试床</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/admin/testbeds</span>
            </div>
            <div class="endpoint-body">
              <h4>创建测试床</h4>
              <p>手动创建新的测试床</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method put">PUT</span>
              <span class="path">/admin/testbeds/{uuid}</span>
            </div>
            <div class="endpoint-body">
              <h4>更新测试床</h4>
              <p>更新测试床信息</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method delete">DELETE</span>
              <span class="path">/admin/testbeds/{uuid}</span>
            </div>
            <div class="endpoint-body">
              <h4>删除测试床</h4>
              <p>删除指定的测试床</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method put">PUT</span>
              <span class="path">/admin/testbeds/{uuid}/maintenance</span>
            </div>
            <div class="endpoint-body">
              <h4>设置维护状态</h4>
              <p>将测试床设置为维护模式</p>
            </div>
          </div>

          <h3>资源实例管理</h3>
          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/admin/resource-instances</span>
            </div>
            <div class="endpoint-body">
              <h4>获取所有资源实例</h4>
              <p>管理员查看所有资源实例</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/admin/resource-instances</span>
            </div>
            <div class="endpoint-body">
              <h4>创建资源实例</h4>
              <p>添加新的资源实例</p>
              <div class="request-example">
                <h5>请求体</h5>
                <pre><code>{
  "resource_type": "virtual_machine",
  "ip_address": "10.4.111.100",
  "port": 22,
  "ssh_user": "root",
  "passwd": "password",
  "snapshot_id": "snapshot-123",
  "description": "测试用虚拟机",
  "is_public": true
}</code></pre>
              </div>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method put">PUT</span>
              <span class="path">/admin/resource-instances/{uuid}</span>
            </div>
            <div class="endpoint-body">
              <h4>更新资源实例</h4>
              <p>更新资源实例信息</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method delete">DELETE</span>
              <span class="path">/admin/resource-instances/{uuid}</span>
            </div>
            <div class="endpoint-body">
              <h4>删除资源实例</h4>
              <p>删除指定的资源实例</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/admin/resource-instances/{uuid}/deploy</span>
            </div>
            <div class="endpoint-body">
              <h4>部署资源实例</h4>
              <p>触发资源实例的自动部署</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/admin/resource-instances/{uuid}/health</span>
            </div>
            <div class="endpoint-body">
              <h4>检查健康状态</h4>
              <p>检查资源实例的 SSH 连接状态</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/admin/resource-instances/{uuid}/restore-snapshot</span>
            </div>
            <div class="endpoint-body">
              <h4>恢复快照</h4>
              <p>将虚拟机恢复到指定快照状态</p>
            </div>
          </div>

          <h3>类别管理</h3>
          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/admin/categories</span>
            </div>
            <div class="endpoint-body">
              <h4>获取所有类别</h4>
              <p>获取所有资源类别</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/admin/categories</span>
            </div>
            <div class="endpoint-body">
              <h4>创建类别</h4>
              <p>创建新的资源类别</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method put">PUT</span>
              <span class="path">/admin/categories/{uuid}</span>
            </div>
            <div class="endpoint-body">
              <h4>更新类别</h4>
              <p>更新类别信息</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method delete">DELETE</span>
              <span class="path">/admin/categories/{uuid}</span>
            </div>
            <div class="endpoint-body">
              <h4>删除类别</h4>
              <p>删除指定的资源类别</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/admin/categories/{uuid}/replenish</span>
            </div>
            <div class="endpoint-body">
              <h4>手动触发补充</h4>
              <p>手动触发类别的资源补充</p>
            </div>
          </div>

          <h3>配额策略管理</h3>
          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/admin/quota-policies</span>
            </div>
            <div class="endpoint-body">
              <h4>获取所有配额策略</h4>
              <p>获取所有配额策略</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/admin/quota-policies</span>
            </div>
            <div class="endpoint-body">
              <h4>创建配额策略</h4>
              <p>创建新的配额策略</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method put">PUT</span>
              <span class="path">/admin/quota-policies/{uuid}</span>
            </div>
            <div class="endpoint-body">
              <h4>更新配额策略</h4>
              <p>更新配额策略</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method delete">DELETE</span>
              <span class="path">/admin/quota-policies/{uuid}</span>
            </div>
            <div class="endpoint-body">
              <h4>删除配额策略</h4>
              <p>删除指定的配额策略</p>
            </div>
          </div>

          <h3>部署任务管理</h3>
          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/admin/deployment-tasks</span>
            </div>
            <div class="endpoint-body">
              <h4>获取部署任务列表</h4>
              <p>获取所有部署任务</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/admin/deployment-tasks/{uuid}</span>
            </div>
            <div class="endpoint-body">
              <h4>获取部署任务详情</h4>
              <p>获取指定部署任务的详细信息</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/admin/deployment-tasks/{uuid}/logs</span>
            </div>
            <div class="endpoint-body">
              <h4>获取部署日志</h4>
              <p>获取部署任务的执行日志</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/admin/deployment-tasks/{uuid}/retry</span>
            </div>
            <div class="endpoint-body">
              <h4>重试部署任务</h4>
              <p>重试失败的部署任务</p>
            </div>
          </div>

          <h3>管道模板管理</h3>
          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/admin/pipeline-templates</span>
            </div>
            <div class="endpoint-body">
              <h4>获取管道模板列表</h4>
              <p>获取所有部署管道模板</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/admin/pipeline-templates</span>
            </div>
            <div class="endpoint-body">
              <h4>创建管道模板</h4>
              <p>创建新的部署管道模板</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method put">PUT</span>
              <span class="path">/admin/pipeline-templates/{id}</span>
            </div>
            <div class="endpoint-body">
              <h4>更新管道模板</h4>
              <p>更新管道模板配置</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method delete">DELETE</span>
              <span class="path">/admin/pipeline-templates/{id}</span>
            </div>
            <div class="endpoint-body">
              <h4>删除管道模板</h4>
              <p>删除指定的管道模板</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/admin/pipeline-templates/{id}/enable</span>
            </div>
            <div class="endpoint-body">
              <h4>启用管道模板</h4>
              <p>启用指定的管道模板</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/admin/pipeline-templates/{id}/disable</span>
            </div>
            <div class="endpoint-body">
              <h4>禁用管道模板</h4>
              <p>禁用指定的管道模板</p>
            </div>
          </div>

          <h3>统计与监控</h3>
          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/admin/metrics</span>
            </div>
            <div class="endpoint-body">
              <h4>获取系统指标</h4>
              <p>获取系统运行指标</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/admin/metrics/usage</span>
            </div>
            <div class="endpoint-body">
              <h4>获取使用统计</h4>
              <p>获取资源使用统计</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method get">GET</span>
              <span class="path">/admin/tasks/statistics</span>
            </div>
            <div class="endpoint-body">
              <h4>获取任务统计</h4>
              <p>获取任务执行统计</p>
            </div>
          </div>

          <h3>清理接口</h3>
          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/admin/cleanup/testbeds</span>
            </div>
            <div class="endpoint-body">
              <h4>清理测试床</h4>
              <p>清理无效的测试床记录</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/admin/cleanup/allocations</span>
            </div>
            <div class="endpoint-body">
              <h4>清理分配记录</h4>
              <p>清理过期的分配记录</p>
            </div>
          </div>

          <div class="api-endpoint">
            <div class="endpoint-header">
              <span class="method post">POST</span>
              <span class="path">/admin/cleanup/all</span>
            </div>
            <div class="endpoint-body">
              <h4>清理所有</h4>
              <p>执行全面清理</p>
            </div>
          </div>
        </section>
      </main>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue'

export default {
  name: 'ApiReference',
  setup() {
    const activeSection = ref('ep-auth')

    const scrollTo = (sectionId) => {
      const element = document.getElementById(sectionId)
      if (element) {
        element.scrollIntoView({ behavior: 'smooth' })
      }
    }

    const handleScroll = () => {
      const sections = [
        'ep-auth', 'ep-events', 'ep-tasks', 'ep-resources',
        'rp-internal', 'rp-external', 'rp-admin'
      ]
      for (const section of sections) {
        const element = document.getElementById(section)
        if (element) {
          const rect = element.getBoundingClientRect()
          if (rect.top <= 150 && rect.bottom >= 150) {
            activeSection.value = section
            break
          }
        }
      }
    }

    onMounted(() => {
      window.addEventListener('scroll', handleScroll)
    })

    onUnmounted(() => {
      window.removeEventListener('scroll', handleScroll)
    })

    return {
      activeSection,
      scrollTo
    }
  }
}
</script>

<style scoped>
.api-reference-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
}

.api-header {
  text-align: center;
  margin-bottom: 3rem;
  padding-bottom: 2rem;
  border-bottom: 1px solid #E2E8F0;
}

.api-header h1 {
  font-size: 2.5rem;
  color: #0C4A6E;
  margin-bottom: 0.5rem;
}

.api-desc {
  color: #64748B;
}

.api-content {
  display: grid;
  grid-template-columns: 220px 1fr;
  gap: 2rem;
}

.api-nav {
  position: sticky;
  top: 2rem;
  height: fit-content;
}

.nav-section {
  margin-bottom: 1.5rem;
}

.nav-section h4 {
  font-size: 0.75rem;
  text-transform: uppercase;
  color: #64748B;
  margin: 0 0 0.5rem;
  padding: 0 1rem;
}

.api-nav ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.api-nav li {
  margin-bottom: 0.25rem;
}

.api-nav a {
  display: block;
  padding: 0.5rem 1rem;
  color: #475569;
  text-decoration: none;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.api-nav a:hover {
  background: #F1F5F9;
  color: #0EA5E9;
}

.api-nav a.active {
  background: #E0F2FE;
  color: #0EA5E9;
  font-weight: 500;
}

.api-main {
  min-width: 0;
}

.api-section {
  margin-bottom: 3rem;
}

.api-section h2 {
  font-size: 1.5rem;
  color: #0C4A6E;
  margin-bottom: 0.5rem;
  padding-bottom: 0.5rem;
  border-bottom: 2px solid #0EA5E9;
}

.api-section h3 {
  font-size: 1.125rem;
  color: #334155;
  margin: 2rem 0 1rem;
  padding-top: 1rem;
  border-top: 1px solid #E2E8F0;
}

.section-desc {
  color: #64748B;
  margin-bottom: 1.5rem;
}

.api-endpoint {
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
  border-radius: 0.5rem;
  margin-bottom: 1rem;
  overflow: hidden;
}

.endpoint-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background: #F1F5F9;
  border-bottom: 1px solid #E2E8F0;
}

.method {
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  min-width: 60px;
  text-align: center;
}

.method.get {
  background: #D1FAE5;
  color: #065F46;
}

.method.post {
  background: #DBEAFE;
  color: #1E40AF;
}

.method.put {
  background: #FEF3C7;
  color: #92400E;
}

.method.delete {
  background: #FEE2E2;
  color: #991B1B;
}

.path {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.875rem;
  color: #334155;
}

.endpoint-body {
  padding: 1rem;
}

.endpoint-body h4 {
  margin: 0 0 0.5rem;
  color: #0C4A6E;
  font-size: 1rem;
}

.endpoint-body > p {
  margin: 0 0 1rem;
  color: #64748B;
  font-size: 0.875rem;
}

.params-table {
  margin-bottom: 1rem;
}

.params-table h5,
.request-example h5,
.response-example h5 {
  font-size: 0.75rem;
  text-transform: uppercase;
  color: #64748B;
  margin: 0 0 0.5rem;
}

.params-table table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
  table-layout: fixed;
}

.params-table th,
.params-table td {
  padding: 0.5rem;
  text-align: left;
  border: 1px solid #E2E8F0;
}

.params-table td {
  /* Apply text overflow to all table cells */
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Ensure inline elements in table cells also truncate */
.params-table td > *,
.params-table td > span,
.params-table td > a {
  display: inline-block;
  max-width: 100%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: bottom;
}

.params-table th {
  background: #F8FAFC;
  font-weight: 500;
  white-space: nowrap;
}

.request-example,
.response-example {
  margin-bottom: 1rem;
}

pre {
  background: #1E293B;
  color: #E2E8F0;
  padding: 1rem;
  border-radius: 0.5rem;
  overflow-x: auto;
  margin: 0.5rem 0 0;
}

code {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.8rem;
  line-height: 1.5;
}

@media (max-width: 768px) {
  .api-content {
    grid-template-columns: 1fr;
  }

  .api-nav {
    position: static;
    margin-bottom: 2rem;
  }

  .nav-section {
    display: inline-block;
    vertical-align: top;
    margin-right: 1rem;
  }

  .api-nav ul {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
  }

  .api-nav li {
    margin: 0;
  }

  .api-nav a {
    padding: 0.375rem 0.75rem;
    font-size: 0.75rem;
  }
}
</style>
