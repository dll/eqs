package handler

import (
	"net/http"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

type CreateProjectRequest struct {
	ProjectType string  `json:"project_type" binding:"required"`
	Title       string  `json:"title" binding:"required"`
	BudgetMin   float64 `json:"budget_min"`
	BudgetMax   float64 `json:"budget_max"`
}

func CreateProject(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project := model.Project{
		UserID:      userID,
		ProjectType: req.ProjectType,
		Title:       req.Title,
		BudgetMin:   req.BudgetMin,
		BudgetMax:   req.BudgetMax,
		Status:      0,
	}

	if err := model.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"project": project})
}

func ListProjects(c *gin.Context) {
	var projects []model.Project
	model.DB.Preload("User").Find(&projects)
	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

func GetProject(c *gin.Context) {
	id := c.Param("id")

	var project model.Project
	if err := model.DB.Preload("User").First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"project": project})
}

func ApplyProject(c *gin.Context) {
	// TODO: Implement project application logic
	c.JSON(http.StatusOK, gin.H{"message": "报名成功"})
}

func SelectSupplier(c *gin.Context) {
	// TODO: Implement supplier selection logic
	c.JSON(http.StatusOK, gin.H{"message": "选择成功"})
}
