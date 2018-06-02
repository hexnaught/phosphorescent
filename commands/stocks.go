package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type avResp struct {
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

func fetchAVData(xFields []string, avAPIKey string) avResp {
	var avData avResp
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=BATCH_STOCK_QUOTES&symbols=%s&apikey=%s", strings.Join(xFields, ","), avAPIKey)

	// Create Request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return avData
	}

	// Create Client
	client := http.Client{}

	// Run request and get response
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return avData
	}

	// Close when done
	defer resp.Body.Close()

	// Sort out the response
	if err := json.NewDecoder(resp.Body).Decode(&avData); err != nil {
		log.Println(err)
	}

	return avData
}

// GetStock ...
func GetStock(symbols []string, avAPIKey string) *discordgo.MessageEmbed {
	var xFields []*discordgo.MessageEmbedField
	var validSymbols []string
	var xData avResp

	if len(symbols) > 0 {
		found := map[string]bool{}
		uniqueSymbols := []string{}

		for j := range symbols {
			if found[symbols[j]] == false {
				found[symbols[j]] = true
				uniqueSymbols = append(uniqueSymbols, symbols[j])
			}
		}

		xData = fetchAVData(uniqueSymbols, avAPIKey)
	} else {
		xData = fetchAVData([]string{"goog", "aapl", "msft", "nvda"}, avAPIKey)
	}

	var xPrice float64
	var xTime time.Time
	for i := 0; i < len(xData.MainData); i++ {
		xTime, _ = time.Parse("2006-01-02 15:04:05", xData.MainData[i].Timestamp)
		xPrice, _ = strconv.ParseFloat(xData.MainData[i].Price, 64)
		xFields = append(xFields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("`%s - %s`", xData.MainData[i].Symbol, xTime.Format("15:04:05")),
			Value:  fmt.Sprintf("Price: %.2f\nVolume: %s", xPrice, xData.MainData[i].Volume),
			Inline: true,
		})
		// Only add/show symbols that are valid/found in the API call
		validSymbols = append(validSymbols, xData.MainData[i].Symbol)
	}

	if len(validSymbols) <= 0 {
		xData = fetchAVData([]string{"goog", "aapl", "msft", "nvda"}, avAPIKey)
	}

	// fmt.Println(validSymbols)

	if len(xFields) == 0 {
		return &discordgo.MessageEmbed{
			Color: 0xF6FF93, // Yellow
			Title: "No Stocks Found...",
		}
	}

	embed := &discordgo.MessageEmbed{
		Color:       0xF6FF93, // Yellow
		Description: fmt.Sprintf("Information for %s", strings.Join(validSymbols, ", ")),
		Fields:      xFields,
		Title:       fmt.Sprintf("Stocks - %s", xTime.Format("02-01-2006")),
	}
	return embed
}
