package project

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/constants/messages"
	"github.com/swarajkumarsingh/turbo-deploy/functions/general"
	validators "github.com/swarajkumarsingh/turbo-deploy/functions/validator"
	model "github.com/swarajkumarsingh/turbo-deploy/models/project"
)

func IsValidGitHubURL(url string) bool {
	regex := `^https:\/\/github\.com\/[a-zA-Z0-9-]+\/[a-zA-Z0-9._-]+$`
	matched, err := regexp.MatchString(regex, url)
	return err == nil && matched
}

func ValidateGitHubURL(url string, resultChan chan<- bool) {
	defer close(resultChan)
	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		resultChan <- false
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		resultChan <- true
	} else {
		resultChan <- false
	}
}


func IsAccessibleGitHubRepo(url string) (bool, error) {
	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, errors.New("repository not found or is private")
	default:
		return false, fmt.Errorf("unexpected status code from source: %d", resp.StatusCode)
	}
}

func getProjectIdFromParam(ctx *gin.Context) (int, bool) {
	userId := ctx.Param("uid")
	valid := general.SQLInjectionValidation(userId)

	if !valid {
		return 0, false
	}
	uid, err := general.IsInt(userId)
	if err != nil {
		return 0, false
	}

	return uid, true
}

func getCreateProjectBody(ctx *gin.Context) (model.ProjectBody, error) {
	var body model.ProjectBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		return body, errors.New(messages.InvalidBodyMessage)
	}

	if err := validators.ValidateStruct(body); err != nil {
		return body, err
	}
	return body, nil
}
