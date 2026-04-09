<template>
  <div class="my-resource-instance-page">
    <div class="page-header">
      <div class="header-icon">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
        </svg>
      </div>
      <div class="header-content">
        <h1>我创建的资源实例</h1>
        <p>管理您的虚拟机和物理机资源</p>
      </div>
      <button class="btn btn-primary" @click="showCreateModal = true">
        <svg class="btn-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        添加资源实例
      </button>
    </div>

    <div class="content-container">
      <div v-if="loading" class="loading-container">
        <div class="spinner"></div>
        <span>加载中...</span>
      </div>

      <div v-else-if="resources.length === 0" class="empty-state">
        <div class="empty-icon">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
          </svg>
        </div>
        <h3>还没有资源实例</h3>
        <p>创建您的第一个资源实例来开始使用</p>
        <button class="btn btn-primary" @click="showCreateModal = true">
          <svg class="btn-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          立即添加
        </button>
      </div>

      <div v-else class="resources-grid">
        <div v-for="resource in resources" :key="resource.uuid" class="resource-card">
          <div class="card-header">
            <div class="card-title-row">
              <div class="resource-icon" :class="getTypeClass(resource.instance_type || resource.resource_type)">
                <svg v-if="(resource.instance_type || resource.resource_type) === 'virtual_machine' || (resource.instance_type || resource.resource_type) === 'VirtualMachine'" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>
                <svg v-else xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
                </svg>
              </div>
              <div class="card-title-info">
                <h3>{{ resource.name || resource.ip_address || resource.host || '未命名' }}</h3>
                <span class="resource-type">{{ getTypeLabel(resource.instance_type || resource.resource_type) }}</span>
              </div>
            </div>
            <span class="status-badge" :class="getStatusClass(resource.status)">
              <span class="status-dot"></span>
              {{ getStatusLabel(resource.status) }}
            </span>
          </div>

          <div class="card-body">
            <div class="info-grid">
              <div class="info-item">
                <span class="info-label">主机地址</span>
                <span class="info-value host-value">{{ resource.ip_address || resource.host || '-' }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">SSH 用户</span>
                <span class="info-value">{{ resource.ssh_user || 'root' }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">SSH 端口</span>
                <span class="info-value">{{ resource.port || resource.ssh_port || '-' }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">可见性</span>
                <span class="info-value">
                  <span v-if="resource.is_public" class="visibility-badge visibility-public">公开</span>
                  <span v-else class="visibility-badge visibility-private">私有</span>
                </span>
              </div>
            </div>

            <div v-if="resource.snapshot_id || resource.snapshot_instance_uuid" class="snapshot-section">
              <div class="snapshot-label">快照信息</div>
              <div class="snapshot-info">
                <div v-if="resource.snapshot_id" class="snapshot-item">
                  <span class="snapshot-key">ID:</span>
                  <span class="snapshot-value">{{ resource.snapshot_id }}</span>
                </div>
                <div v-if="resource.snapshot_instance_uuid" class="snapshot-item">
                  <span class="snapshot-key">实例:</span>
                  <span class="snapshot-value">{{ resource.snapshot_instance_uuid }}</span>
                </div>
              </div>
            </div>

            <div v-if="resource.description" class="description-section">
              <div class="description-label">描述</div>
              <p class="description-text">{{ resource.description }}</p>
            </div>

            <div class="meta-info">
              <span class="meta-item">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
                {{ resource.created_by || '-' }}
              </span>
              <span class="meta-item">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                {{ formatTime(resource.created_at) }}
              </span>
            </div>
          </div>

          <div class="card-footer">
            <button class="action-btn action-btn-edit" @click="editResource(resource)" title="编辑">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
              编辑
            </button>
            <button
              v-if="(resource.instance_type || resource.resource_type) !== 'physical_machine' && (resource.instance_type || resource.resource_type) !== 'Machine'"
              class="action-btn action-btn-info"
              @click="viewTestbeds(resource)"
              title="查看关联的 Testbed"
            >
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
              </svg>
              Testbed
            </button>
            <button
              class="action-btn action-btn-test"
              @click="testConnection(resource)"
              title="测试连接"
              :disabled="testing === resource.uuid"
            >
              <svg v-if="testing !== resource.uuid" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.398 4.914a15 15 0 0121.203 0M12 12h.01" />
              </svg>
              <span v-if="testing === resource.uuid" class="btn-spinner"></span>
              {{ testing === resource.uuid ? '测试中' : '测试' }}
            </button>
            <button
              v-if="resource.testbed_count === 0"
              class="action-btn action-btn-danger"
              @click="deleteResource(resource)"
              title="删除"
            >
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
              删除
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showCreateModal || showEditModal" class="modal-overlay" @click.self="closeModals">
      <div class="modal modal-medium">
        <div class="modal-header">
          <div class="modal-header-content">
            <div class="modal-icon">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
              </svg>
            </div>
            <h3>{{ showEditModal ? '编辑资源实例' : '添加资源实例' }}</h3>
          </div>
          <button class="modal-close" @click="closeModals">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <form @submit.prevent="saveResource" class="modal-form">
          <div class="form-section">
            <div class="section-header">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <h4>基本信息</h4>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label for="name">名称 <span class="required">*</span></label>
                <input id="name" v-model="formData.name" type="text" class="form-control" placeholder="如: 测试数据库服务器" required />
              </div>
              <div class="form-group">
                <label for="type">资源类型 <span class="required">*</span></label>
                <select id="type" v-model="formData.resource_type" class="form-control" required>
                  <option value="">选择类型</option>
                  <option value="virtual_machine">虚拟机</option>
                  <option value="physical_machine">物理机</option>
                </select>
              </div>
            </div>
            <div class="form-group">
              <label for="description">描述</label>
              <textarea id="description" v-model="formData.description" class="form-control" rows="2" placeholder="请简要描述该资源实例的用途"></textarea>
            </div>
          </div>

          <div class="form-section">
            <div class="section-header">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              <h4>连接信息</h4>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label for="host">主机地址 <span class="required">*</span></label>
                <input id="host" v-model="formData.host" type="text" class="form-control" placeholder="192.168.1.100" required />
              </div>
              <div class="form-group">
                <label for="sshPort">SSH 端口 <span class="required">*</span></label>
                <input id="sshPort" v-model.number="formData.ssh_port" type="number" class="form-control" placeholder="22" required />
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label for="sshUser">SSH 用户名 <span class="required">*</span></label>
                <input id="sshUser" v-model="formData.ssh_user" type="text" class="form-control" placeholder="root" required />
              </div>
              <div class="form-group">
                <label for="passwd">SSH 密码 <span class="required">*</span></label>
                <input id="passwd" v-model="formData.passwd" type="password" class="form-control" placeholder="请输入 SSH 密码" required />
              </div>
            </div>
          </div>

          <div v-if="formData.resource_type === 'virtual_machine'" class="form-section vm-section">
            <div class="section-header">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
              </svg>
              <h4>虚拟机设置</h4>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label for="snapshotId">快照 ID</label>
                <input id="snapshotId" v-model="formData.snapshot_id" type="text" class="form-control" placeholder="snapshot-v1.0" />
                <small class="form-hint">用于回滚的快照标识</small>
              </div>
              <div class="form-group">
                <label for="snapshotInstanceUuid">快照实例ID</label>
                <input id="snapshotInstanceUuid" v-model="formData.snapshot_instance_uuid" type="text" class="form-control" placeholder="ins-xxxxxxxx" />
                <small class="form-hint">虚拟机实例的唯一标识</small>
              </div>
            </div>
          </div>

          <div class="form-section permission-section">
            <div class="section-header">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
              <h4>权限设置</h4>
            </div>
            <div class="permission-card">
              <label class="toggle-label">
                <input type="checkbox" v-model="formData.is_public" :disabled="formData.resource_type === 'virtual_machine'" class="toggle-input" />
                <span class="toggle-slider"></span>
                <div class="toggle-content">
                  <div class="toggle-header">
                    <svg class="toggle-icon-public" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <span class="toggle-title">{{ formData.is_public ? '公开资源' : '私有资源' }}</span>
                  </div>
                  <p class="toggle-description">{{ formData.is_public ? '允许其他用户查看并使用此资源实例' : '仅您可以查看和使用此资源实例' }}</p>
                </div>
              </label>
              <p v-if="formData.resource_type === 'virtual_machine'" class="permission-hint">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                虚拟机资源默认为公开资源，不可修改
              </p>
            </div>
          </div>

          <div class="modal-footer">
            <div class="footer-left">
              <button type="button" class="btn btn-secondary" @click="closeModals">取消</button>
            </div>
            <div class="footer-right">
              <button type="button" class="btn btn-outline" :disabled="testingConnection" @click="testNewConnection">
                <svg v-if="!testingConnection" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.398 4.914a15 15 0 0121.203 0M12 12h.01" />
                </svg>
                {{ testingConnection ? '测试中...' : '测试连接' }}
              </button>
              <button type="submit" class="btn btn-primary" :disabled="submitting">
                <svg v-if="!submitting" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                </svg>
                {{ submitting ? '保存中...' : '保存' }}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, watch } from 'vue'
import { externalAPI, adminAPI } from '../../api/resourcePool'
import { useDialog } from '../../composables/useDialog'

export default {
  name: 'MyResourceInstanceList',
  setup() {
    const dialog = useDialog()
    const resources = ref([])
    const loading = ref(false)
    const submitting = ref(false)
    const testing = ref(null)
    const showCreateModal = ref(false)
    const showEditModal = ref(false)
    const editingResource = ref(null)

    const currentUsername = computed(() => localStorage.getItem('username') || '')

    const formData = ref({
      name: '',
      resource_type: 'virtual_machine',
      host: '',
      ssh_port: 22,
      ssh_user: 'root',
      passwd: '',
      snapshot_id: '',
      snapshot_instance_uuid: '',
      description: '',
      is_public: true
    })

    const testingConnection = ref(false)

    const testNewConnection = async () => {
      if (!formData.value.host || !formData.value.ssh_port || !formData.value.ssh_user) {
        dialog.alertWarning('请填写主机地址、SSH端口和用户名')
        return
      }
      testingConnection.value = true
      try {
        const result = await adminAPI.testConnection({
          host: formData.value.host,
          ssh_port: formData.value.ssh_port,
          ssh_user: formData.value.ssh_user,
          password: formData.value.passwd
        })
        if (result.success) {
          dialog.alertSuccess('连接成功！')
        } else {
          dialog.alertError('连接失败: ' + (result.message || '未知错误'))
        }
      } catch (error) {
        dialog.alertError('测试连接失败: ' + error.message)
      } finally {
        testingConnection.value = false
      }
    }

    const fetchResources = async () => {
      loading.value = true
      try {
        const data = await externalAPI.getMyResourceInstances()
        resources.value = data.data || data.instances || []
      } catch (error) {
        console.error('Failed to fetch resources:', error)
        dialog.alertError('获取资源实例列表失败: ' + error.message)
      } finally {
        loading.value = false
      }
    }

    const saveResource = async () => {
      submitting.value = true
      try {
        const payload = {
          name: formData.value.name,
          resource_type: formData.value.resource_type,
          host: formData.value.host,
          ssh_port: formData.value.ssh_port,
          ssh_user: formData.value.ssh_user,
          passwd: formData.value.passwd,
          snapshot_id: formData.value.snapshot_id,
          snapshot_instance_uuid: formData.value.snapshot_instance_uuid,
          description: formData.value.description,
          is_public: formData.value.is_public
        }

        if (editingResource.value) {
          await adminAPI.updateResourceInstance(editingResource.value.uuid, payload)
          dialog.alertSuccess('更新成功')
        } else {
          await adminAPI.createResourceInstance(payload)
          dialog.alertSuccess('创建成功')
        }
        closeModals()
        fetchResources()
      } catch (error) {
        console.error('Failed to save resource:', error)
        dialog.alertError('保存失败: ' + error.message)
      } finally {
        submitting.value = false
      }
    }

    const editResource = (resource) => {
      editingResource.value = resource
      formData.value = {
        name: resource.name || resource.ip_address,
        resource_type: normalizeResourceType(resource.instance_type || resource.resource_type),
        host: resource.ip_address || resource.host,
        ssh_port: resource.port || resource.ssh_port || 22,
        ssh_user: resource.ssh_user || 'root',
        passwd: resource.passwd || '',
        snapshot_id: resource.snapshot_id || '',
        snapshot_instance_uuid: resource.snapshot_instance_uuid || '',
        description: resource.description || '',
        is_public: resource.is_public !== undefined ? resource.is_public : false
      }
      showEditModal.value = true
    }

    const deleteResource = async (resource) => {
      const confirmed = await dialog.confirm(`确定要删除资源实例 "${resource.name}" 吗？`)
      if (!confirmed) return
      try {
        await adminAPI.deleteResourceInstance(resource.uuid)
        dialog.alertSuccess('删除成功')
        fetchResources()
      } catch (error) {
        dialog.alertError('删除失败: ' + error.message)
      }
    }

    const viewTestbeds = (resource) => {
      window.location.href = `/resource-pool/testbeds?resource=${resource.uuid}`
    }

    const testConnection = async (resource) => {
      if (resource.created_by !== currentUsername.value) {
        dialog.alertWarning('您只能测试自己创建的资源实例')
        return
      }
      testing.value = resource.uuid
      try {
        const result = await adminAPI.checkResourceHealth(resource.uuid)
        await fetchResources()
        if (result.success || result.healthy) {
          dialog.alertSuccess(`连接测试成功！\n\n状态: ${result.status || 'healthy'}\nIP: ${result.ip_address}\n端口: ${result.port}\n\n资源实例状态已更新为: 可用`)
        } else {
          dialog.alertWarning(`连接测试失败！\n\n${result.message || '无法连接到资源实例'}\n\n资源实例状态已更新为: unreachable`)
        }
      } catch (error) {
        console.error('Health check failed:', error)
        dialog.alertError(`连接测试失败！\n\n${error.message || '无法连接到资源实例'}`)
      } finally {
        testing.value = null
      }
    }

    const closeModals = () => {
      showCreateModal.value = false
      showEditModal.value = false
      editingResource.value = null
      resetForm()
    }

    const resetForm = () => {
      formData.value = {
        name: '',
        resource_type: 'virtual_machine',
        host: '',
        ssh_port: 22,
        ssh_user: 'root',
        passwd: '',
        snapshot_id: '',
        snapshot_instance_uuid: '',
        description: '',
        is_public: true
      }
    }

    const getStatusClass = (status) => {
      const classes = {
        'pending': 'status-pending',
        'active': 'status-active',
        'terminating': 'status-terminating',
        'terminated': 'status-terminated',
        'maintenance': 'status-maintenance'
      }
      return classes[status] || 'status-pending'
    }

    const getStatusLabel = (status) => {
      const labels = {
        'pending': '检查中',
        'active': '可用',
        'terminating': 'unreachable',
        'terminated': '已终止',
        'maintenance': '维护中'
      }
      return labels[status] || status
    }

    const getTypeLabel = (type) => {
      const labels = {
        'virtual_machine': '虚拟机',
        'physical_machine': '物理机',
        'VirtualMachine': '虚拟机',
        'Machine': '物理机',
        'container': '容器'
      }
      return labels[type] || type
    }

    const getTypeClass = (type) => {
      if (type === 'virtual_machine' || type === 'VirtualMachine') return 'type-vm'
      return 'type-pm'
    }

    const normalizeResourceType = (type) => {
      const mapping = {
        'VirtualMachine': 'virtual_machine',
        'Machine': 'physical_machine',
        'virtual_machine': 'virtual_machine',
        'physical_machine': 'physical_machine',
        'container': 'container'
      }
      return mapping[type] || 'virtual_machine'
    }

    const formatTime = (time) => {
      if (!time) return '-'
      return new Date(time).toLocaleString('zh-CN')
    }

    watch(
      () => formData.value.resource_type,
      (newType) => {
        if (newType === 'virtual_machine') {
          formData.value.is_public = true
        }
      }
    )

    onMounted(fetchResources)

    return {
      resources, loading, submitting, testing, testingConnection,
      showCreateModal, showEditModal, formData, currentUsername,
      saveResource, editResource, deleteResource, viewTestbeds,
      testConnection, testNewConnection, closeModals,
      getStatusClass, getStatusLabel, getTypeLabel, getTypeClass, formatTime
    }
  }
}
</script>

<style scoped>
.my-resource-instance-page {
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.5rem 0;
  margin-bottom: 1.5rem;
  border-bottom: 1px solid #E2E8F0;
}

.header-icon {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #6366F1 0%, #A5B4FC 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.25);
}

.header-icon svg {
  width: 24px;
  height: 24px;
  color: white;
}

.header-content {
  flex: 1;
}

.header-content h1 {
  font-size: 1.5rem;
  font-weight: 700;
  color: #0F172A;
  margin: 0 0 0.25rem;
}

.header-content p {
  font-size: 0.875rem;
  color: #64748B;
  margin: 0;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  border-radius: 8px;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
  line-height: 1;
  white-space: nowrap;
}

.btn-icon {
  width: 18px;
  height: 18px;
}

.btn-primary {
  background: linear-gradient(135deg, #7C3AED 0%, #6366F1 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(124, 58, 237, 0.25);
}

.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(124, 58, 237, 0.35);
}

.btn-secondary {
  background: #F1F5F9;
  color: #64748B;
  border: 1px solid #E2E8F0;
}

.btn-secondary:hover {
  background: #E2E8F0;
  color: #475569;
}

.btn-outline {
  background: white;
  color: #7C3AED;
  border: 1px solid #7C3AED;
}

.btn-outline:hover {
  background: #F5F3FF;
}

.content-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 4px 12px rgba(0, 0, 0, 0.05);
  border: 1px solid #E2E8F0;
  min-height: 400px;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 0;
  color: #64748B;
  gap: 1rem;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #E2E8F0;
  border-top-color: #7C3AED;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 0;
  text-align: center;
}

.empty-icon {
  width: 64px;
  height: 64px;
  background: linear-gradient(135deg, #F5F3FF 0%, #EDE9FE 100%);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
}

.empty-icon svg {
  width: 32px;
  height: 32px;
  color: #7C3AED;
}

.empty-state h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #1E293B;
  margin: 0 0 0.5rem;
}

.empty-state p {
  font-size: 0.875rem;
  color: #64748B;
  margin: 0 0 1.5rem;
}

.resources-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.25rem;
  padding: 1.25rem;
}

