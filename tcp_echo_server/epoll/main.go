package main

import (
	"syscall"
	"fmt"
	"os"
	"net"
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

// イベントハンドラー
// syscall.EpollEventのeventsによって実行する処理を切り替える
// とりあえずechoを動くようにした手抜き実装
func handleConnectedEvent(event syscall.EpollEvent)  {
	fmt.Println("event:", event)
	switch event.Events {
	case syscall.EPOLLIN:
		echo(int(event.Fd))
	case syscall.EPOLLIN | syscall.EPOLLRDHUP:
		fmt.Println("connection close: ", event.Fd)
		syscall.Close(int(event.Fd))
	}
}

// ソケットからreadした内容をそのまま返す
func echo(fd int) {
	var buf [32 * 1024]byte
	nbytes, _ := syscall.Read(fd, buf[:])
	if nbytes > 0 {
		fmt.Printf(">>> %s", buf)
		syscall.Write(fd, buf[:nbytes])
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

	var ep epoll
	ep, err = initEpoll()
	if err != nil {
		exit(err)
	}
	defer ep.close()

	add := ep.add(listenFd, syscall.EPOLLIN,false)
	if err = add; err != nil {
		exit(err)
	}

	var events []syscall.EpollEvent
	for {
		// イベント通知待ち
		events, err = ep.wait()
		if err != nil {
			exit(err)
		}

		for _, event := range events {
			// ファイルディスクリプタが新規接続待ちのやつだったら、新規接続処理
			if int(event.Fd) == listenFd {
				connFd, _, err := syscall.Accept(listenFd)
				if err != nil {
					exit(err)
				}
				defer syscall.Close(connFd)

				// コネクション確立後の通信用ソケットのファイルディスクリプタを epoll に登録する
				if e := ep.add(connFd, syscall.EPOLLIN | syscall.EPOLLRDHUP,true); e != nil {
					exit(err)
				}
			} else {
				// コネクション確率済みの処理は handleConnectedEvent を呼び出す
				handleConnectedEvent(event)
			}
		}
	}
}