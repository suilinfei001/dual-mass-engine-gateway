<template>
  <div class="documentation-page">
    <div class="doc-header">
      <h1>双引擎质量网关 - 使用文档</h1>
      <p class="version">版本 1.0.0 | 更新日期: 2026-03-19</p>
    </div>

    <div class="doc-content">
      <nav class="doc-nav">
        <ul>
          <li><a href="#overview" :class="{ active: activeSection === 'overview' }" @click.prevent="scrollTo('overview')">项目概述</a></li>
          <li><a href="#architecture" :class="{ active: activeSection === 'architecture' }" @click.prevent="scrollTo('architecture')">系统架构</a></li>
          <li><a href="#event-processor" :class="{ active: activeSection === 'event-processor' }" @click.prevent="scrollTo('event-processor')">事件处理器</a></li>
          <li><a href="#resource-pool" :class="{ active: activeSection === 'resource-pool' }" @click.prevent="scrollTo('resource-pool')">资源池管理</a></li>
          <li><a href="#quick-start" :class="{ active: activeSection === 'quick-start' }" @click.prevent="scrollTo('quick-start')">快速开始</a></li>
          <li><a href="#faq" :class="{ active: activeSection === 'faq' }" @click.prevent="scrollTo('faq')">常见问题</a></li>
        </ul>
      </nav>

      <main class="doc-main">
        <section id="overview" class="doc-section">
          <h2>项目概述</h2>
          <p>双引擎质量网关是一个企业级的质量检查与资源管理平台，主要用于自动化处理 GitHub Webhook 事件，执行质量检查任务，并管理测试资源池。</p>
          
          <div class="feature-grid">
            <div class="feature-card">
              <div class="feature-icon">🔄</div>
              <h3>事件驱动</h3>
              <p>自动接收和处理 GitHub Webhook 事件，支持 push、pull_request 等多种事件类型</p>
            </div>
            <div class="feature-card">
              <div class="feature-icon">✅</div>
              <h3>质量检查</h3>
              <p>集成 Azure DevOps 管道，执行代码检查、单元测试、E2E 测试等多种质量检查</p>
            </div>
            <div class="feature-card">
              <div class="feature-icon">🖥️</div>
              <h3>资源管理</h3>
              <p>统一管理虚拟机和物理机资源，支持自动分配、回收和快照恢复</p>
            </div>
            <div class="feature-card">
              <div class="feature-icon">🤖</div>
              <h3>AI 增强</h3>
              <p>集成 AI 分析能力，自动分析测试日志并生成质量报告</p>
            </div>
          </div>
        </section>

        <section id="architecture" class="doc-section">
          <h2>系统架构</h2>
          <p>系统采用双引擎架构，包含两个独立部署的模块：</p>
          
          <div class="architecture-diagram">
            <div class="arch-box event-receiver">
              <h4>Event Receiver</h4>
              <p>部署在外网 (10.4.111.141)</p>
              <ul>
                <li>接收 GitHub Webhook</li>
                <li>存储事件数据</li>
                <li>管理质量检查状态</li>
              </ul>
            </div>
            <div class="arch-arrow">→</div>
            <div class="arch-box event-processor">
              <h4>Event Processor</h4>
              <p>部署在内网 (10.4.174.125)</p>
              <ul>
                <li>调度任务执行</li>
                <li>调用 Azure DevOps</li>
                <li>AI 日志分析</li>
              </ul>
            </div>
            <div class="arch-arrow">→</div>
            <div class="arch-box resource-pool">
              <h4>Resource Pool</h4>
              <p>部署在内网 (10.4.174.125)</p>
              <ul>
                <li>管理测试资源</li>
                <li>自动分配/回收</li>
                <li>快照恢复</li>
              </ul>
            </div>
          </div>

          <h3>服务端口</h3>
          <table class="doc-table">
            <thead>
              <tr>
                <th>服务</th>
                <th>前端端口</th>
                <th>后端端口</th>
                <th>MySQL 端口</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td title="Event Receiver">Event Receiver</td>
                <td title="8081">8081</td>
                <td title="5001">5001</td>
                <td title="3306">3306</td>
              </tr>
              <tr>
                <td title="Event Processor">Event Processor</td>
                <td title="8082">8082</td>
                <td title="5003">5003</td>
                <td title="3307">3307</td>
              </tr>
            </tbody>
          </table>
        </section>

        <section id="event-processor" class="doc-section">
          <h2>事件处理器</h2>
          
          <h3>支持的事件类型</h3>
          <table class="doc-table">
            <thead>
              <tr>
                <th>事件类型</th>
                <th>触发条件</th>
                <th>执行任务</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td title="push">push</td>
                <td title="代码推送到分支">代码推送到分支</td>
                <td title="basic_ci_all → deployment_deployment → specialized_tests_*">basic_ci_all → deployment_deployment → specialized_tests_*</td>
              </tr>
              <tr>
                <td title="pull_request.opened">pull_request.opened</td>
                <td title="创建 PR">创建 PR</td>
                <td title="basic_ci_all → specialized_tests_*">basic_ci_all → specialized_tests_*</td>
              </tr>
              <tr>
                <td title="pull_request.synchronize">pull_request.synchronize</td>
                <td title="PR 更新">PR 更新</td>
                <td title="取消旧任务，执行新任务">取消旧任务，执行新任务</td>
              </tr>
            </tbody>
          </table>

          <h3>任务执行流程</h3>
          <div class="flow-diagram">
            <div class="flow-step">
              <span class="step-num">1</span>
              <span class="step-text">事件接收</span>
            </div>
            <div class="flow-arrow">→</div>
            <div class="flow-step">
              <span class="step-num">2</span>
              <span class="step-text">任务创建</span>
            </div>
            <div class="flow-arrow">→</div>
            <div class="flow-step">
              <span class="step-num">3</span>
              <span class="step-text">资源分配</span>
            </div>
            <div class="flow-arrow">→</div>
            <div class="flow-step">
              <span class="step-num">4</span>
              <span class="step-text">管道执行</span>
            </div>
            <div class="flow-arrow">→</div>
            <div class="flow-step">
              <span class="step-num">5</span>
              <span class="step-text">AI 分析</span>
            </div>
            <div class="flow-arrow">→</div>
            <div class="flow-step">
              <span class="step-num">6</span>
              <span class="step-text">结果更新</span>
            </div>
          </div>

          <h3>质量检查类型</h3>
          <ul class="doc-list">
            <li><strong>basic_ci_all</strong>: 基础 CI 检查，包含代码检查、编译、单元测试等</li>
            <li><strong>deployment_deployment</strong>: 部署任务，将应用部署到测试环境</li>
            <li><strong>specialized_tests_api_test</strong>: API 接口测试</li>
            <li><strong>specialized_tests_module_e2e</strong>: 模块级 E2E 测试</li>
            <li><strong>specialized_tests_agent_e2e</strong>: Agent E2E 测试</li>
            <li><strong>specialized_tests_ai_e2e</strong>: AI 功能 E2E 测试</li>
          </ul>
        </section>

        <section id="resource-pool" class="doc-section">
          <h2>资源池管理</h2>
          
          <h3>核心概念</h3>
          <div class="concept-grid">
            <div class="concept-item">
              <h4>资源实例 (Resource Instance)</h4>
              <p>物理机或虚拟机实例，包含 IP、SSH 凭证等信息</p>
            </div>
            <div class="concept-item">
              <h4>测试床 (Testbed)</h4>
              <p>从资源实例分配的测试环境，用于执行具体任务</p>
            </div>
            <div class="concept-item">
              <h4>类别 (Category)</h4>
              <p>资源分组，支持按服务目标分类管理</p>
            </div>
            <div class="concept-item">
              <h4>配额策略 (Quota Policy)</h4>
              <p>定义资源自动补充规则，确保资源池容量</p>
            </div>
          </div>

          <h3>资源类型</h3>
          <table class="doc-table">
            <thead>
              <tr>
                <th>类型</th>
                <th>说明</th>
                <th>特点</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td title="VirtualMachine">VirtualMachine</td>
                <td title="虚拟机实例">虚拟机实例</td>
                <td title="支持快照恢复、自动重置">支持快照恢复、自动重置</td>
              </tr>
              <tr>
                <td title="Machine">Machine</td>
                <td title="物理机实例">物理机实例</td>
                <td title="性能稳定、适合长期任务">性能稳定、适合长期任务</td>
              </tr>
            </tbody>
          </table>

          <h3>资源状态流转</h3>
          <div class="status-flow">
            <span class="status available">available</span>
            <span class="arrow">→</span>
            <span class="status allocated">allocated</span>
            <span class="arrow">→</span>
            <span class="status in_use">in_use</span>
            <span class="arrow">→</span>
            <span class="status releasing">releasing</span>
            <span class="arrow">→</span>
            <span class="status available">available</span>
          </div>

          <h3>自动补充机制</h3>
          <p>系统支持配置配额策略，当可用资源低于阈值时自动触发补充：</p>
          <ol class="doc-list">
            <li>检查类别下可用资源数量</li>
            <li>与配额策略中的最小阈值比较</li>
            <li>低于阈值时触发自动部署任务</li>
            <li>部署完成后自动注册到资源池</li>
          </ol>
        </section>

        <section id="quick-start" class="doc-section">
          <h2>快速开始</h2>
          
          <h3>1. 登录系统</h3>
          <p>访问 <code>http://10.4.174.125:8082</code>，使用管理员账号登录。</p>

          <h3>2. 配置可执行资源</h3>
          <p>在「可执行资源」页面配置 Azure DevOps 管道信息：</p>
          <ul class="doc-list">
            <li>资源名称：用于任务匹配</li>
            <li>管道 URL：Azure DevOps 管道地址</li>
            <li>资源类型：basic_ci、deployment、e2e_test 等</li>
          </ul>

          <h3>3. 配置资源池</h3>
          <p>在「资源池管理」中：</p>
          <ol class="doc-list">
            <li>创建资源类别（如：robot-testbed）</li>
            <li>添加资源实例（虚拟机或物理机）</li>
            <li>配置配额策略（可选）</li>
          </ol>

          <h3>4. 触发事件</h3>
          <p>向 GitHub 推送代码或创建 PR，系统将自动：</p>
          <ol class="doc-list">
            <li>接收 Webhook 事件</li>
            <li>创建质量检查任务</li>
            <li>分配测试资源</li>
            <li>执行管道并分析结果</li>
          </ol>

          <h3>5. 查看结果</h3>
          <p>在「事件处理」页面查看事件状态和质量检查结果。</p>
        </section>

        <section id="faq" class="doc-section">
          <h2>常见问题</h2>
          
          <div class="faq-item">
            <h4>Q: 任务一直处于 pending 状态？</h4>
            <p>A: 检查以下几点：</p>
            <ul class="doc-list">
              <li>确认 Event Receiver 服务正常运行</li>
              <li>检查任务调度器是否启动</li>
              <li>查看日志中是否有资源分配失败</li>
            </ul>
          </div>

          <div class="faq-item">
            <h4>Q: 部署任务失败，Chart URL 为空？</h4>
            <p>A: 确保 basic_ci_all 任务成功完成，Chart URL 是从编译结果中提取的。</p>
          </div>

          <div class="faq-item">
            <h4>Q: 资源分配失败？</h4>
            <p>A: 检查资源池中是否有可用的测试床，以及资源实例状态是否为 active。</p>
          </div>

          <div class="faq-item">
            <h4>Q: AI 分析结果不准确？</h4>
            <p>A: 检查 AI 配置是否正确，可以在「控制台」页面测试 AI 连接。</p>
          </div>

          <div class="faq-item">
            <h4>Q: 如何手动触发任务？</h4>
            <p>A: 可以通过 Event Receiver 的 Mock API 模拟事件触发：</p>
            <pre><code>POST http://10.4.111.141:5001/api/mock/simulate/push</code></pre>
          </div>
        </section>
      </main>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue'

