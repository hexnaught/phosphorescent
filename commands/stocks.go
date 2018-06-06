package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jokerdan/phosphorescent/util"
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
	Error string
}

func fetchAVData(xFields []string, avAPIKey string) avResp {
	var avData avResp

	url := fmt.Sprintf("https://www.alphavantage.co/query?function=BATCH_STOCK_QUOTES&symbols=%s&apikey=%s", strings.Join(xFields, ","), avAPIKey)
	err := util.DoCallout(url, &avData)
	if err != nil {
		avData.Error = "There was an issue with the callout"
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

	if xData.Error != "" {
		return &discordgo.MessageEmbed{
			Color: 0xF6FF93, // Yellow
			Title: xData.Error,
		}
	}

	var xPrice float64
	var xTime time.Time

	for _, data := range xData.MainData {
		xTime, _ = time.Parse("2006-01-02 15:04:05", data.Timestamp)
		xPrice, _ = strconv.ParseFloat(data.Price, 64)
		xFields = append(xFields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("`%s - %s`", data.Symbol, xTime.Format("15:04:05")),
			Value:  fmt.Sprintf("Price: %.2f\nVolume: %s", xPrice, data.Volume),
			Inline: true,
		})
		// Only add/show symbols that are valid/found in the API call
		validSymbols = append(validSymbols, data.Symbol)
	}

	if len(validSymbols) <= 0 {
		xData = fetchAVData([]string{"goog", "aapl", "msft", "nvda"}, avAPIKey)
		if xData.Error != "" {
			return &discordgo.MessageEmbed{
				Color: 0xF6FF93, // Yellow
				Title: xData.Error,
			}
		}
	}

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
