package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Jason-CKY/telegram-ssbbot/pkg/schemas"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"github.com/vicanso/go-charts/v2"
)

func ListBonds(startDate time.Time, endDate time.Time, rows int) (*[]schemas.SavingsBonds, error) {
	queryParams := fmt.Sprintf("rows=%v&filters=issue_date:[%v+TO+%v]&sort=issue_date+desc", rows, startDate.Format(time.DateOnly), endDate.Format(time.DateOnly))
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

func ListBondInterestRates(bond schemas.SavingsBonds) (*schemas.BondInterest, error) {
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

func FormatSavingsBondNotification(bond schemas.SavingsBonds, interest schemas.BondInterest) string {
	issueCode := bond.IssueCode
	issueDate := time.Time(bond.IssueDate).Format("02 Jan 2006")
	maturityDate := time.Time(bond.MaturityDate).Format("02 Jan 2006")
	lastDayToApply := time.Time(bond.LastDayToApply).Format("02 Jan 2006")

	message := fmt.Sprintf(
		"🇸🇬 *Singapore Savings Bonds \\(%s\\)* 🇸🇬\n\n"+
			"*Issue Code:* %s\n"+
			"*Issue Date:* %s\n"+
			"*Maturity Date:* %s\n"+
			"*Last Day to Apply:* %s\n\n"+
			"*1\\-Year Average Return:* %.2f%%\n"+
			"*10\\-Year Average Return:* %.2f%%\n\n"+
			"*Key Dates:*\n"+
			"\\- First Interest Date: %s\n"+
			"\\- Interest Payment Months: %s\n\n"+
			"*Additional Information:*\n"+
			"\\- Issue Size: %.2f Million SGD\n",
		issueCode,
		issueCode,
		issueDate,
		maturityDate,
		lastDayToApply,
		interest.Year1Return,
		interest.Year10Return,
		time.Time(bond.FirstInterestDate).Format("02 Jan 2006"),
		bond.PaymentMonth,
		bond.IssueSize,
	)
	message = strings.Replace(message, ".", "\\.", -1)
	return message
}

func GenerateSSBInterestRatesChart(interestRates []float64, dates []string) (*[]byte, error) {
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
			Text: "Singapore Savings Bonds 10-Year Average Returns",
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

func GenerateNotificationMessage(chatID int64, timezone *time.Location) (*tgbotapi.PhotoConfig, error) {
	// get the last 12 bonds
	bondsPtr, err := ListBonds(time.Now().In(timezone).AddDate(-1, 0, 0), time.Now().In(timezone).AddDate(0, 1, 0), 12)
	if err != nil {
		return nil, err
	}
	bonds := *bondsPtr

	var latestBond schemas.SavingsBonds
	var latestBondInterests schemas.BondInterest

	latestBond = bonds[0]
	for i := len(bonds)/2 - 1; i >= 0; i-- {
		opp := len(bonds) - 1 - i
		bonds[i], bonds[opp] = bonds[opp], bonds[i]
	}

	var bondReturns []float64
	var bondDates []string

	for _, bond := range bonds {
		bondInterestRate, err := ListBondInterestRates(bond)
		if err != nil {
			return nil, err
		}
		bondReturns = append(bondReturns, bondInterestRate.Year10Return)
		bondDates = append(bondDates, time.Time(bond.IssueDate).Format("Jan 06"))
		if bond.IssueCode == latestBond.IssueCode {
			latestBondInterests = *bondInterestRate
		}
	}
	buf, err := GenerateSSBInterestRatesChart(bondReturns, bondDates)
	if err != nil {
		return nil, err
	}
	photoFileBytes := tgbotapi.FileBytes{
		Name:  "picture",
		Bytes: *buf,
	}
	photoConfig := tgbotapi.NewPhoto(chatID, photoFileBytes)

	// add message information on the latest bond
	photoConfig.Caption = FormatSavingsBondNotification(latestBond, latestBondInterests)
	photoConfig.ParseMode = "MarkdownV2"
	return &photoConfig, nil
}

func ScheduleUpdate(bot *tgbotapi.BotAPI) {
	localTimezone, err := time.LoadLocation("Asia/Singapore") // Look up a location by it's IANA name.
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup

	for {
		time.Sleep(1 * time.Minute)
		chats, err := schemas.GetUsersToNotify(int(time.Now().In(localTimezone).Month()))
		if err != nil {
			panic(err)
		}

		if len(chats) > 0 {
			bondsPtr, err := ListBonds(time.Now().In(localTimezone).AddDate(0, -1, 0), time.Now().AddDate(0, 1, 0).In(localTimezone), 1)
			if err != nil {
				panic(err)
			}
			for _, bond := range *bondsPtr {
				if time.Time(bond.IssueDate).Month() < time.Now().In(localTimezone).Month() {
					// this month's bonds not released yet, skipping loop
					continue
				}
			}
		}

		for _, chat := range chats {
			wg.Add(1)
			go func(bot *tgbotapi.BotAPI, chatSettings *schemas.ChatSettings, timezone *time.Location) {
				defer wg.Done()
				photoConfig, err := GenerateNotificationMessage(chatSettings.ChatId, timezone)
				if err != nil {
					panic(err)
				}
				if _, err := bot.Send(photoConfig); err != nil {
					panic(err)
				}
				chatSettings.LastNotificationTime = schemas.DatetimeWithoutTimezone(time.Now().In(localTimezone))
				chatSettings.Update()
			}(bot, &chat, localTimezone)
			wg.Wait()
		}
	}
}
