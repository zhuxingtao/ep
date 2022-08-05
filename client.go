package ep

import (
	"bytes"
	"encoding/json"
	"ep/hxxp"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

type Client struct {
	Conn     *websocket.Conn
	ID       string
	UID      string
	upChan   chan Req
	Closed   AtomicBool
	downChan chan []byte
	hub      *Hub
}
type IClient interface {
	ReadLoop()
	WriteLoop()
	Send(*Payload)
	Close()
}

func NewClient(conn *websocket.Conn, clientID string) *Client {
	return &Client{
		Conn:     conn,
		ID:       clientID,
		downChan: make(chan []byte, ClientWriteQueueSize),
	}
}

// ReadLoop pumps messages from the websocket connection to the hub.
func (e *Client) ReadLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("c.ReadLoop %v  err=%v", e, r)
		}
	}()
	e.Conn.SetReadLimit(MaxMessageSize)
	_ = e.Conn.SetReadDeadline(time.Now().Add(PongWait))
	e.Conn.SetPongHandler(func(string) error {
		_ = e.Conn.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})
	for {
		messageType, message, err := e.Conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("c.ReadLoop error: %v", err)
			}
			break
		}
		if messageType == websocket.CloseMessage {
			log.Printf("c.ReadLoop: recv active close message. %v", e)
			break
		}
		var req Req
		decoder := json.NewDecoder(bytes.NewReader(message))
		decoder.UseNumber()
		err = decoder.Decode(&req)
		if err != nil {
			log.Printf("c.ReadLoop: unmarshal json %v,%s,%s", e, err, string(message))
			break
		}
		// debug
		log.Printf("req = %+v  cli=%s", req, e)
		switch req.Command {
		case CMDEcho:
			var data []byte
			rsp := NewResp(req.MsgID, req.Data, e.ID)
			data = rsp.Bytes()
			e.downChan <- data
		default:
			e.upChan <- req
		}
	}
}

// WriteLoop pumps messages from the hub to the websocket connection.
func (e *Client) WriteLoop() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		if err := recover(); err != nil {
			log.Printf("c.WriteLoop: recover %s,%s", e, err)
		}
		e.Close()
	}()
	for {
		select {
		case msg, ok := <-e.downChan:
			_ = e.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				_ = e.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			}
			err := e.Conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Printf("c.WriteLoop %v %v", e.ID, err)
			}
		case <-ticker.C:
			_ = e.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := e.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("c.WriteLoop: ticker conn write %s ping err. %s", e.ID, err)
				return
			}
		}
	}
}

func (e *Client) Start() {
	go e.ReadLoop()
	go e.WriteLoop()
	e.downChan <- NewBroadcast(CMDBClientInfo, hxxp.M{"client_id": e.ID}).Bytes()
}

func (e *Client) String() string {
	return fmt.Sprintf("client[id=%s,uid=%s]", e.ID, e.UID)
}

func (e *Client) Close() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("c.close-chan: err", e, err)
		}
	}()
	if e.upChan != nil {
		e.upChan <- Req{
			Command: CMDExit,
			Uid:     e.UID,
			cli:     e,
			MsgID:   StrUUID(),
		}
	}
	e.upChan = nil
	err := e.Conn.Close()
	if err != nil {
		log.Println("close fail ", e, err)
	}
	if e.hub != nil {
		e.hub.unregister <- e
	}
	close(e.downChan)
}

func (e *Client) Send(p *Payload) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("c.send:\t", e, "err=", err, string(p.Data))
		}
	}()
	e.downChan <- p.Data
}
