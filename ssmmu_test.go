package ssmmu

import "testing"
import "time"

func setup() (mmu *SSMMU) {
	mmu = NewSSMMU("udp", "192.168.168.2:2043")
	return
}

func TestAddPort(t *testing.T) {
	for i := 0; i < 10; i++ {
		mmu := setup()
		succ, err := mmu.Add(8592+i, "m")
		if err != nil {
			t.Error(err)
		}
		if !succ {
			t.Fatal("add port should be succ")
		}
		time.Sleep(time.Second * 1)
		mmu.Close()
	}
}

func TestStat(t *testing.T) {
	for i := 0; i < 3; i++ {
		mmu := setup()
		data, err := mmu.Stat(15 * time.Second)
		if err != nil {
			t.Error(err)
		}
		t.Log(string(data))
		time.Sleep(time.Second * 1)
		mmu.Close()
	}
}

func TestRemovePort(t *testing.T) {
	mmu := setup()
	succ, err := mmu.Remove(8592)
	if err != nil {
		t.Error(err)
	}
	if !succ {
		t.Fatal("remove port should be succ")
	}
}

func TestPing(t *testing.T) {
	mmu := setup()
	succ, duration, err := mmu.ping()
	if err != nil {
		t.Error(err)
	}
	if !succ {
		t.Fatal("shoud recv pong")
	}
	t.Logf("ping time: %.3fms", duration.Seconds()*1000)
}
