package storage_test

import (
	"testing"

	"github.com/quality-gateway/shared/pkg/storage"
)

// MockDB for testing
type mockDB struct {
	storage.DB
}

func TestBuildInsertQuery(t *testing.T) {
	tests := []struct {
		name     string
		table    string
		columns  []string
		expected string
	}{
		{
			name:     "single column",
			table:    "users",
			columns:  []string{"name"},
			expected: "INSERT INTO users (name) VALUES (?)",
		},
		{
			name:     "multiple columns",
			table:    "users",
			columns:  []string{"name", "email", "age"},
			expected: "INSERT INTO users (name, email, age) VALUES (?, ?, ?)",
		},
		{
			name:     "empty columns",
			table:    "users",
			columns:  []string{},
			expected: "INSERT INTO users () VALUES ()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.BuildInsertQuery(tt.table, tt.columns)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestBuildUpdateQuery(t *testing.T) {
	tests := []struct {
		name     string
		table    string
		columns  []string
		where    string
		expected string
	}{
		{
			name:     "single column with where",
			table:    "users",
			columns:  []string{"name"},
			where:    "id = ?",
			expected: "UPDATE users SET name = ? WHERE id = ?",
		},
		{
			name:     "multiple columns with where",
			table:    "users",
			columns:  []string{"name", "email"},
			where:    "id = ?",
			expected: "UPDATE users SET name = ?, email = ? WHERE id = ?",
		},
		{
			name:     "multiple columns without where",
			table:    "users",
			columns:  []string{"name", "email"},
			where:    "",
			expected: "UPDATE users SET name = ?, email = ?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.BuildUpdateQuery(tt.table, tt.columns, tt.where)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRowScanner(t *testing.T) {
	t.Run("NewRowScanner creates scanner", func(t *testing.T) {
		scan := func(row storage.Scanner) (int, error) {
			return 42, nil
		}
		scanner := storage.NewRowScanner(scan)
		if scanner == nil {
			t.Error("expected non-nil scanner")
		}
	})

	t.Run("Scan returns value", func(t *testing.T) {
		scan := func(row storage.Scanner) (string, error) {
			return "test", nil
		}
		scanner := storage.NewRowScanner(scan)

		// Mock scanner
		mockScanner := &mockScanner{}
		result, err := scanner.Scan(mockScanner)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "test" {
			t.Errorf("expected 'test', got %s", result)
		}
	})
}

type mockScanner struct{}

func (m *mockScanner) Scan(dest ...interface{}) error {
	// Simple mock that sets first dest to 42 if it's an *int
	if len(dest) > 0 {
		if ptr, ok := dest[0].(*int); ok {
			*ptr = 42
		}
	}
	return nil
}

// Example repository usage test
type TestEntity struct {
	ID   int64
	Name string
}

func ExampleRepository() {
	// This is just a compile-time example to ensure types work
	type any = interface{}
	var _ any = storage.Repository[TestEntity]{}
}

func TestSliceScanner(t *testing.T) {
	// This test verifies the type signature works correctly
	scan := func(row storage.Scanner) (int, error) {
		return 42, nil
	}
	scanner := storage.NewRowScanner(scan)

	// We can't easily test with real sql.Rows without a database,
	// but we can verify the function signature compiles correctly
	_ = scanner
}
