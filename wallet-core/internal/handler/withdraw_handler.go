package handler

import (
	"wallet-core/internal/handler/request"
	"wallet-core/internal/handler/response"
	"wallet-core/internal/model"
	"wallet-core/internal/service"
	"wallet-core/pkg/errno"

	"github.com/gin-gonic/gin"
)

type WithdrawHandler struct{}

var Withdraw = &WithdrawHandler{}

// CreateWithdrawal 申请提现
// @Summary 申请提现
// @Description 用户发起提现申请
// @Tags Wallet
// @Accept json
// @Produce json
// @Param request body request.CreateWithdrawalRequest true "Withdraw Request"
// @Success 200 {object} response.Response
// @Router /api/v1/withdraw [post]
func (h *WithdrawHandler) CreateWithdrawal(c *gin.Context) {
	// 1. 绑定参数
	var req request.CreateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errno.ErrBind)
		return
	}

	// 2. 获取用户 ID (Mock)
	// userID := c.GetUint64("uid")
	userID := uint64(1) // Mock User 1

	// 3. 构造 Model
	w := &model.Withdrawal{
		UserID:    userID,
		ToAddress: req.ToAddress,
		Amount:    req.Amount,
		Chain:     req.Chain,
	}

	// 4. 调用 Service
	if err := service.Withdraw.CreateWithdrawal(c.Request.Context(), userID, w); err != nil {
		response.Error(c, errno.ErrDatabase) // 简单处理
		return
	}

	response.Success(c, w)
}
