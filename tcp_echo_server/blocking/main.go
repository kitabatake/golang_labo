package main

import (
	"syscall"
	"net"
	"fmt"
	"os"
)

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}

// 接続待ちソケットの準備
// 指定されたIPアドレス、ポート番号のソケットのファイルディスクリプタを返す
func initListenFd(ipAddr string, port int) (int, error){
	listenFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return -1, err
	}

	addr := syscall.SockaddrInet4{Port: port}
	copy(addr.Addr[:], net.ParseIP(ipAddr).To4())

	syscall.Bind(listenFd, &addr)
	syscall.Listen(listenFd, 10)
	return listenFd, nil
}

// コネクションごとにgoroutineで呼び出され、
// ソケットからreadした内容をそのまま返す
func echo(fd int) {
	defer syscall.Close(fd)
	var buf [32 * 1024]byte
	for {
		fmt.Println("waiting Read:", fd)
		nbytes, e := syscall.Read(fd, buf[:]) // blocking!
		if nbytes > 0 {
			fmt.Printf(">>> %s", buf)
			syscall.Write(fd, buf[:nbytes]) // blocking!
			fmt.Printf("<<< %s", buf)
		}
		if e != nil {
			fmt.Println("echo error:", e)
			break
		}
	}
}

func main() {
	var listenFd int
	var err error

	listenFd, err = initListenFd("0.0.0.0", 3000)
	if err != nil {
		exit(err)
	}
	defer syscall.Close(listenFd)

	for {
		fmt.Println("waiting new connection")
		connFd, _, err := syscall.Accept(listenFd) // blocking!
		if err != nil {
			exit(err)
		}
		fmt.Println("connection accepted:", connFd)
		go echo(connFd)
	}
}