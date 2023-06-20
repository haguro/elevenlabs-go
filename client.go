//go:generate go run cmd/codegen/main.go

package elevenlabs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	elevenlabsBaseURL = "https://api.elevenlabs.io/v1"
	defaultTimeout    = 30 * time.Second
	contentTypeJSON   = "application/json"
)

var (
	once          sync.Once
	defaultClient *Client
)

type QueryFunc func(*url.Values)

type Client struct {
	baseURL string
	apiKey  string
	timeout time.Duration
	ctx     context.Context
}

func getDefaultClient() *Client {
	once.Do(func() {
		defaultClient = NewClient(context.Background(), "", defaultTimeout)
	})
	return defaultClient
}

func SetAPIKey(apiKey string) {
	getDefaultClient().apiKey = apiKey
}

func SetTimeout(timeout time.Duration) {
	getDefaultClient().timeout = timeout
}

func NewClient(ctx context.Context, apiKey string, reqTimeout time.Duration) *Client {
	return &Client{baseURL: elevenlabsBaseURL, apiKey: apiKey, timeout: reqTimeout, ctx: ctx}
}

func LatencyOptimizations(value int) QueryFunc {
	return func(q *url.Values) {
		q.Add("optimize_streaming_latency", fmt.Sprint(value))
	}
}

func WithSettings() QueryFunc {
	return func(q *url.Values) {
		q.Add("with_settings", "true")
	}
}

func (c *Client) doRequest(ctx context.Context, method, url string, bodyBuf *bytes.Buffer, contentType string, queries ...QueryFunc) ([]byte, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(timeoutCtx, method, url, bodyBuf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", contentType)
	if c.apiKey != "" {
		req.Header.Add("xi-api-key", c.apiKey)
	}

	q := req.URL.Query()
	for _, qf := range queries {
		qf(&q)
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusBadRequest, http.StatusUnauthorized:
			apiErr := &APIError{}
			if err := json.Unmarshal(respBody, apiErr); err != nil {
				return respBody, err
			}
			return respBody, apiErr
		case http.StatusUnprocessableEntity:
			valErr := &ValidationError{}
			if err := json.Unmarshal(respBody, valErr); err != nil {
				return respBody, err
			}
			return respBody, valErr
		default:
			return respBody, fmt.Errorf("unexpected HTTP status \"%d %s\" returned from server", resp.StatusCode, http.StatusText(resp.StatusCode))
		}
	}

	return respBody, nil
}

func (c *Client) TextToSpeech(voiceID string, ttsReq TextToSpeechRequest, queries ...QueryFunc) ([]byte, error) {
	reqBody, err := json.Marshal(ttsReq)
	if err != nil {
		return nil, err
	}

	return c.doRequest(c.ctx, http.MethodPost, fmt.Sprintf("%s/text-to-speech/%s", c.baseURL, voiceID), bytes.NewBuffer(reqBody), contentTypeJSON, queries...)
}

func (c *Client) GetModels() ([]Model, error) {
	body, err := c.doRequest(c.ctx, http.MethodGet, fmt.Sprintf("%s/models", c.baseURL), &bytes.Buffer{}, contentTypeJSON)
	if err != nil {
		return nil, err
	}

	var models []Model
	err = json.Unmarshal(body, &models)
	if err != nil {
		return nil, err
	}

	return models, nil
}

func (c *Client) GetVoices() ([]Voice, error) {
	body, err := c.doRequest(c.ctx, http.MethodGet, fmt.Sprintf("%s/voices", c.baseURL), &bytes.Buffer{}, contentTypeJSON)
	if err != nil {
		return nil, err
	}

	var voiceResp GetVoicesResponse
	err = json.Unmarshal(body, &voiceResp)
	if err != nil {
		return nil, err
	}

	return voiceResp.Voices, nil
}

func (c *Client) GetDefaultVoiceSettings() (VoiceSettings, error) {
	var voiceSettings VoiceSettings
	body, err := c.doRequest(c.ctx, http.MethodGet, fmt.Sprintf("%s/voices/settings/default", c.baseURL), &bytes.Buffer{}, contentTypeJSON)
	if err != nil {
		return VoiceSettings{}, err
	}

	err = json.Unmarshal(body, &voiceSettings)
	if err != nil {
		return VoiceSettings{}, err
	}

	return voiceSettings, nil
}

func (c *Client) GetVoiceSettings(voiceId string) (VoiceSettings, error) {
	var voiceSettings VoiceSettings
	body, err := c.doRequest(c.ctx, http.MethodGet, fmt.Sprintf("%s/voices/%s/settings", c.baseURL, voiceId), &bytes.Buffer{}, contentTypeJSON)
	if err != nil {
		return VoiceSettings{}, err
	}

	err = json.Unmarshal(body, &voiceSettings)
	if err != nil {
		return VoiceSettings{}, err
	}

	return voiceSettings, nil
}

func (c *Client) GetVoice(voiceId string, queries ...QueryFunc) (Voice, error) {
	var voice Voice
	body, err := c.doRequest(c.ctx, http.MethodGet, fmt.Sprintf("%s/voices/%s", c.baseURL, voiceId), &bytes.Buffer{}, contentTypeJSON, queries...)
	if err != nil {
		return Voice{}, err
	}

	err = json.Unmarshal(body, &voice)
	if err != nil {
		return Voice{}, err
	}

	return voice, nil
}

func (c *Client) DeleteVoice(voiceId string) error {
	_, err := c.doRequest(c.ctx, http.MethodDelete, fmt.Sprintf("%s/voices/%s", c.baseURL, voiceId), &bytes.Buffer{}, contentTypeJSON)
	return err
}

