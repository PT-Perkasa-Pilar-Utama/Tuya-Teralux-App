package entities

// MQTTUser represents an MQTT client user in the mqtt_users table
// This table is managed by EMQX Auth Service but accessed directly by the backend
type MQTTUser struct {
	ID          int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username    string `gorm:"type:varchar(255);uniqueIndex;not null" json:"username"`
	Password    string `gorm:"type:varchar(255);not null" json:"password"`
	IsDeleted   bool   `gorm:"column:is_deleted;not_null;default:false" json:"is_deleted"`
	IsSuperuser bool   `gorm:"column:is_superuser;not_null;default:false" json:"is_superuser"`
}

// TableName specifies the table name for GORM
func (MQTTUser) TableName() string {
	return "mqtt_users"
}
