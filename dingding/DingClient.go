package dingding

import (
	"BitCoin/pkg/settings"
	"errors"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/parnurzeal/gorequest"
	"sync"
	"time"
)

type DingClient struct {
	DingUrl         string
	AppKey          string
	AppSecret       string
	AccessToken     AccessTokenRtn
	AccessTokenMute sync.RWMutex
}

var DClient *DingClient

//初始化
func init() {
	DClient = &DingClient{
		DingUrl:   settings.BitConfig.DingDing.Url,
		AppKey:    settings.BitConfig.DingDing.AppKey,
		AppSecret: settings.BitConfig.DingDing.AppSecret,
	}
	DClient.initAccessToken()
}
func (d *DingClient) initAccessToken() error {
	_, _, errs := gorequest.New().Get(d.DingUrl + "/gettoken?appkey=" +
		d.AppKey + "&appsecret=" + d.AppSecret + "").EndStruct(&d.AccessToken)
	if len(errs) > 0 {
		return errors.New("access get error")
	}
	if d.AccessToken.Errcode != 0 {
		return errors.New(d.AccessToken.Errmsg)
	}
	//设置token过期重新获取时间为5000秒
	d.AccessToken.GetTime = time.Now().Add(5000 * time.Second)
	return nil
}
func (d *DingClient) GetAccessToken() string {
	d.AccessTokenMute.Lock()
	defer d.AccessTokenMute.Unlock()
	//获得成功或者获得的时间比现在的时间早，则重新获取
	if d.AccessToken.GetTime.Before(time.Now()) {
		d.initAccessToken()
	}
	return d.AccessToken.AccessToken
}

func (d *DingClient) SendGroupMessage(message string, chatId string) error {
	groupMessage := GroupMessage{
		Chatid:  chatId,
		Msgtype: "text",
		Text: struct {
			Content string `json:"content"`
		}{message},
	}
	//groupMessageJson, _ := jsoniter.Marshal(groupMessage)
	accesstoken := d.GetAccessToken()

	_, body, errs := gorequest.New().Post(d.DingUrl + "/chat/send?access_token=" +
		accesstoken).Send(groupMessage).EndBytes()
	if len(errs) > 0 {
		return errors.New("access get error")
	}

	if jsoniter.Get(body, "errcode").ToInt32() != 0 {
		fmt.Print(string(body))
		return errors.New(string(body))
	}
	return nil
}
