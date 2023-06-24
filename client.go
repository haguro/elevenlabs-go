//go:generate go run cmd/codegen/main.go

// Package elevenlabs provide an interface to interact with the Elevenlabs voice generation API in Go.
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

// QueryFunc represents the type of functions that sets certain query string to
// a given or certain value.
type QueryFunc func(*url.Values)

// Client represents an API client that can be used to make calls to the Elevenlabs API.
// The NewClient function should be used when instantiating a new Client.
//
// This library also includes a default client instance that can be used when it's more convenient or when
// only a single instance of Client will ever be used by the program. The default client's API key and timeout
// (which defaults to 30 seconds) can be modified with SetAPIKey and SetTimeout respectively, but the parent
// context is fixed and is set to context.Background().
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

// SetAPIKey sets the API key for the default client.
//
// It should be called before making any API calls with the default client if
// authentication is needed.
// The function takes a string argument which is the API key to be set.
func SetAPIKey(apiKey string) {
	getDefaultClient().apiKey = apiKey
}

// SetTimeout sets the timeout duration for the default client.
//
// It can be called if a custom timeout settings are required for API calls.
// The function takes a time.Duration argument which is the timeout to be set.
func SetTimeout(timeout time.Duration) {
	getDefaultClient().timeout = timeout
}

// NewClient creates and returns a new Client object with provided settings.
//
// It should be used to instantiate a new client with a specific API key, request timeout, and context.
//
// It takes a context.Context argument which act as the parent context to be used for requests made by this
// client, a string argument that represents the API key to be used for authenticated requests and
// a time.Duration argument that represents the timeout duration for the client's requests.
//
// It returns a pointer to a newly created Client.
func NewClient(ctx context.Context, apiKey string, reqTimeout time.Duration) *Client {
	return &Client{baseURL: elevenlabsBaseURL, apiKey: apiKey, timeout: reqTimeout, ctx: ctx}
}

func (c *Client) doRequest(ctx context.Context, RespBodyWriter io.Writer, method, url string, bodyBuf io.Reader, contentType string, queries ...QueryFunc) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(timeoutCtx, method, url, bodyBuf)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "*/*")
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}
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
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		switch resp.StatusCode {
		case http.StatusBadRequest, http.StatusUnauthorized:
			apiErr := &APIError{}
			if err := json.Unmarshal(respBody, apiErr); err != nil {
				return err
			}
			return apiErr
		case http.StatusUnprocessableEntity:
			valErr := &ValidationError{}
			if err := json.Unmarshal(respBody, valErr); err != nil {
				return err
			}
			return valErr
		default:
			return fmt.Errorf("unexpected HTTP status \"%d %s\" returned from server", resp.StatusCode, http.StatusText(resp.StatusCode))
		}
	}

	_, err = io.Copy(RespBodyWriter, resp.Body)
	return err
}

// LatencyOptimizations is a QueryFunc that sets the http query 'optimize_streaming_latency' to
// a certain value. It is meant to be used used with TextToSpeech and TextToSpeechStream to turn
// on latency optimization.
//
// Possible values:
// 0 - default mode (no latency optimizations).
// 1 - normal latency optimizations.
// 2 - strong latency optimizations.
// 3 - max latency optimizations.
// 4 - max latency optimizations, with text normalizer turned off (best latency, but can mispronounce things like numbers or dates).
func LatencyOptimizations(value int) QueryFunc {
	return func(q *url.Values) {
		q.Add("optimize_streaming_latency", fmt.Sprint(value))
	}
}

// WithSettings is a QueryFunc that sets the http query 'with_settings' to true. It is meant to be used with
// GetVoice to include Voice setting info with the Voice metadata.
func WithSettings() QueryFunc {
	return func(q *url.Values) {
		q.Add("with_settings", "true")
	}
}

