package deployer

import (
	"context"
	"testing"
	"time"
)

// TestMockDeployService_DeployProduct tests mock deployment
func TestMockDeployService_DeployProduct(t *testing.T) {
	deployer := NewMockDeployService()

	ctx := context.Background()
	req := DeployRequest{
		ResourceInstanceUUID: "test-instance-1",
		IPAddress:            "192.168.1.100",
		Port:                 22,
		Passwd:               "password",
		ProductVersion:       "v1.0.0",
		ConfigFile:           "{}",
		EnvVars:              map[string]string{"ENV": "test"},
		Timeout:              100 * time.Millisecond,
	}

	result, err := deployer.DeployProduct(ctx, req)
	if err != nil {
		t.Fatalf("DeployProduct failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected deployment to succeed")
	}

	if result.MariaDBPort == 0 {
		t.Error("Expected MariaDB port to be set")
	}

	if result.MariaDBUser == "" {
		t.Error("Expected MariaDB user to be set")
	}

	if result.MariaDBPasswd == "" {
		t.Error("Expected MariaDB password to be set")
	}

	// Check stats
	stats := deployer.GetStats()
	if stats["deploy_count"] == nil {
		t.Error("Expected deploy_count in stats")
	}
}

// TestMockDeployService_RestoreSnapshot tests mock snapshot restore
func TestMockDeployService_RestoreSnapshot(t *testing.T) {
	deployer := NewMockDeployService()
	deployer.SetDelays(50*time.Millisecond, 50*time.Millisecond)

	ctx := context.Background()

	err := deployer.RestoreSnapshot(ctx, "test-instance", "test-snapshot")
	if err != nil {
		t.Fatalf("RestoreSnapshot failed: %v", err)
	}

	// Check stats
	stats := deployer.GetStats()
	if stats["restore_count"] == nil {
		t.Error("Expected restore_count in stats")
	}
}

// TestMockDeployService_CheckHealth tests mock health check
func TestMockDeployService_CheckHealth(t *testing.T) {
	deployer := NewMockDeployService()

	ctx := context.Background()
	healthy, err := deployer.CheckHealth(ctx, "192.168.1.100", 3306, "root", "password")
	if err != nil {
		t.Fatalf("CheckHealth failed: %v", err)
	}

	if !healthy {
		t.Error("Expected health check to return true")
	}
}

// TestMockDeployService_FailureRate tests mock failure simulation
func TestMockDeployService_FailureRate(t *testing.T) {
	deployer := NewMockDeployService()
	deployer.SetDelays(10*time.Millisecond, 10*time.Millisecond)

	// Set 100% failure rate
	deployer.SetFailureRate(1.0)

	ctx := context.Background()
	req := DeployRequest{
		ResourceInstanceUUID: "test-instance",
		IPAddress:            "192.168.1.100",
		Port:                 22,
		Passwd:               "password",
		ProductVersion:       "v1.0.0",
		Timeout:              50 * time.Millisecond,
	}

	result, err := deployer.DeployProduct(ctx, req)
	if err != nil {
		t.Fatalf("DeployProduct failed: %v", err)
	}

	if result.Success {
		t.Error("Expected deployment to fail with 100% failure rate")
	}
	if result.ErrorMessage == "" {
		t.Error("Expected error message in failed result")
	}
}

// TestMockDeployService_ConcurrentDeployments tests concurrent deployment safety
func TestMockDeployService_ConcurrentDeployments(t *testing.T) {
	deployer := NewMockDeployService()
	deployer.SetDelays(50*time.Millisecond, 50*time.Millisecond)

	ctx := context.Background()

	// Launch multiple concurrent deployments
	concurrency := 5
	results := make(chan *DeployResult, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			req := DeployRequest{
				ResourceInstanceUUID: "test-instance",
				IPAddress:            "192.168.1.100",
				Port:                 22,
				Passwd:               "password",
				ProductVersion:       "v1.0.0",
				Timeout:              200 * time.Millisecond,
			}
			result, err := deployer.DeployProduct(ctx, req)
			if err != nil {
				t.Errorf("Deployment %d failed: %v", index, err)
			}
			results <- result
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < concurrency; i++ {
		result := <-results
		if result.Success {
			successCount++
		}
	}

	if successCount != concurrency {
		t.Errorf("Expected all deployments to succeed, got %d/%d", successCount, concurrency)
	}

	// Check stats
	stats := deployer.GetStats()
	deployCount, ok := stats["deploy_count"].(int)
	if !ok || deployCount != concurrency {
		t.Errorf("Expected deploy_count to be %d, got %v", concurrency, deployCount)
	}
}

// TestDeployRequest_LogConfig tests request logging
func TestDeployRequest_LogConfig(t *testing.T) {
	req := DeployRequest{
		ResourceInstanceUUID: "test-instance",
		IPAddress:            "192.168.1.100",
		Port:                 22,
		Passwd:               "secret",
		ProductVersion:       "v1.0.0",
		EnvVars: map[string]string{
			"ENV1": "value1",
			"ENV2": "value2",
		},
	}

	config := LogConfig(req)
	if config == "" {
		t.Error("Expected config string to be non-empty")
	}

	// Check that sensitive fields are included (as this is for debugging)
	if !contains(config, "secret") {
		t.Error("Expected config to contain password")
	}
	if !contains(config, "test-instance") {
		t.Error("Expected config to contain instance UUID")
	}
}

// TestDeployerFactory_CreateDeployer tests deployer factory
func TestDeployerFactory_CreateDeployer(t *testing.T) {
	// Test mock deployer creation
	mockFactory := NewDeployerFactory("mock")
	deployer, err := mockFactory.CreateDeployer(nil)
	if err != nil {
		t.Fatalf("Failed to create mock deployer: %v", err)
	}

	if _, ok := deployer.(*MockDeployService); !ok {
		t.Error("Expected MockDeployService type")
	}

	// Test tencent deployer creation
	tencentFactory := NewDeployerFactory("tencent")
	_, err = tencentFactory.CreateDeployer(map[string]string{
		"tencent_api_key":    "test-key",
		"tencent_secret_key": "test-secret",
		"tencent_region":     "ap-shanghai",
	})
	if err != nil {
		t.Fatalf("Failed to create tencent deployer: %v", err)
	}

	// Test missing API key
	_, err = tencentFactory.CreateDeployer(map[string]string{
		"tencent_secret_key": "test-secret",
	})
	if err == nil {
		t.Error("Expected error for missing API key")
	}

	// Test invalid deployer type
	invalidFactory := NewDeployerFactory("invalid")
	_, err = invalidFactory.CreateDeployer(nil)
	if err == nil {
		t.Error("Expected error for invalid deployer type")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
