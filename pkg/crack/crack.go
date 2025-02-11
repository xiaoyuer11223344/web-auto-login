package crack

import (
	"context"
	"strings"
	"time"

	"fmt"
	log "github.com/sirupsen/logrus"
	"xiaoyu/pkg/browser"
)

type Task struct {
	URL      string
	Username string
	Password string
	Timeout  int
}

type Result struct {
	Success  bool
	Error    error
	Task     Task
	Attempts int
}

type Cracker struct {
	delay        time.Duration
	maxAttempts  int
	maxCrackNum  int
	maxCrackTime time.Duration
	threads      int
	browser      *browser.Browser
	selector     *browser.Selector
}

func New(delay int, maxAttempts int, maxCrackNum int, maxCrackTime int, threads int, b *browser.Browser, s *browser.Selector) *Cracker {
	return &Cracker{
		delay:        time.Duration(delay) * time.Second,
		maxAttempts:  maxAttempts,
		maxCrackNum:  maxCrackNum,
		maxCrackTime: time.Duration(maxCrackTime) * time.Second,
		threads:      threads,
		browser:      b,
		selector:     s,
	}
}

func ProcessPassword(password string, username string) string {
	return strings.ReplaceAll(password, "%user%", username)
}

func (c *Cracker) SingleTaskCrack(ctx context.Context, task Task) []Result {

	var passwords []string
	if task.Password != "" {
		passwords = []string{task.Password}
	} else {
		passwords = []string{"123456", "1q2w3e4r", "12345", "Aa123456", "admin", "admin123", "admin@123", "Admin@123"}
	}

	log.WithFields(log.Fields{
		"action": "start_password_tests",
		"target": task.URL,
		"count":  len(passwords),
	}).Info("Starting password tests")

	var results []Result
	for _, pass := range passwords {
		select {
		case <-ctx.Done():
			return results
		default:
			_task := Task{
				URL:      task.URL,
				Username: task.Username,
				Password: pass,
			}

			result := c.processTask(ctx, _task)
			results = append(results, result)

			if result.Success {
				log.WithFields(log.Fields{
					"url":      _task.URL,
					"username": _task.Username,
					"password": _task.Password,
					"status":   "success",
				}).Info("Login successful")
				return results
			}

			if result.Error != nil && !strings.Contains(result.Error.Error(), "timed out") {
				log.WithFields(log.Fields{
					"url":      _task.URL,
					"username": _task.Username,
					"password": _task.Password,
					"status":   "fail",
					"error":    result.Error.Error(),
				}).Debug("Login attempt failed")
			}

			time.Sleep(c.delay)
		}
	}
	return results
}

//
//func (c *Cracker) Start(ctx context.Context, tasks []Task) []Result {
//	var passwords []string
//	if len(tasks) > 0 && len(tasks[0].Password) > 0 {
//		passwords = []string{tasks[0].Password}
//	} else {
//		passwords = []string{
//			"123456", "1q2w3e4r", "12345", "Aa123456", "admin",
//			"admin123", "admin@123", "Admin@123",
//		}
//	}
//
//	log.WithFields(log.Fields{
//		"action": "start_password_tests",
//		"count":  len(passwords),
//		"target": tasks[0].URL,
//	}).Info("Starting password tests")
//
//	var results []Result
//	for _, pass := range passwords {
//		select {
//		case <-ctx.Done():
//			return results
//		default:
//			task := Task{
//				URL:      tasks[0].URL,
//				Username: tasks[0].Username,
//				Password: pass,
//			}
//
//			log.WithFields(log.Fields{
//				"username": task.Username,
//				"password": task.Password,
//				"status":   "attempt",
//			}).Debug("Testing credentials")
//
//			result := c.processTask(task)
//			results = append(results, result)
//
//			if result.Success {
//				log.WithFields(log.Fields{
//					"url":    task.URL,
//					"status": "success",
//				}).Info("Login successful")
//				return results
//			}
//
//			if result.Error != nil && !strings.Contains(result.Error.Error(), "timed out") {
//				log.WithFields(log.Fields{
//					"error": result.Error.Error(),
//				}).Debug("Login attempt failed")
//			}
//
//			time.Sleep(c.delay)
//		}
//	}
//	return results
//}

func (c *Cracker) processTask(ctx context.Context, task Task) Result {
	result := Result{Task: task}

	// 登录上下文
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	doneChan := make(chan bool, 1)

	go func() {
		defer close(errChan)
		defer close(doneChan)

		// 密码处理
		password := ProcessPassword(task.Password, task.Username)

		// 登录网站
		if err := c.browser.Login(ctx, c.selector, task.Username, password); err != nil {
			errChan <- fmt.Errorf("login failed: %w", err)
			return
		}

		// 轮询结果
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				errChan <- fmt.Errorf("login verification timed out")
				return
			case <-ticker.C:
				if c.browser.IsLoggedIn() {
					doneChan <- true
					return
				}
			}
		}
	}()

	select {
	case err := <-errChan:
		result.Error = err
		result.Success = false
		result.Attempts = 1
	case <-doneChan:
		result.Success = true
		result.Attempts = 1
	case <-ctx.Done():
		result.Error = fmt.Errorf("login attempt timed out after 15 seconds")
		result.Success = false
		result.Attempts = 1
	}

	return result
}
