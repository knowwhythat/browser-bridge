package native

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"os"
	"sync"
)

// Bridge 管理 Native Messaging 双向通信
type Bridge struct {
	mu      sync.Mutex
	pending map[string]chan *ResponseMsg
}

type RequestMsg struct {
	ID     string                 `json:"id"`
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
}

type ResponseMsg struct {
	ID      string      `json:"id"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func NewBridge() *Bridge {
	return &Bridge{
		pending: make(map[string]chan *ResponseMsg),
	}
}

// Send 向 Extension 发送消息（写入 stdout）
func (b *Bridge) Send(msg *RequestMsg) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// 4 字节 little-endian 长度前缀
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(len(data)))

	if _, err := os.Stdout.Write(buf); err != nil {
		return err
	}
	if _, err := os.Stdout.Write(data); err != nil {
		return err
	}

	log.Printf("[native] sent to extension: action=%s id=%s", msg.Action, msg.ID)
	return nil
}

// AddPending 注册等待响应的 channel
func (b *Bridge) AddPending(id string, ch chan *ResponseMsg) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pending[id] = ch
}

// RemovePending 移除等待响应的 channel
func (b *Bridge) RemovePending(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.pending, id)
}

// PumpStdin 从 Extension 读取消息（stdin），分发到等待的 channel
func (b *Bridge) PumpStdin() {
	for {
		// 读取 4 字节长度
		lenBuf := make([]byte, 4)
		if _, err := io.ReadFull(os.Stdin, lenBuf); err != nil {
			if err == io.EOF {
				log.Println("[native] stdin closed, extension disconnected")
				os.Exit(0)
			}
			log.Printf("[native] read length error: %v", err)
			continue
		}

		msgLen := binary.LittleEndian.Uint32(lenBuf)
		if msgLen == 0 {
			continue
		}

		// 读取消息体
		data := make([]byte, msgLen)
		if _, err := io.ReadFull(os.Stdin, data); err != nil {
			log.Printf("[native] read message error: %v", err)
			continue
		}

		var resp ResponseMsg
		if err := json.Unmarshal(data, &resp); err != nil {
			log.Printf("[native] unmarshal error: %v", err)
			continue
		}

		log.Printf("[native] received from extension: id=%s success=%v", resp.ID, resp.Success)

		// 分发到等待的 channel
		b.mu.Lock()
		if ch, ok := b.pending[resp.ID]; ok {
			ch <- &resp
			delete(b.pending, resp.ID)
		} else {
			log.Printf("[native] no pending request for id=%s", resp.ID)
		}
		b.mu.Unlock()
	}
}

// SendAndWait 发送消息并等待响应
func (b *Bridge) SendAndWait(msg *RequestMsg, timeout <-chan struct{}) (*ResponseMsg, error) {
	ch := make(chan *ResponseMsg, 1)
	b.AddPending(msg.ID, ch)
	defer b.RemovePending(msg.ID)

	if err := b.Send(msg); err != nil {
		return nil, err
	}

	select {
	case resp := <-ch:
		return resp, nil
	case <-timeout:
		return nil, io.ErrNoProgress
	}
}
