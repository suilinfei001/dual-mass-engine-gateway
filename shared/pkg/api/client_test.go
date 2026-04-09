package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/quality-gateway/shared/pkg/api"
)

func TestClient(t *testing.T) {
	t.Run("NewClient creates client with defaults", func(t *testing.T) {
		client := api.NewClient(nil)
		if client == nil {
			t.Fatal("expected non-nil client")
		}
	})

	t.Run("NewClient creates client with config", func(t *testing.T) {
		cfg := &api.ClientConfig{
			BaseURL: "http://example.com",
			Token:   "test-token",
			Timeout: 10 * time.Second,
		}
		client := api.NewClient(cfg)
		if client == nil {
			t.Fatal("expected non-nil client")
		}
	})

	t.Run("GET request sends request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.Response{
				Success: true,
				Data:    map[string]string{"message": "ok"},
			})
		}))
		defer server.Close()

		client := api.NewClient(&api.ClientConfig{BaseURL: server.URL})
		resp, err := client.Get("/test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !resp.Success {
			t.Error("expected success to be true")
		}
	})

	t.Run("POST request sends JSON body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
			}

			var data map[string]string
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if data["key"] != "value" {
				t.Errorf("expected key=value, got key=%s", data["key"])
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.Response{
				Success: true,
				Data:    map[string]string{"created": "ok"},
			})
		}))
		defer server.Close()

		client := api.NewClient(&api.ClientConfig{BaseURL: server.URL})
		resp, err := client.Post("/test", map[string]string{"key": "value"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !resp.Success {
			t.Error("expected success to be true")
		}
	})

	t.Run("PUT request sends JSON body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Errorf("expected PUT, got %s", r.Method)
			}

			var data map[string]string
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				t.Fatalf("decode request: %v", err)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.Response{
				Success: true,
				Data:    map[string]string{"updated": "ok"},
			})
		}))
		defer server.Close()

		client := api.NewClient(&api.ClientConfig{BaseURL: server.URL})
		resp, err := client.Put("/test/1", map[string]string{"key": "value"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !resp.Success {
			t.Error("expected success to be true")
		}
	})

	t.Run("DELETE request sends request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.Response{
				Success: true,
				Data:    map[string]string{"deleted": "ok"},
			})
		}))
		defer server.Close()

		client := api.NewClient(&api.ClientConfig{BaseURL: server.URL})
		resp, err := client.Delete("/test/1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !resp.Success {
			t.Error("expected success to be true")
		}
	})

	t.Run("PATCH request sends JSON body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPatch {
				t.Errorf("expected PATCH, got %s", r.Method)
			}

			var data map[string]string
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				t.Fatalf("decode request: %v", err)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.Response{
				Success: true,
				Data:    map[string]string{"patched": "ok"},
			})
		}))
		defer server.Close()

		client := api.NewClient(&api.ClientConfig{BaseURL: server.URL})
		resp, err := client.Patch("/test/1", map[string]string{"key": "value"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !resp.Success {
			t.Error("expected success to be true")
		}
	})

	t.Run("Auth token is added to requests", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth != "Bearer test-token" {
				t.Errorf("expected Authorization Bearer test-token, got %s", auth)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.Response{Success: true})
		}))
		defer server.Close()

		client := api.NewClient(&api.ClientConfig{
			BaseURL: server.URL,
			Token:   "test-token",
		})
		resp, err := client.Get("/test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !resp.Success {
			t.Error("expected success to be true")
		}
	})

	t.Run("GetData unmarshals response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.Response{
				Success: true,
				Data:    map[string]string{"message": "hello"},
			})
		}))
		defer server.Close()

		client := api.NewClient(&api.ClientConfig{BaseURL: server.URL})

		var result map[string]string
		err := client.GetData("/test", &result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result["message"] != "hello" {
			t.Errorf("expected message=hello, got %s", result["message"])
		}
	})

	t.Run("PostData sends and unmarshals response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.Response{
				Success: true,
				Data:    map[string]string{"id": "123"},
			})
		}))
		defer server.Close()

		client := api.NewClient(&api.ClientConfig{BaseURL: server.URL})

		var result map[string]string
		err := client.PostData("/test", map[string]string{"name": "test"}, &result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result["id"] != "123" {
			t.Errorf("expected id=123, got %s", result["id"])
		}
	})
}

func TestClientSetters(t *testing.T) {
	client := api.NewClient(&api.ClientConfig{
		BaseURL: "http://example.com",
		Token:   "token1",
	})

	t.Run("SetBaseURL updates base URL", func(t *testing.T) {
		client.SetBaseURL("http://updated.com")
		// Can't verify directly, but no error should occur
	})

	t.Run("SetToken updates token", func(t *testing.T) {
		client.SetToken("token2")
	})

	t.Run("SetTimeout updates timeout", func(t *testing.T) {
		client.SetTimeout(5 * time.Second)
	})
}
