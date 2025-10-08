package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	tools "app/handlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	bitrixSpaEntityTypeID = 1050
	bitrixSourceID        = "TELEGRAM"
	bitrixSourceDesc      = "Telegram"
	bitrixAssignedUserID  = 35
)

var (
	bitrixClientOnce sync.Once
	bitrixClientInst *BitrixClient
	bitrixClientErr  error

	bitrixSyncMu sync.Mutex
	bitrixSynced = make(map[int64]bool)

	chatStateMu sync.Mutex
	chatStates  = make(map[int64]*bitrixSession)
)

type BitrixClient struct {
	baseURL    string
	httpClient *http.Client
}

type bitrixSession struct {
	Phone       string
	SpeakerName string
	City        string
	ContactName string
}

func getBitrixClient() (*BitrixClient, error) {
	bitrixClientOnce.Do(func() {
		base := strings.TrimRight(os.Getenv("B24_BASE"), "/")
		if base == "" {
			bitrixClientErr = errors.New("B24_BASE env is empty")
			return
		}
		bitrixClientInst = &BitrixClient{
			baseURL: base,
			httpClient: &http.Client{
				Timeout: 10 * time.Second,
			},
		}
	})
	return bitrixClientInst, bitrixClientErr
}

func trySyncBitrixDeal(bot *tgbotapi.BotAPI, chatID int64) {
	bitrixSyncMu.Lock()
	if bitrixSynced[chatID] {
		bitrixSyncMu.Unlock()
		return
	}
	bitrixSyncMu.Unlock()

	session := snapshotSession(chatID)
	if session == nil {
		return
	}

	phone := strings.TrimSpace(session.Phone)
	courseCity := strings.TrimSpace(session.City)
	if phone == "" || courseCity == "" {
		return
	}

	formattedPhone, err := normalizePhone(phone)
	if err != nil {
		log.Printf("bitrix: phone normalization error (%s): %v", phone, err)
		msg := tgbotapi.NewMessage(chatID, "РќРµ СѓРґР°Р»РѕСЃСЊ РѕР±СЂР°Р±РѕС‚Р°С‚СЊ РЅРѕРјРµСЂ С‚РµР»РµС„РѕРЅР°. РџСЂРѕРІРµСЂСЊ С„РѕСЂРјР°С‚ Рё РїРѕРїСЂРѕР±СѓР№ СЃРЅРѕРІР°.")
		tools.SendAndLog(bot, msg)
		return
	}

	client, err := getBitrixClient()
	if err != nil {
		log.Printf("bitrix: init error: %v", err)
		msg := tgbotapi.NewMessage(chatID, "РќРµ СѓРґР°Р»РѕСЃСЊ РїРѕРґРєР»СЋС‡РёС‚СЊСЃСЏ Рє Bitrix24. РџРѕРїСЂРѕР±СѓР№С‚Рµ РїРѕР·Р¶Рµ.")
		tools.SendAndLog(bot, msg)
		return
	}

	contactName := strings.TrimSpace(session.ContactName)
	if contactName == "" {
		contactName = "РљР»РёРµРЅС‚ Telegram"
	}

	courseTitle := buildCourseTitle(session)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	contactID, itemID, err := client.syncDeal(ctx, formattedPhone, contactName, courseTitle)
	if err != nil {
		log.Printf("bitrix: sync error for chat %d: %v", chatID, err)
		msg := tgbotapi.NewMessage(chatID, "РќРµ СѓРґР°Р»РѕСЃСЊ СЃРѕС…СЂР°РЅРёС‚СЊ РґР°РЅРЅС‹Рµ РІ Bitrix24. РњС‹ СѓР¶Рµ СЂР°Р·Р±РёСЂР°РµРјСЃСЏ, РїРѕРїСЂРѕР±СѓР№ С‡СѓС‚СЊ РїРѕР·Р¶Рµ.")
		tools.SendAndLog(bot, msg)
		return
	}

	bitrixSyncMu.Lock()
	bitrixSynced[chatID] = true
	bitrixSyncMu.Unlock()

	log.Printf("bitrix: synced contact %s and item %s for chat %d", contactID, itemID, chatID)
}

func buildCourseTitle(session *bitrixSession) string {
	speaker := strings.TrimSpace(session.SpeakerName)
	city := strings.TrimSpace(session.City)
	switch {
	case speaker != "" && city != "":
		return speaker + " - " + city
	case city != "":
		return city
	case speaker != "":
		return speaker
	default:
		return "не указан"
	}
}

func normalizePhone(raw string) (string, error) {
	var digits []rune
	for _, r := range raw {
		if unicode.IsDigit(r) {
			digits = append(digits, r)
		}
	}
	if len(digits) == 0 {
		return "", errors.New("phone has no digits")
	}

	value := string(digits)
	switch len(value) {
	case 10:
		value = "7" + value
	case 11:
		if value[0] == '8' {
			value = "7" + value[1:]
		}
	default:
		if len(value) < 10 {
			return "", fmt.Errorf("phone length %d too short", len(value))
		}
	}

	return "+" + value, nil
}