// TextToSpeech converts and returns a given text to speech audio using a certain voice.
//
// It takes a string argument that represents the ID of the voice to be used for the text to speech conversion,
// a TextToSpeechRequest argument that contain the text to be used to generate the audio alongside other settings
// and an optional list of QueryFunc 'queries' to modify the request. The QueryFunc relevant for this function
// is LatencyOptimizations.
//
// It returns a byte slice that contains mpeg encoded audio data in case of success, or an error.
func (c *Client) TextToSpeech(voiceID string, ttsReq TextToSpeechRequest, queries ...QueryFunc) ([]byte, error) {
	reqBody, err := json.Marshal(ttsReq)
	if err != nil {
		return nil, err
	}
	b := bytes.Buffer{}
	err = c.doRequest(c.ctx, &b, http.MethodPost, fmt.Sprintf("%s/text-to-speech/%s", c.baseURL, voiceID), bytes.NewBuffer(reqBody), contentTypeJSON, queries...)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// TextToSpeech converts and streams a given text to speech audio using a certain voice.
//
// It takes an io.Writer argument to which the streamed audio will be copied, a string argument that represents the
// ID of the voice to be used for the text to speech conversion, a TextToSpeechRequest argument that contain the text
// to be used to generate the audio alongside other settings and an optional list of QueryFunc 'queries' to modify the
// request. The QueryFunc relevant for this function is LatencyOptimizations.
//
// It is important to set the timeout of the client to a duration large enough to maintain the desired streaming period.
//
// It returns nil if successful or an error otherwise.
func (c *Client) TextToSpeechStream(streamWriter io.Writer, voiceID string, ttsReq TextToSpeechRequest, queries ...QueryFunc) error {
	reqBody, err := json.Marshal(ttsReq)
	if err != nil {
		return err
	}

	return c.doRequest(c.ctx, streamWriter, http.MethodPost, fmt.Sprintf("%s/text-to-speech/%s/stream", c.baseURL, voiceID), bytes.NewBuffer(reqBody), contentTypeJSON, queries...)
}

// GetModels retrieves the list all available models.
//
// It returns a slice of Model objects or an error.
func (c *Client) GetModels() ([]Model, error) {
	b := bytes.Buffer{}
	err := c.doRequest(c.ctx, &b, http.MethodGet, fmt.Sprintf("%s/models", c.baseURL), &bytes.Buffer{}, contentTypeJSON)
	if err != nil {
		return nil, err
	}

	var models []Model
	if err := json.Unmarshal(b.Bytes(), &models); err != nil {
		return nil, err
	}

	return models, nil
}

// GetVoices retrieves the list of all voices available for use.
//
// It returns a slice of Voice objects or an error.
func (c *Client) GetVoices() ([]Voice, error) {
	b := bytes.Buffer{}
	err := c.doRequest(c.ctx, &b, http.MethodGet, fmt.Sprintf("%s/voices", c.baseURL), &bytes.Buffer{}, contentTypeJSON)
	if err != nil {
		return nil, err
	}

	var voiceResp GetVoicesResponse
	if err := json.Unmarshal(b.Bytes(), &voiceResp); err != nil {
		return nil, err
	}

	return voiceResp.Voices, nil
}

// GetDefaultVoiceSettings retrieves the default settings for voices
//
// It returns a VoiceSettings object or an error.
func (c *Client) GetDefaultVoiceSettings() (VoiceSettings, error) {
	var voiceSettings VoiceSettings
	b := bytes.Buffer{}
	err := c.doRequest(c.ctx, &b, http.MethodGet, fmt.Sprintf("%s/voices/settings/default", c.baseURL), &bytes.Buffer{}, contentTypeJSON)
	if err != nil {
		return VoiceSettings{}, err
	}

	if err := json.Unmarshal(b.Bytes(), &voiceSettings); err != nil {
		return VoiceSettings{}, err
	}

	return voiceSettings, nil
}

// GetVoiceSettings retrieves the settings for a specific voice.
//
// It takes a string argument that represents the ID of the voice for which the settings are retrieved.
//
// It returns a VoiceSettings object or an error.
func (c *Client) GetVoiceSettings(voiceId string) (VoiceSettings, error) {
	var voiceSettings VoiceSettings
	b := bytes.Buffer{}
	err := c.doRequest(c.ctx, &b, http.MethodGet, fmt.Sprintf("%s/voices/%s/settings", c.baseURL, voiceId), &bytes.Buffer{}, contentTypeJSON)
	if err != nil {
		return VoiceSettings{}, err
	}

	if err := json.Unmarshal(b.Bytes(), &voiceSettings); err != nil {
		return VoiceSettings{}, err
	}

	return voiceSettings, nil
}

// GetVoice retrieves metadata about a certain voice.
//
// It takes a string argument that represents the ID of the voice for which the metadata are retrieved
// and an optional list of QueryFunc 'queries' to modify the request. The QueryFunc relevant for this
// function is WithSettings.
//
// It returns a Voice object or an error.
func (c *Client) GetVoice(voiceId string, queries ...QueryFunc) (Voice, error) {
	var voice Voice
	b := bytes.Buffer{}
	err := c.doRequest(c.ctx, &b, http.MethodGet, fmt.Sprintf("%s/voices/%s", c.baseURL, voiceId), &bytes.Buffer{}, contentTypeJSON, queries...)
	if err != nil {
		return Voice{}, err
	}

	if err := json.Unmarshal(b.Bytes(), &voice); err != nil {
		return Voice{}, err
	}

	return voice, nil
}

// DeleteVoice deletes a voice.
//
// It takes a string argument that represents the ID of the voice to be deleted.
//
// It returns a nil if successful, or an error.
func (c *Client) DeleteVoice(voiceId string) error {
	return c.doRequest(c.ctx, &bytes.Buffer{}, http.MethodDelete, fmt.Sprintf("%s/voices/%s", c.baseURL, voiceId), &bytes.Buffer{}, contentTypeJSON)
}

// EditVoiceSettings updates the settings for a specific voice.
//
// It takes a string argument that represents the ID of the voice for which the settings to be
// updated belong, and a VoiceSettings argument that contains the new settings to be applied.
//
// It returns nil if successful or an error otherwise.
func (c *Client) EditVoiceSettings(voiceId string, settings VoiceSettings) error {
	reqBody, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	return c.doRequest(c.ctx, &bytes.Buffer{}, http.MethodPost, fmt.Sprintf("%s/voices/%s/settings/edit", c.baseURL, voiceId), bytes.NewBuffer(reqBody), contentTypeJSON)
}

// AddVoice adds a new voice to the user's VoiceLab.
//
// It takes an AddEditVoiceRequest argument that contains the information of the voice to be added.
//
// It returns the ID of the newly added voice, or an error.
func (c *Client) AddVoice(voiceReq AddEditVoiceRequest) (string, error) {
	reqBodyBuf, contentType, err := voiceReq.buildRequestBody()
	if err != nil {
		return "", err
	}
	b := bytes.Buffer{}
	err = c.doRequest(c.ctx, &b, http.MethodPost, fmt.Sprintf("%s/voices/add", c.baseURL), reqBodyBuf, contentType)
	if err != nil {
		return "", err
	}
	var voiceResp AddVoiceResponse
	if err := json.Unmarshal(b.Bytes(), &voiceResp); err != nil {
		return "", err
	}
	return voiceResp.VoiceId, nil
}

// EditVoice updates an existing voice belonging to the user.
//
// It takes a string argument that represents the ID of the voice to update,
// and an AddEditVoiceRequest argument 'voiceReq' that contains the updated information for the voice.
//
// It returns nil if successful or an error otherwise.
func (c *Client) EditVoice(voiceId string, voiceReq AddEditVoiceRequest) error {
	reqBodyBuf, contentType, err := voiceReq.buildRequestBody()
	if err != nil {
		return err
	}
	return c.doRequest(c.ctx, &bytes.Buffer{}, http.MethodPost, fmt.Sprintf("%s/voices/%s/edit", c.baseURL, voiceId), reqBodyBuf, contentType)
}

// DeleteSample deletes a sample associated with a specific voice.
//
// It takes two string arguments representing the ID of the voice to which the sample belongs
// and the ID of the sample to be deleted respectively.
//
// It returns nil if successful or an error otherwise.
func (c *Client) DeleteSample(voiceId, sampleId string) error {
	return c.doRequest(c.ctx, &bytes.Buffer{}, http.MethodDelete, fmt.Sprintf("%s/voices/%s/samples/%s", c.baseURL, voiceId, sampleId), &bytes.Buffer{}, contentTypeJSON)
}

// GetSampleAudio retrieves the audio data for a specific sample associated with a voice.
//
// It takes two string arguments representing the IDs of the voice and sample respectively.
//
// It returns a byte slice containing the audio data in case of success or an error.
func (c *Client) GetSampleAudio(voiceId, sampleId string) ([]byte, error) {
	b := bytes.Buffer{}
	err := c.doRequest(c.ctx, &b, http.MethodGet, fmt.Sprintf("%s/voices/%s/samples/%s/audio", c.baseURL, voiceId, sampleId), &bytes.Buffer{}, contentTypeJSON)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// PageSize is a QueryFunc that sets the http query 'page_size' to a given value. It is meant to be used
// with GetHistory to set the number of elements returned in the GetHistoryResponse.History slice.
func PageSize(n int) QueryFunc {
	return func(q *url.Values) {
		q.Add("page_size", fmt.Sprint(n))
	}
}

// StartAfter is a QueryFunc that sets the http query 'start_after_history_item_id' to a given item ID.
// It is meant to be used with GetHistory to
func StartAfter(id string) QueryFunc {
	return func(q *url.Values) {
		q.Add("start_after_history_item_id", id)
	}
}

// NextHistoryPageFunc represent functions that can be used to access subsequent history pages. It is
// returned by the GetHistory client method.
//
// A NextHistoryPageFunc function wraps a call to GetHistory which will subsequently return another
// NextHistoryPageFunc until all history pages are retrieved in which case nil will be returned in its pace.
//
// As such, a "while"-style for loop or recursive calls to the returned NextHistoryPageFunc can be employed
// to retrieve all history if needed.
type NextHistoryPageFunc func(...QueryFunc) (GetHistoryResponse, NextHistoryPageFunc, error)

// GetHistory retrieves the history of all created audio and their metadata
//
// It accepts an optional list of QueryFunc 'queries' to modify the request. The QueryFunc functions
// relevant for this function are PageSize and StartAfter.
//
// It returns a GetHistoryResponse object containing the history data, a function of type NextHistoryPageFunc
// to retrieve the next page of history, and an error.
func (c *Client) GetHistory(queries ...QueryFunc) (GetHistoryResponse, NextHistoryPageFunc, error) {
	var historyResp GetHistoryResponse
	b := bytes.Buffer{}
	err := c.doRequest(c.ctx, &b, http.MethodGet, fmt.Sprintf("%s/history", c.baseURL), &bytes.Buffer{}, contentTypeJSON, queries...)
	if err != nil {
		return GetHistoryResponse{}, nil, err
	}

	if err := json.Unmarshal(b.Bytes(), &historyResp); err != nil {
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

// GetHistoryItem retrieves a specific history item by its ID.
//
// It takes a string argument 'representing the ID of the history item to be retrieved.
//
// It returns a HistoryItem object representing the retrieved history item, or an error.
func (c *Client) GetHistoryItem(itemId string) (HistoryItem, error) {
	var historyItem HistoryItem
	b := bytes.Buffer{}
	err := c.doRequest(c.ctx, &b, http.MethodGet, fmt.Sprintf("%s/history/%s", c.baseURL, itemId), &bytes.Buffer{}, contentTypeJSON)
	if err != nil {
		return HistoryItem{}, err
	}

	if err := json.Unmarshal(b.Bytes(), &historyItem); err != nil {
		return HistoryItem{}, err
	}

	return historyItem, nil
}

// DeleteHistoryItem deletes a specific history item by its ID.
//
// It takes a string argument representing the ID of the history item to be deleted.
//
// It returns nil if successful or an error otherwise.
func (c *Client) DeleteHistoryItem(itemId string) error {
	return c.doRequest(c.ctx, &bytes.Buffer{}, http.MethodDelete, fmt.Sprintf("%s/history/%s", c.baseURL, itemId), &bytes.Buffer{}, contentTypeJSON)
}

// GetHistoryItemAudio retrieves the audio data for a specific history item by its ID.
//
// It takes a string argument representing the ID of the history item for which the audio
// data is retrieved.
//
// It returns a byte slice containing the audio data or an error.
func (c *Client) GetHistoryItemAudio(itemId string) ([]byte, error) {
	b := bytes.Buffer{}
	err := c.doRequest(c.ctx, &b, http.MethodGet, fmt.Sprintf("%s/history/%s/audio", c.baseURL, itemId), &bytes.Buffer{}, contentTypeJSON)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// DownloadHistoryAudio downloads the audio data for a one or more history items.
//
// It takes a DownloadHistoryRequest argument that specifies the history item(s) to download.
//
// It returns a byte slice containing the downloaded audio data. If one history item ID was provided
// the byte slice is a mpeg encoded audio file. If multiple item IDs where provided, the byte slice
// is a zip file packing the history items' audio files.
func (c *Client) DownloadHistoryAudio(dlReq DownloadHistoryRequest) ([]byte, error) {
	reqBody, err := json.Marshal(dlReq)
	if err != nil {
		return nil, err
	}

	b := bytes.Buffer{}
	err = c.doRequest(c.ctx, &b, http.MethodPost, fmt.Sprintf("%s/history/download", c.baseURL), bytes.NewBuffer(reqBody), contentTypeJSON)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// GetSubscription retrieves the subscription details for the user.
//
// It returns a Subscription object representing the subscription details, or an error.
func (c *Client) GetSubscription() (Subscription, error) {
	sub := Subscription{}
	b := bytes.Buffer{}
	err := c.doRequest(c.ctx, &b, http.MethodGet, fmt.Sprintf("%s/user/subscription", c.baseURL), &bytes.Buffer{}, contentTypeJSON)
	if err != nil {
		return sub, err
	}

	if err := json.Unmarshal(b.Bytes(), &sub); err != nil {
		return sub, err
	}

	return sub, nil
}

// GetUser retrieves the user information.
//
// It returns a User object representing the user details, or an error.
//
// The Subscription object returned with User will not have the invoicing details populated.
// Use GetSubscription to retrieve the user's full subscription details.
func (c *Client) GetUser() (User, error) {
	user := User{}
	b := bytes.Buffer{}
	err := c.doRequest(c.ctx, &b, http.MethodGet, fmt.Sprintf("%s/user", c.baseURL), &bytes.Buffer{}, contentTypeJSON)
	if err != nil {
		return user, err
	}

	if err := json.Unmarshal(b.Bytes(), &user); err != nil {
		return user, err
	}

	return user, nil
}
