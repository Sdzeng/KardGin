package model

type EssayModel struct {
	BaseModel   `json:"-"`
	UserName    string `gorm:"column:user_name" json:"user_name"`
	Pass        string `json:"-"`
	Phone       string `json:"phone"`
	RealName    string `gorm:"column:real_name" json:"real_name"`
	Status      int    `json:"status"`
	Token       string `json:"token"`
	LastLoginIp string `gorm:"column:last_login_ip" json:"last_login_ip"`
}
