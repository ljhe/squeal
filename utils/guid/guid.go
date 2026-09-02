package guid

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// 64-bit layout:
// [63:32] unix seconds (32)
// [31:16] sequence     (16)
// [15:0]  serverId     (16)
const (
	timestampBits = 32
	sequenceBits  = 16
	serverBits    = 16

	sequenceShift  = serverBits                   // 16
	timestampShift = sequenceShift + sequenceBits // 32

	sequenceMask = (1 << sequenceBits) - 1 // 0xffff, 0-65535
	serverMask   = (1 << serverBits) - 1   // 0xffff, 0-65535
)

var (
	ErrInvalidParam        = errors.New("guid: invalid serverId")
	ErrSequenceOverflow    = errors.New("guid: sequence overflow in current second")
	ErrClockMovedBackwards = errors.New("guid: clock moved backwards")
)

var (
	mu            sync.Mutex
	lastTimeStamp uint64
	sequence      int
	nowUnix       = func() int64 { return time.Now().Unix() }
)

func NewGuid(serverId int) (uint64, error) {
	if serverId < 0 || serverId > serverMask {
		return 0, fmt.Errorf("%w: serverId=%d[0-%d]", ErrInvalidParam, serverId, serverMask)
	}

	mu.Lock()
	defer mu.Unlock()

	current := uint64(nowUnix())
	if lastTimeStamp != 0 && current < lastTimeStamp {
		return 0, fmt.Errorf("%w: current=%d last=%d", ErrClockMovedBackwards, current, lastTimeStamp)
	}

	if current == lastTimeStamp {
		if sequence >= sequenceMask {
			return 0, ErrSequenceOverflow
		}
		sequence++
	} else {
		lastTimeStamp = current
		sequence = 0
	}

	uid := lastTimeStamp << timestampShift
	uid |= uint64(sequence) << sequenceShift
	uid |= uint64(serverId)
	return uid, nil
}
