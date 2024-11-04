package project

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/constants"
	"github.com/swarajkumarsingh/turbo-deploy/constants/messages"
	"github.com/swarajkumarsingh/turbo-deploy/functions/general"
	"github.com/swarajkumarsingh/turbo-deploy/functions/logger"
	validators "github.com/swarajkumarsingh/turbo-deploy/functions/validator"
	model "github.com/swarajkumarsingh/turbo-deploy/models/project"
)

func getUserIdFromReq(ctx *gin.Context) (string, bool) {
	uid, valid := ctx.Get(constants.UserIdMiddlewareConstant)
	if !valid || uid == nil ||  fmt.Sprintf("%v", uid) == "" {
		return "", false
	}

	return fmt.Sprintf("%v", uid), true
}

func getCurrentPageValue(ctx *gin.Context) int {
	val, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil {
		logger.WithRequest(ctx).Errorln("error while extracting current page value: ", err)
		return 1
	}
	return val
}

func getOffsetValue(page int, itemsPerPage int) int {
	return (page - 1) * itemsPerPage
}

func getItemPerPageValue(ctx *gin.Context) int {
	val, err := strconv.Atoi(ctx.DefaultQuery("per_page", strconv.Itoa(constants.DefaultPerPageSize)))
	if err != nil {
		logger.WithRequest(ctx).Errorln("error while extracting item per-page value: ", err)
		return constants.DefaultPerPageSize
	}
	return val
}

func calculateTotalPages(page, itemsPerPage int) int {
	if page <= 0 {
		return 1
	}
	return (page + itemsPerPage - 1) / itemsPerPage
}

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
