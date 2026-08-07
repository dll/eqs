package handler

import (
	"net/http"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

func Recharge(c *gin.Context) {
	// TODO: Integrate payment gateway
	c.JSON(http.StatusOK, gin.H{"message": "充值成功"})
}

func Withdraw(c *gin.Context) {
	// TODO: Implement withdrawal logic
	c.JSON(http.StatusOK, gin.H{"message": "提现申请已提交"})
}

func GetPayRecords(c *gin.Context) {
	userID := c.GetUint("user_id")

	var records []model.Payout
	model.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&records)

	c.JSON(http.StatusOK, gin.H{"records": records})
}
