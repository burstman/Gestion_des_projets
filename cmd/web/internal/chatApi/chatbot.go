package chatapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Message struct {
	Id      int    `json:"id"`
	Message string `json:"message"`
	Url     string `json:"url"`
}

type SenderReceiver interface {
	SendReceive(id int, message string) (*Message, error)
}

func NewSenderReceive(Url string) SenderReceiver {
	return &Message{Url: Url}
}

func (m *Message) SendReceive(Id int, data string) (*Message, error) {
	m = &Message{
		Id:      Id,
		Message: data,
		Url:     m.Url,
	}
	jsonPayload, err := json.Marshal(m)

	fmt.Println(string(jsonPayload))

	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, m.Url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to send Api chatbot message: %s", resp.Status)
	}
	var receivedMessage Message
	err = json.NewDecoder(resp.Body).Decode(&receivedMessage)

	if err != nil {
		return nil, err
	}
	//fmt.Println(receivedMessage)

	return &receivedMessage, nil
}
