package vendor

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

var shipmentServiceURL = os.Getenv("SHIPMENT_SERVICE_URL")

func forwardToShipmentService(method, path string, body interface{}, jwt string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(b)
	}
	url := shipmentServiceURL + path
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", jwt)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{}
	return client.Do(req)
}
