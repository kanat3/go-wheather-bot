package main

import (
	"bytes"
	"log"

	"encoding/json"
	"fmt"

	"net/http"
	"strconv"
	"strings"

	weather "github.com/briandowns/openweathermap"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const botToken = "YourToken"
const apiKey = "YourApi"

type Location struct {
	Ip       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
	Readme   string `json:"readme"`
}

func ReadJSON(url string) (*Location, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var ipinfo *Location
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	respByte := buf.Bytes()
	if err := json.Unmarshal(respByte, &ipinfo); err != nil {
		return nil, err
	}

	return ipinfo, nil
}

func CoordConverting(loc string) (lat float64, lon float64, err error) {
	str := strings.Split(loc, ",")
	lat, err = strconv.ParseFloat(str[0], 64)
	if err != nil {
		fmt.Println("Координаты не получены")
	}
	lon, err = strconv.ParseFloat(str[1], 64)
	if err != nil {
		fmt.Println("Координаты не получены")
	}
	return lat, lon, nil
}

func GetWeatherInfoByCoords(lat float64, lon float64) (info *weather.CurrentWeatherData, err error) {
	// simple weather get
	fmt.Printf("%f %f", lat, lon)
	info, err = weather.NewCurrent("C", "RU", apiKey)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	info.CurrentByCoordinates(
		&weather.Coordinates{
			Longitude: lon,
			Latitude:  lat,
		},
	)
	return info, err
}

func GetWeatherInfo() (info *weather.CurrentWeatherData, err error) {
	url := "https://ipinfo.io/json"
	user_loc, err := ReadJSON(url)
	if err != nil {
		return nil, err
	}
	// get location infloat type
	var lat, lon float64
	lat, lon, err = CoordConverting(user_loc.Loc)
	if err != nil {
		return nil, err
	}
	// simple weather get
	info, err = weather.NewCurrent("C", "RU", apiKey)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	info.CurrentByCoordinates(
		&weather.Coordinates{
			Longitude: lon,
			Latitude:  lat,
		},
	)
	return info, err
}

func main() {
	// bot
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		//log.Panic(err)
		return
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		btn1 := tgbotapi.KeyboardButton{
			RequestLocation: true,
			Text:            "Отправить своё местоположение",
		}
		btn2 := tgbotapi.KeyboardButton{
			RequestLocation: false,
			Text:            "Что по погоде?",
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{btn1, btn2})
		bot.Send(msg)
		if update.Message.Text == "" && update.Message.Location != nil {
			var weather_info *weather.CurrentWeatherData
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Так, сейчас прогноз будет поточнее...")
			weather_info, err = GetWeatherInfoByCoords(update.Message.Location.Latitude, update.Message.Location.Longitude)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ой, что-то пошло не так.")
				bot.Send(msg)
			} else {
				bot.Send(msg)
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprint("По данным сервиса OpenWeather сейчас ", weather_info.Name,
					": ", weather_info.Weather[0].Description, "\n",
					"Температура: ", weather_info.Main.Temp, "° ", "Ощущается как: ", weather_info.Main.FeelsLike, "°\n"))
				bot.Send(msg)
			}
		}
		if update.Message.Text == btn2.Text {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Так, сейчас посмотрим...")
			bot.Send(msg)
			var weather_info *weather.CurrentWeatherData
			weather_info, err := GetWeatherInfo()
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ой, что-то пошло не так.")
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprint("По данным сервиса OpenWeather в городе ",
					": ", weather_info.Weather[0].Description, "\n",
					"Температура: ", weather_info.Main.Temp, "° ", "Ощущается как: ", weather_info.Main.FeelsLike, "°\n"))
				bot.Send(msg)
			}
		}
	}
}
