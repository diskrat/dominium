package p2p

import (
	"encoding/json"
	"fmt"
)

type MessageType string

const (
	MsgBlockNew      MessageType = "BLOCK_NEW"
	MsgBlockRequest  MessageType = "BLOCK_REQUEST"
	MsgBlockResponse MessageType = "BLOCK_RESPONSE"
	MsgTxNew         MessageType = "TX_NEW"
	MsgMinerRegister MessageType = "MINER_REGISTER"
	MsgMinerList     MessageType = "MINER_LIST"
	MsgChainSync     MessageType = "CHAIN_SYNC"
	MsgChainResponse MessageType = "CHAIN_RESPONSE"
	MsgPing          MessageType = "PING"
	MsgPong          MessageType = "PONG"
)

type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
	Sender  string      `json:"sender"`
}

type PeerInfo struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Port    int    `json:"port"`
	IsMiner bool   `json:"is_miner"`
}

type BlockMessage struct {
	BlockJSON string `json:"block_json"`
}

type BlockRequestMessage struct {
	Index int `json:"index"`
}

type ChainSyncMessage struct {
	FromIndex int `json:"from_index"`
}

type MinerRegisterMessage struct {
	PeerInfo PeerInfo `json:"peer_info"`
}

type TxMessage struct {
	TxJSON string `json:"tx_json"`
}

func NewMessage(msgType MessageType, payload interface{}, sender string) *Message {
	return &Message{
		Type:    msgType,
		Payload: payload,
		Sender:  sender,
	}
}

func (m *Message) ToJSON() (string, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}
	return string(data), nil
}

func MessageFromJSON(data string) (*Message, error) {
	var msg Message
	if err := json.Unmarshal([]byte(data), &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}
	return &msg, nil
}

func (m *Message) GetPayloadJSON() (string, error) {
	data, err := json.Marshal(m.Payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}
	return string(data), nil
}
