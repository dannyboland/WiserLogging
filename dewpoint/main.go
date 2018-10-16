package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/segmentio/go-athena"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
		panic(e)
	}
}

// Return the coldest temperature forecast for the rest of the day
func coldForecast(locationID int) float64 {
	datapoint_apikey := ""
	url := fmt.Sprintf("http://datapoint.metoffice.gov.uk/public/data/val/wxfcs/all/json/%d?res=3hourly&key=%s",locationID,datapoint_apikey)
	response, err := http.Get(url)
	check(err)
	responseData, err := ioutil.ReadAll(response.Body)
	check(err)
	var result map[string]interface{}
	json.Unmarshal(responseData, &result)
	forecasts := result["SiteRep"].(map[string]interface{})["DV"].(map[string]interface{})["Location"].(map[string]interface{})["Period"].([]interface{})
	today := forecasts[0].(map[string]interface{})["Rep"].([]interface{})
	minTemperature := 99
	for value := range today {
		forecastTemperature, err := strconv.Atoi(today[value].(map[string]interface{})["T"].(string))
		check(err)
		if (forecastTemperature < minTemperature) { minTemperature = forecastTemperature }
	}
	tomorrow := forecasts[1].(map[string]interface{})["Rep"].([]interface{})
	for value := range tomorrow {
		// Check for the rest of the night, stop at morning
		if value >= 4 { break }
		forecastTemperature, err := strconv.Atoi(tomorrow[value].(map[string]interface{})["T"].(string))
		check(err)
		if (forecastTemperature < minTemperature) { minTemperature = forecastTemperature }
	}
	return float64(minTemperature)
}

func dewpoint(T float64, RH float64) float64 {
	b := 17.62
	c := 243.12
	H := math.Log(RH/100) + (b*T)/(c+T)
	DP := c * H / (b - H)
	return DP
}

func windowSurfaceTemperature(Texterior float64, Tinterior float64, Uvalue float64, Hinterior float64) float64 {
	return (Hinterior*Tinterior + Uvalue*Texterior) / (Hinterior+Uvalue)
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var messageBuilder strings.Builder
	var alert bool
	coldestExteriorTemperature := coldForecast(1234567)
	query, err := ioutil.ReadFile("athena.sql")
	check(err)
	db, err := sql.Open("athena", "db=YOURDB&output_location=s3://aws-athena-query-results-YOURACCOUNT-eu-west-1&region=eu-west-1")
	check(err)
	_, err = db.Query("MSCK REPAIR TABLE YOURDB.YOURTABLE")
	check(err)
	rows, err := db.Query(fmt.Sprintf("%s", query))
	check(err)
	for rows.Next() {
		var room string
		var lag int64
		var TargetTemperature float64
		var Temperature float64
		var Humidity float64
		rows.Scan(&room, &lag, &TargetTemperature, &Temperature, &Humidity)
		Temperature /= 10
		TargetTemperature /= 10
		if (TargetTemperature == -20) {TargetTemperature = 10}
		windowTemperature := windowSurfaceTemperature(coldestExteriorTemperature, TargetTemperature, 4, 20)
		dewTemperature := dewpoint(Temperature, Humidity)
		if windowTemperature <= dewTemperature {
			alert = true
			fmt.Fprintln(&messageBuilder, "Condensation expected!")
			fmt.Fprintf(&messageBuilder, "Room: %s, Temperature: %.2f, set to temperature: %.2f, humidity: %.2f. Forecast external temperature: %.2f, calculated window temperature: %.2f, dew point temperature: %.2f.\n", room, Temperature, TargetTemperature, Humidity, coldestExteriorTemperature, windowTemperature, dewTemperature)
		}
	}
	if alert {
		sess, err := session.NewSession()
		check(err)
		svc := sns.New(sess, aws.NewConfig().WithRegion("eu-west-1"))
		topicArn := "arn:aws:sns:eu-west-1:YOURACCOUNT:aws_monitor"
		messageString := messageBuilder.String()
		subjectString := "Condensation Alert"
		SNSMessage := sns.PublishInput{
			Message: &messageString,
			Subject: &subjectString,
			TopicArn: &topicArn,
		}
		_, err = svc.Publish(&SNSMessage)
		check(err)
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		}, nil
}

func main() {
	lambda.Start(Handler)
}
