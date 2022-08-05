package ep

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	ErrClientNil      = errors.New("client=nil")
	ErrClientAttached = errors.New("client already attached")
)

type IHub interface {
	Add(*Client) error
	ReadLoop(int)
	WriteLoop()
	ControlLoop()
}
type Hub struct {
	clients    sync.Map // <clientID, *Client>, 全集
	csUIDs     sync.Map // uid  map[clientID]*Client
	nPoll      int      // 并发度
	up         chan Req
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		nPoll:      80,
		up:         make(chan Req, ClientUpQueueSize),
		unregister: make(chan *Client, HubUnregisterQueueSize),
	}
}

func (e *Hub) Add(cli *Client) error {
	if e == nil {
		return nil
	}
	if cli == nil {
		return ErrClientNil
	}

	if _, ok := e.clients.Load(cli.ID); ok {
		return ErrClientAttached
	}
	e.clients.Store(cli.ID, cli)
	cli.upChan = e.up
	if cli.UID != "" {
		if c, ok := e.csUIDs.Load(cli.UID); ok {
			if cc, ok := c.(*sync.Map); ok {
				cc.Store(cli.ID, cli)
			}
		} else {
			cc := &sync.Map{}
			cc.Store(cli.ID, cli)
			e.csUIDs.Store(cli.UID, cc)
		}
	}
	cli.hub = e
	return nil
}

// ReadLoop receive message from kafka
func (e *Hub) ReadLoop(id int) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("hub.ReadLoop[%d] exit: err=%v", id, err)
		}
	}()

	for {
		select {
		default:
			// read message from kafka
			time.Sleep(1 * time.Second)
			msg := &WSXChgMessage{}
			e.dispatch(msg)
		}
	}

}

// WriteLoop write message to kafka to give upstream services
func (e *Hub) WriteLoop() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("hub.WriteLoop: err %s", err)
		}
	}()
	log.Printf("hub.writeloop start")
	for {
		select {
		case req, ok := <-e.up:
			if ok {
				fmt.Printf("send req to up service %v", req)
			}
		}
	}
	log.Printf("hub.writeloop end")

}

func (e *Hub) ControlLoop() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("hub.ControlLoop_exit: err %s", err)
		}
	}()
	for {
		select {
		case c, ok := <-e.unregister:
			if ok {
				e.Delete(c)
			}
		}
	}
}

func (e *Hub) broadcast(p *Payload, msgID string) {
	total := 0
	e.clients.Range(func(key, value interface{}) bool {
		if c, ok := value.(*Client); ok && !c.Closed.Val() {
			c.Send(p)
			total += 1
		}
		return true
	})
	log.Println("OUT\thub.broadcast: trackid=", msgID, " total=", total)
}

func (e *Hub) send2uid(uids []string, p *Payload, msgid string) {
	for _, uid := range uids {
		if c, ok := e.csUIDs.Load(uid); ok {
			if cc, ok := c.(*sync.Map); ok {
				cc.Range(func(vcid, vcli interface{}) bool {
					if cli, ok := vcli.(*Client); ok && !cli.Closed.Val() {
						cli.Send(p)
					}
					return true
				})
			}
		} else {
			// log.Println("OUT\thub.send2uid:\ttrackid=", msgid, "uid", uid, "not found", "data", string(data))
		}
	}
}

func (e *Hub) send2client(cids []string, p *Payload, msgid string, extra interface{}) {
	for _, cid := range cids {
		if c, ok := e.clients.Load(cid); ok {
			if cc, ok := c.(*Client); ok && !cc.Closed.Val() {
				cc.Send(p)
			}
		} else {
			// log.Printf("hub.send2client(): cid %s not found in hub", cid)
		}
	}
}

func (e *Hub) Delete(cli *Client) error {
	if e == nil {
		return nil
	}
	if cli == nil {
		return ErrClientNil
	}
	e.clients.Delete(cli.ID)
	cli.upChan = nil
	if c, ok := e.csUIDs.Load(cli.UID); ok {
		if cc, ok := c.(*sync.Map); ok {
			cc.Delete(cli.ID)
			l := SyncLen(cc)
			if l == 0 {
				e.csUIDs.Delete(cli.UID)
			}
		}
	}
	cli.hub = nil
	return nil
}

func (e *Hub) dispatch(msg *WSXChgMessage) {
	//log.Println("OUT\thub.dispatch:\t", msg)

	cmd := msg.Command
	rsp := NewBroadcast(msg.Command, msg.Payload)
	data := rsp.Bytes()
	p := &Payload{Command: cmd, Data: data}
	msgID := msg.ID()
	switch msg.Strategy {
	case StrategyAll:
		go e.broadcast(p, msgID)
	case StrategyUID:
		go e.send2uid(msg.To, p, msgID)
	case StrategyClient:
		go e.send2client(msg.To, p, msgID, msg.Resp.Data["extra"])
	}
}

func (e *Hub) Start() {
	for i := 0; i < e.nPoll; i++ {
		go e.ReadLoop(i)
	}

	go e.WriteLoop()
	go e.ControlLoop()
}
