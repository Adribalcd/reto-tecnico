package statistics

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStatisticsReturnsStatisticsObject(t *testing.T) {
	var received requestBody
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/statistics" {
			t.Errorf("ruta = %q, se esperaba %q", r.URL.Path, "/api/v1/statistics")
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Errorf("authorization = %q, se esperaba %q", got, "Bearer token")
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("no se pudo decodificar la petición: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"statistics":{"max":1,"min":-1}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, time.Second)
	matrices := map[string][][]float64{"q": {{1, 0}, {0, -1}}}

	stats, err := client.Statistics(context.Background(), "Bearer token", matrices)
	if err != nil {
		t.Fatalf("Statistics devolvió un error: %v", err)
	}
	if string(stats) != `{"max":1,"min":-1}` {
		t.Fatalf("statistics = %s", stats)
	}
	if len(received.Matrices["q"]) != 2 {
		t.Fatalf("las matrices no se reenviaron como se esperaba: %v", received.Matrices)
	}
}

func TestStatisticsMapsUpstreamFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"code":"invalid_matrix","message":"la matriz no puede estar vacía"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, time.Second)

	_, err := client.Statistics(context.Background(), "", map[string][][]float64{"q": {{1}}})
	if err == nil {
		t.Fatal("se esperaba un error del servicio de estadísticas")
	}

	var upstream *UpstreamError
	if !errors.As(err, &upstream) {
		t.Fatalf("tipo de error = %T, se esperaba *UpstreamError", err)
	}
	if upstream.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba %d", upstream.StatusCode, http.StatusBadRequest)
	}
	if upstream.Message != "la matriz no puede estar vacía" {
		t.Fatalf("message = %q", upstream.Message)
	}
}

func TestStatisticsReportsEmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, time.Second)

	if _, err := client.Statistics(context.Background(), "", map[string][][]float64{"q": {{1}}}); err == nil {
		t.Fatal("se esperaba un error cuando la respuesta no trae el objeto statistics")
	}
}