export default {
  name: 'Documentation',
  setup() {
    const activeSection = ref('overview')

    const scrollTo = (sectionId) => {
      const element = document.getElementById(sectionId)
      if (element) {
        element.scrollIntoView({ behavior: 'smooth' })
      }
    }

    const handleScroll = () => {
      const sections = ['overview', 'architecture', 'event-processor', 'resource-pool', 'quick-start', 'faq']
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
.documentation-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
}

.doc-header {
  text-align: center;
  margin-bottom: 3rem;
  padding-bottom: 2rem;
  border-bottom: 1px solid #E2E8F0;
}

.doc-header h1 {
  font-size: 2.5rem;
  color: #0C4A6E;
  margin-bottom: 0.5rem;
}

.doc-header .version {
  color: #64748B;
  font-size: 0.875rem;
}

.doc-content {
  display: grid;
  grid-template-columns: 200px 1fr;
  gap: 2rem;
}

.doc-nav {
  position: sticky;
  top: 2rem;
  height: fit-content;
}

.doc-nav ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.doc-nav li {
  margin-bottom: 0.5rem;
}

.doc-nav a {
  display: block;
  padding: 0.5rem 1rem;
  color: #64748B;
  text-decoration: none;
  border-radius: 0.5rem;
  transition: all 0.2s;
}

.doc-nav a:hover {
  background: #F1F5F9;
  color: #0EA5E9;
}

.doc-nav a.active {
  background: #E0F2FE;
  color: #0EA5E9;
  font-weight: 500;
}

.doc-main {
  min-width: 0;
}

.doc-section {
  margin-bottom: 3rem;
  padding-bottom: 2rem;
  border-bottom: 1px solid #E2E8F0;
}

.doc-section:last-child {
  border-bottom: none;
}

.doc-section h2 {
  font-size: 1.75rem;
  color: #0C4A6E;
  margin-bottom: 1.5rem;
  padding-bottom: 0.5rem;
  border-bottom: 2px solid #0EA5E9;
}

.doc-section h3 {
  font-size: 1.25rem;
  color: #334155;
  margin: 1.5rem 0 1rem;
}

.doc-section p {
  color: #475569;
  line-height: 1.7;
  margin-bottom: 1rem;
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
  margin: 1.5rem 0;
}

.feature-card {
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
  border-radius: 0.75rem;
  padding: 1.5rem;
  text-align: center;
}

.feature-icon {
  font-size: 2.5rem;
  margin-bottom: 1rem;
}

.feature-card h3 {
  margin: 0 0 0.5rem;
  color: #0C4A6E;
}

.feature-card p {
  margin: 0;
  font-size: 0.875rem;
}

.architecture-diagram {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  margin: 2rem 0;
  flex-wrap: wrap;
}

.arch-box {
  background: #F8FAFC;
  border: 2px solid #E2E8F0;
  border-radius: 0.75rem;
  padding: 1.5rem;
  min-width: 200px;
  text-align: center;
}

.arch-box.event-receiver {
  border-color: #10B981;
}

.arch-box.event-processor {
  border-color: #0EA5E9;
}

.arch-box.resource-pool {
  border-color: #8B5CF6;
}

.arch-box h4 {
  margin: 0 0 0.5rem;
  color: #0C4A6E;
}

.arch-box p {
  margin: 0 0 1rem;
  font-size: 0.75rem;
  color: #64748B;
}

.arch-box ul {
  list-style: none;
  padding: 0;
  margin: 0;
  text-align: left;
}

.arch-box li {
  font-size: 0.875rem;
  padding: 0.25rem 0;
  color: #475569;
}

.arch-arrow {
  font-size: 1.5rem;
  color: #94A3B8;
}

.doc-table {
  width: 100%;
  border-collapse: collapse;
  margin: 1rem 0;
  table-layout: fixed;
}

.doc-table th,
.doc-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border: 1px solid #E2E8F0;
}

.doc-table td {
  color: #475569;
  /* Apply text overflow to all table cells */
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Ensure inline elements in table cells also truncate */
.doc-table td > *,
.doc-table td > span,
.doc-table td > a {
  display: inline-block;
  max-width: 100%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: bottom;
}

.doc-table th {
  background: #F8FAFC;
  font-weight: 600;
  color: #0C4A6E;
  white-space: nowrap;
}

.flow-diagram {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  margin: 2rem 0;
  flex-wrap: wrap;
}

.flow-step {
  display: flex;
  flex-direction: column;
  align-items: center;
  background: #E0F2FE;
  border-radius: 0.5rem;
  padding: 0.75rem 1rem;
}

.step-num {
  width: 24px;
  height: 24px;
  background: #0EA5E9;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 600;
  margin-bottom: 0.25rem;
}

.step-text {
  font-size: 0.75rem;
  color: #0C4A6E;
  white-space: nowrap;
}

.flow-arrow {
  color: #94A3B8;
  font-size: 1.25rem;
}

.doc-list {
  padding-left: 1.5rem;
  margin: 1rem 0;
}

.doc-list li {
  margin-bottom: 0.5rem;
  color: #475569;
  line-height: 1.6;
}

.concept-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1rem;
  margin: 1rem 0;
}

.concept-item {
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
  border-radius: 0.5rem;
  padding: 1rem;
}

.concept-item h4 {
  margin: 0 0 0.5rem;
  color: #0C4A6E;
  font-size: 1rem;
}

.concept-item p {
  margin: 0;
  font-size: 0.875rem;
}

.status-flow {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  margin: 1.5rem 0;
  flex-wrap: wrap;
}

.status {
  padding: 0.5rem 1rem;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  font-weight: 500;
}

.status.available {
  background: #D1FAE5;
  color: #065F46;
}

.status.allocated {
  background: #FEF3C7;
  color: #92400E;
}

.status.in_use {
  background: #DBEAFE;
  color: #1E40AF;
}

.status.releasing {
  background: #FEE2E2;
  color: #991B1B;
}

.status-flow .arrow {
  color: #94A3B8;
}

.faq-item {
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
  border-radius: 0.5rem;
  padding: 1.5rem;
  margin-bottom: 1rem;
}

.faq-item h4 {
  margin: 0 0 0.75rem;
  color: #0C4A6E;
}

.faq-item p {
  margin: 0 0 0.5rem;
}

pre {
  background: #1E293B;
  color: #E2E8F0;
  padding: 1rem;
  border-radius: 0.5rem;
  overflow-x: auto;
  margin: 1rem 0;
}

code {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.875rem;
}

p code {
  background: #F1F5F9;
  color: #0EA5E9;
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
}

@media (max-width: 768px) {
  .doc-content {
    grid-template-columns: 1fr;
  }

  .doc-nav {
    position: static;
    margin-bottom: 2rem;
  }

  .doc-nav ul {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .doc-nav li {
    margin: 0;
  }

  .doc-nav a {
    padding: 0.375rem 0.75rem;
    font-size: 0.875rem;
  }

  .architecture-diagram {
    flex-direction: column;
  }

  .arch-arrow {
    transform: rotate(90deg);
  }
}
</style>
