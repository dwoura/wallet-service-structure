package main

import (
	"fmt"
	"log"
	"time"

	"wallet-core/internal/model"
	"wallet-core/pkg/config"
	"wallet-core/pkg/database"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 这个脚本模拟完整的 MultiSig 业务流程
// 1. User 提现
// 2. Admin1 同意
// 3. Admin2 同意
// 4. Broadcaster 扫描并执行

func main() {
	// 0. Init
	// 强制设置 Config 路径，因为 go run 在 cmd/test 下运行时找不到上级目录的 config.yaml
	// 或者直接硬编码配置用于测试
	config.Global.DB.Host = "localhost"
	config.Global.DB.User = "gorm"
	config.Global.DB.Password = "gorm"
	config.Global.DB.Name = "gorm"
	config.Global.DB.Port = "9920"
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		config.Global.DB.Host, config.Global.DB.User, config.Global.DB.Password, config.Global.DB.Name, config.Global.DB.Port)
	db, err := database.ConnectPostgres(dsn)
	if err != nil {
		log.Fatal(err)
	}

	// 1. 模拟用户提现
	log.Println("=== Step 1: User Request Withdrawal ===")
	withdraw := &model.Withdrawal{
		UserID:            1001,
		ToAddress:         "0xUserAddress...",
		Amount:            decimal.NewFromFloat(1.5),
		Chain:             "ETH",
		Status:            "pending_review",
		RequiredApprovals: 2,
		CurrentApprovals:  0,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	db.Create(withdraw)
	log.Printf("Created Withdrawal ID: %d, Status: %s\n", withdraw.ID, withdraw.Status)

	// 2. Admin 1 Review
	log.Println("\n=== Step 2: Admin 1 Approve ===")
	review(db, withdraw.ID, 1, "approve")

	// 3. Admin 2 Review
	log.Println("\n=== Step 3: Admin 2 Approve (Reach Threshold) ===")
	review(db, withdraw.ID, 2, "approve")

	// 4. 模拟 Broadcaster 轮询
	log.Println("\n=== Step 4: Broadcaster Polling ===")
	var target model.Withdrawal
	db.First(&target, withdraw.ID)
	if target.Status == "pending_broadcast" {
		log.Printf("✅ Withdrawal %d is ready for broadcast!\n", target.ID)

		// 模拟上链
		target.Status = "completed"
		target.TxHash = "0xMockedTxHashOnChain"
		db.Save(&target)
		log.Println("🚀 Broadcast simulated. Status changed to completed.")
	} else {
		log.Printf("❌ Unexpected status: %s\n", target.Status)
	}
}

func review(db *gorm.DB, withdrawID uint64, adminID uint64, action string) {
	err := db.Transaction(func(tx *gorm.DB) error {
		var w model.Withdrawal
		// Lock
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&w, withdrawID).Error; err != nil {
			return err
		}

		// Create Review
		review := model.WithdrawalReview{
			WithdrawalID: withdrawID,
			AdminID:      adminID,
			Status:       action,
			Remark:       "Integration Test",
			CreatedAt:    time.Now(),
		}
		if err := tx.Create(&review).Error; err != nil {
			return fmt.Errorf("Admin %d already reviewed or error: %v", adminID, err)
		}

		// Update Withdrawal
		if action == "approve" {
			w.CurrentApprovals++
			if w.CurrentApprovals >= w.RequiredApprovals {
				w.Status = "pending_broadcast"
			}
		}
		tx.Save(&w)
		log.Printf("Admin %d approved. Current Approvals: %d, Status: %s\n", adminID, w.CurrentApprovals, w.Status)
		return nil
	})

	if err != nil {
		log.Printf("Review failed: %v\n", err)
	}
}
