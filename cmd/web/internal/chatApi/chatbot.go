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
	SendReceive(id int, message, url string) error
}

func NewSenderReceive() SenderReceiver {
	return &Message{}
}

func (m *Message) SendReceive(Id int, data, url string) error {
	m = &Message{
		Id:      Id,
		Message: data,
		Url:     url,
	}
	jsonPayload, err := json.Marshal(m)

	fmt.Println(string(jsonPayload))

	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, m.Url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send Api chatbot message: %s", resp.Status)
	}
	var receivedMessage Message
	err = json.NewDecoder(resp.Body).Decode(&receivedMessage)

	if err != nil {
		return err
	}
	fmt.Println(receivedMessage)

	return nil
}
