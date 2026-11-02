package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/remotetalk/backend/internal/auth"
)

const testSecret = "test-secret-with-at-least-32-chars-long!"

func newTestRouter(t *testing.T) (*gin.Engine, *auth.TokenIssuer) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	issuer := auth.NewTokenIssuer(testSecret, time.Hour)

	r := gin.New()
	r.GET("/protected", AuthMiddleware(issuer), func(c *gin.Context) {
		id, ok := CurrentUserID(c)
		if !ok {
			c.String(http.StatusInternalServerError, "no user id")
			return
		}
		c.String(http.StatusOK, id.String())
	})
	return r, issuer
}

func TestAuthMiddlewareMissingHeader(t *testing.T) {
	r, _ := newTestRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"unauthorized"`) {
		t.Fatalf("expected unauthorized code in body, got %s", w.Body.String())
	}
}

func TestAuthMiddlewareWrongScheme(t *testing.T) {
	r, _ := newTestRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic abc")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	r, _ := newTestRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddlewareValidToken(t *testing.T) {
	r, issuer := newTestRouter(t)
	userID := uuid.New()
	tok, err := issuer.Generate(userID)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok.Token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%s)", w.Code, w.Body.String())
	}
	if w.Body.String() != userID.String() {
		t.Fatalf("expected body %q, got %q", userID.String(), w.Body.String())
	}
}
