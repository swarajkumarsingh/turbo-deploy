// Package errorHandler - contains error handling files
package errorHandler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/swarajkumarsingh/turbo-deploy/conf"
	"github.com/swarajkumarsingh/turbo-deploy/constants"
	"github.com/swarajkumarsingh/turbo-deploy/constants/messages"
	"github.com/swarajkumarsingh/turbo-deploy/functions/logger"
	models "github.com/swarajkumarsingh/turbo-deploy/models/error"

	sentry "github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

var log = logger.Log

func stringToStruct(str string) (models.CustomError, error) {
	var s models.CustomError
	err := json.Unmarshal([]byte(str), &s)
	if err != nil {
		return models.CustomError{}, err
	}
	return s, nil
}

func respondWithJSON(c *gin.Context, code int, response map[string]interface{}) {
	c.Set("Content-Type", "application/json")
	c.Status(code)
	c.JSON(code, response)
}

// Recovery handles the panic happening on any function, this is to be called by defer in functions
func Recovery(c *gin.Context, httpStatusCode int) {
	if r := recover(); r != nil {
		msg := messages.SomethingWentWrongMessage
		code := http.StatusInternalServerError

		rawStruct, ok := r.(string)
		definedStruct, _ := stringToStruct(rawStruct)
		
		msg = definedStruct.Message
		code = definedStruct.Status_code

		if ok {
			// string message passed - no need to report to sentry
			CustomError(c, code, msg)
		} else {
			err, ok := r.(error)
			msg = messages.SomethingWentWrongMessage
			if ok {
				// if error object found report to sentry
				logger.WithRequest(c).Errorln("recovered: ", r)
				logger.WithRequest(c).Errorln(StackTrace(Wrap(err)))
				CustomErrorSentry(c, httpStatusCode, msg, err, strconv.Itoa(httpStatusCode))
			} else {
				// when string or error cannot be recovered (rare case)
				CustomError(c, http.StatusInternalServerError, messages.SomethingWentWrongMessage)
			}
		}
	}
}

// CustomErrorSentry returns an error message after reporting to sentry (if environment is not local)
func CustomErrorSentry(c *gin.Context, httpStatusCode int, msg string, err error, errorCode string) {
	if conf.ENV != constants.ENV_LOCAL {
		// report to sentry first if environment is prod, uat or dev
		ReportToSentry(c, err)
	}
	CustomError(c, httpStatusCode, msg)
}

// CustomError returns an error message without reporting to sentry
func CustomError(c *gin.Context, httpStatusCode int, msg string) {
	errJSON := map[string]interface{}{
		"error":   true,
		"message": msg,
	}
	respondWithJSON(c, httpStatusCode, errJSON)
}

// CustomErrorJSON returns a JSON without reporting to sentry
func CustomErrorJSON(c *gin.Context, httpStatusCode int, errJSON map[string]interface{}) {
	respondWithJSON(c, httpStatusCode, errJSON)
}

func ReportToSentry(c *gin.Context, err error) {
	hub := sentry.GetHubFromContext(c)
	hub.CaptureException(err)
}

// RecoveryNoResponse handles the panic happening on any function, this is to be called by defer in functions
func RecoveryNoResponse() {
	if r := recover(); r != nil {
		log.Errorln("recovered: ", r)
		_, ok := r.(string)
		if !ok {
			err, ok := r.(error)
			if ok {
				// if error object found report to sentry
				log.Errorln("recovered: ", r)
				log.Errorln(StackTrace(Wrap(err)))
				localHub := sentry.CurrentHub().Clone()
				localHub.CaptureException(err)
			}
		}
	}
}
