package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/quality-gateway/shared/pkg/api"
	"github.com/quality-gateway/shared/pkg/logger"
)

func TestClientErrorHandling(t *testing.T) {
	t.Run("GetData handles error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.Response{
				Success: false,
				Error: &api.ErrorInfo{
					Code:    "INVALID_INPUT",
					Message: "invalid data",
				},
			})
		}))
		defer server.Close()

		client := api.NewClient(&api.ClientConfig{BaseURL: server.URL})

		var result map[string]string
		err := client.GetData("/test", &result)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "invalid data") {
			t.Errorf("expected error to contain 'invalid data', got %s", err.Error())
		}
	})

	t.Run("GetData handles non-JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("plain text response"))
		}))
		defer server.Close()

		client := api.NewClient(&api.ClientConfig{BaseURL: server.URL})

		// For non-JSON response, use raw Get and check response
		resp, err := client.Get("/test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Response data should contain the raw text
		if resp.Data == nil {
			t.Error("expected data to contain raw response")
		}
	})

	t.Run("PostData handles error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(api.Response{
				Success: false,
				Error: &api.ErrorInfo{
					Code:    "CONFLICT",
					Message: "resource exists",
				},
			})
		}))
		defer server.Close()

		client := api.NewClient(&api.ClientConfig{BaseURL: server.URL})

		var result map[string]string
		err := client.PostData("/test", map[string]string{"name": "test"}, &result)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "resource exists") {
			t.Errorf("expected error to contain 'resource exists', got %s", err.Error())
		}
	})

	t.Run("Request handles network error", func(t *testing.T) {
		client := api.NewClient(&api.ClientConfig{
			BaseURL:    "http://localhost:9999", // Non-existent server
			Timeout:    100 * time.Millisecond,
			MaxRetries: 1,
		})

		_, err := client.Get("/test")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Handles JSON unmarshal error in response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		client := api.NewClient(&api.ClientConfig{BaseURL: server.URL})

		resp, err := client.Get("/test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should return raw body as string when JSON parsing fails
		if resp.Data == nil {
			t.Error("expected data to contain raw response")
		}
	})
}

func TestClientContextMethods(t *testing.T) {
	t.Run("GetWithContext cancels request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := api.NewClient(&api.ClientConfig{BaseURL: server.URL})

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_, err := client.GetWithContext(ctx, "/test")
		if err == nil {
			t.Error("expected timeout error, got nil")
		}
	})

	t.Run("PostWithContext works with valid context", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.Response{Success: true})
		}))
		defer server.Close()

		client := api.NewClient(&api.ClientConfig{BaseURL: server.URL})

		ctx := context.Background()
		resp, err := client.PostWithContext(ctx, "/test", map[string]string{"key": "value"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !resp.Success {
			t.Error("expected success to be true")
		}
	})

	t.Run("PutWithContext works", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.Response{Success: true})
		}))
		defer server.Close()

		client := api.NewClient(&api.ClientConfig{BaseURL: server.URL})

		ctx := context.Background()
		resp, err := client.PutWithContext(ctx, "/test", map[string]string{"key": "value"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !resp.Success {
			t.Error("expected success to be true")
		}
	})

	t.Run("DeleteWithContext works", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.Response{Success: true})
		}))
		defer server.Close()

		client := api.NewClient(&api.ClientConfig{BaseURL: server.URL})

		ctx := context.Background()
		resp, err := client.DeleteWithContext(ctx, "/test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !resp.Success {
			t.Error("expected success to be true")
		}
	})

	t.Run("PatchWithContext works", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(api.Response{Success: true})
		}))
		defer server.Close()

		client := api.NewClient(&api.ClientConfig{BaseURL: server.URL})

		ctx := context.Background()
		resp, err := client.PatchWithContext(ctx, "/test", map[string]string{"key": "value"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !resp.Success {
			t.Error("expected success to be true")
		}
	})
}

func TestServerWithMiddleware(t *testing.T) {
	log := logger.New(logger.Config{Level: logger.InfoLevel})

	t.Run("Server with middleware chain", func(t *testing.T) {
		cfg := api.DefaultConfig()
		srv := api.New(cfg, log)

		// Register middleware
		srv.RegisterMiddleware(
			api.RecoveryMiddleware(log),
			api.CORSMiddleware([]string{"*"}),
		)

		// Register handler
		srv.GET("/test", func(w http.ResponseWriter, r *http.Request) {
			api.OK(w, map[string]string{"status": "ok"})
		})

		// Test request with Origin header (required for CORS)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()
		srv.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Check CORS headers - when "*" is allowed, it returns the origin
		cors := w.Header().Get("Access-Control-Allow-Origin")
		if cors != "http://example.com" {
			t.Errorf("expected CORS header http://example.com, got %s", cors)
		}
	})

	t.Run("Server with auth middleware", func(t *testing.T) {
		cfg := api.DefaultConfig()
		srv := api.New(cfg, log)

		srv.RegisterMiddleware(api.AuthMiddleware("secret-token"))
		srv.GET("/protected", func(w http.ResponseWriter, r *http.Request) {
			api.OK(w, map[string]string{"message": "access granted"})
		})

		// Test without token
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()
		srv.Router().ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}

		// Test with valid token
		req = httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer secret-token")
		w = httptest.NewRecorder()
		srv.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestResponseDecodeJSON(t *testing.T) {
	t.Run("DecodeJSON successfully decodes", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		jsonData, _ := json.Marshal(data)

		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(string(jsonData)))
		req.Header.Set("Content-Type", "application/json")

		var result map[string]string
		err := api.DecodeJSON(req, &result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result["key"] != "value" {
			t.Errorf("expected key=value, got key=%s", result["key"])
		}
	})

	t.Run("DecodeJSON rejects unknown fields", func(t *testing.T) {
		type TestStruct struct {
			Key string `json:"key"`
		}
		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"key":"value","unknown":"field"}`))
		req.Header.Set("Content-Type", "application/json")

		var result TestStruct
		err := api.DecodeJSON(req, &result)
		if err == nil {
			t.Error("expected error for unknown field, got nil")
		}
	})

	t.Run("DecodeJSON handles invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`invalid json`))
		req.Header.Set("Content-Type", "application/json")

		var result map[string]string
		err := api.DecodeJSON(req, &result)
		if err == nil {
			t.Error("expected error for invalid JSON, got nil")
		}
	})
}

func TestHandleErrorWithAppError(t *testing.T) {
	t.Run("HandleError with AppError", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/test", nil)

		// Create an AppError-like error
		err := errors.New("NOT_FOUND: resource not found")

		api.HandleError(w, r, err)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}

		var resp api.Response
		json.NewDecoder(w.Body).Decode(&resp)

		if resp.Success {
			t.Error("expected success to be false")
		}
	})
}
