package statistics

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

type UpstreamError struct {
	StatusCode int
	Message    string
}

func (e *UpstreamError) Error() string {
	return fmt.Sprintf("el servicio de estadísticas respondió %d: %s", e.StatusCode, e.Message)
}

type requestBody struct {
	Matrices map[string][][]float64 `json:"matrices"`
}

type responseBody struct {
	Statistics json.RawMessage `json:"statistics"`
}

// Statistics envía las matrices a la API en Node.js y devuelve su objeto de
// estadísticas sin interpretarlo. El Authorization original viaja con la llamada
// para que el servicio de destino vea la misma identidad.
func (c *Client) Statistics(ctx context.Context, authorization string, matrices map[string][][]float64) (json.RawMessage, error) {
	payload, err := json.Marshal(requestBody{Matrices: matrices})
	if err != nil {
		return nil, fmt.Errorf("no se pudo codificar la petición de estadísticas: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/statistics", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("no se pudo construir la petición de estadísticas: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("no se pudo llamar al servicio de estadísticas: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, upstreamError(res)
	}

	var body responseBody
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&body); err != nil {
		return nil, fmt.Errorf("no se pudo decodificar la respuesta de estadísticas: %w", err)
	}
	if len(body.Statistics) == 0 {
		return nil, errors.New("la respuesta de estadísticas no trae el objeto statistics")
	}

	return body.Statistics, nil
}

func upstreamError(res *http.Response) error {
	message := "error inesperado"
	var body struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(io.LimitReader(res.Body, 64<<10)).Decode(&body); err == nil && body.Error.Message != "" {
		message = body.Error.Message
	}

	return &UpstreamError{StatusCode: res.StatusCode, Message: message}
}
