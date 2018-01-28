package main

import (
	//	"io"
	"net/http"

	"golang.org/x/net/websocket"
	"flag"
	"fmt"
	"log"
)

var addr = flag.String("addr", ":8082", "http service address")
var root = flag.String("root", ".", "http root path")
var clients = make(map[*websocket.Conn]*websocket.Conn) // connected clients

func main() {
	flag.Parse()
	log.SetFlags(0)

	fmt.Println("begin")
	http.Handle("/", http.FileServer(http.Dir(*root))) // <-- note this line
	http.Handle("/wsmpeg1", websocket.Handler(echoHandler))

	http.HandleFunc("/publish", publishHandler)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}

	fmt.Println("end")

}

func echoHandler(ws *websocket.Conn) {

	fmt.Printf("Handle url: %s\n", ws.Request().URL.String())

	ws.PayloadType = websocket.BinaryFrame
	clients[ws] = ws

	msg := make([]byte, 512)
	n, err := ws.Read(msg)
	if err != nil {
		//log.Fatal(err)
		return
	}
	fmt.Printf("Receive: %s\n", msg[:n])

	send_msg := "[" + string(msg[:n]) + "]"
	m, err := ws.Write([]byte(send_msg))
	if err != nil {
		//log.Fatal(err)
		return
	}
	fmt.Printf("Send: %s\n", msg[:m])
}

func write2clients(msg []byte) {
	for _, ws := range clients {
		_, err := ws.Write(msg)
		if err != nil {
			//log.Fatal(err)
			delete(clients, ws)
		}
	}

}

func publishHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		request := make([]byte, 1024 * 1024)
		for {
			read_len, err := r.Body.Read(request)
			if (err != nil ) {
				fmt.Println(err)
				break
			} else {
				if read_len == 0 {
					break
				} else {
					//conn.Write([]byte("OK"))
					//fmt.Printf("body size:%d\n", read_len)
					write2clients(request[:read_len])
				}
			}
		}
	}
}