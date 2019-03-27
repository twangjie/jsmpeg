package main

import (
	//	"io"
	"net/http"
	//"encoding/json"
	"golang.org/x/net/websocket"
	"flag"
	"fmt"
	"log"
	"encoding/json"
)

var addr = flag.String("addr", ":8083", "http service address")
var root = flag.String("root", ".", "http root path")
var clients = make(map[string]*websocket.Conn) // connected clients
var publishers = make(map[string]string) // connected clients


func main() {
	flag.Parse()
	log.SetFlags(0)

	fmt.Println("begin")
	http.Handle("/", http.FileServer(http.Dir(*root))) // <-- note this line
	http.Handle("/play", websocket.Handler(streamingHandler))

	http.HandleFunc("/stat", statHandler)
	http.HandleFunc("/publish", publishHandler)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}

	fmt.Println("end")

}

func streamingHandler(ws *websocket.Conn) {

	fmt.Printf("Handle url: %s\n", ws.Request().URL.String())

	ws.PayloadType = websocket.BinaryFrame

	var r = ws.Request()
	r.ParseForm()
	var clientId = r.Form.Get("clientId")

	clients[clientId] = ws

	msg := make([]byte, 512)
	n, err := ws.Read(msg)
	if err != nil {
		//log.Fatal(err)
        delete(clients, clientId)
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
	for key, ws := range clients {
		_, err := ws.Write(msg)
		if err != nil {
			//log.Fatal(err)
			delete(clients, key)
		}
	}

}

func publishHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
        
        //url := r.URL.String()

        r.ParseForm()
        var cameraId = r.Form.Get("cameraId")
        
        if _, ok := publishers[cameraId]; ok {  
            return
        }
        
        publishers[cameraId] = cameraId
                
        fmt.Printf("Handle camera: %s\n", cameraId)
        
		request := make([]byte, 1024 * 1024)
		for {
			read_len, err := r.Body.Read(request)
			if (err != nil ) {
                delete(publishers, cameraId)
				fmt.Println(err)
				break
			} else {
				if read_len == 0 {
                    delete(publishers, cameraId)
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

func statHandler(w http.ResponseWriter, r *http.Request) {

    var clientIds []string
    for kc := range clients {
        clientIds = append(clientIds, kc)
    }
	//b, _ := json.Marshal(clientIds)
    //clientsJson := string(b)

    var pIds []string
    for kp := range publishers {
        pIds = append(pIds, kp)
    }
    
    var stats = make(map[string][]string)
    stats["clients"] = clientIds 
    stats["publishers"] = pIds
    
    statJson, _ := json.Marshal(stats)
    
	w.Header().Add("Content-Type", "application/json")

	w.Write(statJson)
}