package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"net/http"
	"os"
	"time"
)

var (
	red    = color.Red
	green  = color.Green
	yellow = color.Yellow
	client = &http.Client{}
	azal   = &Azal{url: "https://www.azal.az/az", apiURL: "https://api.azal.travel/azal/searchFlight"}
)

func init() {
	azal.authKey = os.Getenv("TG_AUTH_KEY")
}

func main() {
	bot, err := tgbotapi.NewBotAPI(azal.authKey)
	if err != nil {
		red(err.Error())
	}

	bot.Debug = false
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		green(msg.Text)

		switch msg.Text {
		case "/start":
			msg.Text = "I'm Azal bot. I can search tickets for you."
			bot.Send(msg)
			fallthrough
		case "/hello", "/hi":
			msg.Text = "Hello, " + update.Message.From.FirstName + "!"
			bot.Send(msg)
			continue
		case "/today":
			azal.date = time.Now().Format("2006-01-02")
		case "/tomorrow":
			azal.date = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
		case "/stop":
			msg.Text = "Bye, " + update.Message.From.FirstName + "!"
			bot.Send(msg)
			continue
		default:
			if !CheckDate(msg.Text) {
				msg.Text = "Please enter valid date format: YYYY-MM-DD Ex: " + time.Now().Format("2006-01-02")
				bot.Send(msg)
				continue
			}
			azal.date = msg.Text
		}

		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Please wait..."))
		body := fmt.Sprintf(`{"searchType":"NORMAL","lang":"AZ","departurePort":"GYD","arrivalPort":"NAJ","classType":"ECONOMY","directionType":"OW","adultCount":"1","infantCount":"0","childCount":"0","departureDate":"%s","arrivalDate":"0--","isAZ":true}`, azal.date)

		flight := sendRequest(body)

		msg.Text = flight.beautify()
		if _, err := bot.Send(msg); err != nil {
			red(err.Error())
		}
	}

}

func sendRequest(body string) Flight {
	req, err := http.NewRequest("POST", azal.apiURL, bytes.NewBuffer([]byte(body)))
	if err != nil {
		red("Error: " + err.Error())
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		red("Error: " + err.Error())
		os.Exit(1)
	}

	defer resp.Body.Close()

	var flightResponse Flight
	err = json.NewDecoder(resp.Body).Decode(&flightResponse)

	if flightResponse.Message.Message != "Ok" {
		yellow("I can't found tickets for this time")
		green("but I can found tickets for next time")
	}

	return flightResponse
}

func (f Flight) beautify() string {
	var buffer bytes.Buffer
	if f.Message.Message != "Ok" {
		buffer.WriteString(f.Message.Message + "\n\n")
	}

	for _, flight := range f.Data.FlightsOfDeparture {
		buffer.WriteString("Date: " + flight.FlightDate + "\n")
		buffer.WriteString("Departure: " + flight.FlightDate + "\n")
		if flight.Fare != "" {
			buffer.WriteString("Price: " + flight.Fare + flight.Currency + "\n")
		}
		for _, flightInfo := range flight.FlightInfoDetailList {
			buffer.WriteString("-----------------------------------------------------\n")
			buffer.WriteString("Departure: " + flightInfo.DepartureTime + "\n")
			buffer.WriteString("Arrival: " + flightInfo.ArrivalTime + "\n")
		}

		buffer.WriteString("Buy ticket: " + azal.url + "\n\n")
		buffer.WriteString("================================" + "\n\n")
	}
	return buffer.String()
}

func CheckDate(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false
	}
	return true
}

type Flight struct {
	Status  int     `json:"status"`
	Message Message `json:"message"`
	Data    Data    `json:"data"`
}

type Message struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Data struct {
	FlightsOfDeparture []FlightsOfDeparture `json:"flightsOfDeparture"`
	FlightsOfArrival   interface{}          `json:"flightsOfArrival"`
	DepartureCityName  []string             `json:"departureCityName"`
	ArrivalCityName    []string             `json:"arrivalCityName"`
	UUID               string               `json:"uuid"`
}

type FlightsOfDeparture struct {
	FlightDate           string             `json:"flightDate"`
	Fare                 string             `json:"fare"`
	Currency             string             `json:"currency"`
	FlightInfoDetailList []FlightInfoDetail `json:"flightInfoDetailList"`
}

type FlightInfoDetail struct {
	DepartureTime           string         `json:"departureTime"`
	DepartureDate           string         `json:"departureDate"`
	ArrivalTime             string         `json:"arrivalTime"`
	ArrivalDate             string         `json:"arrivalDate"`
	DepartureTimeUTC        string         `json:"departureTimeUTC"`
	DepartureDateUTC        string         `json:"departureDateUTC"`
	DepartureStation        interface{}    `json:"departureStation"`
	DepartureCityName       []string       `json:"departureCityName"`
	DepartureAirportCode3A  string         `json:"departureAirportCode3A"`
	ArrivalStation          interface{}    `json:"arrivalStation"`
	ArrivalCityName         []string       `json:"arrivalCityName"`
	ArrivalAirportCode3A    string         `json:"arrivalAirportCode3A"`
	MarketingCompany        string         `json:"marketingCompany"`
	OperatingCompany        string         `json:"operatingCompany"`
	FlightNumber            string         `json:"flightNumber"`
	TypeOfAircraft          string         `json:"typeOfAircraft"`
	StopsCount              int            `json:"stopsCount"`
	LegDuration             string         `json:"legDuration"`
	CarrierCode             int            `json:"carrierCode"`
	FareCityPairList        []FareCityPair `json:"fareCityPairList"`
	TransitFlightDetailList interface{}    `json:"transitFlightDetailList"`
}

type FareCityPair struct {
	RateClass               string                `json:"rateClass"`
	GlobalClass             interface{}           `json:"globalClass"`
	BookingClass            string                `json:"bookingClass"`
	GlobalSubClass          interface{}           `json:"globalSubClass"`
	CodeShare               interface{}           `json:"codeShare"`
	Company                 string                `json:"company"`
	DirectionCode           string                `json:"directionCode"`
	TotalFare               float64               `json:"totalFare"`
	AllTotalFare            float64               `json:"allTotalFare"`
	Fare                    float64               `json:"fare"`
	Taxes                   float64               `json:"taxes"`
	Currency                string                `json:"currency"`
	FlightClassTypeCode     int                   `json:"flightClassTypeCode"`
	Exists                  bool                  `json:"exists"`
	PassengerFareDetailList []PassengerFareDetail `json:"passengerFareDetailList"`
}

type PassengerFareDetail struct {
	PassengerCategory string  `json:"passengerCategory"`
	Count             int     `json:"count"`
	Fare              float64 `json:"fare"`
	Taxes             float64 `json:"taxes"`
	TotalFare         float64 `json:"totalFare"`
	Currency          string  `json:"currency"`
}

type Azal struct {
	authKey string
	url     string
	date    string
	apiURL  string
}
