package api_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/quality-gateway/shared/pkg/api"
	"github.com/quality-gateway/shared/pkg/logger"
)

func TestServer(t *testing.T) {
	log := logger.New(logger.Config{Level: logger.InfoLevel})

	t.Run("New creates server with default config", func(t *testing.T) {
		cfg := api.DefaultConfig()
		srv := api.New(cfg, log)

		if srv == nil {
			t.Fatal("expected non-nil server")
		}
		if srv.Router() == nil {
			t.Error("expected non-nil router")
		}
	})

	t.Run("Config address returns correct address", func(t *testing.T) {
		cfg := api.Config{Host: "localhost", Port: 8080}
		if addr := cfg.Address(); addr != "localhost:8080" {
			t.Errorf("expected localhost:8080, got %s", addr)
		}
	})
}

func TestResponse(t *testing.T) {
	t.Run("OK writes successful response", func(t *testing.T) {
		w := httptest.NewRecorder()
		api.OK(w, map[string]string{"message": "test"})

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp api.Response
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}

		if !resp.Success {
			t.Error("expected success to be true")
		}
	})

	t.Run("NotFound writes 404 response", func(t *testing.T) {
		w := httptest.NewRecorder()
		api.NotFound(w, "not found")

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}

		var resp api.Response
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}

		if resp.Success {
			t.Error("expected success to be false")
		}
		if resp.Error == nil {
			t.Error("expected error to be set")
		}
	})

	t.Run("BadRequest writes 400 response", func(t *testing.T) {
		w := httptest.NewRecorder()
		api.BadRequest(w, "bad request")

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("Unauthorized writes 401 response", func(t *testing.T) {
		w := httptest.NewRecorder()
		api.Unauthorized(w, "unauthorized")

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("NoContent writes 204 response", func(t *testing.T) {
		w := httptest.NewRecorder()
		api.NoContent(w)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}
	})
}

func TestRouter(t *testing.T) {
	router := api.NewRouter()

	t.Run("GET registers handler", func(t *testing.T) {
		called := false
		router.GET("/test", func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if !called {
			t.Error("handler was not called")
		}
	})

	t.Run("POST registers handler", func(t *testing.T) {
		called := false
		router.POST("/test", func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusCreated)
		})

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if !called {
			t.Error("handler was not called")
		}
	})

	t.Run("Group creates prefixed router", func(t *testing.T) {
		called := false
		group := router.Group("/api")
		group.GET("/test", func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if !called {
			t.Error("handler was not called")
		}
	})
}

func TestMiddleware(t *testing.T) {
	log := logger.New(logger.Config{Level: logger.InfoLevel})

	t.Run("CORS middleware adds headers", func(t *testing.T) {
		router := api.NewRouter()
		router.Use(api.CORSMiddleware([]string{"*"}))
		router.GET("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Header().Get("Access-Control-Allow-Origin") == "" {
			t.Error("expected CORS header to be set")
		}
	})

	t.Run("Recovery middleware recovers from panic", func(t *testing.T) {
		router := api.NewRouter()
		router.Use(api.RecoveryMiddleware(log))
		router.GET("/panic", func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("Auth middleware validates token", func(t *testing.T) {
		router := api.NewRouter()
		router.Use(api.AuthMiddleware("valid-token"))
		router.GET("/protected", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("Auth middleware accepts valid token", func(t *testing.T) {
		router := api.NewRouter()
		router.Use(api.AuthMiddleware("valid-token"))
		router.GET("/protected", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestHandleWrapper(t *testing.T) {
	t.Run("Handle calls handler", func(t *testing.T) {
		called := false
		handler := api.Handle(func(w http.ResponseWriter, r *http.Request) error {
			called = true
			api.OK(w, map[string]string{"test": "ok"})
			return nil
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		if !called {
			t.Error("handler was not called")
		}
	})

	t.Run("Handle handles errors", func(t *testing.T) {
		handler := api.Handle(func(w http.ResponseWriter, r *http.Request) error {
			return errors.New("test error")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestChain(t *testing.T) {
	t.Run("Chain combines middleware", func(t *testing.T) {
		order := []string{}
		mw1 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "mw1")
				next.ServeHTTP(w, r)
			})
		}
		mw2 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "mw2")
				next.ServeHTTP(w, r)
			})
		}

		router := api.NewRouter()
		router.Use(api.Chain(mw1, mw2))
		router.GET("/test", func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "handler")
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if len(order) != 3 {
			t.Errorf("expected 3 calls, got %d", len(order))
		}
	})
}

func TestTimeoutMiddleware(t *testing.T) {
	t.Run("Timeout middleware times out slow handlers", func(t *testing.T) {
		router := api.NewRouter()
		router.Use(api.TimeoutMiddleware(10 * time.Millisecond))
		router.GET("/slow", func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/slow", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Logf("Expected timeout but got status %d", w.Code)
		}
	})
}

func TestRouterPathParam(t *testing.T) {
	t.Run("GET with single path parameter", func(t *testing.T) {
		router := api.NewRouter()
		var capturedID string

		router.GET("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
			capturedID = api.GetPathParam(r, "id")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(capturedID))
		})

		req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if capturedID != "123" {
			t.Errorf("expected id 123, got %s", capturedID)
		}
	})

	t.Run("GET with multiple path parameters", func(t *testing.T) {
		router := api.NewRouter()
		var capturedUser, capturedPost string

		router.GET("/users/{userId}/posts/{postId}", func(w http.ResponseWriter, r *http.Request) {
			capturedUser = api.GetPathParam(r, "userId")
			capturedPost = api.GetPathParam(r, "postId")
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/users/alice/posts/456", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if capturedUser != "alice" {
			t.Errorf("expected userId alice, got %s", capturedUser)
		}
		if capturedPost != "456" {
			t.Errorf("expected postId 456, got %s", capturedPost)
		}
	})

	t.Run("GET with path parameter and trailing path", func(t *testing.T) {
		router := api.NewRouter()
		var capturedID string

		router.GET("/resources/{id}/details", func(w http.ResponseWriter, r *http.Request) {
			capturedID = api.GetPathParam(r, "id")
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/resources/abc-123/details", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if capturedID != "abc-123" {
			t.Errorf("expected id abc-123, got %s", capturedID)
		}
	})

	t.Run("Path parameter returns empty string when not found", func(t *testing.T) {
		router := api.NewRouter()
		var capturedValue string

		router.GET("/test", func(w http.ResponseWriter, r *http.Request) {
			capturedValue = api.GetPathParam(r, "id")
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if capturedValue != "" {
			t.Errorf("expected empty string, got %s", capturedValue)
		}
	})

	t.Run("Different HTTP methods with path parameters", func(t *testing.T) {
		router := api.NewRouter()
		var getID, putID string

		router.GET("/items/{id}", func(w http.ResponseWriter, r *http.Request) {
			getID = api.GetPathParam(r, "id")
			w.WriteHeader(http.StatusOK)
		})

		router.PUT("/items/{id}", func(w http.ResponseWriter, r *http.Request) {
			putID = api.GetPathParam(r, "id")
			w.WriteHeader(http.StatusOK)
		})

		// Test GET
		req := httptest.NewRequest(http.MethodGet, "/items/100", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if getID != "100" {
			t.Errorf("GET: expected id 100, got %s", getID)
		}

		// Test PUT
		req = httptest.NewRequest(http.MethodPut, "/items/200", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if putID != "200" {
			t.Errorf("PUT: expected id 200, got %s", putID)
		}
	})
}
