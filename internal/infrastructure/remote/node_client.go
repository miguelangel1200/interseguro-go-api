// Package remote contiene los adaptadores de salida (secundarios) que se
// comunican con servicios externos vía HTTP.
package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"interseguro/go-api/internal/domain"
)

// NodeClient es el adaptador de salida que implementa StatisticsRepository
// comunicándose con la API Node.js.
type NodeClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewNodeClient crea el adaptador hacia la API Node.js.
func NewNodeClient(baseURL string) *NodeClient {
	return &NodeClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// sendStatisticsRequest es el DTO de salida (request) hacia la API Node.js.
type sendStatisticsRequest struct {
	Original domain.Matrix `json:"original"`
	Rotated  domain.Matrix `json:"rotated"`
	Q        domain.Matrix `json:"Q"`
	R        domain.Matrix `json:"R"`
}

// SendStatistics envía el resultado de la factorización QR a la API Node.js
// y devuelve las estadísticas calculadas. authToken se propaga como Bearer
// para autenticar la llamada al servicio remoto.
func (c *NodeClient) SendStatistics(ctx context.Context, res domain.QRResult, authToken string) (domain.Statistics, error) {
	payload, err := json.Marshal(sendStatisticsRequest{
		Original: res.Original,
		Rotated:  res.Rotated,
		Q:        res.Q,
		R:        res.R,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}

	url := c.baseURL + "/statistics"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call node api: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read node api response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("node api returned status %d: %s", resp.StatusCode, string(body))
	}

	var out domain.Statistics
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("parse node api response: %w", err)
	}
	return out, nil
}
