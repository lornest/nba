package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

const port = "3001"

const boxScoreSummaryURL = "https://stats.nba.com/stats/boxscoresummaryv2?GameID=0021700807"

type ScoreboardResponse struct {
	Date       time.Time   `json:"date"`
	LineScores []LineScore `json:"lineScores"`
}

type LineScore struct {
	TeamAbbr     string
	TeamCity     string
	TeamNickname string
	PtsQtr1      int
	PtsQtr2      int
	PtsQtr3      int
	PtsQtr4      int
	PtsOt1       int
	PtsOt2       int
	PtsOt3       int
	PtsOt4       int
	PtsOt5       int
	PtsOt6       int
	PtsOt7       int
	PtsOt8       int
	PtsOt9       int
	PtsOt10      int
	Pts          int
}

type BoxScoreSummary struct {
	ResultSets []ResultSets `json:"resultSets"`
}

type ResultSets struct {
	Name    string        `json:"name"`
	Headers []string      `json:"headers"`
	RowSet  []interface{} `json:"rowSet"` // Use interface{} because the type varies
}

func main() {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func routes() http.Handler {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			req, err := http.NewRequest("GET", boxScoreSummaryURL, nil)
			if err != nil {
				fmt.Println("Error creating the request:", err)
				return
			}

			// Set headers:w
			req.Header.Set("Host", "stats.nba.com")
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:72.0) Gecko/20100101 Firefox/72.0")
			req.Header.Set("Accept", "application/json, text/plain, */*")
			req.Header.Set("Accept-Language", "en-US,en;q=0.5")
			req.Header.Set("Accept-Encoding", "gzip, deflate, br")
			req.Header.Set("x-nba-stats-origin", "stats")
			req.Header.Set("x-nba-stats-token", "true")
			req.Header.Set("Connection", "keep-alive")
			req.Header.Set("Referer", "https://stats.nba.com/")
			req.Header.Set("Pragma", "no-cache")
			req.Header.Set("Cache-Control", "no-cache")
			// Create an HTTP client and send the request

			client := &http.Client{}
			response, err := client.Do(req)
			if err != nil {
				fmt.Println("Error making the request:", err)
				return
			}
			defer response.Body.Close()

			var reader = response.Body
			if response.Header.Get("Content-Encoding") == "gzip" {
				gzipReader, err := gzip.NewReader(response.Body)
				if err != nil {
					fmt.Println("Error creating gzip reader:", err)
					return
				}
				defer gzipReader.Close()
				reader = gzipReader
			}

			// Read the response body
			body, err := io.ReadAll(reader)
			if err != nil {
				fmt.Println("Error reading response body:", err)
				return
			}

			// Unmarshal the JSON into your struct
			var boxScoreSummary BoxScoreSummary
			err = json.Unmarshal(body, &boxScoreSummary)
			if err != nil {
				fmt.Println("Error unmarshaling JSON:", err)
				return
			}

			scoreboard := ScoreboardResponse{}

			for _, resultSet := range boxScoreSummary.ResultSets {
				if resultSet.Name == "LineScore" {
					for _, row := range resultSet.RowSet {
						row := row.([]interface{})

						if scoreboard.Date.IsZero() {
							if val, ok := row[0].(string); ok {
								layout := "2006-01-02T15:04:05"
								t, err := time.Parse(layout, val)
								if err != nil {
									fmt.Println("Error parsing the date:", err)
									return
								}
								scoreboard.Date = t
							} else {
								fmt.Println("The first element of RowSet is not a string.")
							}
						}

						lineScore := LineScore{
							TeamAbbr:     getStringValue(row[4]),
							TeamCity:     getStringValue(row[5]),
							TeamNickname: getStringValue(row[6]),
							PtsQtr1:      getIntValue(row[8]),
							PtsQtr2:      getIntValue(row[9]),
							PtsQtr3:      getIntValue(row[10]),
							PtsQtr4:      getIntValue(row[11]),
							PtsOt1:       getIntValue(row[12]),
							PtsOt2:       getIntValue(row[13]),
							PtsOt3:       getIntValue(row[14]),
							PtsOt4:       getIntValue(row[15]),
							PtsOt5:       getIntValue(row[16]),
							PtsOt6:       getIntValue(row[17]),
							PtsOt7:       getIntValue(row[18]),
							PtsOt8:       getIntValue(row[19]),
							PtsOt9:       getIntValue(row[20]),
							PtsOt10:      getIntValue(row[21]),
							Pts:          getIntValue(row[22]),
						}

						scoreboard.LineScores = append(scoreboard.LineScores, lineScore)
					}
				}
			}

			responseJSON, err := json.Marshal(scoreboard)
			if err != nil {
				fmt.Println("Error marshaling the response:", err)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(responseJSON)
		})
	})

	return r
}

func getStringValue(val interface{}) string {
	if str, ok := val.(string); ok {
		return str
	}
	fmt.Printf("Failed to convert the value to a string, is type %T.\n", val)
	return ""
}

func getIntValue(val interface{}) int {
	if num, ok := val.(float64); ok {
		return int(num)
	}
	fmt.Printf("Failed to convert the value to an int, is type %T.\n", val)
	return 0
}
