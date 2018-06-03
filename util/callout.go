package util

import (
	"encoding/json"
	"log"
	"net/http"
)

// DoCallout ...
func DoCallout(callURL string, result interface{}) error {

	var err error
	// Create Request
	req, err := http.NewRequest("GET", callURL, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return err
	}

	// Create Client
	client := http.Client{}

	// Run request and get response
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return err
	}

	// Close when done
	defer resp.Body.Close()

	// Sort out the response
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
