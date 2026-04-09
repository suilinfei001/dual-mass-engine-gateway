<template>
  <div class="resources-page">
    <div class="page-header">
      <div class="header-content">
        <div class="header-icon">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>
        </div>
        <div class="header-text">
          <h1 class="page-title">可执行资源</h1>
          <p class="page-subtitle">管理 Azure DevOps 流水线资源和检查项</p>
        </div>
      </div>
      <button
        v-if="isLoggedIn && !isAdmin"
        class="btn btn-primary"
        @click="showCreateModal = true"
      >
        <svg class="btn-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        创建资源
      </button>
    </div>

    <div class="card">
      <div class="card-header">
        <div class="tabs-wrapper">
          <button
            class="tab"
            :class="{ active: activeTab === 'all' }"
            @click="activeTab = 'all'"
          >
            <svg class="tab-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
            所有资源
            <span class="tab-count">{{ resources.length }}</span>
          </button>
          <button
            v-if="isLoggedIn && !isAdmin"
            class="tab"
            :class="{ active: activeTab === 'my' }"
            @click="activeTab = 'my'"
          >
            <svg class="tab-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
            我的资源
            <span class="tab-count">{{ myResourcesCount }}</span>
          </button>
        </div>
      </div>

      <div v-if="loading" class="loading-state">
        <div class="spinner"></div>
        <p>加载中...</p>
      </div>

      <div v-else-if="resources.length === 0" class="empty-state">
        <div class="empty-icon-wrapper">
          <svg class="empty-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
          </svg>
        </div>
        <p class="empty-title">没有找到资源</p>
        <p class="empty-desc">{{ activeTab === 'my' ? '您还没有创建任何资源' : '暂无可用资源' }}</p>
      </div>

      <div v-else class="table-container">
        <table class="data-table">
          <thead>
            <tr>
              <th class="col-name">
                <div class="th-content">
                  <svg class="th-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
                  </svg>
                  <span>名称</span>
                </div>
              </th>
              <th class="col-type">
                <div class="th-content">
                  <svg class="th-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" />
                  </svg>
                  <span>类型</span>
                </div>
              </th>
              <th class="col-skip">
                <div class="th-content">
                  <svg class="th-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                  <span>跳过</span>
                </div>
              </th>
              <th v-if="activeTab === 'all'" class="col-creator">
                <div class="th-content">
                  <svg class="th-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                  <span>创建者</span>
                </div>
              </th>
              <th class="col-pipeline">
                <div class="th-content">
                  <svg class="th-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                  <span>流水线 ID</span>
                </div>
              </th>
              <th class="col-micro">
                <div class="th-content">
                  <svg class="th-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                  </svg>
                  <span>微服务</span>
                </div>
              </th>
              <th class="col-repo">
                <div class="th-content">
                  <svg class="th-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                  </svg>
                  <span>仓库路径</span>
                </div>
              </th>
              <th v-if="activeTab === 'my'" class="col-actions">
                <div class="th-content">
                  <svg class="th-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
                  </svg>
                  <span>操作</span>
                </div>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="resource in resources"
              :key="resource.id"
              class="data-row"
              :class="{ 'row-highlight': hoveredRow === resource.id }"
              @mouseenter="hoveredRow = resource.id"
              @mouseleave="hoveredRow = null"
            >
              <td class="cell-name" :title="resource.resource_name">
                <div class="resource-name-cell">
                  <span class="resource-name">{{ resource.resource_name }}</span>
                </div>
              </td>
              <td class="cell-type" :title="formatResourceType(resource.resource_type)">
                <span class="type-badge" :class="getTypeClass(resource.resource_type)">
                  <span class="type-dot"></span>
                  {{ formatResourceType(resource.resource_type) }}
                </span>
              </td>
              <td class="cell-skip" :title="resource.allow_skip ? '可跳过' : '—'">
                <span v-if="resource.allow_skip" class="skip-badge">
                  <svg class="skip-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                  </svg>
                  可跳过
                </span>
                <span v-else class="text-muted">—</span>
              </td>
              <td v-if="activeTab === 'all'" class="cell-creator" :title="resource.creator_name">
                <div class="creator-cell">
                  <div class="creator-avatar">{{ getInitial(resource.creator_name) }}</div>
                  <span class="creator-name">{{ resource.creator_name }}</span>
                </div>
              </td>
              <td class="cell-pipeline" :title="resource.pipeline_id">
                <code class="pipeline-code">{{ resource.pipeline_id }}</code>
              </td>
              <td class="cell-micro" :title="resource.microservice_name || '—'">
                <span v-if="resource.microservice_name" class="micro-badge">{{ resource.microservice_name }}</span>
                <span v-else class="text-muted">—</span>
              </td>
              <td class="cell-repo" :title="resource.repo_path">
                <div class="repo-cell">
                  <svg class="repo-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                  </svg>
                  <span class="repo-path">{{ resource.repo_path }}</span>
                </div>
              </td>
              <td v-if="activeTab === 'my'" class="cell-actions">
                <template v-if="canEdit(resource)">
                  <div class="action-buttons">
                    <button
                      class="action-btn action-btn-edit"
                      @click="editResource(resource)"
                      title="编辑资源"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                      </svg>
                    </button>
                    <button
                      class="action-btn action-btn-delete"
                      @click="confirmDelete(resource)"
                      title="删除资源"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>
                  </div>
                </template>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <Transition name="modal">
      <div v-if="showCreateModal || editingResource" class="modal-overlay" @click="closeModal">
        <div class="modal" @click.stop>
          <div class="modal-header">
            <div class="modal-title-section">
              <div class="modal-icon-wrapper" :class="{ 'modal-icon-edit': editingResource }">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path v-if="editingResource" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                  <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                </svg>
              </div>
              <div>
                <h3 class="modal-title">{{ editingResource ? '编辑资源' : '创建资源' }}</h3>
                <p class="modal-subtitle">配置 Azure DevOps 流水线资源</p>
              </div>
            </div>
            <button class="modal-close" @click="closeModal" aria-label="关闭">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <form @submit.prevent="handleSubmit" class="modal-form">
            <div class="modal-body">
              <div class="form-section">
                <div class="section-header">
                  <svg class="section-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                  <span>快速设置</span>
                </div>
                <div class="form-group skip-option">
                  <label class="checkbox-label">
                    <input
                      type="checkbox"
                      id="allowSkip"
                      v-model="form.allow_skip"
                    />
                    <span class="checkbox-custom"></span>
                    <span class="checkbox-text">允许跳过此检查项</span>
                  </label>
                  <p class="form-hint">勾选后，此检查项可以被跳过执行，无需配置 Azure DevOps</p>
                </div>
              </div>

              <template v-if="form.allow_skip">
                <div class="skip-mode-notice">
                  <svg class="notice-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <div>
                    <p class="notice-title">跳过模式</p>
                    <p class="notice-desc">只需填写基本信息，Azure 配置将自动设置为空</p>
                  </div>
                </div>

                <div class="form-section">
                  <div class="section-header">
                    <svg class="section-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    <span>基本信息</span>
                  </div>

                  <div class="form-group">
                    <label for="resourceType">资源类型 <span class="required">*</span></label>
                    <select
                      id="resourceType"
                      v-model="form.resource_type"
                      class="form-control"
                      :disabled="editingResource"
                      :class="{ 'form-control-disabled': editingResource }"
                      required
                    >
                      <option value="">选择类型</option>
                      <option value="basic_ci_all">基础 CI 全量</option>
                      <option value="deployment_deployment">部署</option>
                      <option value="specialized_tests_api_test">API 测试</option>
                      <option value="specialized_tests_module_e2e">模块 E2E</option>
                      <option value="specialized_tests_agent_e2e">代理 E2E</option>
                      <option value="specialized_tests_ai_e2e">AI E2E</option>
                    </select>
                    <p v-if="editingResource" class="form-hint">资源类型不可更改</p>
                  </div>

                  <div class="form-group">
                    <label for="repoPath">仓库路径 <span class="required">*</span></label>
                    <input
                      type="text"
                      id="repoPath"
                      v-model="form.repo_path"
                      class="form-control"
                      required
                      placeholder="例如：github.com/org/repo"
                    />
                  </div>

                  <div class="form-group">
                    <label for="description">描述</label>
                    <textarea
                      id="description"
                      v-model="form.description"
                      class="form-control"
                      rows="3"
                      placeholder="此检查项可以跳过"
                    ></textarea>
                  </div>
                </div>
              </template>

              <template v-else>
                <div class="form-section">
                  <div class="section-header">
                    <svg class="section-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    <span>基本信息</span>
                  </div>

                  <div class="form-row">
                    <div class="form-group">
                      <label for="resourceType">资源类型 <span class="required">*</span></label>
                      <select
                        id="resourceType"
                        v-model="form.resource_type"
                        class="form-control"
                        :disabled="editingResource"
                        :class="{ 'form-control-disabled': editingResource }"
                        required
                      >
                        <option value="">选择类型</option>
                        <option value="basic_ci_all">基础 CI 全量</option>
                        <option value="deployment_deployment">部署</option>
                        <option value="specialized_tests_api_test">API 测试</option>
                        <option value="specialized_tests_module_e2e">模块 E2E</option>
                        <option value="specialized_tests_agent_e2e">代理 E2E</option>
                        <option value="specialized_tests_ai_e2e">AI E2E</option>
                      </select>
                      <p v-if="editingResource" class="form-hint">资源类型不可更改</p>
                    </div>
                    <div class="form-group">
                      <label for="repoPath">仓库路径 <span class="required">*</span></label>
                      <input
                        type="text"
                        id="repoPath"
                        v-model="form.repo_path"
                        class="form-control"
                        required
                        placeholder="例如：github.com/org/repo"
                      />
                    </div>
                  </div>
                </div>

                <div class="form-section">
                  <div class="section-header">
                    <svg class="section-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                    </svg>
                    <span>Azure DevOps 配置</span>
                  </div>

                  <div class="form-row">
                    <div class="form-group">
                      <label for="organization">组织 <span class="required">*</span></label>
                      <input
                        type="text"
                        id="organization"
                        v-model="form.organization"
                        class="form-control"
                        required
                        placeholder="例如：MyOrg"
                      />
                    </div>
                    <div class="form-group">
                      <label for="project">项目 <span class="required">*</span></label>
                      <input
                        type="text"
                        id="project"
                        v-model="form.project"
                        class="form-control"
                        required
                        placeholder="例如：MyProject"
                      />
                    </div>
                  </div>

                  <div class="form-row">
                    <div class="form-group">
                      <label for="pipelineId">流水线 ID <span class="required">*</span></label>
                      <input
                        type="number"
                        id="pipelineId"
                        v-model.number="form.pipeline_id"
                        class="form-control"
                        required
                        placeholder="例如：123"
                      />
                    </div>
                    <div class="form-group">
                      <label for="microserviceName">微服务名称</label>
                      <input
                        type="text"
                        id="microserviceName"
                        v-model="form.microservice_name"
                        class="form-control"
                        placeholder="可选"
                      />
                    </div>
                  </div>

                  <div class="form-group">
                    <label for="pipelineParams">流水线参数 (JSON)</label>
                    <textarea
                      id="pipelineParams"
                      v-model="form.pipeline_params"
                      class="form-control form-control-code"
                      rows="5"
                      placeholder='{"TRIVY_EXIT_CODE": "1", "BUILD_TYPE": "opensource"}'
                    ></textarea>
                    <p class="form-hint">JSON 格式的流水线参数配置</p>
                  </div>

                  <div class="form-row">
                    <div class="form-group">
                      <label for="podName">Pod 名称</label>
                      <input
                        type="text"
                        id="podName"
                        v-model="form.pod_name"
                        class="form-control"
                        placeholder="可选"
                      />
                    </div>
                  </div>

                  <div class="form-group">
                    <label for="description">描述</label>
                    <textarea
                      id="description"
                      v-model="form.description"
                      class="form-control"
                      rows="3"
                      placeholder="可选，输入资源的详细描述"
                    ></textarea>
                  </div>
                </div>
              </template>
            </div>

            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="closeModal">取消</button>
              <button type="submit" class="btn btn-primary" :disabled="submitting">
                <svg v-if="submitting" class="btn-spinner" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="spinner-circle" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" stroke-dasharray="32" stroke-dashoffset="32" />
                </svg>
                <span>{{ submitting ? '保存中...' : '保存' }}</span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </Transition>

    <Transition name="modal">
      <div v-if="deletingResource" class="modal-overlay" @click="deletingResource = null">
        <div class="modal modal-small" @click.stop>
          <div class="modal-header modal-header-danger">
            <div class="modal-title-section">
              <div class="modal-icon-wrapper modal-icon-danger">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
              </div>
              <div>
                <h3 class="modal-title">确认删除</h3>
                <p class="modal-subtitle">此操作不可撤销</p>
              </div>
            </div>
            <button class="modal-close" @click="deletingResource = null" aria-label="关闭">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div class="modal-body">
            <div class="confirm-content">
              <p class="confirm-message">确定要删除资源 <strong>"{{ deletingResource.resource_name }}"</strong> 吗？</p>
              <div class="confirm-warning">
                <svg class="warning-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
                <span>删除后将无法恢复</span>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="deletingResource = null">取消</button>
            <button class="btn btn-danger" @click="deleteResource" :disabled="submitting">
              <svg v-if="submitting" class="btn-spinner" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="spinner-circle" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" stroke-dasharray="32" stroke-dashoffset="32" />
              </svg>
              <span>{{ submitting ? '删除中...' : '确认删除' }}</span>
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script>
import { ref, computed, onMounted, watch, inject } from 'vue'
import { useDialog } from '../composables/useDialog'

