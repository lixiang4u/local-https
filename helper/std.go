package helper

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func ExitMsg(msg string) {
	log.Println(msg)
	_, _ = os.Stdin.Read(make([]byte, 1))
	os.Exit(1)
}

func ExitWithCtrlC() {
	// 创建一个信号通道
	var sigChan = make(chan os.Signal, 1)
	// 监听 SIGINT 信号
	signal.Notify(sigChan, syscall.SIGINT)
	// 阻塞，直到接收到信号
	<-sigChan
	// 停止监听信号
	signal.Stop(sigChan)
	close(sigChan)
	os.Exit(1)
}