func (c *Client) EditVoiceSettings(voiceId string, settings VoiceSettings) error {
	reqBody, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	_, err = c.doRequest(c.ctx, http.MethodPost, fmt.Sprintf("%s/voices/%s/settings/edit", c.baseURL, voiceId), bytes.NewBuffer(reqBody), contentTypeJSON)
	return err
}

func (c *Client) AddVoice(voiceReq AddEditVoiceRequest) (string, error) {
	reqBodyBuf, contentType, err := voiceReq.buildRequestBody()
	if err != nil {
		return "", err
	}
	body, err := c.doRequest(c.ctx, http.MethodPost, fmt.Sprintf("%s/voices/add", c.baseURL), reqBodyBuf, contentType)
	if err != nil {
		return "", err
	}
	var voiceResp AddVoiceResponse
	err = json.Unmarshal(body, &voiceResp)
	if err != nil {
		return "", err
	}
	return voiceResp.VoiceId, nil
}

func (c *Client) EditVoice(voiceId string, voiceReq AddEditVoiceRequest) error {
	reqBodyBuf, contentType, err := voiceReq.buildRequestBody()
	if err != nil {
		return err
	}
	_, err = c.doRequest(c.ctx, http.MethodPost, fmt.Sprintf("%s/voices/%s/edit", c.baseURL, voiceId), reqBodyBuf, contentType)
	return err
}

func (c *Client) DeleteSample(voiceId, sampleId string) error {
	_, err := c.doRequest(c.ctx, http.MethodDelete, fmt.Sprintf("%s/voices/%s/samples/%s", c.baseURL, voiceId, sampleId), &bytes.Buffer{}, contentTypeJSON)
	return err
}

func (c *Client) GetSampleAudio(voiceId, sampleId string) ([]byte, error) {
	return c.doRequest(c.ctx, http.MethodGet, fmt.Sprintf("%s/voices/%s/samples/%s/audio", c.baseURL, voiceId, sampleId), &bytes.Buffer{}, contentTypeJSON)
}

func PageSize(n int) QueryFunc {
	return func(q *url.Values) {
		q.Add("page_size", fmt.Sprint(n))
	}
}

func StartAfter(id string) QueryFunc {
	return func(q *url.Values) {
		q.Add("start_after_history_item_id", id)
	}
}

type NextHistoryPageFunc func(...QueryFunc) (GetHistoryResponse, NextHistoryPageFunc, error)

func (c *Client) GetHistory(queries ...QueryFunc) (GetHistoryResponse, NextHistoryPageFunc, error) {
	var historyResp GetHistoryResponse
	body, err := c.doRequest(c.ctx, http.MethodGet, fmt.Sprintf("%s/history", c.baseURL), &bytes.Buffer{}, contentTypeJSON, queries...)
	if err != nil {
		return GetHistoryResponse{}, nil, err
	}

	err = json.Unmarshal(body, &historyResp)
	if err != nil {
		return GetHistoryResponse{}, nil, err
	}

	if !historyResp.HasMore {
		return historyResp, nil, nil
	}

	nextPageFunc := func(qf ...QueryFunc) (GetHistoryResponse, NextHistoryPageFunc, error) {
		qf = append(qf, StartAfter(historyResp.LastHistoryItemId))
		return c.GetHistory(qf...)
	}
	return historyResp, nextPageFunc, nil
}

func (c *Client) GetHistoryItem(itemId string) (HistoryItem, error) {
	var historyItem HistoryItem
	body, err := c.doRequest(c.ctx, http.MethodGet, fmt.Sprintf("%s/history/%s", c.baseURL, itemId), &bytes.Buffer{}, contentTypeJSON)
	if err != nil {
		return HistoryItem{}, err
	}

	err = json.Unmarshal(body, &historyItem)
	if err != nil {
		return HistoryItem{}, err
	}

	return historyItem, nil
}

func (c *Client) DeleteHistoryItem(itemId string) error {
	_, err := c.doRequest(c.ctx, http.MethodDelete, fmt.Sprintf("%s/history/%s", c.baseURL, itemId), &bytes.Buffer{}, contentTypeJSON)
	return err
}

func (c *Client) GetHistoryItemAudio(itemId string) ([]byte, error) {
	return c.doRequest(c.ctx, http.MethodGet, fmt.Sprintf("%s/history/%s/audio", c.baseURL, itemId), &bytes.Buffer{}, contentTypeJSON)
}

func (c *Client) DownloadHistoryAudio(dlReq DownloadHistoryRequest) ([]byte, error) {
	reqBody, err := json.Marshal(dlReq)
	if err != nil {
		return nil, err
	}

	return c.doRequest(c.ctx, http.MethodPost, fmt.Sprintf("%s/history/download", c.baseURL), bytes.NewBuffer(reqBody), contentTypeJSON)
}

func (c *Client) GetSubscription() (Subscription, error) {
	sub := Subscription{}
	body, err := c.doRequest(c.ctx, http.MethodGet, fmt.Sprintf("%s/user/subscription", c.baseURL), &bytes.Buffer{}, contentTypeJSON)
	if err != nil {
		return sub, err
	}

	err = json.Unmarshal(body, &sub)
	if err != nil {
		return sub, err
	}

	return sub, nil
}

func (c *Client) GetUser() (User, error) {
	user := User{}
	body, err := c.doRequest(c.ctx, http.MethodGet, fmt.Sprintf("%s/user", c.baseURL), &bytes.Buffer{}, contentTypeJSON)
	if err != nil {
		return user, err
	}

	err = json.Unmarshal(body, &user)
	if err != nil {
		return user, err
	}

	return user, nil
}
