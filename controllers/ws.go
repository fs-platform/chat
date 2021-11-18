package controllers

import (
	"chatServer/pkg/websocket"
	"fmt"
	"github.com/astaxie/beego"
	"time"
)

type WsController struct {
	beego.Controller
}

// @Title ws connect
// @Description socket handShake
// @Success 200 {string} get success
// @router / [get]
func (Ws *WsController) Get() {
	con, err := websocket.HanShake(Ws.Ctx.ResponseWriter, Ws.Ctx.Request)
	if err != nil {
		fmt.Println("handShake 握手失败", err.Error())
		return
	}
	user := new(websocket.User)
	user.Id = Ws.Ctx.Request.RemoteAddr
	user.Name = Ws.Ctx.Request.RemoteAddr
	user.Ws = con
	user.UpdateTime = time.Now().Unix()
	websocket.Join(user)
	for {
		_, data, err := con.ReadMessage()
		if err != nil {
			websocket.Exit(user)
			return
		}
		userChat := websocket.UserChat{
			User: user,
			Data: string(data),
		}
		go func() {
			websocket.UserChan <- userChat
		}()
	}
	Ws.ServeJSON()
}
