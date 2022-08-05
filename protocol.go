package ep

import (
	"encoding/json"
	"log"
	"time"
)

type Payload struct {
	Command int
	Data    []byte
	ZData   []byte // 压缩后数据
}

// Req 表示ws消息请求结构
type Req struct {
	Command int                    `json:"command"`
	Uid     string                 `json:"uid"`
	Data    map[string]interface{} `json:"data"`
	Ctime   int64                  `json:"ctime"`
	MsgID   string                 `json:"msg_id"`
	cli     *Client
}

// Resp 表示ws消息响应结构
type Resp struct {
	Command   int                    `json:"command"`
	MessageID string                 `json:"message_id"`
	TrackID   string                 `json:"track_id"`
	Data      map[string]interface{} `json:"data"`
	ClientId  string                 `json:"client_id,omitempty"`
	Ctime     int64                  `json:"ctime"`
}

func (r Resp) Bytes() []byte {
	bs, err := json.Marshal(r)
	if err != nil {
		log.Printf("ws: %s marshal err %s", r, err)
		return nil
	}
	return bs
}

func NewBroadcast(cmd int, data map[string]interface{}) Resp {
	return Resp{Command: cmd, TrackID: StrUUID(), Data: data, Ctime: time.Now().UnixMilli()}
}

func NewResp(msgID string, data map[string]interface{}, clientID string) Resp {
	return Resp{
		MessageID: msgID,
		TrackID:   StrUUID(),
		Data:      data,
		ClientId:  clientID,
		Ctime:     time.Now().UnixMilli(),
	}
}

// WSXChgMessage 表示kafka消息结构
type WSXChgMessage struct {
	MessageType int                    `json:"message_type"`
	Strategy    string                 `json:"strategy,omitempty"`
	To          []string               `json:"to,omitempty"`
	UIDs        []string               `json:"uids,omitempty"` // 兼容字段
	Command     int                    `json:"command,omitempty"`
	Payload     map[string]interface{} `json:"payload,omitempty"` // 兼容字段
	Extra       map[string]interface{} `json:"extra,omitempty"`
	Req         *Req                   `json:"req,omitempty"`
	Resp        *Resp                  `json:"resp,omitempty"`
	MsgID       string                 `json:"msg_id,omitempty"` // 兼容字段
	Key         string                 `json:"-"`                // 分区参考key, 默认为""
	Compressed  bool                   `json:"compressed"`       // 是否压缩过
}

func (m WSXChgMessage) ID() string {
	if m.MessageType == MessageTypeReq && m.Req != nil {
		return m.Req.MsgID
	} else if m.Resp != nil {
		return m.Resp.TrackID
	}
	return ""
}
