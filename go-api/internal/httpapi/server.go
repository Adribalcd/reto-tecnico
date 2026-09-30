package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"reto-tecnico/go-api/internal/auth"
	"reto-tecnico/go-api/internal/matrix"
	"reto-tecnico/go-api/internal/statistics"
)

type Server struct {
	Tokens       *auth.Verifier
	Statistics   *statistics.Client
	MaxMatrixDim int
}

type requestBody struct {
	Matrix [][]float64 `json:"matrix"`
}

type responseBody struct {
	Matrix     [][]float64     `json:"matrix"`
	Q          [][]float64     `json:"q"`
	R          [][]float64     `json:"r"`
	Statistics json.RawMessage `json:"statistics"`
}

type errorBody struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (s *Server) Register(app *fiber.App) {
	app.Get("/health", s.health)

	api := app.Group("/api/v1")
	api.Post("/qr", s.authenticate, s.qr)
}

func (s *Server) health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok", "service": "go-api"})
}

func (s *Server) qr(c *fiber.Ctx) error {
	var body requestBody
	if err := c.BodyParser(&body); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid_body", "el cuerpo debe ser un objeto JSON con el campo matrix")
	}

	rows, cols, err := matrix.Dimensions(body.Matrix)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid_matrix", err.Error())
	}
	if rows > s.MaxMatrixDim || cols > s.MaxMatrixDim {
		return fail(c, fiber.StatusRequestEntityTooLarge, "matrix_too_large",
			fmt.Sprintf("la matriz no puede superar %d filas ni columnas", s.MaxMatrixDim))
	}

	q, r, err := matrix.QRDecompose(body.Matrix)
	if err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid_matrix", err.Error())
	}

	stats, err := s.Statistics.Statistics(c.UserContext(), c.Get(fiber.HeaderAuthorization), map[string][][]float64{
		"q": q,
		"r": r,
	})
	if err != nil {
		var upstream *statistics.UpstreamError
		if errors.As(err, &upstream) {
			return fail(c, fiber.StatusBadGateway, "statistics_unavailable", upstream.Error())
		}
		return fail(c, fiber.StatusBadGateway, "statistics_unavailable", "el servicio de estadísticas no está disponible")
	}

	return c.JSON(responseBody{Matrix: body.Matrix, Q: q, R: r, Statistics: stats})
}

func (s *Server) authenticate(c *fiber.Ctx) error {
	const prefix = "Bearer "

	header := c.Get(fiber.HeaderAuthorization)
	if !strings.HasPrefix(header, prefix) {
		return fail(c, fiber.StatusUnauthorized, "unauthorized", "falta el header Authorization con un token Bearer")
	}

	if _, err := s.Tokens.Verify(strings.TrimSpace(strings.TrimPrefix(header, prefix))); err != nil {
		return fail(c, fiber.StatusUnauthorized, "unauthorized", "el token no es válido o ya expiró")
	}

	return c.Next()
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	message := "error inesperado del servidor"

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		status = fiberErr.Code
		message = fiberErr.Message
	}

	return fail(c, status, codeForStatus(status), message)
}

func fail(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(errorBody{Error: apiError{Code: code, Message: message}})
}

func codeForStatus(status int) string {
	switch status {
	case fiber.StatusNotFound:
		return "not_found"
	case fiber.StatusMethodNotAllowed:
		return "method_not_allowed"
	case fiber.StatusBadRequest:
		return "bad_request"
	case fiber.StatusUnauthorized:
		return "unauthorized"
	default:
		return "internal_error"
	}
}