export default {
  name: 'Resources',
  setup() {
    const dialog = useDialog()

    const resources = ref([])
    const loading = ref(false)
    const submitting = ref(false)
    const activeTab = ref('all')
    const showCreateModal = ref(false)
    const editingResource = ref(null)
    const deletingResource = ref(null)
    const hoveredRow = ref(null)
    
    const isLoggedIn = computed(() => localStorage.getItem('isLoggedIn') === 'true')
    const isAdmin = computed(() => localStorage.getItem('userRole') === 'admin')
    
    const form = ref({
      resource_name: '',
      resource_type: '',
      allow_skip: false,
      organization: '',
      project: '',
      pipeline_id: null,
      pipeline_params: '',
      microservice_name: '',
      pod_name: '',
      repo_path: '',
      description: ''
    })
    
    const fetchResources = async () => {
      loading.value = true
      try {
        const url = activeTab.value === 'my' ? '/api/resources/my' : '/api/resources'
        const response = await fetch(url, { credentials: 'include' })
        const data = await response.json()
        resources.value = data.data || []
      } catch (error) {
        console.error('Failed to fetch resources:', error)
      } finally {
        loading.value = false
      }
    }
    
    const canEdit = (resource) => {
      if (!isLoggedIn.value) return false
      if (isAdmin.value) return true
      const username = localStorage.getItem('username')
      return resource.creator_name === username
    }
    
    const editResource = (resource) => {
      editingResource.value = resource
      const pipelineParamsJson = resource.pipeline_params
        ? JSON.stringify(resource.pipeline_params, null, 2)
        : ''

      form.value = {
        resource_name: resource.resource_name,
        resource_type: resource.resource_type,
        allow_skip: resource.allow_skip || false,
        organization: resource.organization || '',
        project: resource.project || '',
        pipeline_id: resource.pipeline_id,
        pipeline_params: pipelineParamsJson,
        microservice_name: resource.microservice_name || '',
        pod_name: resource.pod_name || '',
        repo_path: resource.repo_path,
        description: resource.description || ''
      }
    }

    const closeModal = () => {
      showCreateModal.value = false
      editingResource.value = null
      form.value = {
        resource_name: '',
        resource_type: '',
        allow_skip: false,
        organization: '',
        project: '',
        pipeline_id: null,
        pipeline_params: '',
        microservice_name: '',
        pod_name: '',
        repo_path: '',
        description: ''
      }
    }
    
    const handleSubmit = async () => {
      submitting.value = true
      try {
        const url = editingResource.value
          ? `/api/resources/${editingResource.value.id}`
          : '/api/resources'
        
        const method = editingResource.value ? 'PUT' : 'POST'
        
        const payload = { ...form.value }
        
        if (payload.pipeline_params) {
          try {
            payload.pipeline_params = JSON.parse(payload.pipeline_params)
          } catch (e) {
            dialog.alertError('流水线参数 JSON 格式不正确')
            submitting.value = false
            return
          }
        }
        
        const response = await fetch(url, {
          method,
          headers: {
            'Content-Type': 'application/json'
          },
          credentials: 'include',
          body: JSON.stringify(payload)
        })
        
        const data = await response.json()
        
        if (data.success) {
          dialog.alertSuccess(editingResource.value ? '资源更新成功' : '资源创建成功')
          closeModal()
          fetchResources()
        } else {
          dialog.alertError(data.message || '操作失败')
        }
      } catch (error) {
        console.error('Failed to submit:', error)
        dialog.alertError('操作失败: ' + error.message)
      } finally {
        submitting.value = false
      }
    }
    
    const confirmDelete = (resource) => {
      deletingResource.value = resource
    }
    
    const deleteResource = async () => {
      submitting.value = true
      try {
        const response = await fetch(`/api/resources/${deletingResource.value.id}`, {
          method: 'DELETE',
          credentials: 'include'
        })
        
        const data = await response.json()
        
        if (data.success) {
          dialog.alertSuccess('资源删除成功')
          deletingResource.value = null
          fetchResources()
        } else {
          dialog.alertError(data.message || '删除失败')
        }
      } catch (error) {
        console.error('Failed to delete:', error)
        dialog.alertError('删除失败: ' + error.message)
      } finally {
        submitting.value = false
      }
    }
    
    const getTypeClass = (type) => {
      const classes = {
        'basic_ci_all': 'type-basic',
        'deployment_deployment': 'type-deployment',
        'specialized_tests_api_test': 'type-api',
        'specialized_tests_module_e2e': 'type-module',
        'specialized_tests_agent_e2e': 'type-agent',
        'specialized_tests_ai_e2e': 'type-ai'
      }
      return classes[type] || 'type-default'
    }
    
    const formatResourceType = (type) => {
      const names = {
        'basic_ci_all': '基础 CI',
        'deployment_deployment': '部署',
        'specialized_tests_api_test': 'API 测试',
        'specialized_tests_module_e2e': '模块 E2E',
        'specialized_tests_agent_e2e': '代理 E2E',
        'specialized_tests_ai_e2e': 'AI E2E'
      }
      return names[type] || type
    }
    
    const getInitial = (name) => {
      if (!name) return '?'
      return name.charAt(0).toUpperCase()
    }
    
    const myResourcesCount = computed(() => {
      if (!isLoggedIn.value) return 0
      const username = localStorage.getItem('username')
      return resources.value.filter(r => r.creator_name === username).length
    })
    
    watch(activeTab, () => {
      fetchResources()
    })
    
    onMounted(() => {
      fetchResources()
    })
    
    return {
      resources,
      loading,
      submitting,
      activeTab,
      showCreateModal,
      editingResource,
      deletingResource,
      hoveredRow,
      form,
      isLoggedIn,
      isAdmin,
      myResourcesCount,
      fetchResources,
      canEdit,
      editResource,
      closeModal,
      handleSubmit,
      confirmDelete,
      deleteResource,
      getTypeClass,
      formatResourceType,
      getInitial
    }
  }
}
</script>

