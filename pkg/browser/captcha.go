package browser

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/go-rod/rod/lib/proto"
	log "github.com/sirupsen/logrus"
	"zp-weblogin/pkg/ocr"
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
	if err := config.ValidateConfig(); err != nil {
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
	
	imgEl, err := h.browser.findElement(selector.CaptchaImg, "captcha image")
	if err != nil {
		return "", fmt.Errorf("captcha image not found: %w", err)
	}

	screenshot, err := imgEl.Screenshot(proto.PageCaptureScreenshotFormat("png"), 1)
	if err != nil {
		return "", fmt.Errorf("failed to capture captcha: %w", err)
	}

	base64Data := base64.StdEncoding.EncodeToString(screenshot)
	if len(base64Data) < ocr.MinImageSize || len(base64Data) > ocr.MaxImageSize {
		return "", ocr.ErrInvalidImage
	}

	result, err := h.client.RecognizeCaptcha(ctx, []byte(base64Data))
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
