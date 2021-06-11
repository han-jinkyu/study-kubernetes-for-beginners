package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

var version = os.Getenv("version")
var connection int32

func main() {
	log.Printf("%s / starting process on %v", version, os.Getpid())

	var status int

	if version == "v1" {
		status = 201
	} else if version == "v2" {
		status = 202
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		log.Println(version, req.URL.Path)
		defer func() {
			atomic.AddInt32(&connection, -1)
		}()
		atomic.AddInt32(&connection, 1)

		// /sleep/N 요청에는 N초간 슬립 모드
		if strings.HasPrefix(req.URL.Path, "/sleep") {
			id := strings.TrimPrefix(req.URL.Path, "/sleep")
			i, _ := strconv.Atoi(id)
			time.Sleep(time.Second * time.Duration(i))
		}
		w.WriteHeader(status)
	})

	// SIGTERM, SIGINT 무시
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for {
			sig := <-signalChannel
			log.Println("received ", sig)
		}
	}()

	// 매 초마다 연결 상태를 출력
	go func() {
		for {
			log.Println(version, "/ connection", atomic.LoadInt32(&connection))
			time.Sleep(time.Second)
		}
	}()

	http.ListenAndServe(":5000", nil)
}
