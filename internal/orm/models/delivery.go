package models

type Delivery struct {
	ID      uint   `json:"id" gorm:"primaryKey;autoIncrement;type:bigint"`
	OrderID uint   `json:"order_id" gorm:"not null;index;type:bigint"`
	Name    string `json:"name" gorm:"not null;size:255"`
	Phone   string `json:"phone" gorm:"not null;size:20"`
	Zip     string `json:"zip" gorm:"not null;size:20"`
	City    string `json:"city" gorm:"not null;size:100"`
	Address string `json:"address" gorm:"not null;size:255"`
	Region  string `json:"region" gorm:"not null;size:100"`
	Email   string `json:"email" gorm:"not null;size:255"`
}

func (Delivery) TableName() string {
	return "deliveries"
}
