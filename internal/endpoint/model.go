package endpoints

// Request structures for JSON body binding
type DepositRequest struct {
	Amount float64 `json:"amount" binding:"required"`
}

type WithdrawRequest struct {
	Amount float64 `json:"amount" binding:"required"`
}

type TransferRequest struct {
	FromUserID int     `json:"from_user_id" binding:"required"`
	ToUserID   int     `json:"to_user_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required"`
}
