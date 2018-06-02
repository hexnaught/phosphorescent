package commands

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

type wotUserDetailsResp struct {
	Username string `json:"username"`
	Stats    struct {
		Jobs          string `json:"jobs"`
		Mass          string `json:"mass"`
		DistanceTotal string `json:"totalDistance"`
		DistanceAvg   string `json:"averageDistance"`
		TimeOnDuty    string `json:"timeOnDuty"`
	}
	Achievements []string `json:"achievements"`
	Error        string   `json:"error"`
}

type wotUserSearchResp struct {
	Records []struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Avatar   string `json:"avatar_img"`
	} `json:"records"`
	Error string `json:"error"`
}

// TruckerInfo ...
func TruckerInfo(username string) *discordgo.MessageEmbed {

	var userInfo wotUserDetailsResp

	userInfo = userDetailCallout(userSearchCallout(username))

	if userInfo.Error != "" {
		return &discordgo.MessageEmbed{Title: "No user information found for " + username}
	}

	return &discordgo.MessageEmbed{
		Title: "World Of Trucks information for " + userInfo.Username,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://1.bp.blogspot.com/-Lq52OMRx1NU/UmBZRqKZGII/AAAAAAAAAQE/YvX8iUmXoHs_Nm3iz9O7nvB1raysLby6ACKgB/s1600/wotr_logo.jpg",
		},
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "Total Jobs Taken",
				Value:  userInfo.Stats.Jobs,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Total Mass Hauled",
				Value:  userInfo.Stats.Mass,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Total Distance Driven",
				Value:  userInfo.Stats.DistanceTotal,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Average Job Distance",
				Value:  userInfo.Stats.DistanceAvg,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Time On Duty",
				Value:  userInfo.Stats.TimeOnDuty,
				Inline: true,
			},
		},
		Color: 0xF6FF93, // Yellow
	}
}

func userSearchCallout(username string) wotUserSearchResp {
	var userSearchResults wotUserSearchResp

	// Create Request
	req, err := http.NewRequest("GET", "https://www.worldoftrucks.com/en/ajax/search.php?type=users&text="+username, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return userSearchResults
	}

	// Create Client
	client := http.Client{}

	// Run request and get response
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return userSearchResults
	}

	// Close when done
	defer resp.Body.Close()

	// Sort out the response
	if err := json.NewDecoder(resp.Body).Decode(&userSearchResults); err != nil {
		log.Println(err)
	}

	return userSearchResults
}

func userDetailCallout(userSearchRes wotUserSearchResp) wotUserDetailsResp {
	var userDetailResults wotUserDetailsResp

	// Create Request
	req, err := http.NewRequest("GET", "https://wotapi.thor.re/api/wot/player/"+strconv.Itoa(userSearchRes.Records[0].ID), nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return userDetailResults
	}

	// Create Client
	client := http.Client{}

	// Run request and get response
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return userDetailResults
	}

	// Close when done
	defer resp.Body.Close()

	// Sort out the response
	if err := json.NewDecoder(resp.Body).Decode(&userDetailResults); err != nil {
		log.Println(err)
	}

	return userDetailResults
}
