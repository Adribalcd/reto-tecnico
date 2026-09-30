package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testSecret = "test-secret"
	testIssuer = "reto-tecnico"
)

func TestVerifyAcceptsValidToken(t *testing.T) {
	verifier := NewVerifier(testSecret, testIssuer)
	token := sign(t, testSecret, testIssuer, time.Now().Add(time.Hour))

	claims, err := verifier.Verify(token)
	if err != nil {
		t.Fatalf("Verify devolvió un error: %v", err)
	}
	if claims.Subject != "admin" {
		t.Fatalf("subject = %q, se esperaba %q", claims.Subject, "admin")
	}
}

func TestVerifyRejectsInvalidTokens(t *testing.T) {
	verifier := NewVerifier(testSecret, testIssuer)

	cases := map[string]string{
		"secreto incorrecto": sign(t, "another-secret", testIssuer, time.Now().Add(time.Hour)),
		"emisor incorrecto":  sign(t, testSecret, "another-issuer", time.Now().Add(time.Hour)),
		"expirado":           sign(t, testSecret, testIssuer, time.Now().Add(-time.Minute)),
		"malformado":         "not.a.token",
	}

	for name, token := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := verifier.Verify(token); err == nil {
				t.Fatal("se esperaba que Verify rechazara el token")
			}
		})
	}
}

func sign(t *testing.T, secret, issuer string, expiresAt time.Time) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "admin",
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	})

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("no se pudo firmar el token: %v", err)
	}
	return signed
}
