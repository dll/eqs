package handler

import (
	"fmt"
	"math"
	"strconv"
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
		Longitude:      req.Longitude,
		Latitude:       req.Latitude,
		DistanceMeters: req.DistanceMeters,
		EvidenceFileID: req.EvidenceFileID,
	}
	// P1-09：经纬度加密存储
	lonEnc, _ := model.EncryptField(fmt.Sprintf("%.6f", req.Longitude))
	latEnc, _ := model.EncryptField(fmt.Sprintf("%.6f", req.Latitude))
	record.LongitudeEnc = lonEnc
	record.LatitudeEnc = latEnc
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
	// P1-09：若存在加密经纬度则解密回填
	for i := range records {
		if records[i].LongitudeEnc != "" {
			if v, err := model.DecryptField(records[i].LongitudeEnc); err == nil {
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					records[i].Longitude = f
				}
			}
		}
		if records[i].LatitudeEnc != "" {
			if v, err := model.DecryptField(records[i].LatitudeEnc); err == nil {
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					records[i].Latitude = f
				}
			}
		}
	}
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