package model

import (
	"testing"

	"github.com/eqs/server/internal/config"
)

func TestAutoMigrate_AllTables(t *testing.T) {
	db := InitTestDB()
	if db == nil {
		t.Fatal("初始化测试库失败")
	}

	// 验证关键表存在
	tables := []string{
		"users", "projects", "supplier_qualifications", "bids", "orders",
		"contracts", "payment_milestones", "deliverables", "project_files",
		"file_annotations", "payment_transactions", "attendance_records",
		"delivery_templates", "contract_templates", "disputes",
		"dispute_evidences", "dispute_expert_assignments", "reviews",
		"messages", "notifications", "audit_logs",
	}

	for _, tbl := range tables {
		if !db.Migrator().HasTable(tbl) {
			t.Errorf("表 %s 未创建", tbl)
		}
	}
}

func TestCreateAndQueryUser(t *testing.T) {
	db := InitTestDB()
	phone := "13000000000"

	user := User{Phone: phone, UserType: 1, Status: 1, CreditScore: 100}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	var fetched User
	if err := db.Where("phone = ?", phone).First(&fetched).Error; err != nil {
		t.Fatalf("查询用户失败: %v", err)
	}
	if fetched.ID != user.ID {
		t.Fatalf("ID不匹配: %d != %d", fetched.ID, user.ID)
	}
}

func TestCreateUser_WxOpenIDOptional(t *testing.T) {
	db := InitTestDB()

	// 两个手机号用户均可不填wx_openid（来源指针为空），不应触发唯一冲突
	for _, phone := range []string{"13800000001", "13800000002"} {
		u := User{Phone: phone, UserType: 2, Status: 1}
		if err := db.Create(&u).Error; err != nil {
			t.Fatalf("创建手机号用户 %s 失败: %v", phone, err)
		}
	}

	// 微信用户唯一（手机号留空走微信登录）
	openid := "openid_x1"
	u1 := User{UserType: 2, Status: 1}
	u1.WxOpenID = &openid
	if err := db.Create(&u1).Error; err != nil {
		t.Fatalf("创建微信用户失败: %v", err)
	}

	// 第二位微信用户不填 openid，验证差异
	openid2 := "openid_x2"
	u2 := User{Phone: "13500000009", UserType: 2, Status: 1}
	u2.WxOpenID = &openid2
	if err := db.Create(&u2).Error; err != nil {
		t.Fatalf("创建第二位微信用户失败: %v", err)
	}

	// 重复openid应冲突
	dup := User{UserType: 2}
	dup.WxOpenID = &openid
	if err := db.Create(&dup).Error; err == nil {
		t.Fatal("重复 openid 应违反唯一索引")
	}
}

func TestCreateOrderWithMilestone(t *testing.T) {
	db := InitTestDB()

	order := Order{ProjectID: 1, SupplierID: 2, Amount: 10000, Status: 0}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("创建订单失败: %v", err)
	}

	ms := PaymentMilestone{OrderID: order.ID, Name: "首付", Sequence: 1, Ratio: 50, Amount: 5000, Status: "pending"}
	if err := db.Create(&ms).Error; err != nil {
		t.Fatalf("创建节点失败: %v", err)
	}

	var count int64
	db.Model(&PaymentMilestone{}).Where("order_id = ?", order.ID).Count(&count)
	if count != 1 {
		t.Fatalf("期望1个节点，得到 %d", count)
	}
}

func TestDeliverableVersionIncrement(t *testing.T) {
	db := InitTestDB()

	order := Order{ProjectID: 1, SupplierID: 2, Amount: 8000, Status: 1}
	db.Create(&order)
	ms := PaymentMilestone{OrderID: order.ID, Name: "设计", Sequence: 1, Ratio: 100, Amount: 8000}
	db.Create(&ms)

	var maxVer int
	db.Model(&Deliverable{}).Where("order_id = ? AND milestone_id = ?", order.ID, ms.ID).
		Select("COALESCE(MAX(version),0)").Scan(&maxVer)
	if maxVer != 0 {
		t.Fatalf("初始版本应为0，得到 %d", maxVer)
	}
}

// 模型创建后再执行一次自迁移，验证幂等性（不重复建表不报错）
func TestAutoMigrate_Idempotent(t *testing.T) {
	db := InitTestDB()
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("重复执行迁移应无错误: %v", err)
	}
}

func TestInitRedis(t *testing.T) {
	cfg := &config.Config{ServerPort: "8080"}
	if err := InitRedis(cfg); err != nil {
		t.Fatalf("初始化 Redis 配置应无错误: %v", err)
	}
	if RDB == nil {
		t.Fatal("RDB 不应为 nil")
	}
}

func TestInitDB_ConnectionFailure(t *testing.T) {
	cfg := &config.Config{
		DBHost: "127.0.0.1", DBPort: "1", // 端口1通常不可达，快速连接拒绝
		DBUser: "root", DBPassword: "", DBName: "eqs",
	}
	if _, err := InitDB(cfg); err == nil {
		t.Fatal("连接失败应返回错误")
	}
}