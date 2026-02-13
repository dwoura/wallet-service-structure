package validator

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func Init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		validate = v
		// Register custom validations or translations here if needed
		// For now, we use default binding validator from Gin which is also go-playground/validator
	}
}

// GetErrorMsg translates validation errors into user-friendly messages
func GetErrorMsg(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errMsgs []string
		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()
			param := e.Param()

			switch tag {
			case "required":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 不能为空", field))
			case "email":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 格式不正确", field))
			case "min":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 长度至少为 %s", field, param))
			case "max":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 长度不能超过 %s", field, param))
			case "oneof":
				errMsgs = append(errMsgs, fmt.Sprintf("%s 必须是 [%s] 之一", field, param))
			default:
				errMsgs = append(errMsgs, fmt.Sprintf("%s 校验失败 (%s)", field, tag))
			}
		}
		return strings.Join(errMsgs, "; ")
	}
	return "请求参数错误"
}
