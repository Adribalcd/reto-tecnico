package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"reto-tecnico/go-api/internal/auth"
	"reto-tecnico/go-api/internal/statistics"
)

const (
	testSecret = "test-secret"
	testIssuer = "reto-tecnico"
)

type statisticsStub struct {
	status int
	body   string
}

func newTestApp(t *testing.T, stub statisticsStub) *fiber.App {
	t.Helper()

	downstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(stub.status)
		_, _ = io.WriteString(w, stub.body)
	}))
	t.Cleanup(downstream.Close)

	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	(&Server{
		Tokens:       auth.NewVerifier(testSecret, testIssuer),
		Statistics:   statistics.NewClient(downstream.URL, time.Second),
		MaxMatrixDim: 10,
	}).Register(app)

	return app
}

func TestQRReturnsFactorizationAndStatistics(t *testing.T) {
	app := newTestApp(t, statisticsStub{
		status: http.StatusOK,
		body:   `{"statistics":{"max":0,"min":-1,"average":0,"sum":0,"diagonal":{"any":false,"matrices":{"q":false,"r":false}}}}`,
	})

	status, body := post(t, app, "/api/v1/qr", token(t), `{"matrix":[[1,0,0],[-1,0,1]]}`)
	if status != http.StatusOK {
		t.Fatalf("status = %d, body = %s", status, body)
	}

	var response struct {
		Matrix     [][]float64     `json:"matrix"`
		Q          [][]float64     `json:"q"`
		R          [][]float64     `json:"r"`
		Statistics json.RawMessage `json:"statistics"`
	}
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		t.Fatalf("no se pudo decodificar la respuesta: %v (%s)", err, body)
	}

	if len(response.Q) != 2 || len(response.Q[0]) != 2 {
		t.Fatalf("Q tiene una forma inesperada: %v", response.Q)
	}
	if len(response.R) != 2 || len(response.R[0]) != 3 {
		t.Fatalf("R tiene una forma inesperada: %v", response.R)
	}
	if len(response.Statistics) == 0 {
		t.Fatal("la respuesta no incluye estadísticas")
	}
}

func TestQRRequiresValidToken(t *testing.T) {
	app := newTestApp(t, statisticsStub{status: http.StatusOK, body: `{"statistics":{}}`})

	cases := map[string]string{
		"ausente":  "no-header",
		"inválido": "Bearer not-a-token",
	}

	for name, header := range cases {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/qr", nil)
			if header != "no-header" {
				req.Header.Set(fiber.HeaderAuthorization, header)
			}

			res, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test falló: %v", err)
			}
			if res.StatusCode != http.StatusUnauthorized {
				t.Fatalf("status = %d, se esperaba %d", res.StatusCode, http.StatusUnauthorized)
			}
		})
	}
}

func TestQRRejectsInvalidMatrix(t *testing.T) {
	app := newTestApp(t, statisticsStub{status: http.StatusOK, body: `{"statistics":{}}`})

	status, body := post(t, app, "/api/v1/qr", token(t), `{"matrix":[[1,2],[3]]}`)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", status, body)
	}
}

func TestQRReportsDownstreamFailure(t *testing.T) {
	app := newTestApp(t, statisticsStub{
		status: http.StatusBadRequest,
		body:   `{"error":{"code":"invalid_matrix","message":"rechazada"}}`,
	})

	status, body := post(t, app, "/api/v1/qr", token(t), `{"matrix":[[1,2],[3,4]]}`)
	if status != http.StatusBadGateway {
		t.Fatalf("status = %d, body = %s", status, body)
	}
}

func TestHealth(t *testing.T) {
	app := newTestApp(t, statisticsStub{status: http.StatusOK, body: `{"statistics":{}}`})

	res, err := app.Test(httptest.NewRequest(http.MethodGet, "/health", nil))
	if err != nil {
		t.Fatalf("app.Test falló: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, se esperaba %d", res.StatusCode, http.StatusOK)
	}
}

func post(t *testing.T, app *fiber.App, path, authorization, payload string) (int, string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(payload))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	req.Header.Set(fiber.HeaderAuthorization, authorization)

	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test falló: %v", err)
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("no se pudo leer la respuesta: %v", err)
	}

	return res.StatusCode, string(raw)
}

func token(t *testing.T) string {
	t.Helper()

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "admin",
		Issuer:    testIssuer,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("no se pudo firmar el token: %v", err)
	}

	return "Bearer " + signed
}