.resource-card {
  background: #FAFAFC;
  border: 1px solid #E2E8F0;
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.2s ease;
}

.resource-card:hover {
  border-color: #CBD5E1;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  background: white;
  border-bottom: 1px solid #E2E8F0;
}

.card-title-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.resource-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.resource-icon svg {
  width: 20px;
  height: 20px;
  color: white;
}

.type-vm {
  background: linear-gradient(135deg, #7C3AED 0%, #A78BFA 100%);
}

.type-pm {
  background: linear-gradient(135deg, #0EA5E9 0%, #7DD3FC 100%);
}

.card-title-info h3 {
  font-size: 0.9375rem;
  font-weight: 600;
  color: #1E293B;
  margin: 0;
}

.resource-type {
  font-size: 0.75rem;
  color: #64748B;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 500;
  border-radius: 6px;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.status-pending {
  background: #FEF3C7;
  color: #D97706;
}

.status-pending .status-dot {
  background: #F59E0B;
}

.status-active {
  background: #ECFDF5;
  color: #059669;
}

.status-active .status-dot {
  background: #10B981;
}

.status-terminating {
  background: #FEF2F2;
  color: #DC2626;
}

.status-terminating .status-dot {
  background: #EF4444;
}

.status-terminated {
  background: #F1F5F9;
  color: #64748B;
}

.status-terminated .status-dot {
  background: #94A3B8;
}

.status-maintenance {
  background: #EFF6FF;
  color: #2563EB;
}

.status-maintenance .status-dot {
  background: #3B82F6;
}

.card-body {
  padding: 1.25rem;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.75rem;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.info-label {
  font-size: 0.75rem;
  color: #64748B;
  font-weight: 500;
}

.info-value {
  font-size: 0.875rem;
  color: #1E293B;
}

.host-value {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.8125rem;
}

.visibility-badge {
  display: inline-block;
  padding: 0.125rem 0.5rem;
  font-size: 0.6875rem;
  font-weight: 500;
  border-radius: 4px;
}

.visibility-public {
  background: #ECFDF5;
  color: #059669;
}

.visibility-private {
  background: #F1F5F9;
  color: #64748B;
}

.snapshot-section {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px dashed #E2E8F0;
}

.snapshot-label {
  font-size: 0.75rem;
  color: #64748B;
  font-weight: 500;
  margin-bottom: 0.5rem;
}

.snapshot-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.snapshot-item {
  font-size: 0.8125rem;
  font-family: 'Monaco', 'Menlo', monospace;
}

.snapshot-key {
  color: #64748B;
  margin-right: 0.5rem;
}

.snapshot-value {
  color: #7C3AED;
}

.description-section {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px dashed #E2E8F0;
}

.description-label {
  font-size: 0.75rem;
  color: #64748B;
  font-weight: 500;
  margin-bottom: 0.375rem;
}

.description-text {
  font-size: 0.8125rem;
  color: #475569;
  margin: 0;
  line-height: 1.5;
}

.meta-info {
  display: flex;
  gap: 1rem;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px dashed #E2E8F0;
}

.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.75rem;
  color: #64748B;
}

.meta-item svg {
  width: 14px;
  height: 14px;
}

.card-footer {
  display: flex;
  gap: 0.5rem;
  padding: 1rem 1.25rem;
  background: white;
  border-top: 1px solid #E2E8F0;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  font-size: 0.75rem;
  font-weight: 500;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn svg {
  width: 14px;
  height: 14px;
}

.action-btn-edit {
  background: #F1F5F9;
  color: #475569;
}

.action-btn-edit:hover {
  background: #E2E8F0;
}

.action-btn-info {
  background: #EFF6FF;
  color: #2563EB;
}

.action-btn-info:hover {
  background: #DBEAFE;
}

.action-btn-test {
  background: linear-gradient(135deg, #7C3AED 0%, #6366F1 100%);
  color: white;
}

.action-btn-test:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(124, 58, 237, 0.3);
}

.action-btn-danger {
  background: #FEF2F2;
  color: #DC2626;
}

.action-btn-danger:hover {
  background: #FEE2E2;
}

.btn-spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  border-radius: 16px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-medium {
  width: 640px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  background: linear-gradient(135deg, #F5F3FF 0%, #EDE9FE 100%);
  border-bottom: 1px solid #E2E8F0;
}

.modal-header-content {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.modal-icon {
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, #7C3AED 0%, #6366F1 100%);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-icon svg {
  width: 18px;
  height: 18px;
  color: white;
}

.modal-header h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #1E293B;
  margin: 0;
}

.modal-close {
  width: 32px;
  height: 32px;
  border: none;
  background: white;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.modal-close svg {
  width: 18px;
  height: 18px;
  color: #64748B;
}

.modal-close:hover {
  background: #F1F5F9;
}

.modal-form {
  padding: 1.5rem;
  overflow-y: auto;
}

.form-section {
  margin-bottom: 1.5rem;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid #E2E8F0;
}

.section-header svg {
  width: 18px;
  height: 18px;
  color: #7C3AED;
}

.section-header h4 {
  font-size: 0.875rem;
  font-weight: 600;
  color: #1E293B;
  margin: 0;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-group label {
  display: block;
  font-size: 0.8125rem;
  font-weight: 500;
  color: #374151;
  margin-bottom: 0.375rem;
}

.required {
  color: #DC2626;
}

.form-control {
  width: 100%;
  padding: 0.625rem 0.875rem;
  font-size: 0.875rem;
  border: 1px solid #E2E8F0;
  border-radius: 8px;
  background: white;
  transition: all 0.2s ease;
  color: #1E293B;
}

.form-control:focus {
  outline: none;
  border-color: #7C3AED;
  box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.1);
}

.form-control::placeholder {
  color: #94A3B8;
}

textarea.form-control {
  resize: vertical;
  min-height: 60px;
}

.form-hint {
  display: block;
  font-size: 0.75rem;
  color: #64748B;
  margin-top: 0.25rem;
}

.form-hint-warning {
  color: #D97706;
}

.vm-section {
  background: #F8FAFC;
  border-radius: 8px;
  padding: 1rem;
  border: 1px solid #E2E8F0;
}

.permission-section {
  margin-bottom: 0;
}

.permission-card {
  background: #F8FAFC;
  border-radius: 8px;
  padding: 1rem;
  border: 1px solid #E2E8F0;
}

.toggle-label {
  display: flex;
  align-items: flex-start;
  gap: 0.875rem;
  cursor: pointer;
  width: 100%;
}

.toggle-input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: relative;
  width: 44px;
  height: 24px;
  background: #CBD5E1;
  border-radius: 12px;
  transition: all 0.2s ease;
  flex-shrink: 0;
  margin-top: 0.125rem;
}

.toggle-slider::after {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 20px;
  height: 20px;
  background: white;
  border-radius: 50%;
  transition: all 0.2s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
}

.toggle-input:checked + .toggle-slider {
  background: linear-gradient(135deg, #7C3AED 0%, #6366F1 100%);
}

.toggle-input:checked + .toggle-slider::after {
  transform: translateX(20px);
}

.toggle-input:disabled + .toggle-slider {
  opacity: 0.6;
  cursor: not-allowed;
}

.toggle-content {
  flex: 1;
}

.toggle-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.toggle-icon-public {
  width: 18px;
  height: 18px;
  color: #7C3AED;
}

.toggle-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #1E293B;
}

.toggle-description {
  font-size: 0.8125rem;
  color: #64748B;
  margin: 0.25rem 0 0 0;
  line-height: 1.4;
}

.permission-hint {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  margin-top: 0.75rem;
  padding: 0.625rem 0.875rem;
  background: linear-gradient(135deg, #FEF3C7 0%, #FDE68A 100%);
  border-radius: 6px;
  font-size: 0.8125rem;
  color: #92400E;
}

.permission-hint svg {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

.modal-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.5rem;
  background: #F8FAFC;
  border-top: 1px solid #E2E8F0;
}

.footer-left {
  display: flex;
  gap: 0.75rem;
}

.footer-left .btn {
  height: 38px;
  min-height: 38px;
}

.footer-right {
  display: flex;
  gap: 0.75rem;
}

.footer-right .btn {
  height: 38px;
  min-height: 38px;
}

.footer-right .btn svg {
  width: 16px;
  height: 16px;
}

@media (max-width: 1024px) {
  .resources-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }

  .form-row {
    grid-template-columns: 1fr;
  }

  .modal-medium {
    width: 95%;
    max-width: 95%;
  }
}
</style>
