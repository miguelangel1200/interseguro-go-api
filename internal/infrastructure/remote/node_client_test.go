package remote

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"interseguro/go-api/internal/domain"
)

func sampleResult() domain.QRResult {
	return domain.QRResult{
		Original: domain.Matrix{{1, 2}, {3, 4}},
		Rotated:  domain.Matrix{{3, 1}, {4, 2}},
		Q:        domain.Matrix{{0.3, 0.9}, {0.9, -0.3}},
		R:        domain.Matrix{{3.2, 4.4}, {0, 0.6}},
	}
}

func TestSendStatistics_Success(t *testing.T) {
	var gotAuth, gotPath, gotCT string
	var gotBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		gotCT = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"mean": 2.5, "count": 4}`))
	}))
	defer srv.Close()

	client := NewNodeClient(srv.URL)
	stats, err := client.SendStatistics(context.Background(), sampleResult(), "jwt-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/statistics" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	if gotCT != "application/json" {
		t.Errorf("unexpected content type: %s", gotCT)
	}
	if gotAuth != "Bearer jwt-token" {
		t.Errorf("expected bearer token propagation, got %q", gotAuth)
	}
	if stats["mean"] != 2.5 {
		t.Fatalf("unexpected stats: %v", stats)
	}

	// El body debe incluir las cuatro matrices y estar bien formado.
	for _, key := range []string{"original", "rotated", "Q", "R"} {
		if _, ok := gotBody[key]; !ok {
			t.Errorf("request body missing %q: %v", key, gotBody)
		}
	}
}

func TestSendStatistics_Non200Status(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	defer srv.Close()

	client := NewNodeClient(srv.URL)
	_, err := client.SendStatistics(context.Background(), sampleResult(), "")
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Fatalf("error should include status code: %v", err)
	}
}

func TestSendStatistics_InvalidJSONResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("esto no es json"))
	}))
	defer srv.Close()

	client := NewNodeClient(srv.URL)
	if _, err := client.SendStatistics(context.Background(), sampleResult(), ""); err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func TestSendStatistics_ConnectionError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	url := srv.URL
	srv.Close() // fuerza un error de conexión

	client := NewNodeClient(url)
	if _, err := client.SendStatistics(context.Background(), sampleResult(), ""); err == nil {
		t.Fatal("expected connection error")
	}
}

func TestSendStatistics_ContextCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := NewNodeClient(srv.URL)
	if _, err := client.SendStatistics(ctx, sampleResult(), ""); err == nil {
		t.Fatal("expected error for canceled context")
	}
}

func TestNewNodeClient_Defaults(t *testing.T) {
	client := NewNodeClient("http://localhost:3000")
	if client.baseURL != "http://localhost:3000" {
		t.Fatalf("unexpected baseURL: %s", client.baseURL)
	}
	if client.httpClient == nil {
		t.Fatal("httpClient must not be nil")
	}
}
