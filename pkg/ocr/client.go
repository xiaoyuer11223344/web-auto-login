package ocr

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Client struct {
	baseURL     string
	httpClient  *http.Client
	initialized bool // Ensure proper initialization
}

type OCRRequest struct {
	Image []byte `json:"image"`
}

type OCRResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func NewClient(baseURL string) (*Client, error) {
	if baseURL == "" {
		return nil, errors.New("OCR base URL must be provided through environment configuration")
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: DefaultConfig().Timeout,
		},
		initialized: true,
	}, nil
}

func (c *Client) RecognizeCaptcha(ctx context.Context, imageData []byte) (string, error) {
	if !c.initialized {
		return "", errors.New("OCR client not properly initialized")
	}

	logger := log.WithField("action", "ocr_request")

	reqBody := OCRRequest{
		Image: imageData,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/ocr/b64/json", c.baseURL)
	if c.baseURL == "" {
		return "", errors.New("OCR endpoint URL not configured")
	}
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OCR request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var ocrResp OCRResponse
	if err = json.Unmarshal(body, &ocrResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if ocrResp.Code != 0 {
		return "", fmt.Errorf("OCR service error: %s", ocrResp.Message)
	}

	logger.WithField("result", ocrResp.Data).Debug("OCR request successful")
	return ocrResp.Data, nil
}
