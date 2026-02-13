package handler

import (
	"strconv"

	"wallet-core/internal/handler/request"
	"wallet-core/internal/handler/response"
	"wallet-core/internal/service"
	"wallet-core/pkg/errno"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct{}

var Admin = &AdminHandler{}

// ReviewWithdrawal 审核提现
// @Summary 审核提现
// @Description 管理员对提现申请进行审批 (Approve/Reject)
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "Withdrawal ID"
// @Param request body request.ReviewWithdrawalRequest true "Review Request"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/withdrawals/{id}/review [post]
func (h *AdminHandler) ReviewWithdrawal(c *gin.Context) {
	// 1. 获取 ID
	idStr := c.Param("id")
	// 校验 ID 格式? 这里直接传给 Service，Service 也没转 int。
	// Service 接收 string，但在 Gorm 查询时用了 "id = ?"，Gorm 会尝试转换。
	// 最好还是转一下。

	// 2. 绑定参数
	var req request.ReviewWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 这里可以直接返回 err，如果 errno.Decode 能处理
		// 或者返回 errno.ErrBind，这样更规范
		response.Error(c, errno.ErrBind)
		return
	}

	// 3. 获取 Admin ID (Mock)
	// 正常应该从 JWT 中获取 c.Get("uid")
	// 为了演示，这里假设 header 中有 x-admin-id，或者先写死。
	// 既然是教学，写个 Mock 获取逻辑。
	adminIDStr := c.GetHeader("X-Admin-ID")
	adminID, _ := strconv.ParseUint(adminIDStr, 10, 64)
	if adminID == 0 {
		adminID = 1 // 默认 Admin 1
	}

	// 4. 调用 Service
	if err := service.Admin.ReviewWithdrawal(c.Request.Context(), idStr, adminID, req.Action, req.Remark); err != nil {
		response.Error(c, err) // 这里应该区分 error 类型，简单处理
		return
	}

	response.Success(c, nil)
}
