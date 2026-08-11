package handler

import (
	"math"
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

type CheckInRequest struct {
	OrderID        uint    `json:"order_id" binding:"required"`
	Longitude     float64 `json:"longitude"`
	Latitude      float64 `json:"latitude"`
	EvidenceFileID uint    `json:"evidence_file_id"`
	DistanceMeters int    `json:"distance_meters"`
}

// CheckIn 现场打卡：开始履约计时
func CheckIn(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	var order model.Order
	if err := model.DB.First(&order, req.OrderID).Error; err != nil {
		notFound(c, "订单不存在")
		return
	}
	if order.SupplierID != userID {
		forbidden(c, "仅服务方可现场打卡")
		return
	}

	record := model.AttendanceRecord{
		OrderID:        req.OrderID,
		UserID:         userID,
		CheckInAt:      time.Now(),
		Longitude:      model.EncryptedFloat(req.Longitude),
		Latitude:       model.EncryptedFloat(req.Latitude),
		DistanceMeters: req.DistanceMeters,
		EvidenceFileID: req.EvidenceFileID,
	}
	if err := model.DB.Create(&record).Error; err != nil {
		serverError(c, err)
		return
	}

	// 距离校验：超出阈值标记异常，需人工审核
	if req.DistanceMeters > 500 {
		model.DB.Model(&record).Update("verification_status", "exception")
	}

	WriteAudit(c, "attendance.checkin", "order", req.OrderID, gin.H{"record_id": record.ID, "distance_meters": req.DistanceMeters})
	ok(c, gin.H{"attendance": record})
}

// ListAttendance 查询某订单的打卡记录
func ListAttendance(c *gin.Context) {
	orderID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "订单ID无效")
		return
	}

	var records []model.AttendanceRecord
	model.DB.Where("order_id = ?", orderID).Order("check_in_at DESC").Find(&records)
	ok(c, gin.H{"attendance": records, "total_hours": computeTotalHours(records)})
}

// computeTotalHours 按打卡记录粗略估算履约时长（MVP 简化，两个月点近似）
func computeTotalHours(records []model.AttendanceRecord) float64 {
	if len(records) == 0 {
		return 0
	}
	first := records[len(records)-1].CheckInAt
	last := records[0].CheckInAt
	return math.Round(last.Sub(first).Hours()*100) / 100
}