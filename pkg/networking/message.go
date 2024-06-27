package networking

import (
    "encoding/json"
)

type MessageType int

const (
    PUT MessageType = iota + 650
    GET
    SUCCESS
    FAILURE
)

type Message struct {
    Type          MessageType
    Key           string
    Value         []byte
    Success       bool
    SenderAddress string // Address of the message sender
}

func EncodeMessage(msg Message) ([]byte, error) {
    return json.Marshal(msg)
}

func DecodeMessage(data []byte) (Message, error) {
    var msg Message
    err := json.Unmarshal(data, &msg)
    return msg, err
}
