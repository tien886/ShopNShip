package dto

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=PENDING ASSIGNED IN_TRANSIT DELIVERED CANCELLED"`
}
