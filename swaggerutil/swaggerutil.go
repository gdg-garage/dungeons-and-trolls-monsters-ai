package swaggerutil

import (
	"net/http"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
	"go.uber.org/zap"
)

func LogResponse(logger *zap.SugaredLogger, err error, httpResp *http.Response, method string, request interface{}) {
	if err != nil {
		LogError(logger, err, httpResp, method, request)
	} else {
		LogSuccess(logger, httpResp, method, request)
	}
}

func LogSuccess(logger *zap.SugaredLogger, httpResp *http.Response, method string, request interface{}) {
	logger.Infow("Successfully sent request to server",
		"statusCode", httpResp.StatusCode,
		"method", method,
		"requestPayload", request,
	)
}

func LogError(logger *zap.SugaredLogger, err error, httpResp *http.Response, method string, request interface{}) {
	loggerWErr := logger.With(
		zap.Error(err),
		"method", method,
		"requestPayload", request,
	)
	if httpResp != nil {
		loggerWErr = loggerWErr.With(
			"statusCode", httpResp.StatusCode,
		)
	}
	swaggerErr, ok := err.(swagger.GenericSwaggerError)
	if ok {
		loggerWErr.Errorw("Server responded with error",
			zap.Any("responseBody", string(swaggerErr.Body())),
		)
		return
	}
	loggerWErr.Errorw("Server responded with error. FAILED TO PARSE SWAGGER ERROR!")
}
