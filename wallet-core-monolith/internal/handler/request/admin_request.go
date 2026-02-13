package request

type ReviewWithdrawalRequest struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
	Remark string `json:"remark"`
}
