package browser

import (
	"fmt"
	"sort"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	log "github.com/sirupsen/logrus"
)

type FormDesc struct {
	Form      *rod.Element
	Score     int
	HasLogin  bool
	HasPass   bool
	HasSubmit bool
	Position  proto.Point
	selector  *Selector
}

type Selector struct {
	UserInput     string `yaml:"userInput" json:"userInput"`
	PasswordInput string `yaml:"passwordInput" json:"passwordInput"`
	LoginBtn      string `yaml:"loginBtn" json:"loginBtn"`
	RememberMe    string `yaml:"rememberMe" json:"rememberMe"`
	CaptchaInput  string `yaml:"captchaInput" json:"captchaInput"`
	CaptchaImg    string `yaml:"captchaImg" json:"captchaImg"`
	form          *rod.Element
}

var (

	// Common selectors for form elements
	userInputSelectors = []string{
		"input[id='uid']",
		"input[id='usernameIpt']",
		"input[id='account']",
		"input[id='username']",
		"input[id*='user']",
		"input[id='loginid']",
		"input[name*='user']",
		"input[class*='user']",

		"input[placeholder*='账号']",
		"input[placeholder*='用户']",
		"input[placeholder*='工号']",
		"input[placeholder*='邮箱']",

		"input[name='user[login]']",
		"input[name='username']",
		"input[name='uid']",
		"input[name='account']",
	}

	passInputSelectors = []string{
		"input[type='password']",
		"input[placeholder='密码']",
		"input[placeholder*='密码']",
		"input[name='user[password]']",
		"input[name='pwd']",
		"input[id='pwd']",
		"input[id*='pass']",
		"input[name*='pass']",
		"input[class*='pass']",
	}

	loginBtnSelectors = []string{
		//"button[name='submit']",
		"button[type='submit']",
		"button[type='button']",
		//"button[id*='login-btn']",
		//"button[class*='login_button']",
		//"button[class*='btn-login']",
		//"button[class*='submit-btn']",
		//"button[class*='el-button']",
		//"button[value*='登录']",
		//"button[id*='commit']",
		//
		//"input[value='Login']",
		//"input[value*='Sign in']",
		//"input[id*='login-btn']",
		//"input[class*='login_button']",
		//"input[class*='btn-login']",
		//"input[name='commit']",
		//"input[type='submit']",
		//
		//"div[class*='login_button']",
		//"div[class*='btn-login']",
		//
		//"#login-btn",
		//"#loginBtn",
		//".btn-login",
		//".login-btn",
		//".btn-primary",
	}

	checkBoxSelectors = []string{
		"input[type='checkbox']",
	}

	captchaInputSelectors = []string{
		"input[placeholder*='验证码']",
		"input[placeholder*='verification']",
		"input[placeholder*='Verification']",
	}

	captchaImageSelectors = []string{
		"img",
		//"input[id='checkCode']",
		//"img[id*='captcha']",
		//"img[id*='Captcha']",
		//"img[alt*='验证码']",
		//"img[alt*='captcha']",
		//"img[src*='captcha']",
		//"img[src*='verify']",
		//"img[class*='captcha']",
		//"img[id*='captcha']",
		//"img[title*='验证码']",
		//"img[title*='captcha']",

		".captcha-img",
		".verify-img",

		// ElementUI
		".el-image img[src*='captcha']",
		".el-image[alt*='验证码']",
		".el-image[alt*='captcha']",
	}
)

func (b *Browser) scoreLoginForm(form *rod.Element) (*FormDesc, error) {
	logger := log.WithField("action", "socre_login_form")
	_ = logger

	formDesc := &FormDesc{Form: form}

	_selector, err := b.findFormElements(form, true)
	if err != nil {
		return nil, err
	}

	formDesc.selector = _selector

	if _selector.UserInput != "" {
		formDesc.Score++
	}
	if _selector.PasswordInput != "" {
		formDesc.Score++
	}
	if _selector.LoginBtn != "" {
		formDesc.Score++
	}
	if _selector.CaptchaInput != "" {
		formDesc.Score++
	}
	if _selector.CaptchaImg != "" {
		formDesc.Score++
	}

	return formDesc, nil
}

// findFormElements
// @Description: 匹配表单内元素
// @receiver b
// @param form
// @return *Selector
// @return error
func (b *Browser) findFormElements(form *rod.Element, enhance bool) (*Selector, error) {
	logger := log.WithField("action", "find_form_elements")

	selector := &Selector{form: form}

	// Find username input with retry
	for i := 0; i < MaxRetries; i++ {
		for _, sel := range userInputSelectors {
			fmt.Println(sel)
			if el, err := form.Element(sel); err == nil && el != nil {
				if visible, _ := el.Visible(); visible {
					selector.UserInput = el.MustGetXPath(false)
					logger.WithField("xpath", selector.UserInput).Debug("Found username input")
					goto foundPass
				}
			}
		}
		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			logger.WithField("attempt", i+1).Debug("Username input not found, retrying...")
		}
	}

