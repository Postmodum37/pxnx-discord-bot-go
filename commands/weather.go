package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"pxnx-discord-bot/services"
)

// getWeatherIcon returns an appropriate emoji for the weather condition
func getWeatherIcon(condition string) string {
	lowerCondition := strings.ToLower(condition)
	switch {
	case strings.Contains(lowerCondition, "clear"):
		return "☀️"
	case strings.Contains(lowerCondition, "cloud"):
		return "☁️"
	case strings.Contains(lowerCondition, "rain"):
		return "🌧️"
	case strings.Contains(lowerCondition, "snow"):
		return "❄️"
	case strings.Contains(lowerCondition, "thunder"):
		return "⛈️"
	case strings.Contains(lowerCondition, "mist") || strings.Contains(lowerCondition, "fog"):
		return "🌫️"
	default:
		return "🌤️"
	}
}

// createErrorEmbed creates a standardized error embed
func createErrorEmbed(title, description, errorMsg string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       0xe74c3c, // ColorRed
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Error",
				Value: errorMsg,
			},
		},
	}
}

// HandleWeatherCommand handles the weather slash command
func HandleWeatherCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	options := i.ApplicationCommandData().Options
	city := options[0].StringValue()

	weatherData, err := services.GetWeatherData(city)
	if err != nil {
		// Return error embed if API call fails
		errorEmbed := createErrorEmbed(
			"❌ Weather Error",
			fmt.Sprintf("Unable to fetch weather data for **%s**", city),
			"City not found or API error. Please check the city name and try again.",
		)
		errorEmbed.Footer = &discordgo.MessageEmbedFooter{
			Text: "Powered by OpenWeatherMap",
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{errorEmbed},
			},
		})
	}

	// Format temperature
	temp := fmt.Sprintf("%.1f°C", weatherData.Main.Temp)
	feelsLike := fmt.Sprintf("%.1f°C", weatherData.Main.FeelsLike)

	// Get weather condition and icon
	condition := "Unknown"
	description := "No description available"
	if len(weatherData.Weather) > 0 {
		condition = weatherData.Weather[0].Main
		titleCaser := cases.Title(language.English)
		description = titleCaser.String(weatherData.Weather[0].Description)
	}

	weatherIcon := getWeatherIcon(condition)
	location := weatherData.Name
	if weatherData.Sys.Country != "" {
		location = fmt.Sprintf("%s, %s", weatherData.Name, weatherData.Sys.Country)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Weather in %s", weatherIcon, location),
		Description: description,
		Color:       0x3498db, // ColorBlue
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🌡️ Temperature",
				Value:  temp,
				Inline: true,
			},
			{
				Name:   "🤏 Feels Like",
				Value:  feelsLike,
				Inline: true,
			},
			{
				Name:   "💧 Humidity",
				Value:  fmt.Sprintf("%d%%", weatherData.Main.Humidity),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Powered by OpenWeatherMap",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Add wind information if available
	if weatherData.Wind.Speed > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "💨 Wind Speed",
			Value:  fmt.Sprintf("%.1f m/s", weatherData.Wind.Speed),
			Inline: true,
		})
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}