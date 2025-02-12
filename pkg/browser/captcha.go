package browser

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/go-rod/rod/lib/proto"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"xiaoyu/pkg/ocr"
)

type CaptchaHandler struct {
	browser *Browser
	client  *ocr.Client
	config  *ocr.Config
}

func NewCaptchaHandler(browser *Browser, ocrBaseURL string) (*CaptchaHandler, error) {
	if ocrBaseURL == "" {
		return nil, errors.New("OCR base URL must be configured")
	}

	client, err := ocr.NewClient(ocrBaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OCR client: %w", err)
	}

	config := ocr.NewConfigWithURL(ocrBaseURL)
	if err = config.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid OCR config: %w", err)
	}

	return &CaptchaHandler{
		browser: browser,
		client:  client,
		config:  config,
	}, nil
}

func (h *CaptchaHandler) HandleCaptcha(ctx context.Context, selector *Selector) (string, error) {
	if selector.CaptchaImg == "" || selector.CaptchaInput == "" {
		return "", nil
	}

	logger := log.WithField("action", "handle_captcha")

	imgEL, err := h.browser.findElement(selector.CaptchaImg, "captcha image")
	if err != nil {
		return "", fmt.Errorf("captcha image not found: %w", err)
	}

	data, err := imgEL.Screenshot(proto.PageCaptureScreenshotFormat("png"), 1)
	if err != nil {
		return "", fmt.Errorf("captcha image data failed: %w", err)
	}

	err = os.WriteFile("1.png", data, 0644)
	if err != nil {
		return "", fmt.Errorf("captcha image data failed: %w", err)
	}

	base64Data := base64.StdEncoding.EncodeToString(data)
	if len(base64Data) < ocr.MinImageSize || len(base64Data) > ocr.MaxImageSize {
		return "", ocr.ErrInvalidImage
	}

	result, err := h.client.RecognizeCaptcha(ctx, base64Data)
	if err != nil {
		return "", fmt.Errorf("OCR failed: %w", err)
	}

	result = strings.TrimSpace(result)
	if result == "" {
		return "", ocr.ErrInvalidResponse
	}

	logger.WithField("result", result).Debug("Captcha recognized")
	return result, nil
}
