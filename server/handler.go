package server

import (
	"github.com/gorilla/websocket"
	"time"
)

func (w *wsConn) heartbeat() {
	timer := time.NewTimer(time.Second * 60)
	for {
		select {
		case <-timer.C:
			if !w.IsAlive() {
				w.Close()
				break
			}
			timer.Reset(time.Second * 60)
		case <-w.closeChan:
			timer.Stop()
			break
		}
	}
}

func (w *wsConn) handlePing() *wsMessage {
	w.KeepAlive()
	return &wsMessage{MsgType: websocket.PongMessage, MsgData: nil}
}

func (w *wsConn) handleTextMsg(message *wsMessage) *wsMessage {
	w.KeepAlive()
	return message
}

func (w *wsConn) Handle() {
	var (
		err     error
		message *wsMessage
		resp    *wsMessage
	)
	go w.heartbeat()
	for {
		if message, err = w.readMsg(); err != nil {
			w.Close()
			break
		}

		// 只处理文本消息
		//if message.MsgType != websocket.TextMessage {
		//	continue
		//}

		switch message.MsgType {
		// 处理ping请求
		case websocket.PingMessage:
			resp = w.handlePing()
		case websocket.TextMessage:
			resp = w.handleTextMsg(message)
		default:
			continue
		}

		if resp != nil {
			// socket缓冲区写满不是致命错误
			if err = w.sendMsg(resp); err != nil {
				w.Close()
				break
			}
		}
	}
	return
}
