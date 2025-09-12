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

	// Default to current weather if no duration specified
	duration := "current"
	if len(options) > 1 && options[1].Name == "duration" {
		duration = options[1].StringValue()
	}

	switch duration {
	case "current":
		return handleCurrentWeather(s, i, city)
	case "1-day":
		return handleForecast(s, i, city, 1)
	case "5-day":
		return handleForecast(s, i, city, 5) // OpenWeatherMap free tier supports up to 5 days
	default:
		return handleCurrentWeather(s, i, city)
	}
}

// handleCurrentWeather handles current weather requests
func handleCurrentWeather(s SessionInterface, i *discordgo.InteractionCreate, city string) error {
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

// handleForecast handles forecast requests (1-day or multi-day)
func handleForecast(s SessionInterface, i *discordgo.InteractionCreate, city string, days int) error {
	forecastData, err := services.GetForecastData(city, days)
	if err != nil {
		// Return error embed if API call fails
		errorEmbed := createErrorEmbed(
			"❌ Forecast Error",
			fmt.Sprintf("Unable to fetch forecast data for **%s**", city),
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

	// Process forecast data into daily summaries
	dailyForecasts := services.ProcessDailyForecasts(forecastData)

	// Limit based on requested days
	if len(dailyForecasts) > days {
		dailyForecasts = dailyForecasts[:days]
	}

	location := forecastData.City.Name
	if forecastData.City.Country != "" {
		location = fmt.Sprintf("%s, %s", forecastData.City.Name, forecastData.City.Country)
	}

	// Create forecast embed
	titleCaser := cases.Title(language.English)
	durationText := "1-Day"
	if days > 1 {
		durationText = fmt.Sprintf("%d-Day", days)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📅 %s Forecast for %s", durationText, location),
		Description: fmt.Sprintf("Weather forecast for the next %d day(s)", len(dailyForecasts)),
		Color:       0x3498db, // ColorBlue
		Fields:      []*discordgo.MessageEmbedField{},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Powered by OpenWeatherMap",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Add daily forecast fields
	for i, daily := range dailyForecasts {
		weatherIcon := getWeatherIcon(daily.Condition)
		dateStr := daily.Date.Format("Mon, Jan 2")
		switch i {
		case 0:
			dateStr = "Today"
		case 1:
			dateStr = "Tomorrow"
		}

		value := fmt.Sprintf("%s %s\n🌡️ %.1f°C - %.1f°C\n💧 %d%% humidity",
			weatherIcon,
			titleCaser.String(daily.Description),
			daily.TempMin,
			daily.TempMax,
			daily.Humidity,
		)

		if daily.WindSpeed > 0 {
			value += fmt.Sprintf("\n💨 %.1f m/s", daily.WindSpeed)
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   dateStr,
			Value:  value,
			Inline: days == 1, // For 1-day forecast, don't inline to show more detail
		})
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}