func (c *BitrixClient) syncDeal(ctx context.Context, phone, name, courseTitle string) (string, string, error) {
	contactID, err := c.findOrCreateContact(ctx, phone, name)
	if err != nil {
		return "", "", err
	}

	itemID, err := c.createSpaItem(ctx, contactID, courseTitle)
	if err != nil {
		return "", "", err
	}

	return contactID, itemID, nil
}

func (c *BitrixClient) findOrCreateContact(ctx context.Context, phone, name string) (string, error) {
	if id, err := c.findContact(ctx, phone); err != nil {
		return "", err
	} else if id != "" {
		return id, nil
	}
	return c.createContact(ctx, phone, name)
}

func (c *BitrixClient) findContact(ctx context.Context, phone string) (string, error) {
	payload := map[string]any{
		"filter": map[string]string{
			"PHONE": phone,
		},
		"select": []string{"ID", "NAME", "LAST_NAME", "PHONE"},
	}

	var response struct {
		Result []struct {
			ID string `json:"ID"`
		} `json:"result"`
	}

	if err := c.post(ctx, "crm.contact.list", payload, &response); err != nil {
		return "", err
	}

	if len(response.Result) == 0 {
		return "", nil
	}
	return response.Result[0].ID, nil
}

func (c *BitrixClient) createContact(ctx context.Context, phone, name string) (string, error) {
	payload := map[string]any{
		"fields": map[string]any{
			"NAME":               name,
			"OPENED":             "Y",
			"SOURCE_ID":          bitrixSourceID,
			"SOURCE_DESCRIPTION": bitrixSourceDesc,
			"ASSIGNED_BY_ID":     bitrixAssignedUserID,
			"PHONE": []map[string]string{
				{
					"VALUE":      phone,
					"VALUE_TYPE": "WORK",
				},
			},
		},
	}

	var response struct {
		Result any `json:"result"`
	}
	if err := c.post(ctx, "crm.contact.add", payload, &response); err != nil {
		return "", err
	}

	switch v := response.Result.(type) {
	case float64:
		return strconv.Itoa(int(v)), nil
	case string:
		return v, nil
	default:
		raw, _ := json.Marshal(response.Result)
		return "", fmt.Errorf("unexpected contact add result: %s", raw)
	}
}

func (c *BitrixClient) createSpaItem(ctx context.Context, contactID, courseTitle string) (string, error) {
	idInt, err := strconv.Atoi(contactID)
	if err != nil {
		return "", fmt.Errorf("invalid contact id %s: %w", contactID, err)
	}

	payload := map[string]any{
		"entityTypeId": bitrixSpaEntityTypeID,
		"fields": map[string]any{
			"title":             fmt.Sprintf("РћРїР»Р°С‚Р° РёР· Telegram вЂ” РєСѓСЂСЃ %s", courseTitle),
			"opened":            "Y",
			"contactIds":        []int{idInt},
			"sourceId":          bitrixSourceID,
			"sourceDescription": bitrixSourceDesc,
			"assignedById":      bitrixAssignedUserID,
		},
	}

	var response struct {
		Result struct {
			Item struct {
				ID any `json:"id"`
			} `json:"item"`
		} `json:"result"`
	}

	if err := c.post(ctx, "crm.item.add", payload, &response); err != nil {
		return "", err
	}

	switch v := response.Result.Item.ID.(type) {
	case float64:
		return strconv.Itoa(int(v)), nil
	case string:
		return v, nil
	default:
		raw, _ := json.Marshal(response.Result.Item.ID)
		return "", fmt.Errorf("unexpected item id type: %s", raw)
	}
}

func (c *BitrixClient) post(ctx context.Context, endpoint string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := c.baseURL + "/" + strings.TrimLeft(endpoint, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		log.Printf("bitrix: http %d response: %s. request: %s", resp.StatusCode, string(respBody), string(body))
		return fmt.Errorf("bitrix http %d", resp.StatusCode)
	}

	var apiErr struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Error != "" {
		log.Printf("bitrix: api error %s (%s). request: %s", apiErr.Error, apiErr.ErrorDescription, string(body))
		return fmt.Errorf("%s: %s", apiErr.Error, apiErr.ErrorDescription)
	}

	if out != nil {
		if err := json.Unmarshal(respBody, out); err != nil {
			return err
		}
	}
	return nil
}

func snapshotSession(chatID int64) *bitrixSession {
	chatStateMu.Lock()
	defer chatStateMu.Unlock()
	state := chatStates[chatID]
	if state == nil {
		return nil
	}
	copy := *state
	return &copy
}

func updateSession(chatID int64, fn func(*bitrixSession)) {
	chatStateMu.Lock()
	defer chatStateMu.Unlock()
	state := chatStates[chatID]
	if state == nil {
		state = &bitrixSession{}
		chatStates[chatID] = state
	}
	fn(state)
}

func setSessionContact(chatID int64, phone, contactName string) {
	updateSession(chatID, func(s *bitrixSession) {
		if phone != "" {
			s.Phone = phone
		}
		if contactName != "" {
			s.ContactName = contactName
		}
	})
}

func setSessionCourse(chatID int64, speaker, city string) {
	updateSession(chatID, func(s *bitrixSession) {
		if speaker != "" {
			s.SpeakerName = speaker
		}
		if city != "" {
			s.City = city
		}
	})
}
