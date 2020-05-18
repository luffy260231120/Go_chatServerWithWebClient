package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

var (
	upgrader = websocket.Upgrader {
		// 读取存储空间大小
		ReadBufferSize:1024,
		// 写入存储空间大小
		WriteBufferSize:1024,
		// 允许跨域
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel

// Define our message object
type Message struct {
	Name     string    `json:"name"`
	Time     time.Time `json:"time"`
	Data     string    `json:"data"`
}


func wsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handler")
	var (
		wbsCon *websocket.Conn
		err error
		data []byte
	)
	// 完成http应答，在httpheader中放下如下参数
	if wbsCon, err = upgrader.Upgrade(w, r, nil);err != nil {
		return // 获取连接失败直接返回
	}

	clients[wbsCon] = true
	//fmt.Printf("wbsCon = %p\r\n", wbsCon)
	for {

		// 1. 通知客户端
		// 只能发送Text, Binary 类型的数据

		if _, data, err = wbsCon.ReadMessage();err != nil {
			goto ERR // 跳转到关闭连接
		}
		fmt.Printf("收到了消息%s\r\n", data)
		msg := Message{wbsCon.RemoteAddr().String(), time.Now(), string(data)}
		fmt.Println(msg)
		broadcast <- msg
	}
ERR:
	// 关闭连接
	fmt.Println(wbsCon.RemoteAddr(), "断开啦")
	wbsCon.Close()
	delete(clients, wbsCon)
}

func listening() {
	// 要一直监听着，所以加个for
	for {
		select {
		case msg := <- broadcast:
			for con := range clients {
				if err := con.WriteJSON(msg); err != nil {
					// 如果出现了连接错误，那么就断开连接
					log.Printf("error: %v", err)
					con.Close()
					delete(clients, con)
				}
			}
		}
	}
}

func main()  {
	// 这个协程一直监听着信息，信息来广发给所有连接
	go listening()
	// 当有请求访问ws时，执行此回调方法
	http.HandleFunc("/ws",wsHandler)
	// 监听127.0.0.1:7777
	err := http.ListenAndServe("0.0.0.0:7777", nil)
	if err != nil {
		log.Fatal("ListenAndServe", err.Error())
	}
}