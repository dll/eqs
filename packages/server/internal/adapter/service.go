package adapter

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/eqs/server/internal/model"
	"gorm.io/gorm"
)

var errNotFound = errors.New("resource not found")

type service struct {
	db  *gorm.DB
	now func() time.Time
}

func (s *service) available() error {
	if s.db == nil {
		return errors.New("database is not configured")
	}
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

func (s *service) projects(page, pageSize int, status string) (pageData[projectItem], error) {
	var projects []model.Project
	query := s.db.Model(&model.Project{})
	if codes := projectStatusCodes(status); status != "" {
		if len(codes) == 0 {
			return pageData[projectItem]{}, fmt.Errorf("unsupported status: %s", status)
		}
		query = query.Where("status IN ?", codes)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return pageData[projectItem]{}, err
	}
	if err := query.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&projects).Error; err != nil {
		return pageData[projectItem]{}, err
	}

	items := make([]projectItem, 0, len(projects))
	for _, project := range projects {
		var orderCount int64
		if err := s.db.Model(&model.Order{}).Where("project_id = ?", project.ID).Count(&orderCount).Error; err != nil {
			return pageData[projectItem]{}, err
		}
		items = append(items, projectItem{
			ID: fmt.Sprintf("proj_%d", project.ID), Title: sanitizeText(project.Title),
			Type: project.ServiceType, Status: projectStatus(project.Status),
			BudgetRange: budgetRange(project.BudgetMin, project.BudgetMax), CreatedAt: project.CreatedAt,
			OwnerID: "user_***", OrderCount: orderCount,
		})
	}
	return pageData[projectItem]{Total: total, Page: page, PageSize: pageSize, Items: items}, nil
}

func (s *service) projectSummary(id uint) (projectSummary, error) {
	var project model.Project
	if err := s.db.First(&project, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return projectSummary{}, errNotFound
		}
		return projectSummary{}, err
	}
	var orders []model.Order
	if err := s.db.Where("project_id = ?", id).Order("id ASC").Find(&orders).Error; err != nil {
		return projectSummary{}, err
	}
	orderItems := make([]orderItem, 0, len(orders))
	for _, order := range orders {
		var supplier model.User
		provider := ""
		if order.SupplierID != 0 {
			if err := s.db.Select("company_name").First(&supplier, order.SupplierID).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return projectSummary{}, err
			}
			provider = supplier.CompanyName
		}
		orderItems = append(orderItems, orderItem{
			ID: fmt.Sprintf("order_%d", order.ID), Status: orderStatus(order.Status),
			AmountRange: moneyRange(order.Amount), ServiceProvider: maskProvider(provider), CompletedAt: order.CompletedAt,
		})
	}
	var attachments int64
	if err := s.db.Model(&model.ProjectFile{}).Where("project_id = ?", id).Count(&attachments).Error; err != nil {
		return projectSummary{}, err
	}
	return projectSummary{
		ID: fmt.Sprintf("proj_%d", project.ID), Title: sanitizeText(project.Title), Type: project.ServiceType,
		Status: projectStatus(project.Status), Description: sanitizeText(project.Description),
		BudgetRange: budgetRange(project.BudgetMin, project.BudgetMax), Address: maskAddress(project.Address),
		CreatedAt: project.CreatedAt, Orders: orderItems, AttachmentsCount: attachments,
	}, nil
}

func (s *service) alerts(page, pageSize int, severity string) (pageData[alertItem], error) {
	now := s.now()
	items := make([]alertItem, 0)
	var orders []model.Order
	if err := s.db.Preload("Project").Order("id ASC").Find(&orders).Error; err != nil {
		return pageData[alertItem]{}, err
	}
	for _, order := range orders {
		if order.Project.Deadline != nil && order.Project.Deadline.Before(now) && order.Status != 3 && order.Status != 6 {
			items = append(items, alertItem{ID: fmt.Sprintf("delivery_%d", order.ID), Type: "delivery_delay", Severity: "warning", ProjectID: fmt.Sprintf("proj_%d", order.ProjectID), OrderID: fmt.Sprintf("order_%d", order.ID), Message: "订单交付已延期", CreatedAt: *order.Project.Deadline})
		}
	}
	var payments []model.PaymentTransaction
	if err := s.db.Where("status = ?", 0).Order("id ASC").Find(&payments).Error; err != nil {
		return pageData[alertItem]{}, err
	}
	for _, payment := range payments {
		var order model.Order
		if err := s.db.Select("id", "project_id").First(&order, payment.OrderID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return pageData[alertItem]{}, err
		}
		items = append(items, alertItem{ID: fmt.Sprintf("payment_%d", payment.ID), Type: "payment_pending", Severity: "warning", ProjectID: fmt.Sprintf("proj_%d", order.ProjectID), OrderID: fmt.Sprintf("order_%d", order.ID), Message: "付款待处理", CreatedAt: payment.CreatedAt})
	}
	var qualifications []model.SupplierQualification
	if err := s.db.Where("valid_to IS NOT NULL AND valid_to >= ? AND valid_to <= ?", now, now.Add(30*24*time.Hour)).Order("id ASC").Find(&qualifications).Error; err != nil {
		return pageData[alertItem]{}, err
	}
	for _, qualification := range qualifications {
		items = append(items, alertItem{ID: fmt.Sprintf("qualification_%d", qualification.ID), Type: "qualification_expire", Severity: "info", ProjectID: ProjectID, Message: "服务方资质即将过期", CreatedAt: *qualification.ValidTo})
	}
	var disputes []model.Dispute
	if err := s.db.Where("status <> ?", "closed").Order("id ASC").Find(&disputes).Error; err != nil {
		return pageData[alertItem]{}, err
	}
	for _, dispute := range disputes {
		var order model.Order
		if err := s.db.Select("id", "project_id").First(&order, dispute.OrderID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return pageData[alertItem]{}, err
		}
		items = append(items, alertItem{ID: fmt.Sprintf("dispute_%d", dispute.ID), Type: "dispute_raised", Severity: "error", ProjectID: fmt.Sprintf("proj_%d", order.ProjectID), OrderID: fmt.Sprintf("order_%d", order.ID), Message: "订单产生纠纷", CreatedAt: dispute.CreatedAt})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	if severity != "" {
		filtered := items[:0]
		for _, item := range items {
			if item.Severity == severity {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	total := len(items)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return pageData[alertItem]{Total: int64(total), Page: page, PageSize: pageSize, Items: items[start:end]}, nil
}

func (s *service) eqsVersion() (string, error) {
	var version model.SystemVersion
	err := s.db.Order("released_at DESC, id DESC").First(&version).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "0.2.0", nil
	}
	if err != nil {
		return "", err
	}
	return version.Version, nil
}

func projectStatus(status int) string {
	switch status {
	case 4:
		return "completed"
	case 5, 6, 7:
		return "cancelled"
	default:
		return "active"
	}
}

func projectStatusCodes(status string) []int {
	switch status {
	case "active":
		return []int{0, 1, 2, 3}
	case "completed":
		return []int{4}
	case "cancelled":
		return []int{5, 6, 7}
	default:
		return nil
	}
}

func orderStatus(status int) string {
	switch status {
	case 0:
		return "pending_contract"
	case 1:
		return "in_progress"
	case 2:
		return "pending_acceptance"
	case 3:
		return "completed"
	case 4:
		return "disputed"
	case 5:
		return "offline"
	case 6:
		return "cancelled"
	default:
		return "unknown"
	}
}
