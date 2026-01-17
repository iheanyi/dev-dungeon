package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSPAHandler_ServesIndexForSPARoutes(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Create index.html
	indexContent := "<html><body>SPA Index</body></html>"
	if err := os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte(indexContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a static asset
	if err := os.WriteFile(filepath.Join(tmpDir, "style.css"), []byte("body {}"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create _app directory with a JS file
	appDir := filepath.Join(tmpDir, "_app")
	if err := os.Mkdir(appDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "app.js"), []byte("console.log('app')"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create SPA handler
	s := &Server{config: Config{StaticDir: tmpDir}}
	handler := s.spaHandlerDir(tmpDir)

	tests := []struct {
		name           string
		path           string
		wantStatus     int
		wantBody       string
		wantBodySubstr string
	}{
		{
			name:       "root serves index.html",
			path:       "/",
			wantStatus: http.StatusOK,
			wantBody:   indexContent,
		},
		{
			name:           "static file serves directly",
			path:           "/style.css",
			wantStatus:     http.StatusOK,
			wantBodySubstr: "body {}",
		},
		{
			name:           "nested static file serves directly",
			path:           "/_app/app.js",
			wantStatus:     http.StatusOK,
			wantBodySubstr: "console.log",
		},
		{
			name:       "SPA route /leaderboard serves index.html",
			path:       "/leaderboard",
			wantStatus: http.StatusOK,
			wantBody:   indexContent,
		},
		{
			name:       "SPA route /about serves index.html",
			path:       "/about",
			wantStatus: http.StatusOK,
			wantBody:   indexContent,
		},
		{
			name:       "SPA route /players/testuser serves index.html",
			path:       "/players/testuser",
			wantStatus: http.StatusOK,
			wantBody:   indexContent,
		},
		{
			name:       "SPA route /daily serves index.html",
			path:       "/daily",
			wantStatus: http.StatusOK,
			wantBody:   indexContent,
		},
		{
			name:       "nonexistent route serves index.html for SPA",
			path:       "/nonexistent",
			wantStatus: http.StatusOK,
			wantBody:   indexContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			body := rec.Body.String()
			if tt.wantBody != "" && body != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
			if tt.wantBodySubstr != "" && !contains(body, tt.wantBodySubstr) {
				t.Errorf("body %q does not contain %q", body, tt.wantBodySubstr)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	if start+len(substr) > len(s) {
		return false
	}
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
