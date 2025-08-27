package models

type OrderItem struct {
	ID          uint   `json:"id" gorm:"primaryKey;autoIncrement;type:bigint"`
	OrderID     uint   `json:"order_id" gorm:"not null;index;type:bigint"`
	ChrtID      int    `json:"chrt_id" gorm:"not null"`
	TrackNumber string `json:"track_number" gorm:"not null;size:100"`
	Price       int    `json:"price" gorm:"not null"`
	Rid         string `json:"rid" gorm:"not null;size:100"`
	Name        string `json:"name" gorm:"not null;size:255"`
	Sale        int    `json:"sale" gorm:"not null"`
	Size        string `json:"size" gorm:"not null;size:20"`
	TotalPrice  int    `json:"total_price" gorm:"not null"`
	NmID        int    `json:"nm_id" gorm:"not null"`
	Brand       string `json:"brand" gorm:"not null;size:100"`
	Status      int    `json:"status" gorm:"not null"`
}

func (OrderItem) TableName() string {
	return "order_items"
}
