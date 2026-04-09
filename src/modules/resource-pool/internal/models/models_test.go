package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestTestbedStatusTransition 测试 Testbed 状态转换通过 Mark 方法
func TestTestbedStatusTransition(t *testing.T) {
	t.Run("available -> allocated", func(t *testing.T) {
		testbed := &Testbed{Status: TestbedStatusAvailable}
		allocUUID := uuid.New().String()
		testbed.MarkAllocated(allocUUID)

		if testbed.Status != TestbedStatusAllocated {
			t.Errorf("expected status %s, got %s", TestbedStatusAllocated, testbed.Status)
		}
		if testbed.CurrentAllocUUID == nil {
			t.Errorf("CurrentAllocUUID should be set")
		}
	})

	t.Run("allocated -> in_use", func(t *testing.T) {
		testbed := &Testbed{Status: TestbedStatusAllocated}
		testbed.MarkInUse()

		if testbed.Status != TestbedStatusInUse {
			t.Errorf("expected status %s, got %s", TestbedStatusInUse, testbed.Status)
		}
	})

	t.Run("in_use -> releasing", func(t *testing.T) {
		testbed := &Testbed{Status: TestbedStatusInUse}
		testbed.MarkReleasing()

		if testbed.Status != TestbedStatusReleasing {
			t.Errorf("expected status %s, got %s", TestbedStatusReleasing, testbed.Status)
		}
	})

	t.Run("releasing -> available", func(t *testing.T) {
		allocUUID := uuid.New().String()
		testbed := &Testbed{Status: TestbedStatusReleasing, CurrentAllocUUID: &allocUUID}
		testbed.MarkAvailable()

		if testbed.Status != TestbedStatusAvailable {
			t.Errorf("expected status %s, got %s", TestbedStatusAvailable, testbed.Status)
		}
		if testbed.CurrentAllocUUID != nil {
			t.Errorf("CurrentAllocUUID should be cleared")
		}
	})

	t.Run("available -> deleted", func(t *testing.T) {
		testbed := &Testbed{Status: TestbedStatusAvailable}
		testbed.MarkDeleted()

		if testbed.Status != TestbedStatusDeleted {
			t.Errorf("expected status %s, got %s", TestbedStatusDeleted, testbed.Status)
		}
		// Note: deleted should NOT transition back to available
		// Testbed is one-time use, deleted is terminal state
	})
}

