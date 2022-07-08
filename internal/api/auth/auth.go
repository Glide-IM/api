package auth

import (
	"fmt"
	comm2 "github.com/glide-im/api/internal/api/comm"
	"github.com/glide-im/api/internal/api/router"
	"github.com/glide-im/api/internal/auth"
	"github.com/glide-im/api/internal/dao/common"
	"github.com/glide-im/api/internal/dao/userdao"
	"github.com/glide-im/api/internal/dao/wrapper/app"
	"github.com/glide-im/api/internal/dao/wrapper/collect"
	"github.com/glide-im/api/internal/im"
	"github.com/glide-im/glide/pkg/messages"
	"math/rand"
	"strconv"
	"time"
)

var avatars = []string{
	"http://dengzii.com/static/a.webp",
	"http://dengzii.com/static/b.webp",
	"http://dengzii.com/static/c.webp",
	"http://dengzii.com/static/d.webp",
	"http://dengzii.com/static/e.webp",
	"http://dengzii.com/static/f.webp",
	"http://dengzii.com/static/g.webp",
	"http://dengzii.com/static/h.webp",
	"http://dengzii.com/static/i.webp",
	"http://dengzii.com/static/j.webp",
	"http://dengzii.com/static/k.webp",
	"http://dengzii.com/static/l.webp",
	"http://dengzii.com/static/m.webp",
	"http://dengzii.com/static/n.webp",
	"http://dengzii.com/static/o.webp",
	"http://dengzii.com/static/p.webp",
	"http://dengzii.com/static/q.webp",
	"http://dengzii.com/static/r.webp",
}

var nicknames = []string{"佐菲", "赛文", "杰克", "艾斯", "泰罗", "雷欧", "阿斯特拉", "艾迪", "迪迦", "杰斯", "奈克斯", "梦比优斯", "盖亚", "戴拿"}

type Interface interface {
	AuthToken(info *route.Context, req *AuthTokenRequest) error
	SignIn(info *route.Context, req *SignInRequest) error
	Logout(info *route.Context) error
	Register(info *route.Context, req *RegisterRequest) error
}

var (
	ErrInvalidToken      = comm2.NewApiBizError(1001, "token is invalid, plz sign in")
	ErrSignInAccountInfo = comm2.NewApiBizError(1002, "check your account and password")
	ErrReplicatedLogin   = comm2.NewApiBizError(1003, "replicated login")
)

var (
	host = []string{
		fmt.Sprintf("ws://%s/ws", "127.0.0.1:8080"),
	}
)

type AuthApi struct {
}

func (*AuthApi) AuthToken(ctx *route.Context, req *AuthTokenRequest) error {

	result, err := auth.Auth(ctx.Uid, ctx.Device, req.Token)
	if err != nil {
		return ErrInvalidToken
	}
	uid, err := strconv.ParseInt(result.Uid, 10, 64)
	resp := AuthResponse{
		Token:   result.Token,
		Uid:     uid,
		Servers: host,
	}
	ctx.Response(messages.NewMessage(ctx.Seq, comm2.ActionSuccess, resp))
	return nil
}

func (*AuthApi) SignIn(ctx *route.Context, request *SignInRequest) error {
	if len(request.Account) == 0 || len(request.Password) == 0 {
		return ErrSignInAccountInfo
	}
	user, err := userdao.Dao.GetUidInfoByLogin(request.Account, request.Password)
	if err != nil || user.Uid == 0 {
		if err == common.ErrNoRecordFound || user.Uid == 0 {
			return ErrSignInAccountInfo
		}
		return comm2.NewDbErr(err)
	}

	token, err := auth.GenerateTokenExpire(user.Uid, request.Device, 24*3)
	if err != nil {
		return comm2.NewDbErr(err)
	}

	tk := AuthResponse{
		Uid:      user.Uid,
		Token:    token,
		Servers:  host,
		NickName: user.Nickname,
	}
	resp := messages.NewMessage(ctx.Seq, comm2.ActionSuccess, tk)

	ctx.Uid = user.Uid
	ctx.Device = request.Device
	ctx.Response(resp)
	return nil
}

