package model

// AllModels 返回所有需要迁移的数据库模型对象
// 新增表时，只需要在这里添加即可，不需要修改 main.go
func AllModels() []interface{} {
	return []interface{}{
		&User{},
		&Account{},
		&Address{},
		&Deposit{},
		&Withdrawal{},
		&Collection{},
		&OutboxMessage{},
	}
}
