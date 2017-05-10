// ssmmu is package for ShadowSocks manage tool via UDP or unix
package ssmmu

import (
	"errors"
	"fmt"
	"net"
	"runtime"
	"time"
)

// SSMMU struct presents the ShadowSocks server
type SSMMU struct {
	MAType  string // manage address type, should be `udp` or `unix`
	MAAddr  string
	Timeout time.Duration // conn/read/write timeout
	conn    net.Conn
}

var statData chan []byte

func NewSSMMU(MAType, MAAddr string, Timeouts ...time.Duration) *SSMMU {
	Timeouts = append(Timeouts, 1500*time.Millisecond)
	statData = make(chan []byte, 1024*1024)
	return &SSMMU{
		MAType:  MAType,
		MAAddr:  MAAddr,
		Timeout: Timeouts[0],
	}
}

func (self *SSMMU) dial() (err error) {
	self.conn, err = net.DialTimeout(self.MAType, self.MAAddr, self.Timeout)
	if err == nil {
		runtime.SetFinalizer(self, func(self *SSMMU) { self.conn.Close() })
	}
	return
}

func (self *SSMMU) send(cmd string) (err error) {
	if self.conn == nil {
		err = self.dial()
		if err != nil {
			return
		}
	}

	_, err = self.conn.Write([]byte(cmd))
	return
}

func (self *SSMMU) recv() (rsp []byte, err error) {
	if self.conn == nil {
		err = self.dial()
		if err != nil {
			return
		}
	}

	rsp = make([]byte, 1506)
	n, err := self.conn.Read(rsp)
	return rsp[0:n], err
}

func (self *SSMMU) command(cmd string, shouldRecv ...string) (succ bool, err error) {
	shouldRecv = append(shouldRecv, "ok")

	err = self.send(cmd)
	if err != nil {
		return
	}

	rsp, err := self.recv()
	if err != nil {
		return
	}

	if len(rsp) > 3 && string(rsp)[:5] == "stat:" {
		statData <- rsp
		succ = true
	}

	if string(rsp) == shouldRecv[0] {
		succ = true
	}
	return
}

func (self *SSMMU) Close() error {
	return self.conn.Close()
}

func (self *SSMMU) Add(port int, passwd string) (succ bool, err error) {
	cmd := fmt.Sprintf(`add: {"server_port": %d, "password": "%s"}`, port, passwd)
	succ, err = self.command(cmd)
	return
}

func (self *SSMMU) Remove(port int) (succ bool, err error) {
	cmd := fmt.Sprintf(`remove: {"server_port": %d}`, port)
	succ, err = self.command(cmd)
	return
}

func (self *SSMMU) ping() (succ bool, duration time.Duration, err error) {
	st := time.Now()
	succ, err = self.command("ping")
	duration = time.Since(st)
	return
}

func (self *SSMMU) Stat(timeout time.Duration) (resp []byte, err error) {
	recvC := make(chan bool)
	go func() {
		self.ping()
		resp = <-statData
		recvC <- true
	}()

	select {
	case <-recvC:
	case <-time.After(timeout):
		err = errors.New("Stat timeout")
	}

	return
}
