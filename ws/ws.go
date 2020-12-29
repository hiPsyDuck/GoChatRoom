package ws

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

//定义一个websocket连接对象，连接中包含每个连接的信息
type User struct {
	conn *websocket.Conn
	msg  chan []byte
}

//定义一个websocket处理器，用于收集消息和广播消息
type Hub struct {
	//用户列表,保存所有用户
	userList map[*User]bool
	//注册chan,用户注册时添加到chan中
	register chan *User
	//注销chan,用户退出时添加到chan中,再从map中删除
	unregister chan *User
	//广播消息,将消息广播给所有连接
	broadcast chan []byte
}

//定义一个升级器，将普通的http连接升级为websocket连接
var upgrade = websocket.Upgrader{
	//定义读写缓冲区大小
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//初始化处理中心，以便调用
var InitHub = &Hub{
	userList:   make(map[*User]bool),
	register:   make(chan *User),
	unregister: make(chan *User),
	broadcast:  make(chan []byte),
}

//处理请求
func Handle(w http.ResponseWriter, r *http.Request) {
	//通过升级器得到Websocket链接
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("获取连接失败:", err)
	}
	//连接成功后注册用户
	user := &User{
		conn: conn,
		msg:  make(chan []byte),
	}
	InitHub.register <- user
	defer func() {
		InitHub.unregister <- user
	}()
	//得到连接后，就可以开始读写数据了
	go read(user)
	write(user)
}

func read(user *User) {
	//从连接中循环读取信息
	for {
		_, msg, err := user.conn.ReadMessage()
		if err != nil {
			fmt.Println("用户退出:", user.conn.RemoteAddr().String())
			InitHub.unregister <- user
			break
		}
		//将读取到的信息传入websocket处理器中的broadcast中，
		InitHub.broadcast <- msg
	}
}

func write(user *User) {
	for data := range user.msg {
		err := user.conn.WriteMessage(1, data)
		//fmt.Printf("%s sent: %s\n", user.conn.RemoteAddr(), string(data))
		if err != nil {
			fmt.Println("写入失败")
			break
		}
	}
}

//处理中心处理获取到的信息
func (h *Hub) Run() {
	for {
		select {
		case user := <-h.register:
			h.userList[user] = true
		case user := <-h.unregister:
			//从注销列表中取数据，判断用户列表中是否存在这个用户，存在就删掉
			if _, ok := h.userList[user]; ok {
				delete(h.userList, user)
			}
		case data := <-h.broadcast:
			//从广播chan中取消息，然后遍历给每个用户，发送到用户的msg中
			for u := range h.userList {
				select {
				case u.msg <- data:
				default:
					delete(h.userList, u)
					close(u.msg)
				}
			}
		}
	}
}
