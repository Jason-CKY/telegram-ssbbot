package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Jason-CKY/telegram-ssbbot/pkg/schemas"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"github.com/vicanso/go-charts/v2"
)

func listBonds(startDate time.Time, endDate time.Time, rows int) (*[]schemas.SavingsBonds, error) {
	queryParams := fmt.Sprintf("rows=%v&filters=issue_date:[%v+TO+%v]&sort=issue_date+asc", rows, startDate.Format(time.DateOnly), endDate.Format(time.DateOnly))
	endpoint := fmt.Sprintf("%v?%v", "https://eservices.mas.gov.sg/statistics/api/v1/bondsandbills/m/listsavingbonds", queryParams)

	log.Debugf("querying %v", endpoint)

	req, httpErr := http.NewRequest(http.MethodGet, endpoint, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0") // need to set user-agent if not will throw 403 error
	if httpErr != nil {
		return nil, httpErr
	}
	client := &http.Client{}
	res, httpErr := client.Do(req)
	if httpErr != nil {
		return nil, httpErr
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code %v error listing bonds from mas api: %v", res.StatusCode, string(body))
	}
	var savingsBondsAPIResponse schemas.ListSavingsBondsResponse
	jsonErr := json.Unmarshal(body, &savingsBondsAPIResponse)
	// error handling for json unmarshaling
	if jsonErr != nil {
		return nil, jsonErr
	}

	return &savingsBondsAPIResponse.Result.Records, nil
}

func listBondInterestRates(bond schemas.SavingsBonds) (*schemas.BondInterest, error) {
	endpoint := fmt.Sprintf("https://eservices.mas.gov.sg/statistics/api/v1/bondsandbills/m/savingbondsinterest?filters=issue_code:%v", bond.IssueCode)

	log.Debugf("querying %v", endpoint)

	req, httpErr := http.NewRequest(http.MethodGet, endpoint, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0") // need to set user-agent if not will throw 403 error
	if httpErr != nil {
		return nil, httpErr
	}
	client := &http.Client{}
	res, httpErr := client.Do(req)
	if httpErr != nil {
		return nil, httpErr
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code %v error listing bonds from mas api: %v", res.StatusCode, string(body))
	}
	var savingsBondsInterestsAPIResponse schemas.ListSavingsBondsInterestResponse
	jsonErr := json.Unmarshal(body, &savingsBondsInterestsAPIResponse)
	// error handling for json unmarshaling
	if jsonErr != nil {
		return nil, jsonErr
	}

	if len(savingsBondsInterestsAPIResponse.Result.Records) == 0 {
		return nil, fmt.Errorf("savings bonds with issue code: %v not found", bond.IssueCode)
	}
	return &savingsBondsInterestsAPIResponse.Result.Records[0], nil
}

func writeFile(buf []byte) error {
	tmpPath := "./tmp"
	err := os.MkdirAll(tmpPath, 0700)
	if err != nil {
		return err
	}

	file := filepath.Join(tmpPath, "line-chart.png")
	err = os.WriteFile(file, buf, 0600)
	if err != nil {
		return err
	}
	return nil
}

func generateSSBInterestRatesChart(interestRates []float64, dates []string) (*[]byte, error) {
	chartOption := charts.ChartOption{
		Width:  1000,
		Height: 400,
		SeriesList: []charts.Series{
			{
				Type:  charts.ChartTypeLine,
				Data:  charts.NewSeriesDataFromValues(interestRates),
				Label: charts.SeriesLabel{Show: *charts.TrueFlag()},
			}},
		Title: charts.TitleOption{
			Text: "Singapore Savings Bonds 10-Year Returns",
		},
		Padding: charts.Box{
			Top:    20,
			Left:   20,
			Right:  20,
			Bottom: 20,
		},
		Legend: charts.NewLegendOption([]string{
			"Interest Rates",
		}, charts.PositionRight),
		XAxis: charts.NewXAxisOption(dates),
		ValueFormatter: func(f float64) string {
			return fmt.Sprintf("%.00f", f) + "%"
		},
	}
	p, err := charts.Render(chartOption)

	if err != nil {
		return nil, err
	}

	buf, err := p.Bytes()
	if err != nil {
		return nil, err
	}
	return &buf, nil
}

func ScheduleUpdate(bot *tgbotapi.BotAPI) {
	localTimezone, err := time.LoadLocation("Asia/Singapore") // Look up a location by it's IANA name.
	if err != nil {
		panic(err)
	}
	// extract the last 12 months of bonds information
	bonds, err := listBonds(time.Now().In(localTimezone).AddDate(-1, 0, 0), time.Now().In(localTimezone), 12)
	if err != nil {
		panic(err)
	}

	var bondReturns []float64
	var bondDates []string

	for _, bond := range *bonds {
		bondInterestRate, err := listBondInterestRates(bond)
		if err != nil {
			panic(err)
		}
		bondReturns = append(bondReturns, bondInterestRate.Year10Return)
		bondDates = append(bondDates, time.Time(bond.IssueDate).Format("Jan 06"))
	}
	buf, err := generateSSBInterestRatesChart(bondReturns, bondDates)
	if err != nil {
		panic(err)
	}
	err = writeFile(*buf)
	if err != nil {
		panic(err)
	}
}
