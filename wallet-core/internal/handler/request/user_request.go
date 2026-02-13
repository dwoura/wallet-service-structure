package request

// RegisterRequest 用户注册请求参数
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=4,max=20"` // 必填，4-20字符
	Password string `json:"password" binding:"required,min=8,max=32"` // 必填，8-32字符
	Email    string `json:"email" binding:"required,email"`           // 必填，且必须是邮箱格式
}
