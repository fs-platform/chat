package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

// User 单个User
type User struct {
	Id         string
	Name       string
	Ws         *websocket.Conn
	UpdateTime int64
}
type UserListMap map[string]*User

var lock = &sync.Mutex{}

// UserList User列表
var UserList = make(UserListMap)

//公用广播信息
var publish = make(chan string)

// UserChat 客户聊天通道
type UserChat struct {
	User *User
	Data string
}

var UserChan = make(chan UserChat)

// HanShake  握手
func HanShake(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	con, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		return nil, err
	}
	return con, nil
}

// BroadCast 广播信息,通知有客户加入
func BroadCast(message string) {
	publish <- message
}

// Join 加入聊天
func Join(user *User) {
	lock.Lock()
	defer lock.Unlock()
	UserList[user.Id] = user
	go BroadCast("有客户加入:" + user.Id)
}
// Exit 退出聊天
func Exit(user *User) {
	defer func() {
		_ = user.Ws.Close()
	}()
	if _, ok := UserList[user.Id]; ok {
		lock.Lock()
		delete(UserList, user.Id)
		lock.Unlock()
		go BroadCast("有客户退出:" + user.Id)
	}
}
// Listen 监听通道
func Listen() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		//监听广播通知,给所有的客户端发送消息
		case data := <-publish:
			send(UserList, data)
		case data := <-UserChan:
			copyData := make(UserListMap)
			for k, v := range UserList {
				if k == data.User.Id {
					continue
				}
				copyData[k] = v
			}
			send(copyData, data.Data)
		//如果客户端30s内,没有ping 则直接登出当前客户
		case <-ticker.C:
			stamps := time.Now().Unix()
			for _, val := range UserList {
				if stamps-val.UpdateTime > 30 {
					fmt.Println("exit")
					Exit(val)
				}
			}
		}
	}
}

//给客户发送消息
func send(userList UserListMap, data string) {
	if num := len(userList); num > 0 {
		for _, val := range userList {
			err := val.Ws.WriteMessage(1, []byte(data))
			if err != nil {
				fmt.Printf("%s:全局通知失败%s\n", val.Name, err.Error())
				Exit(val)
			}
		}
	}
}
//初始化通道监听
func init() {
	go Listen()
}
