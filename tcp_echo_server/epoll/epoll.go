// epoll周りの処理をまとめた感じ
package main

import (
	"syscall"
	"fmt"
)

const (
	EPOLLET        = 1 << 31
	MaxEpollEvents = 32
)

type epoll struct {
	fd int
}

func initEpoll() (epoll, error) {
	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		return epoll{}, err
	}
	return epoll{fd: epfd}, nil
}

func (ep *epoll) close() {
	syscall.Close(ep.fd)
}

func (ep *epoll) wait() ([]syscall.EpollEvent, error) {
	var events [MaxEpollEvents]syscall.EpollEvent
	nevents, err := syscall.EpollWait(ep.fd, events[:], -1)
	if err != nil {
		return []syscall.EpollEvent{}, err
	}

	return events[:nevents], nil
}

func (ep *epoll) add(fd int, eventOperations uint32, edgeMode bool) error {
	fmt.Println("epoll add:", fd)
	var event syscall.EpollEvent
	event.Events = eventOperations
	if edgeMode {
		event.Events |= EPOLLET
	}
	event.Fd = int32(fd)
	if err := syscall.EpollCtl(ep.fd, syscall.EPOLL_CTL_ADD, fd, &event); err != nil {
		return err
	}
	return nil
}