<style scoped>
.resources-page {
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
  gap: 1rem;
  flex-wrap: wrap;
}

.header-content {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.header-icon {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #7C3AED 0%, #A78BFA 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(124, 58, 237, 0.25);
}

.header-icon svg {
  width: 24px;
  height: 24px;
  color: white;
}

.header-text {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.page-title {
  font-size: 1.75rem;
  font-weight: 700;
  color: #0F172A;
  margin: 0;
}

.page-subtitle {
  font-size: 0.875rem;
  color: #64748B;
  margin: 0;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.625rem 1.25rem;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 600;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn:focus-visible {
  outline: 2px solid #7C3AED;
  outline-offset: 2px;
}

.btn-primary {
  background: linear-gradient(135deg, #7C3AED 0%, #6D28D9 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(124, 58, 237, 0.25);
}

.btn-primary:hover {
  background: linear-gradient(135deg, #6D28D9 0%, #5B21B6 100%);
  box-shadow: 0 4px 12px rgba(124, 58, 237, 0.35);
  transform: translateY(-1px);
}

.btn-secondary {
  background: #F1F5F9;
  color: #475569;
}

.btn-secondary:hover {
  background: #E2E8F0;
  color: #334155;
}

.btn-danger {
  background: linear-gradient(135deg, #EF4444 0%, #DC2626 100%);
  color: white;
}

.btn-danger:hover {
  background: linear-gradient(135deg, #DC2626 0%, #B91C1C 100%);
}

.btn-icon {
  width: 18px;
  height: 18px;
}

.btn-spinner {
  width: 16px;
  height: 16px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.spinner-circle {
  animation: spinner-dash 1.5s ease-in-out infinite;
}

@keyframes spinner-dash {
  0% { stroke-dashoffset: 32; }
  50% { stroke-dashoffset: 8; }
  100% { stroke-dashoffset: 32; }
}

.card {
  background: white;
  border-radius: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 4px 12px rgba(0, 0, 0, 0.05);
  overflow: hidden;
}

.card-header {
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #F1F5F9;
  background: #FAFBFC;
}

.tabs-wrapper {
  display: flex;
  gap: 0.5rem;
}

.tab {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 1rem;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: #64748B;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.tab:hover {
  background: #F1F5F9;
  color: #334155;
}

.tab.active {
  background: linear-gradient(135deg, #7C3AED 0%, #6D28D9 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(124, 58, 237, 0.25);
}

.tab-icon {
  width: 18px;
  height: 18px;
}

.tab-count {
  padding: 0.125rem 0.5rem;
  border-radius: 10px;
  font-size: 0.75rem;
  font-weight: 600;
  background: rgba(255, 255, 255, 0.2);
}

.tab:not(.active) .tab-count {
  background: #E2E8F0;
  color: #64748B;
}

.loading-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  text-align: center;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #E2E8F0;
  border-top-color: #7C3AED;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

.loading-state p {
  color: #64748B;
  margin: 0;
}

.empty-icon-wrapper {
  width: 80px;
  height: 80px;
  background: linear-gradient(135deg, #F1F5F9 0%, #E2E8F0 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.5rem;
}

.empty-icon {
  width: 40px;
  height: 40px;
  color: #94A3B8;
}

.empty-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: #334155;
  margin: 0 0 0.5rem;
}

.empty-desc {
  font-size: 0.875rem;
  color: #94A3B8;
  margin: 0;
}

.table-container {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
}

.data-table th {
  padding: 0.875rem 1rem;
  text-align: left;
  font-size: 0.75rem;
  font-weight: 600;
  color: #64748B;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  background: #FAFBFC;
  border-bottom: 1px solid #E2E8F0;
  white-space: nowrap;
}

.th-content {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.th-icon {
  width: 14px;
  height: 14px;
  opacity: 0.6;
}

.data-table td {
  padding: 1rem;
  border-bottom: 1px solid #F1F5F9;
  vertical-align: middle;
  /* Apply text overflow to all table cells */
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Ensure inline elements in table cells also truncate */
.data-table td > span,
.data-table td > a {
  display: inline-block;
  max-width: 100%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: bottom;
}

/* Ensure action buttons keep flex layout */
.data-table td > .action-buttons {
  display: flex;
}

.data-row {
  transition: background-color 0.15s ease;
}

.data-row:hover,
.row-highlight {
  background: linear-gradient(90deg, rgba(124, 58, 237, 0.04) 0%, rgba(124, 58, 237, 0.01) 100%);
}

.resource-name-cell {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.resource-name {
  font-weight: 600;
  color: #1E293B;
  font-size: 0.9375rem;
}

.type-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 600;
}

.type-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

.type-basic {
  background: #EFF6FF;
  color: #2563EB;
}

.type-deployment {
  background: #FEF2F2;
  color: #DC2626;
}

.type-api {
  background: #ECFDF5;
  color: #059669;
}

.type-module {
  background: #FFFBEB;
  color: #D97706;
}

.type-agent {
  background: #F5F3FF;
  color: #7C3AED;
}

.type-ai {
  background: #FDF2F8;
  color: #DB2777;
}

.type-default {
  background: #F8FAFC;
  color: #64748B;
}

.skip-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  background: linear-gradient(135deg, #FFFBEB 0%, #FEF3C7 100%);
  color: #D97706;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 600;
}

.skip-icon {
  width: 14px;
  height: 14px;
}

.text-muted {
  color: #94A3B8;
}

.creator-cell {
  display: flex;
  align-items: center;
  gap: 0.625rem;
}

.creator-avatar {
  width: 28px;
  height: 28px;
  background: linear-gradient(135deg, #7C3AED 0%, #A78BFA 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 0.75rem;
  font-weight: 600;
}

.creator-name {
  font-size: 0.875rem;
  color: #475569;
}

.pipeline-code {
  display: inline-block;
  padding: 0.25rem 0.625rem;
  background: #F8FAFC;
  color: #475569;
  border-radius: 4px;
  font-family: 'SF Mono', 'Fira Code', monospace;
  font-size: 0.8125rem;
  border: 1px solid #E2E8F0;
}

.micro-badge {
  display: inline-block;
  padding: 0.25rem 0.625rem;
  background: #F0F9FF;
  color: #0369A1;
  border-radius: 4px;
  font-size: 0.8125rem;
  font-weight: 500;
}

.repo-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.repo-icon {
  width: 16px;
  height: 16px;
  color: #94A3B8;
  flex-shrink: 0;
}

.repo-path {
  font-family: 'SF Mono', 'Fira Code', monospace;
  font-size: 0.8125rem;
  color: #64748B;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 200px;
}

.action-buttons {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  gap: 0.5rem;
  align-items: center;
}

.action-btn {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.action-btn svg {
  width: 16px;
  height: 16px;
}

.action-btn-edit {
  background: #F1F5F9;
  color: #64748B;
}

.action-btn-edit:hover {
  background: #7C3AED;
  color: white;
}

.action-btn-delete {
  background: #FEF2F2;
  color: #DC2626;
}

.action-btn-delete:hover {
  background: #DC2626;
  color: white;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  padding: 1rem;
}

.modal {
  background: white;
  border-radius: 20px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  width: 100%;
  max-width: 640px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-small {
  max-width: 440px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid #F1F5F9;
  background: linear-gradient(180deg, #FAFBFC 0%, white 100%);
}

.modal-header-danger {
  background: linear-gradient(180deg, #FEF2F2 0%, white 100%);
}

.modal-title-section {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.modal-icon-wrapper {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #7C3AED 0%, #6D28D9 100%);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(124, 58, 237, 0.25);
}

.modal-icon-wrapper svg {
  width: 20px;
  height: 20px;
  color: white;
}

.modal-icon-edit {
  background: linear-gradient(135deg, #3B82F6 0%, #2563EB 100%);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.25);
}

.modal-icon-danger {
  background: linear-gradient(135deg, #EF4444 0%, #DC2626 100%);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.25);
}

.modal-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: #0F172A;
  margin: 0;
}

.modal-subtitle {
  font-size: 0.8125rem;
  color: #64748B;
  margin: 0.25rem 0 0;
}

.modal-close {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  border: none;
  background: transparent;
  color: #94A3B8;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.modal-close:hover {
  background: #F1F5F9;
  color: #334155;
}

.modal-close svg {
  width: 20px;
  height: 20px;
}

.modal-form {
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
}

.modal-body {
  padding: 1.5rem;
  overflow-y: auto;
  flex: 1;
}

.form-section {
  margin-bottom: 1.5rem;
}

.form-section:last-child {
  margin-bottom: 0;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid #F1F5F9;
}

.section-icon {
  width: 18px;
  height: 18px;
  color: #7C3AED;
}

.section-header span {
  font-size: 0.875rem;
  font-weight: 600;
  color: #334155;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-row .form-group {
  margin-bottom: 0;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-group label {
  display: block;
  font-size: 0.8125rem;
  font-weight: 600;
  color: #374151;
  margin-bottom: 0.375rem;
}

.required {
  color: #EF4444;
}

.form-control {
  width: 100%;
  padding: 0.625rem 0.875rem;
  border: 1px solid #E2E8F0;
  border-radius: 8px;
  font-size: 0.875rem;
  color: #1F2937;
  background: white;
  transition: all 0.2s ease;
}

.form-control:focus {
  outline: none;
  border-color: #7C3AED;
  box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.1);
}

.form-control::placeholder {
  color: #9CA3AF;
}

.form-control-code {
  font-family: 'SF Mono', 'Fira Code', monospace;
  font-size: 0.8125rem;
}

textarea.form-control {
  resize: vertical;
  min-height: 100px;
}

.form-control:disabled,
.form-control-disabled {
  background: #F8FAFC;
  color: #6B7280;
  cursor: not-allowed;
  border-color: #E5E7EB;
}

.form-hint {
  font-size: 0.75rem;
  color: #6B7280;
  margin: 0.375rem 0 0;
}

.skip-option {
  padding: 1rem;
  background: linear-gradient(135deg, #FFFBEB 0%, #FEF3C7 100%);
  border: 2px solid #FCD34D;
  border-radius: 12px;
  margin-bottom: 0;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  cursor: pointer;
  user-select: none;
}

.checkbox-label input[type="checkbox"] {
  display: none;
}

.checkbox-custom {
  width: 22px;
  height: 22px;
  border: 2px solid #F59E0B;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  background: white;
}

.checkbox-label input[type="checkbox"]:checked + .checkbox-custom {
  background: #F59E0B;
  border-color: #F59E0B;
}

.checkbox-label input[type="checkbox"]:checked + .checkbox-custom::after {
  content: '';
  width: 6px;
  height: 10px;
  border: solid white;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
  margin-bottom: 2px;
}

.checkbox-text {
  font-weight: 600;
  font-size: 0.9375rem;
  color: #92400E;
}

.skip-mode-notice {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 1rem;
  background: linear-gradient(135deg, #EFF6FF 0%, #DBEAFE 100%);
  border: 1px solid #93C5FD;
  border-radius: 12px;
  margin-bottom: 1.5rem;
}

.notice-icon {
  width: 20px;
  height: 20px;
  color: #3B82F6;
  flex-shrink: 0;
  margin-top: 0.125rem;
}

.notice-title {
  font-weight: 600;
  color: #1D4ED8;
  margin: 0 0 0.25rem;
  font-size: 0.875rem;
}

.notice-desc {
  color: #3B82F6;
  margin: 0;
  font-size: 0.8125rem;
}

.confirm-content {
  text-align: center;
}

.confirm-message {
  font-size: 1rem;
  color: #334155;
  margin: 0 0 1rem;
}

.confirm-message strong {
  color: #0F172A;
}

.confirm-warning {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: #FEF2F2;
  border-radius: 8px;
  color: #DC2626;
  font-size: 0.8125rem;
  font-weight: 500;
}

.warning-icon {
  width: 16px;
  height: 16px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid #F1F5F9;
  background: #FAFBFC;
}

.modal-enter-active,
.modal-leave-active {
  transition: all 0.25s ease;
}

.modal-enter-active .modal,
.modal-leave-active .modal {
  transition: all 0.25s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal,
.modal-leave-to .modal {
  transform: scale(0.95) translateY(-10px);
  opacity: 0;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: stretch;
  }

  .header-content {
    flex-direction: column;
    align-items: flex-start;
    text-align: left;
  }

  .btn {
    width: 100%;
  }

  .form-row {
    grid-template-columns: 1fr;
  }

  .modal {
    max-height: 95vh;
    border-radius: 16px;
  }

  .modal-header,
  .modal-body,
  .modal-footer {
    padding-left: 1rem;
    padding-right: 1rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .btn,
  .tab,
  .action-btn,
  .data-row,
  .modal-enter-active,
  .modal-leave-active {
    transition: none;
  }

  .spinner {
    animation: none;
  }
}
</style>