// TestTestbedIsAvailable 测试 IsAvailable 方法
func TestTestbedIsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		status   TestbedStatus
		expected bool
	}{
		{"available status", TestbedStatusAvailable, true},
		{"allocated status", TestbedStatusAllocated, false},
		{"in_use status", TestbedStatusInUse, false},
		{"releasing status", TestbedStatusReleasing, false},
		{"maintenance status", TestbedStatusDeleted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testbed := &Testbed{Status: tt.status}
			if got := testbed.IsAvailable(); got != tt.expected {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestTestbedIsAllocated 测试 IsAllocated 方法
func TestTestbedIsAllocated(t *testing.T) {
	tests := []struct {
		name     string
		status   TestbedStatus
		expected bool
	}{
		{"available status", TestbedStatusAvailable, false},
		{"allocated status", TestbedStatusAllocated, true},
		{"in_use status", TestbedStatusInUse, true},
		{"releasing status", TestbedStatusReleasing, false},
		{"maintenance status", TestbedStatusDeleted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testbed := &Testbed{Status: tt.status}
			if got := testbed.IsAllocated(); got != tt.expected {
				t.Errorf("IsAllocated() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestTestbedMarkMethods 测试 Mark 系列方法
func TestTestbedMarkMethods(t *testing.T) {
	testbed := &Testbed{Status: TestbedStatusAvailable}
	allocUUID := uuid.New().String()

	// Test MarkAllocated
	testbed.MarkAllocated(allocUUID)
	if testbed.Status != TestbedStatusAllocated {
		t.Errorf("MarkAllocated() failed, status = %s", testbed.Status)
	}
	if testbed.CurrentAllocUUID == nil || *testbed.CurrentAllocUUID != allocUUID {
		t.Errorf("MarkAllocated() failed to set CurrentAllocUUID")
	}

	// Test MarkInUse
	testbed.MarkInUse()
	if testbed.Status != TestbedStatusInUse {
		t.Errorf("MarkInUse() failed, status = %s", testbed.Status)
	}

	// Test MarkReleasing
	testbed.MarkReleasing()
	if testbed.Status != TestbedStatusReleasing {
		t.Errorf("MarkReleasing() failed, status = %s", testbed.Status)
	}

	// Test MarkAvailable
	testbed.MarkAvailable()
	if testbed.Status != TestbedStatusAvailable {
		t.Errorf("MarkAvailable() failed, status = %s", testbed.Status)
	}
	if testbed.CurrentAllocUUID != nil {
		t.Errorf("MarkAvailable() should clear CurrentAllocUUID")
	}

	// Test MarkDeleted
	testbed.MarkDeleted()
	if testbed.Status != TestbedStatusDeleted {
		t.Errorf("MarkDeleted() failed, status = %s", testbed.Status)
	}
	// Deleted testbed should have CurrentAllocUUID cleared
	if testbed.CurrentAllocUUID != nil {
		t.Errorf("MarkDeleted() should clear CurrentAllocUUID")
	}
}

// TestTestbedToResponse 测试 ToResponse 方法
func TestTestbedToResponse(t *testing.T) {
	now := time.Now()
	testbed := &Testbed{
		ID:                   1,
		UUID:                 uuid.New().String(),
		Name:                 "test-testbed",
		CategoryUUID:         uuid.New().String(),
		ResourceInstanceUUID: uuid.New().String(),
		CurrentAllocUUID:     nil,
		MariaDBPort:          3306,
		MariaDBUser:          "root",
		MariaDBPasswd:        "secret123",
		Status:               TestbedStatusAvailable,
		LastHealthCheck:      now,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	response := testbed.ToResponse()

	if response.UUID != testbed.UUID {
		t.Errorf("UUID mismatch: got %v, want %v", response.UUID, testbed.UUID)
	}
	if response.Name != testbed.Name {
		t.Errorf("Name mismatch: got %v, want %v", response.Name, testbed.Name)
	}
	if response.MariaDBPasswd != testbed.MariaDBPasswd {
		t.Errorf("MariaDBPasswd should be visible when no alloc: got %v", response.MariaDBPasswd)
	}
	if response.Status != testbed.Status {
		t.Errorf("Status mismatch: got %v, want %v", response.Status, testbed.Status)
	}
	if response.LastHealthCheck != now.Format(time.RFC3339) {
		t.Errorf("LastHealthCheck format mismatch")
	}
}

// TestTestbedToResponseWithMaskedPassword 测试密码掩码
func TestTestbedToResponseWithMaskedPassword(t *testing.T) {
	now := time.Now()
	testbed := &Testbed{
		ID:                   1,
		UUID:                 uuid.New().String(),
		Name:                 "test-testbed",
		CategoryUUID:         uuid.New().String(),
		ResourceInstanceUUID: uuid.New().String(),
		MariaDBPort:          3306,
		MariaDBUser:          "root",
		MariaDBPasswd:        "secret123",
		Status:               TestbedStatusInUse,
		LastHealthCheck:      now,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// Show password should show actual password
	response := testbed.ToResponseWithMaskedPassword(true)
	if response.MariaDBPasswd != "secret123" {
		t.Errorf("Should show actual password when showPassword=true, got %v", response.MariaDBPasswd)
	}

	// Mask password should show stars
	response = testbed.ToResponseWithMaskedPassword(false)
	if response.MariaDBPasswd != "****" {
		t.Errorf("Should show masked password when showPassword=false, got %v", response.MariaDBPasswd)
	}
}

// TestNewTestbed 测试 NewTestbed 构造函数
func TestNewTestbed(t *testing.T) {
	categoryUUID := uuid.New().String()
	instanceUUID := uuid.New().String()
	port := 3306
	user := "root"
	passwd := "secret"

	testbed := NewTestbed("test-bed", categoryUUID, ServiceTargetNormal, instanceUUID, port, user, passwd)

	if testbed.Name != "test-bed" {
		t.Errorf("Name mismatch: got %v", testbed.Name)
	}
	if testbed.CategoryUUID != categoryUUID {
		t.Errorf("CategoryUUID mismatch")
	}
	if testbed.ResourceInstanceUUID != instanceUUID {
		t.Errorf("ResourceInstanceUUID mismatch")
	}
	if testbed.MariaDBPort != port {
		t.Errorf("MariaDBPort mismatch")
	}
	if testbed.MariaDBUser != user {
		t.Errorf("MariaDBUser mismatch")
	}
	if testbed.MariaDBPasswd != passwd {
		t.Errorf("MariaDBPasswd mismatch")
	}
	if testbed.Status != TestbedStatusAvailable {
		t.Errorf("Initial status should be available, got %v", testbed.Status)
	}
	if testbed.UUID == "" {
		t.Errorf("UUID should be generated")
	}
}

// TestParseTestbedStatus 测试 ParseTestbedStatus
func TestParseTestbedStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected TestbedStatus
		hasError bool
	}{
		{"available", TestbedStatusAvailable, false},
		{"allocated", TestbedStatusAllocated, false},
		{"in_use", TestbedStatusInUse, false},
		{"releasing", TestbedStatusReleasing, false},
		{"deleted", TestbedStatusDeleted, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseTestbedStatus(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error for input %s", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("got %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

// TestResourceInstanceType 测试 InstanceType
func TestResourceInstanceType(t *testing.T) {
	vm := &ResourceInstance{InstanceType: InstanceTypeVirtualMachine}
	machine := &ResourceInstance{InstanceType: InstanceTypeMachine}

	if !vm.IsVirtualMachine() {
		t.Errorf("VirtualMachine should return true for IsVirtualMachine()")
	}
	if vm.IsMachine() {
		t.Errorf("VirtualMachine should return false for IsMachine()")
	}

	if !machine.IsMachine() {
		t.Errorf("Machine should return true for IsMachine()")
	}
	if machine.IsVirtualMachine() {
		t.Errorf("Machine should return false for IsVirtualMachine()")
	}
}

// TestNewVirtualMachine 测试 NewVirtualMachine
func TestNewVirtualMachine(t *testing.T) {
	ip := "192.168.1.100"
	port := 22
	passwd := "vm-pass"
	snapshotID := "snap-123"
	createdBy := "admin"

	vm := NewVirtualMachine(ip, port, passwd, snapshotID, createdBy)

	if vm.InstanceType != InstanceTypeVirtualMachine {
		t.Errorf("InstanceType should be VirtualMachine")
	}
	if vm.IPAddress != ip {
		t.Errorf("IPAddress mismatch")
	}
	if vm.SnapshotID == nil || *vm.SnapshotID != snapshotID {
		t.Errorf("SnapshotID mismatch")
	}
	if !vm.IsPublic {
		t.Errorf("VirtualMachine should always be public")
	}
	if vm.Status != ResourceInstanceStatusPending {
		t.Errorf("Initial status should be pending")
	}
}

// TestNewMachine 测试 NewMachine
func TestNewMachine(t *testing.T) {
	ip := "192.168.1.101"
	port := 22
	passwd := "machine-pass"
	createdBy := "admin"
	isPublic := true

	machine := NewMachine(ip, port, passwd, createdBy, isPublic)

	if machine.InstanceType != InstanceTypeMachine {
		t.Errorf("InstanceType should be Machine")
	}
	if machine.IPAddress != ip {
		t.Errorf("IPAddress mismatch")
	}
	if machine.SnapshotID != nil {
		t.Errorf("Machine should not have SnapshotID")
	}
	if machine.IsPublic != isPublic {
		t.Errorf("IsPublic mismatch")
	}
}

// TestResourceInstanceCanParticipateInPool 测试 CanParticipateInPool
func TestResourceInstanceCanParticipateInPool(t *testing.T) {
	vm := &ResourceInstance{
		InstanceType: InstanceTypeVirtualMachine,
		Status:       ResourceInstanceStatusActive,
	}
	if !vm.CanParticipateInPool() {
		t.Errorf("Active VirtualMachine should participate in pool")
	}

	machine := &ResourceInstance{
		InstanceType: InstanceTypeMachine,
		Status:       ResourceInstanceStatusActive,
	}
	if machine.CanParticipateInPool() {
		t.Errorf("Machine should not participate in pool")
	}

	inactiveVM := &ResourceInstance{
		InstanceType: InstanceTypeVirtualMachine,
		Status:       ResourceInstanceStatusUnreachable,
	}
	if inactiveVM.CanParticipateInPool() {
		t.Errorf("Terminated VirtualMachine should not participate in pool")
	}
}

// TestAllocationMarkMethods 测试 Allocation Mark 方法
func TestAllocationMarkMethods(t *testing.T) {
	alloc := &Allocation{Status: AllocationStatusPending}

	// Test MarkActive
	expiresAt := time.Now().Add(1 * time.Hour)
	alloc.MarkActive(&expiresAt)
	if alloc.Status != AllocationStatusActive {
		t.Errorf("MarkActive() failed, status = %s", alloc.Status)
	}
	if alloc.ExpiresAt == nil {
		t.Errorf("ExpiresAt should be set")
	}

	// Test MarkReleased
	alloc.MarkReleased()
	if alloc.Status != AllocationStatusReleased {
		t.Errorf("MarkReleased() failed, status = %s", alloc.Status)
	}
	if alloc.ReleasedAt == nil {
		t.Errorf("ReleasedAt should be set")
	}

	// Reset for MarkExpired test
	alloc.Status = AllocationStatusActive
	alloc.ReleasedAt = nil
	alloc.MarkExpired()
	if alloc.Status != AllocationStatusExpired {
		t.Errorf("MarkExpired() failed, status = %s", alloc.Status)
	}
	if alloc.ReleasedAt == nil {
		t.Errorf("ReleasedAt should be set after expired")
	}
}

// TestAllocationIsActive 测试 IsActive 方法
func TestAllocationIsActive(t *testing.T) {
	tests := []struct {
		status   AllocationStatus
		expected bool
	}{
		{AllocationStatusPending, false},
		{AllocationStatusActive, true},
		{AllocationStatusReleased, false},
		{AllocationStatusExpired, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			alloc := &Allocation{Status: tt.status}
			if got := alloc.IsActive(); got != tt.expected {
				t.Errorf("IsActive() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestAllocationIsExpired 测试 IsExpired 方法
func TestAllocationIsExpired(t *testing.T) {
	now := time.Now()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)

	tests := []struct {
		name      string
		expiresAt *time.Time
		expected  bool
	}{
		{"future expiry", &future, false},
		{"past expiry", &past, true},
		{"no expiry", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alloc := &Allocation{ExpiresAt: tt.expiresAt}
			result := alloc.IsExpired()
			if result != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestAllocationIsReleased 测试 IsReleased 方法
func TestAllocationIsReleased(t *testing.T) {
	tests := []struct {
		status   AllocationStatus
		expected bool
	}{
		{AllocationStatusPending, false},
		{AllocationStatusActive, false},
		{AllocationStatusReleased, true},
		{AllocationStatusExpired, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			alloc := &Allocation{Status: tt.status}
			if got := alloc.IsReleased(); got != tt.expected {
				t.Errorf("IsReleased() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestAllocationGetRemainingSeconds 测试 GetRemainingSeconds
func TestAllocationGetRemainingSeconds(t *testing.T) {
	now := time.Now()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)

	tests := []struct {
		name      string
		expiresAt *time.Time
		expected  int64
	}{
		{"future expiry", &future, 3600},
		{"past expiry", &past, 0},
		{"no expiry", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alloc := &Allocation{ExpiresAt: tt.expiresAt}
			result := alloc.GetRemainingSeconds()
			if tt.expected > 0 && (result < tt.expected-10 || result > tt.expected+10) {
				t.Errorf("GetRemainingSeconds() = %v, want ~%v", result, tt.expected)
			}
			if tt.expected == 0 && result > 100 {
				t.Errorf("GetRemainingSeconds() = %v, want ~0", result)
			}
		})
	}
}

// TestNewAllocation 测试 NewAllocation 构造函数
func TestNewAllocation(t *testing.T) {
	testbedUUID := uuid.New().String()
	categoryUUID := uuid.New().String()
	requester := "user1"
	maxLifetime := 3600

	alloc := NewAllocation(testbedUUID, categoryUUID, requester, maxLifetime)

	if alloc.TestbedUUID != testbedUUID {
		t.Errorf("TestbedUUID mismatch")
	}
	if alloc.CategoryUUID != categoryUUID {
		t.Errorf("CategoryUUID mismatch")
	}
	if alloc.Requester != requester {
		t.Errorf("Requester mismatch")
	}
	if alloc.Status != AllocationStatusPending {
		t.Errorf("Initial status should be pending")
	}
	if alloc.ExpiresAt == nil {
		t.Errorf("ExpiresAt should be set")
	}

	// Test with maxLifetimeSeconds = 0
	alloc2 := NewAllocation(testbedUUID, categoryUUID, requester, 0)
	if alloc2.ExpiresAt != nil {
		t.Errorf("ExpiresAt should be nil when maxLifetimeSeconds is 0")
	}
}

// TestCategoryEnableDisable 测试 Category 启用/禁用
func TestCategoryEnableDisable(t *testing.T) {
	category := NewCategory("test", "Test category")

	// Initially enabled
	if !category.Enabled {
		t.Errorf("New category should be enabled by default")
	}

	// Disable
	category.Disable()
	if category.Enabled {
		t.Errorf("Category should be disabled")
	}

	// Enable
	category.Enable()
	if !category.Enabled {
		t.Errorf("Category should be enabled")
	}
}

// TestNewCategory 测试 NewCategory 构造函数
func TestNewCategory(t *testing.T) {
	name := "test-category"
	description := "Test description"

	category := NewCategory(name, description)

	if category.Name != name {
		t.Errorf("Name mismatch")
	}
	if category.Description != description {
		t.Errorf("Description mismatch")
	}
	if !category.Enabled {
		t.Errorf("New category should be enabled")
	}
	if category.UUID == "" {
		t.Errorf("UUID should be generated")
	}
}

// TestQuotaPolicyMethods 测试 QuotaPolicy 方法
func TestQuotaPolicyMethods(t *testing.T) {
	policy := &QuotaPolicy{
		CategoryUUID:       uuid.New().String(),
		MinInstances:       2,
		MaxInstances:       10,
		AutoReplenish:      true,
		ReplenishThreshold: 5,
		Priority:           1,
	}

	// Test ShouldReplenish
	tests := []struct {
		name           string
		availableCount int
		expected       bool
	}{
		{"below threshold", 3, true},
		{"at threshold", 5, false},
		{"above threshold", 6, false},
		{"auto replenish off", 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "auto replenish off" {
				policy.AutoReplenish = false
			} else {
				policy.AutoReplenish = true
			}

			result := policy.ShouldReplenish(tt.availableCount)
			if result != tt.expected {
				t.Errorf("ShouldReplenish() = %v, want %v", result, tt.expected)
			}
		})
	}

	// Reset AutoReplenish for subsequent tests
	policy.AutoReplenish = true

	// Test IsOverQuota
	overQuota := policy.IsOverQuota(11) // 11 > MaxInstances(10)
	if !overQuota {
		t.Errorf("Should be over quota with 11 instances")
	}

	overQuota = policy.IsOverQuota(10) // 10 == MaxInstances
	if !overQuota {
		t.Errorf("Should be over quota at max instances")
	}

	overQuota = policy.IsOverQuota(9) // 9 < MaxInstances
	if overQuota {
		t.Errorf("Should not be over quota with 9 instances")
	}

	// Test CanAllocate
	canAllocate := policy.CanAllocate(8) // 8 < MaxInstances(10)
	if !canAllocate {
		t.Errorf("Should be able to allocate when count is 8")
	}

	canAllocate = policy.CanAllocate(10) // 10 == MaxInstances
	if canAllocate {
		t.Errorf("Should not be able to allocate when at max")
	}
}

// TestNewQuotaPolicy 测试 NewQuotaPolicy 构造函数
func TestNewQuotaPolicy(t *testing.T) {
	categoryUUID := uuid.New().String()
	minInstances := 2
	maxInstances := 10
	priority := 1
	maxLifetime := 7200

	policy := NewQuotaPolicy(categoryUUID, minInstances, maxInstances, priority, maxLifetime)

	if policy.CategoryUUID != categoryUUID {
		t.Errorf("CategoryUUID mismatch")
	}
	if policy.MinInstances != minInstances {
		t.Errorf("MinInstances mismatch")
	}
	if policy.MaxInstances != maxInstances {
		t.Errorf("MaxInstances mismatch")
	}
	if policy.Priority != priority {
		t.Errorf("Priority mismatch")
	}
	if policy.MaxLifetimeSeconds != maxLifetime {
		t.Errorf("MaxLifetimeSeconds mismatch")
	}
	if !policy.AutoReplenish {
		t.Errorf("AutoReplenish should be true by default")
	}
}

// TestSerialization 序列化测试
func TestSerialization(t *testing.T) {
	// Test TestbedResponse serialization
	testbedResp := TestbedResponse{
		UUID:            uuid.New().String(),
		Name:            "test-bed",
		CategoryUUID:    uuid.New().String(),
		Status:          TestbedStatusAvailable,
		LastHealthCheck: time.Now().Format(time.RFC3339),
		CreatedAt:       time.Now().Format(time.RFC3339),
		UpdatedAt:       time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(testbedResp)
	if err != nil {
		t.Errorf("Failed to marshal TestbedResponse: %v", err)
	}

	var decoded TestbedResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Errorf("Failed to unmarshal TestbedResponse: %v", err)
	}

	if decoded.UUID != testbedResp.UUID {
		t.Errorf("UUID mismatch after round-trip")
	}

	// Test AllocationResponse serialization
	allocResp := AllocationResponse{
		UUID:         uuid.New().String(),
		Requester:    "user1",
		CategoryUUID: uuid.New().String(),
		Status:       AllocationStatusActive,
		CreatedAt:    time.Now().Format(time.RFC3339),
		UpdatedAt:    time.Now().Format(time.RFC3339),
	}

	data, err = json.Marshal(allocResp)
	if err != nil {
		t.Errorf("Failed to marshal AllocationResponse: %v", err)
	}

	var decodedAlloc AllocationResponse
	err = json.Unmarshal(data, &decodedAlloc)
	if err != nil {
		t.Errorf("Failed to unmarshal AllocationResponse: %v", err)
	}
}

// TestResourceInstanceToResponse 测试 ResourceInstance ToResponse
func TestResourceInstanceToResponse(t *testing.T) {
	instance := &ResourceInstance{
		UUID:         uuid.New().String(),
		InstanceType: InstanceTypeVirtualMachine,
		IPAddress:    "192.168.1.1",
		Port:         22,
		Passwd:       "secret",
		Status:       ResourceInstanceStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	response := instance.ToResponse()
	if response.UUID != instance.UUID {
		t.Errorf("UUID mismatch")
	}
	if response.InstanceType != instance.InstanceType {
		t.Errorf("InstanceType mismatch")
	}
	if response.Passwd != instance.Passwd {
		t.Errorf("Password should be visible in ToResponse")
	}
}

// TestResourceInstanceToResponseWithMaskedPassword 测试密码掩码
func TestResourceInstanceToResponseWithMaskedPassword(t *testing.T) {
	instance := &ResourceInstance{
		UUID:         uuid.New().String(),
		InstanceType: InstanceTypeVirtualMachine,
		IPAddress:    "192.168.1.1",
		Port:         22,
		Passwd:       "secret",
		Status:       ResourceInstanceStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Show password
	response := instance.ToResponseWithMaskedPassword(true)
	if response.Passwd != "secret" {
		t.Errorf("Password should be visible when showPassword=true")
	}

	// Mask password
	response = instance.ToResponseWithMaskedPassword(false)
	if response.Passwd != "****" {
		t.Errorf("Password should be masked when showPassword=false")
	}
}

// TestAllocationToResponse 测试 Allocation ToResponse
func TestAllocationToResponse(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)

	alloc := &Allocation{
		UUID:         uuid.New().String(),
		TestbedUUID:  uuid.New().String(),
		CategoryUUID: uuid.New().String(),
		Requester:    "user1",
		Status:       AllocationStatusActive,
		ExpiresAt:    &expiresAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	response := alloc.ToResponse()
	if response.UUID != alloc.UUID {
		t.Errorf("UUID mismatch")
	}
	if response.Status != alloc.Status {
		t.Errorf("Status mismatch")
	}
	if response.ExpiresAt == nil {
		t.Errorf("ExpiresAt should be set")
	}
	if response.RemainingSeconds == nil {
		t.Errorf("RemainingSeconds should be set for active allocation")
	}
}

// TestAllocationToResponseForReleased 测试已释放分配的响应
func TestAllocationToResponseForReleased(t *testing.T) {
	now := time.Now()
	releasedAt := now

	alloc := &Allocation{
		UUID:         uuid.New().String(),
		TestbedUUID:  uuid.New().String(),
		CategoryUUID: uuid.New().String(),
		Requester:    "user1",
		Status:       AllocationStatusReleased,
		ReleasedAt:   &releasedAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	response := alloc.ToResponse()
	if response.Status != alloc.Status {
		t.Errorf("Status mismatch")
	}
	if response.ReleasedAt == nil {
		t.Errorf("ReleasedAt should be set")
	}
	if response.RemainingSeconds != nil {
		t.Errorf("RemainingSeconds should be nil for released allocation")
	}
}

// TestParseInstanceType 测试 ParseInstanceType
func TestParseInstanceType(t *testing.T) {
	tests := []struct {
		input    string
		expected InstanceType
		hasError bool
	}{
		{"VirtualMachine", InstanceTypeVirtualMachine, false},
		{"Machine", InstanceTypeMachine, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseInstanceType(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error for input %s", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("got %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

// TestParseAllocationStatus 测试 ParseAllocationStatus
func TestParseAllocationStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected AllocationStatus
		hasError bool
	}{
		{"pending", AllocationStatusPending, false},
		{"active", AllocationStatusActive, false},
		{"released", AllocationStatusReleased, false},
		{"expired", AllocationStatusExpired, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseAllocationStatus(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error for input %s", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("got %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

// TestValidateVirtualMachine 测试 ValidateVirtualMachine
func TestValidateVirtualMachine(t *testing.T) {
	snapshotID := "snap-123"

	validVM := &ResourceInstance{
		InstanceType: InstanceTypeVirtualMachine,
		SnapshotID:   &snapshotID,
		IsPublic:     true,
	}

	err := validVM.ValidateVirtualMachine()
	if err != nil {
		t.Errorf("Valid VM should pass validation: %v", err)
	}

	// Missing SnapshotID
	invalidVM1 := &ResourceInstance{
		InstanceType: InstanceTypeVirtualMachine,
		SnapshotID:   nil,
		IsPublic:     true,
	}

	err = invalidVM1.ValidateVirtualMachine()
	if err == nil {
		t.Errorf("VM without SnapshotID should fail validation")
	}

	// Not public
	invalidVM2 := &ResourceInstance{
		InstanceType: InstanceTypeVirtualMachine,
		SnapshotID:   &snapshotID,
		IsPublic:     false,
	}

	err = invalidVM2.ValidateVirtualMachine()
	if err == nil {
		t.Errorf("VM that is not public should fail validation")
	}

	// Wrong type
	wrongType := &ResourceInstance{
		InstanceType: InstanceTypeMachine,
	}

	err = wrongType.ValidateVirtualMachine()
	if err == nil {
		t.Errorf("Machine should fail VM validation")
	}
}

// TestValidateMachine 测试 ValidateMachine
func TestValidateMachine(t *testing.T) {
	validMachine := &ResourceInstance{
		InstanceType: InstanceTypeMachine,
		SnapshotID:   nil,
	}

	err := validMachine.ValidateMachine()
	if err != nil {
		t.Errorf("Valid machine should pass validation: %v", err)
	}

	// Has SnapshotID
	invalidMachine := &ResourceInstance{
		InstanceType: InstanceTypeMachine,
		SnapshotID:   ptr("snap-123"),
	}

	err = invalidMachine.ValidateMachine()
	if err == nil {
		t.Errorf("Machine with SnapshotID should fail validation")
	}

	// Wrong type
	wrongType := &ResourceInstance{
		InstanceType: InstanceTypeVirtualMachine,
	}

	err = wrongType.ValidateMachine()
	if err == nil {
		t.Errorf("VirtualMachine should fail machine validation")
	}
}

// Helper function
func ptr(s string) *string {
	return &s
}
