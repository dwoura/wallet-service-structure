package handler

import (
	"wallet-core/internal/handler/request"
	"wallet-core/internal/handler/response"
	"wallet-core/pkg/errno"
	"wallet-core/pkg/monitor"
	"wallet-core/pkg/validator"

	"github.com/gin-gonic/gin"
)

// Register 用户注册接口
// @Summary 用户注册
// @Description 注册新用户，演示输入参数校验
// @Tags User
// @Accept json
// @Produce json
// @Param request body request.RegisterRequest true "注册参数"
// @Success 200 {object} response.Response
// @Router /api/v1/register [post]
func Register(c *gin.Context) {
	var req request.RegisterRequest

	// 1. Bind & Validate
	// ShouldBindJSON 会自动根据 struct tag 校验，如果有错误会返回 error
	if err := c.ShouldBindJSON(&req); err != nil {
		// 使用我们封装的 validator 包来翻译错误信息
		errMsg := validator.GetErrorMsg(err)

		// 构造一个 ErrValidation
		// 假设 errno 有一个通用的 ErrValidation (code: 20001, msg: "Validation Error")
		// 或者我们可以临时用 fmt.Errorf, 等 errno.Decode 处理
		// 为了简单，我们直接返回 ParamErr 并带上具体信息
		errResponse := errno.ErrBind.WithMessage(errMsg)
		response.Error(c, errResponse)
		return
	}

	// 2. 模拟业务逻辑
	// 在真实场景中，这里会调用 internal/service/user_service.go
	if req.Username == "admin" {
		response.Error(c, errno.ErrUserAlreadyExist)
		return
	}

	// 5. 返回成功
	// [Metric] 记录注册用户数
	monitor.Business.UserRegisteredTotal.Inc()

	resp := gin.H{
		"user_id":  12345,
		"username": req.Username,
		"email":    req.Email,
	}
	response.Success(c, resp)
}
