package push

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Gateway struct {
	client  *http.Client
	baseURL string
}

func NewGateway(baseURL string, timeout time.Duration) *Gateway {
	return &Gateway{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		baseURL: baseURL,
	}
}

type pushRequest struct {
	RequestId   string            `json:"request_id"`
	RecipientId string            `json:"recipient_id"`
	Kind        string            `json:"kind"`
	Data        map[string]string `json:"data"`
}

func (g *Gateway) RequestPosition(ctx context.Context, tripId, driverId uuid.UUID) error {
	body := pushRequest{
		RequestId:   uuid.New().String(),
		RecipientId: driverId.String(),
		Kind:        "REQUEST_POSITION",
		Data:        map[string]string{"trip_id": tripId.String()},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal push request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/api/v1/push", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create push request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("push request: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusAccepted:
		return nil
	case resp.StatusCode == http.StatusTooManyRequests:
		return fmt.Errorf("push rate limited: retry-after=%s", resp.Header.Get("Retry-After"))
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		return fmt.Errorf("push client error: status=%d", resp.StatusCode)
	case resp.StatusCode >= 500:
		return fmt.Errorf("push server error: status=%d", resp.StatusCode)
	default:
		return fmt.Errorf("push unexpected status: %d", resp.StatusCode)
	}
}

func (g *Gateway) Close() {
	if g.client != nil {
		g.client.CloseIdleConnections()
	}
}
