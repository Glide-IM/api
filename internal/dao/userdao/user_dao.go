package userdao

import (
	"github.com/glide-im/api/internal/dao/common"
	"github.com/glide-im/api/internal/pkg/db"
	"time"
)

const (
	ContactsTypeUser  = 1
	ContactsTypeGroup = 2
)

const (
	ContactsStatusKeep     = 1 // 正常关系
	ContactsStatusApproval = 2 // 等待同意
	ContactsStatusInvalid  = 3 // 双方不存在好友关系
)

var UserInfoDao = &UserInfoDaoImpl{}

type UserInfoDaoImpl struct{}

func (d *UserInfoDaoImpl) AddUser(u *User) error {
	u.Uid = 0
	u.CreateAt = time.Now().Unix()
	query := db.DB.Create(u)
	return common.ResolveError(query)
}

func (d *UserInfoDaoImpl) DelUser(uid int64) error {
	query := db.DB.Where("uid = ?", uid).Delete(&User{})
	return common.ResolveError(query)
}

func (d *UserInfoDaoImpl) AccountExists(account string) (bool, error) {
	var count int64
	query := db.DB.Model(&User{}).Where("account = ?", account).Count(&count)
	if err := common.ResolveError(query); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *UserInfoDaoImpl) HasUser(uid int64) (bool, error) {
	var count int64
	query := db.DB.Model(&User{}).Where("uid = ?", uid).Count(&count)
	if err := common.ResolveError(query); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *UserInfoDaoImpl) UpdateNickname(uid int64, nickname string) error {
	return d.update(uid, "nickname", nickname)
}

func (d *UserInfoDaoImpl) UpdateAvatar(uid int64, avatar string) error {
	return d.update(uid, "avatar", avatar)
}

func (d *UserInfoDaoImpl) UpdatePassword(uid int64, password string) error {
	return d.update(uid, "password", password)
}

func (d *UserInfoDaoImpl) GetUidInfoByLogin(account string, password string) (int64, error) {
	var uid int64
	query := db.DB.Model(&User{}).
		Where("account = ? AND password = ?", account, password).
		Select("uid").
		Find(&uid)
	if query.Error != nil {
		return 0, query.Error
	}
	if query.RowsAffected == 0 {
		return 0, common.ErrNoRecordFound
	}
	return uid, nil
}

func (d *UserInfoDaoImpl) GetPassword(uid int64) (string, error) {
	var password string
	query := db.DB.Model(&User{}).Where("uid = ?", uid).Select("password").Find(&password)
	if err := common.ResolveError(query); err != nil {
		return "", err
	}
	return password, nil
}

func (d *UserInfoDaoImpl) GetUser(uid int64) (*User, error) {
	user := &User{}
	query := db.DB.Model(user).Where("uid = ?", uid).Find(user)
	if err := common.ResolveError(query); err != nil {
		return nil, err
	}
	return user, nil
}

func (d *UserInfoDaoImpl) GetUserSimpleInfo(uid ...int64) ([]*User, error) {
	var us []*User
	query := db.DB.Model(&User{}).Where("uid IN (?)", uid).Select("uid, account, nickname, avatar").Find(&us)
	if err := common.MustFind(query); err != nil {
		return nil, err
	}
	return us, nil
}

func (d *UserInfoDaoImpl) update(uid int64, field string, value interface{}) error {
	query := db.DB.Model(&User{}).Where("uid = ?", uid).Update(field, value)
	return common.ResolveError(query)
}

func (d *UserInfoDaoImpl) UpdateProfile(uid int64, profile UpdateProfile) error {
	query := db.DB.Model(&User{}).Where("uid = ?", uid).Updates(User{
		Account:  profile.Avatar,
		Nickname: profile.Nickname,
		Password: profile.Password,
	})
	return common.ResolveError(query)
}
