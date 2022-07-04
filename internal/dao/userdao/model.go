package userdao

type User struct {
	AppID    int64  `json:"app_id,omitempty"`
	Uid      int64  `gorm:"primaryKey"`
	Account  string `gorm:"unique"`
	Nickname string
	Password string
	Avatar   string
	CreateAt int64
	Role     int
	//CategoryUser []category.CategoryUser `gorm:"foreignKey:Uid;references:Uid"`
}

type Contacts struct {
	Fid     string `gorm:"primaryKey"`
	Uid     int64
	Id      int64
	Remark  string
	Type    int8
	Status  int8
	LastMid int64
}

type LoginState struct {
	Device int64
	Token  string
}