func (*AuthApi) GuestRegister(ctx *route.Context, req *GuestRegisterRequest) error {

	avatar := req.Avatar
	nickname := req.Nickname

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	if len(avatar) == 0 {
		avatar = avatars[rnd.Intn(len(avatars))]
	}
	if len(nickname) == 0 {
		nickname = nicknames[rnd.Intn(len(nicknames))]
	}

	account := "guest_" + randomStr(32)

	u := &userdao.User{
		Account:  account,
		Password: "",
		Nickname: nickname,
		Avatar:   avatar,
	}
	err := userdao.UserInfoDao.AddUser(u)
	if err != nil {
		return comm2.NewDbErr(err)
	}

	user, err := userdao.Dao.GetUidInfoByLogin(account, "")
	if err != nil || user.Uid == 0 {
		if err == common.ErrNoRecordFound || user.Uid == 0 {
			return ErrSignInAccountInfo
		}
		return comm2.NewDbErr(err)
	}

	token, err := auth.GenerateTokenExpire(user.Uid, auth.GUEST_DEVICE, 24*7)

	tk := AuthResponse{
		Uid:      user.Uid,
		Token:    token,
		Servers:  host,
		NickName: user.Nickname,
	}
	ctx.ReturnSuccess(&tk)
	return nil
}

func (*AuthApi) GuestRegisterV2(ctx *route.Context, req *GuestRegisterV2Request) error {
	fingerprintId := req.FingerprintId
	var err error
	var isAccount bool

	fmt.Println("ctx.Context.Request.Header", ctx.Context.Request.Header)
	app_id := app.AppDao.GetAppID(ctx.Context.GetHeader("Host-A"))
	if app_id == 0 {
		return comm2.NewApiBizError(4001, "访问异常")
	}

	u := &userdao.User{
		Account:  fingerprintId,
		Password: "",
		Nickname: fingerprintId,
		Avatar:   "",
		Role:     2,
	}

	isAccount, err = userdao.UserInfoDao.AccountExists(fingerprintId)
	if err != nil {
		return comm2.NewDbErr(err)
	}

	if !isAccount {
		err = userdao.UserInfoDao.AddGuestUser(u)
		if err != nil {
			return comm2.NewDbErr(err)
		}
	}

	user, err := userdao.Dao.GetUidInfoByLogin(fingerprintId, "")
	if err != nil || user.Uid == 0 {
		if err == common.ErrNoRecordFound || user.Uid == 0 {
			return ErrSignInAccountInfo
		}
		return comm2.NewDbErr(err)
	}

	collectData := collect.GetUserUa(ctx)
	collectData.AppID = app_id
	collectData.Device = "phone"
	collectData.Origin = req.Origin
	collectData.Uid = user.Uid
	collect.CollectDataDao.UpdateOrCreate(collectData)

	token, err := auth.GenerateTokenExpire(user.Uid, 3, 24*7)

	tk := GuestAuthResponse{
		Uid:      user.Uid,
		Token:    token,
		Servers:  host,
		AppID:    app_id,
		NickName: user.Nickname,
	}
	ctx.ReturnSuccess(&tk)
	return nil
}

func (*AuthApi) Register(ctx *route.Context, req *RegisterRequest) error {

	exists, err := userdao.UserInfoDao.AccountExists(req.Account)
	if err != nil {
		return comm2.NewDbErr(err)
	}
	if exists {
		return comm2.NewApiBizError(1004, "account already exists")
	}

	//rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	u := &userdao.User{
		Account:  req.Account,
		Password: req.Password,
		Nickname: req.Nickname,
		//Avatar:   nil,
	}
	err = userdao.UserInfoDao.AddUser(u)
	if err != nil {
		return comm2.NewDbErr(err)
	}
	ctx.Response(messages.NewMessage(ctx.Seq, comm2.ActionSuccess, ""))
	return err
}

func (a *AuthApi) Logout(ctx *route.Context) error {
	err := userdao.Dao.DelAuthToken(ctx.Uid, ctx.Device)
	if err != nil {
		return comm2.NewDbErr(err)
	}
	ctx.Response(messages.NewMessage(ctx.Seq, comm2.ActionSuccess, ""))
	_ = im.IM.Logout(strconv.FormatInt(ctx.Uid, 10), strconv.FormatInt(ctx.Device, 10))
	return nil
}

func randomStr(n int) string {
	var l = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	length := len(l)
	for i := range b {
		b[i] = l[rand.Intn(length)]
	}
	return string(b)
}
