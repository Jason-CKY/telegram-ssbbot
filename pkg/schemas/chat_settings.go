package schemas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Jason-CKY/telegram-ssbbot/pkg/utils"
)

type DatetimeWithoutTimezone time.Time

func (t DatetimeWithoutTimezone) MarshalJSON() ([]byte, error) {
	formattedTime := time.Time(t).Format("2006-01-02T15:04:05")
	return json.Marshal(formattedTime)
}

func (t *DatetimeWithoutTimezone) UnmarshalJSON(data []byte) error {
	var timeStr string
	if err := json.Unmarshal(data, &timeStr); err != nil {
		return err
	}
	parsedTime, err := time.Parse("2006-01-02T15:04:05", timeStr)
	if err != nil {
		return err
	}
	*t = DatetimeWithoutTimezone(parsedTime)
	return nil
}

type ChatSettings struct {
	ChatId                 int64                   `json:"chat_id"`
	LastNotificationTime   DatetimeWithoutTimezone `json:"last_notification_time"`
	LatestSSBMonthNotified int                     `json:"latest_ssb_month_notified"`
}

// MarshalJSON implements the json.Marshaler interface.
func (cs ChatSettings) MarshalJSON() ([]byte, error) {
	type Alias ChatSettings // Prevent recursion

	aux := &struct {
		ChatId string `json:"chat_id"`
		*Alias
	}{
		ChatId: strconv.FormatInt(cs.ChatId, 10),
		Alias:  (*Alias)(&cs),
	}
	return json.Marshal(aux)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (cs *ChatSettings) UnmarshalJSON(data []byte) error {
	type Alias ChatSettings // Prevent recursion

	aux := &struct {
		ChatId string `json:"chat_id"`
		*Alias
	}{
		Alias: (*Alias)(cs),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	chatId, err := strconv.ParseInt(aux.ChatId, 10, 64)
	if err != nil {
		return err
	}
	cs.ChatId = chatId
	return nil
}

func (chatSettings ChatSettings) Create() error {
	endpoint := fmt.Sprintf("%v/items/ssbbot_chat_settings", utils.DirectusHost)
	reqBody, _ := json.Marshal(chatSettings)
	req, httpErr := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", utils.DirectusToken))
	if httpErr != nil {
		return httpErr
	}
	client := &http.Client{}
	res, httpErr := client.Do(req)
	if httpErr != nil {
		return httpErr
	}
	body, _ := io.ReadAll(res.Body)
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("error inserting chat settings to directus: %v", string(body))
	}

	return nil
}

func (chatSettings ChatSettings) Update() error {
	endpoint := fmt.Sprintf("%v/items/ssbbot_chat_settings/%v", utils.DirectusHost, chatSettings.ChatId)
	reqBody, _ := json.Marshal(chatSettings)
	req, httpErr := http.NewRequest(http.MethodPatch, endpoint, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", utils.DirectusToken))
	if httpErr != nil {
		return httpErr
	}
	client := &http.Client{}
	res, httpErr := client.Do(req)
	if httpErr != nil {
		return httpErr
	}
	body, _ := io.ReadAll(res.Body)
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("error updating chat settings to directus: %v", string(body))
	}

	return nil

}

func (chatSettings ChatSettings) Delete() error {
	endpoint := fmt.Sprintf("%v/items/ssbbot_chat_settings/%v", utils.DirectusHost, chatSettings.ChatId)
	req, httpErr := http.NewRequest(http.MethodDelete, endpoint, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", utils.DirectusToken))
	if httpErr != nil {
		return httpErr
	}
	client := &http.Client{}
	res, httpErr := client.Do(req)
	if httpErr != nil {
		return httpErr
	}
	body, _ := io.ReadAll(res.Body)
	defer res.Body.Close()
	if res.StatusCode != 204 {
		return fmt.Errorf("error deleting chat settings in directus: %v", string(body))
	}
	return nil
}

func GetChatSettings(chatId int64) (*ChatSettings, error) {
	endpoint := fmt.Sprintf("%v/items/ssbbot_chat_settings", utils.DirectusHost)
	reqBody := []byte(fmt.Sprintf(`{
		"query": {
			"filter": {
				"chat_id": {
					"_eq": "%v"
				}
			}
		}
	}`, chatId))
	req, httpErr := http.NewRequest("SEARCH", endpoint, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", utils.DirectusToken))
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
		return nil, fmt.Errorf("error getting chat settings in directus: %v", string(body))
	}
	var chatSettingsResponse map[string][]ChatSettings
	jsonErr := json.Unmarshal(body, &chatSettingsResponse)
	// error handling for json unmarshaling
	if jsonErr != nil {
		return nil, jsonErr
	}

	if len(chatSettingsResponse["data"]) == 0 {
		return nil, nil
	}

	return &chatSettingsResponse["data"][0], nil
}

func InsertChatSettingsIfNotPresent(chatId int64) (*ChatSettings, bool, error) {
	chatSettings, err := GetChatSettings(chatId)
	if err != nil {
		return nil, false, err
	}
	if chatSettings == nil {
		localTimezone, err := time.LoadLocation("Asia/Singapore") // Look up a location by it's IANA name.
		if err != nil {
			panic(err)
		}
		chatSettings = &ChatSettings{
			ChatId:               chatId,
			LastNotificationTime: DatetimeWithoutTimezone(time.Now().In(localTimezone)),
		}
		err = chatSettings.Create()
		if err != nil {
			return nil, false, err
		}

		return chatSettings, false, nil
	}
	return chatSettings, true, nil
}

func MigrateChatSettingsChatId(fromChatId int64, toChatId int64) error {
	oldChatSettings, err := GetChatSettings(fromChatId)
	if err != nil {
		return err
	}
	err = oldChatSettings.Delete()
	if err != nil {
		return err
	}
	oldChatSettings.ChatId = toChatId
	err = oldChatSettings.Create()
	if err != nil {
		return err
	}
	return nil
}

func GetUsersToNotify(month int) ([]ChatSettings, error) {
	endpoint := fmt.Sprintf("%v/items/ssbbot_chat_settings", utils.DirectusHost)
	reqBody := fmt.Appendf(nil, `{
		"query": {
			"filter": {
				"latest_ssb_month_notified": {
					"_neq": %v
				}
			}
		}
	}`, month)
	req, httpErr := http.NewRequest("SEARCH", endpoint, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", utils.DirectusToken))
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
		return nil, fmt.Errorf("error searching for reminder in directus: %v", string(body))
	}
	var reminderResponse map[string][]ChatSettings
	jsonErr := json.Unmarshal(body, &reminderResponse)
	// error handling for json unmarshaling
	if jsonErr != nil {
		return nil, jsonErr
	}

	return reminderResponse["data"], nil
}
