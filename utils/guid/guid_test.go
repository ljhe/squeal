package guid

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func resetState() {
	mu.Lock()
	lastTimeStamp = 0
	sequence = 0
	nowUnix = func() int64 { return time.Now().Unix() }
	mu.Unlock()
}

func TestNewGuidLayout(t *testing.T) {
	t.Cleanup(resetState)
	resetState()

	nowUnix = func() int64 { return 1_700_000_000 }
	serverID := 12345

	uid, err := NewGuid(serverID)
	if err != nil {
		t.Fatal(err)
	}

	gotTS := uid >> timestampShift
	gotSeq := int((uid >> sequenceShift) & sequenceMask)
	gotServer := int(uid & serverMask)

	if gotTS != 1_700_000_000 {
		t.Fatalf("timestamp=%d", gotTS)
	}
	if gotSeq != 0 {
		t.Fatalf("sequence=%d", gotSeq)
	}
	if gotServer != serverID {
		t.Fatalf("serverId=%d", gotServer)
	}
}

func TestNewGuidInvalidParam(t *testing.T) {
	t.Cleanup(resetState)
	resetState()

	cases := []int{-1, serverMask + 1, 1 << 20}
	for _, serverID := range cases {
		uid, err := NewGuid(serverID)
		if !errors.Is(err, ErrInvalidParam) {
			t.Fatalf("serverId=%d err=%v", serverID, err)
		}
		if uid != 0 {
			t.Fatalf("uid=%d", uid)
		}
	}
}

func TestNewGuidClockMovedBackwards(t *testing.T) {
	t.Cleanup(resetState)
	resetState()

	nowUnix = func() int64 { return 100 }
	if _, err := NewGuid(1); err != nil {
		t.Fatal(err)
	}

	nowUnix = func() int64 { return 99 }
	uid, err := NewGuid(1)
	if !errors.Is(err, ErrClockMovedBackwards) {
		t.Fatalf("err=%v", err)
	}
	if uid != 0 {
		t.Fatalf("uid=%d", uid)
	}
}

func TestNewGuidSequenceOverflow(t *testing.T) {
	t.Cleanup(resetState)
	resetState()

	nowUnix = func() int64 { return 200 }
	seen := make(map[uint64]struct{}, sequenceMask+1)
	for i := 0; i <= sequenceMask; i++ {
		uid, err := NewGuid(1)
		if err != nil {
			t.Fatalf("i=%d err=%v", i, err)
		}
		if _, ok := seen[uid]; ok {
			t.Fatalf("duplicate %d", uid)
		}
		seen[uid] = struct{}{}
	}

	uid, err := NewGuid(1)
	if !errors.Is(err, ErrSequenceOverflow) {
		t.Fatalf("err=%v", err)
	}
	if uid != 0 {
		t.Fatalf("uid=%d", uid)
	}
}

func TestNewGuidConcurrentUnique(t *testing.T) {
	t.Cleanup(resetState)
	resetState()

	const n = 2000
	var wg sync.WaitGroup
	ch := make(chan uint64, n)
	errCh := make(chan error, n)

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			uid, err := NewGuid(1)
			if err != nil {
				errCh <- err
				return
			}
			ch <- uid
		}()
	}
	wg.Wait()
	close(ch)
	close(errCh)

	for err := range errCh {
		t.Fatal(err)
	}

	seen := make(map[uint64]struct{}, n)
	for uid := range ch {
		if _, ok := seen[uid]; ok {
			t.Fatalf("duplicate %d", uid)
		}
		seen[uid] = struct{}{}
	}
	if len(seen) != n {
		t.Fatalf("got %d unique ids", len(seen))
	}
}

func TestNewGuid(t *testing.T) {
	for {
		guid, _ := NewGuid(1)
		fmt.Println(guid)
		time.Sleep(time.Second)
	}
}