foundPass:
	// Find password input with retry
	for i := 0; i < MaxRetries; i++ {
		for _, sel := range passInputSelectors {
			fmt.Println(sel)
			if el, err := form.Element(sel); err == nil && el != nil {
				if visible, _ := el.Visible(); visible {
					selector.PasswordInput = el.MustGetXPath(false)
					logger.WithField("xpath", selector.PasswordInput).Debug("Found password input")
					goto foundButton
				}
			}
		}
		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			logger.WithField("attempt", i+1).Debug("Password input not found, retrying...")
		}
	}

foundButton:
	// Find login button with retry
	for i := 0; i < MaxRetries; i++ {
		for _, sel := range loginBtnSelectors {
			if el, err := form.Element(sel); err == nil && el != nil {
				if visible, _ := el.Visible(); visible {
					selector.LoginBtn = el.MustGetXPath(false)
					logger.WithField("xpath", selector.LoginBtn).Debug("Found login button")
					goto foundRememberCheckBox
				}
			}
		}
		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			logger.WithField("attempt", i+1).Debug("Login button not found, retrying...")
		}
	}

	// enhance login button
	if enhance {
		logger.WithField("attempt", 0).Debug("Login button not found, enhance retrying...")
		for _, sel := range loginBtnSelectors {
			if el, err := b.page.Element(sel); err == nil && el != nil {
				if visible, _ := el.Visible(); visible {
					selector.LoginBtn = el.MustGetXPath(false)
					logger.WithField("xpath", selector.LoginBtn).Debug("Enhance Found login button")
					goto foundRememberCheckBox
				}
			}
		}
	}

foundRememberCheckBox:
	for i := 0; i < MaxRetries; i++ {
		for _, sel := range checkBoxSelectors {
			// Find checkboxes (both remember me and agreement types)
			checkboxes, err := form.Elements(sel)
			if err == nil && len(checkboxes) > 0 {
				for _, checkbox := range checkboxes {
					//if visible, _ := checkbox.Visible(); visible {}
					selector.RememberMe = checkbox.MustGetXPath(false)
					logger.WithField("xpath", selector.RememberMe).Debug("Found rememberMe checkbox")
					goto foundCaptchaInput
				}
			}
		}

		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			logger.WithField("attempt", i+1).Debug("rememberMe checkbox not found, retrying...")
		}
	}

foundCaptchaInput:
	if b.captchaHandler != nil {
		for i := 0; i < MaxRetries; i++ {
			for _, sel := range captchaInputSelectors {
				if el, err := form.Element(sel); err == nil && el != nil {
					if visible, _ := el.Visible(); visible {
						selector.CaptchaInput = el.MustGetXPath(false)
						logger.WithField("xpath", selector.CaptchaInput).Debug("Found Captcha Input")
						goto foundCaptchaImage
					}
				}
			}

			if i < MaxRetries-1 {
				time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
				logger.WithField("attempt", i+1).Debug("captcha input not found, retrying...")
			}
		}
	}

foundCaptchaImage:
	if b.captchaHandler != nil && selector.CaptchaInput != "" {
		for i := 0; i < MaxRetries; i++ {
			for _, sel := range captchaImageSelectors {
				if el, err := form.Element(sel); err == nil && el != nil {
					if visible, _ := el.Visible(); visible {
						selector.CaptchaImg = el.MustGetXPath(false)
						logger.WithField("xpath", selector.CaptchaImg).Debug("Found Captcha Image")
						goto over
					}
				}
			}

			if i < MaxRetries-1 {
				time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
				logger.WithField("attempt", i+1).Debug("captcha image not found, retrying...")
			}
		}
	}

over:
	//return selector, nil
	if selector.UserInput != "" && selector.PasswordInput != "" && selector.LoginBtn != "" {
		return selector, nil
	}

	return nil, fmt.Errorf("not form all elements found")
}

// findFormElements
// @Description: 匹配表单内元素
// @receiver b
// @param form
// @return *Selector
// @return error
func (b *Browser) findElements() (*Selector, error) {
	logger := log.WithField("action", "find_form_elements")

	selector := &Selector{}

	// Find username input with retry
	for i := 0; i < MaxRetries; i++ {
		for _, sel := range userInputSelectors {
			if el, err := b.page.Element(sel); err == nil && el != nil {
				if visible, _ := el.Visible(); visible {
					selector.UserInput = el.MustGetXPath(false)
					logger.WithField("xpath", selector.UserInput).Debug("Found username input")
					goto foundPass
				}
			}
		}
		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			logger.WithField("attempt", i+1).Debug("Username input not found, retrying...")
		}
	}

foundPass:
	// Find password input with retry
	for i := 0; i < MaxRetries; i++ {
		for _, sel := range passInputSelectors {
			if el, err := b.page.Element(sel); err == nil && el != nil {
				if visible, _ := el.Visible(); visible {
					selector.PasswordInput = el.MustGetXPath(false)
					logger.WithField("xpath", selector.PasswordInput).Debug("Found password input")
					goto foundButton
				}
			}
		}
		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			logger.WithField("attempt", i+1).Debug("Password input not found, retrying...")
		}
	}

