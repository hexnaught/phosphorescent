package main

// AVResp ...
type AVResp struct {
	MetaData struct {
		Information string `json:"1. Information"`
		Notes       string `json:"2. Notes"`
		TimeZone    string `json:"3. Time Zone"`
	}
	MainData []struct {
		Symbol    string `json:"1. symbol"`
		Price     string `json:"2. price"`
		Volume    string `json:"3. volume"`
		Timestamp string `json:"4. timestamp"`
	} `json:"Stock Quotes"`
}
