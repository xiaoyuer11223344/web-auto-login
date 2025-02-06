package crack

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"

	"zp-weblogin/pkg/browser"
)

type Task struct {
	URL      string
	Username string
	Password string
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

func (c *Cracker) Start(ctx context.Context, tasks []Task) []Result {
	results := make([]Result, 0)
	taskChan := make(chan Task, len(tasks))
	resultChan := make(chan Result, len(tasks))

	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < c.threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				select {
				case <-ctx.Done():
					return
				default:
					result := c.processTask(task)
					resultChan <- result
					time.Sleep(c.delay)
				}
			}
		}()
	}

	// Send tasks
	go func() {
		for _, task := range tasks {
			taskChan <- task
		}
		close(taskChan)
	}()

	// Wait for completion and collect results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for result := range resultChan {
		results = append(results, result)
		if result.Success {
			logrus.WithFields(logrus.Fields{
				"url":      result.Task.URL,
				"username": result.Task.Username,
				"attempts": result.Attempts,
			}).Info("Login successful")
		}
		if c.maxCrackNum > 0 && len(results) >= c.maxCrackNum {
			break
		}
	}

	// Log summary
	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}
	logrus.WithFields(logrus.Fields{
		"total":    len(results),
		"success":  successCount,
		"failures": len(results) - successCount,
	}).Info("Cracking completed")

	return results
}

func (c *Cracker) processTask(task Task) Result {
	result := Result{Task: task}

	for attempt := 1; attempt <= c.maxAttempts; attempt++ {
		err := c.browser.Navigate(task.URL)
		if err != nil {
			result.Error = fmt.Errorf("navigation failed: %v", err)
			continue
		}

		password := ProcessPassword(task.Password, task.Username)
		err = c.browser.Login(c.selector, task.Username, password)
		if err != nil {
			result.Error = fmt.Errorf("login failed: %v", err)
			continue
		}

		result.Success = true
		result.Attempts = attempt
		break
	}

	return result
}