foundButton:
	// Find login button with retry
	fmt.Println(b.page.HTML())
	for i := 0; i < MaxRetries; i++ {
		for _, sel := range loginBtnSelectors {
			if el, err := b.page.Element(sel); err == nil && el != nil {
				if visible, _ := el.Visible(); visible {
					selector.LoginBtn = el.MustGetXPath(false)
					logger.WithField("xpath", selector.LoginBtn).Debug("Found login button")
					goto foundRememberCheckBox
				}
			}
		}
		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			logger.WithField("attempt", i+1).Debug("Login button not found, retrying...")
		}
	}

foundRememberCheckBox:
	for i := 0; i < MaxRetries; i++ {
		for _, sel := range checkBoxSelectors {
			// Find checkboxes (both remember me and agreement types)
			checkboxes, err := b.page.Elements(sel)
			if err == nil && len(checkboxes) > 0 {
				for _, checkbox := range checkboxes {
					//if visible, _ := checkbox.Visible(); visible {}
					selector.RememberMe = checkbox.MustGetXPath(false)
					logger.WithField("xpath", selector.RememberMe).Debug("Found rememberMe checkbox")
					goto foundCaptchaInput
				}
			}
		}

		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			logger.WithField("attempt", i+1).Debug("rememberMe checkbox not found, retrying...")
		}
	}

foundCaptchaInput:
	if b.captchaHandler != nil {
		for i := 0; i < MaxRetries; i++ {
			for _, sel := range captchaInputSelectors {
				if el, err := b.page.Element(sel); err == nil && el != nil {
					if visible, _ := el.Visible(); visible {
						selector.CaptchaInput = el.MustGetXPath(false)
						logger.WithField("xpath", selector.CaptchaInput).Debug("Found Captcha Input")
						goto foundCaptchaImage
					}
				}
			}

			if i < MaxRetries-1 {
				time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
				logger.WithField("attempt", i+1).Debug("captcha input not found, retrying...")
			}
		}
	}

foundCaptchaImage:
	if b.captchaHandler != nil && selector.CaptchaInput != "" {
		for i := 0; i < MaxRetries; i++ {
			for _, sel := range captchaImageSelectors {
				if el, err := b.page.Element(sel); err == nil && el != nil {
					if visible, _ := el.Visible(); visible {
						selector.CaptchaImg = el.MustGetXPath(false)
						logger.WithField("xpath", selector.CaptchaImg).Debug("Found Captcha Image")
						goto over
					}
				}
			}

			if i < MaxRetries-1 {
				time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
				logger.WithField("attempt", i+1).Debug("captcha image not found, retrying...")
			}
		}
	}

over:
	//return selector, nil
	if selector.UserInput != "" && selector.PasswordInput != "" && selector.LoginBtn != "" {
		return selector, nil
	}

	return nil, fmt.Errorf("not form all elements found")
}

// DetectFormSelectors
// @Description: 自动探测Form表单以及内部相关的其他标签元素
// @receiver b
// @return *Selector
// @return error
func (b *Browser) DetectFormSelectors() (*Selector, error) {
	logger := log.WithField("action", "detect_form_and_selectors")
	logger.Debug("Starting selector detection")

	var err error
	var s *Selector

	//s, err = b.findElements(b.page)
	//if err != nil {
	//	return nil, fmt.Errorf("no visible form found")
	//}

	var formScores []*FormDesc
	var forms rod.Elements
	var formEL *rod.Element

	if forms, err = b.page.Elements("form"); err == nil {
		if len(forms) > 0 && forms != nil {
			// 多form情况
			for _, formEL = range forms {
				score, formErr := b.scoreLoginForm(formEL)
				if formErr == nil {
					formScores = append(formScores, score)
				}
			}
		} else {
			// 单form情况
			if formEL, err = b.page.Element("form"); err == nil && formEL != nil {
				if visible, _ := formEL.Visible(); visible {
					score, formErr := b.scoreLoginForm(formEL)
					if formErr == nil {
						formScores = append(formScores, score)
					}
				}
			}
		}
	}

	if len(formScores) == 0 {
		// 没匹配到表单的话则进行结束
		return nil, fmt.Errorf("no visible form found")
	}

	formEL = formScores[0].Form
	s = formScores[0].selector

	// 得分排序
	sort.Slice(formScores, func(i, j int) bool {
		return formScores[i].Score > formScores[j].Score
	})

	// 匹配到表单的话则进行打印信息
	formDetails := map[string]interface{}{
		"tag":      formEL.MustEval("() => this.tagName ").String(),
		"id":       formEL.MustEval("() => this.id ").String(),
		"action":   formEL.MustEval("() => this.action ").String(),
		"method":   formEL.MustEval("() => this.method ").String(),
		"position": formEL.MustShape().Box(),
	}

	logger.WithFields(log.Fields{"formDetails": formDetails}).Info("Form Details")
	logger.WithFields(log.Fields{
		"score":       formScores[0].Score,
		"hasLogin":    formScores[0].HasLogin,
		"hasPass":     formScores[0].HasPass,
		"hasSubmit":   formScores[0].HasSubmit,
		"position":    formScores[0].Position,
		"formDetails": formDetails,
	}).Debug("Form scored with details")

	return s, nil
}
