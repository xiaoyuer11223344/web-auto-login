package ocr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Client struct {
	baseURL     string
	httpClient  *http.Client
	initialized bool // Ensure proper initialization
}

type OCRRequest struct {
	Image       string `json:"image"`
	Probability bool   `json:"probability"`
	PngFix      bool   `json:"png_fix"`
}

// {"code":200,"message":"Success","data":"sw9f"}
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

func (c *Client) RecognizeCaptcha(ctx context.Context, base64Image string) (string, error) {
	if !c.initialized {
		return "", errors.New("OCR client not properly initialized")
	}

	logger := log.WithField("action", "ocr_request")

	//reqBody := OCRRequest{
	//	Image:       imageData,
	//	Probability: false,
	//	PngFix:      false,
	//}
	//
	//jsonData, err := json.Marshal(reqBody)
	//if err != nil {
	//	return "", fmt.Errorf("failed to marshal request: %w", err)
	//}

	endPoint := fmt.Sprintf("%s/ocr", c.baseURL)
	if c.baseURL == "" {
		return "", errors.New("OCR endpoint URL not configured")
	}

	data := url.Values{}
	data.Set("image", base64Image)
	data.Set("probability", "false")
	data.Set("png_fix", "false")

	req, err := http.NewRequestWithContext(ctx, "POST", endPoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

	if ocrResp.Code != 200 {
		return "", fmt.Errorf("OCR service error: %s", ocrResp.Message)
	}

	logger.WithField("result", ocrResp).Info("OCR request successful")
	return ocrResp.Data, nil
}
