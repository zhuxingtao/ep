package ep

import "time"

const (
	ClientWriteQueueSize   = 100
	MaxMessageSize         = 4096 // 4kb
	PongWait               = 60 * time.Second
	PingPeriod             = (PongWait * 9) / 10
	WriteWait              = 10 * time.Second
	ClientUpQueueSize      = 10000
	HubUnregisterQueueSize = 1000

	StrategyAll     = "all"
	StrategyPartial = "partial"
	StrategyUID     = "uid"
	StrategyClient  = "client"

	MessageTypeResp = 0
	MessageTypeReq  = 1
)
