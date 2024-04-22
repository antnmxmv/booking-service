package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func Test_server(t *testing.T) {
	defer runTestingApp().Stop()
	resp, err := http.Get("http://localhost:8080/readyz")
	if err != nil {
		t.Error("server is not running", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("server is not ready. code: %d", resp.StatusCode)
	}
}

func Test_bookingHandler(t *testing.T) {
	defer runTestingApp().Stop()

	request := `{
		"id": "1",
		"hotel_id": "aa500b05-98b6-4792-8378-9e46c1a1033d",
		"rooms": [
			{
				"type": "eco",
				"count": 1
			}
		],
		"payment_type": "card",
		"payment_details": {
			"card_id": 123
		},
		"start_date": "2024-06-14T14:00:00Z",
		"end_date": "2024-06-14T14:00:00Z"
	}`

	resp, err := http.Post("http://localhost:8080/reservation/", "application/json", strings.NewReader(request))
	if err != nil {
		t.Error("server is not running", err.Error())
		return
	}

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("bad response. code: %d respone: %s", resp.StatusCode, body)
	}
}
