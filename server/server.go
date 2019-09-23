package server

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有CORS跨域请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// websocket的Message对象
type wsMessage struct {
	MsgType int
	MsgData []byte
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		wsSocket *websocket.Conn
		wsConn   *wsConn
	)
	if wsSocket, err = upgrader.Upgrade(w, r, nil); err != nil {
		panic(err)
	}

	// 开启读写循环
	wsConn = InitWsConn(wsSocket)

	// 开始处理
	wsConn.Handle()
}

func InitServer() (err error) {
	g := gin.New()
	g.Use(gin.Recovery())

	g.GET("/", func(c *gin.Context) {
		wsHandler(c.Writer, c.Request)
	})
	err = http.ListenAndServe(":7777", g)
	return
}
