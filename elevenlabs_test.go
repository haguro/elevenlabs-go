package elevenlabs_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/haguro/elevenlabs-go"
)

const (
	mockAPIKey       = "MockAPIKey"
	mockTimeout      = 60 * time.Second
	contentTypeJSON  = "application/json"
	contentMultipart = "multipart/form-data"
)

type testServerConfig struct {
	KeyOptional         bool
	expectedMethod      string
	expectedContentType string
	expectedAccept      string
	expectedQueryStr    string
	statusCode          int
	responseBody        []byte
	returnErr           error
	responseDelay       time.Duration
}

func testServer(t *testing.T, config testServerConfig) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !config.KeyOptional {
			gotAPIKey := r.Header.Get("xi-api-key")
			if gotAPIKey != mockAPIKey {
				t.Errorf("Server: expected API Key %q, got %q", mockAPIKey, gotAPIKey)
			}
		}

		if r.Method != config.expectedMethod {
			t.Errorf("Server: expected HTTP Method to be %q, got %q", config.expectedMethod, r.Method)
		}

		if config.expectedContentType != "" {
			if !strings.Contains(r.Header.Get("Content-Type"), config.expectedContentType) {
				t.Errorf("Server: expected Content-Type %q to contain %q", r.Header.Get("Content-Type"), config.expectedContentType)
			}
			w.Header().Add("Content-Type", config.expectedContentType)
		}

		if config.expectedAccept != "" {
			if r.Header.Get("Accept") != config.expectedAccept {
				t.Errorf("Server: expected Accept header %q, got %q", config.expectedAccept, r.Header.Get("Accept"))
			}
		}

		if config.expectedQueryStr != "" {
			gotQueryStr := r.URL.RawQuery
			if gotQueryStr != config.expectedQueryStr {
				t.Errorf("Server: expected query string %q, got %q", config.expectedQueryStr, gotQueryStr)
			}
		}

		if config.responseDelay > 0 {
			time.Sleep(config.responseDelay)
		}

		w.WriteHeader(config.statusCode)
		if config.returnErr != nil {
			b, err := json.Marshal(config.returnErr)
			if err != nil {
				t.Fatal("Failed to marshal returnErr")
			}
			w.Write(b)
			return
		}
		w.Write(config.responseBody)
	}))
}

func TestDefaultClientSetup(t *testing.T) {
	baseURL := "http://localhost:1234/"
	defaultClient := elevenlabs.MockDefaultClient(baseURL)
	elevenlabs.SetAPIKey(mockAPIKey)
	elevenlabs.SetTimeout(mockTimeout)
	expected := elevenlabs.NewMockClient(context.Background(), baseURL, mockAPIKey, mockTimeout)
	if !reflect.DeepEqual(expected, defaultClient) {
		t.Errorf("Default client set up is incorrect %+v", defaultClient)
	}
}

func TestRequestTimeout(t *testing.T) {
	t.Parallel()
	server := testServer(t, testServerConfig{
		expectedMethod:      http.MethodPost,
		expectedContentType: contentTypeJSON,
		expectedAccept:      "application/json",
		statusCode:          http.StatusOK,
		responseDelay:       500 * time.Millisecond,
	})
	defer server.Close()
	client := elevenlabs.NewMockClient(context.Background(), server.URL, mockAPIKey, 100*time.Millisecond)
	_, err := client.TextToSpeech("TestVoiceID", elevenlabs.TextToSpeechRequest{})
	if err == nil {
		t.Fatalf("Expected context deadline exceeded error returned, got nil")
	}
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context deadline exceeded error returned, got err")
	}
}
func TestAPIErrorOnBadRequestAndUnauthorized(t *testing.T) {
	for _, code := range [2]int{http.StatusBadRequest, http.StatusUnauthorized} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			server := testServer(t, testServerConfig{
				expectedMethod:      http.MethodGet,
				expectedContentType: contentTypeJSON,
				expectedAccept:      "application/json",
				statusCode:          code,
				returnErr:           &elevenlabs.APIError{},
			})
			defer server.Close()
			client := elevenlabs.NewMockClient(context.Background(), server.URL, mockAPIKey, mockTimeout)
			_, err := client.GetModels()
			if err == nil {
				t.Errorf("Expected error of type %T with status code %d, got nil", &elevenlabs.APIError{}, code)
				return
			}
			if _, ok := err.(*elevenlabs.APIError); !ok {
				t.Errorf("Expected error of type %T with status code %d, got %T: %q", &elevenlabs.APIError{}, code, err, err)
			}
		})
	}
}

func TestValidationErrorOnUnprocessableEntity(t *testing.T) {
	server := testServer(t, testServerConfig{
		expectedMethod:      http.MethodPost,
		expectedContentType: contentTypeJSON,
		expectedAccept:      "application/json",
		statusCode:          http.StatusUnprocessableEntity,
		returnErr:           &elevenlabs.ValidationError{},
	})
	defer server.Close()
	client := elevenlabs.NewMockClient(context.Background(), server.URL, mockAPIKey, mockTimeout)
	_, err := client.TextToSpeech("TestVoiceID", elevenlabs.TextToSpeechRequest{})
	if err == nil {
		t.Errorf("Expected error of type %T, got nil", &elevenlabs.ValidationError{})
		return
	}
	if _, ok := err.(*elevenlabs.ValidationError); !ok {
		t.Errorf("Expected error of type %T, got %T: %q", &elevenlabs.ValidationError{}, err, err)
	}
}

func TestErrorOnUnexpectedStatusCode(t *testing.T) {
	server := testServer(t, testServerConfig{
		expectedMethod:      http.MethodPost,
		expectedContentType: contentTypeJSON,
		expectedAccept:      "application/json",
		statusCode:          http.StatusInternalServerError,
	})
	defer server.Close()
	client := elevenlabs.NewMockClient(context.Background(), server.URL, mockAPIKey, mockTimeout)
	_, err := client.TextToSpeech("TestVoiceID", elevenlabs.TextToSpeechRequest{})
	if err == nil {
		t.Error("Expected error, got nil")
		return
	}
	want := "unexpected HTTP status"
	if !strings.Contains(err.Error(), want) {
		t.Errorf("Expected error to contain %q, got %q", want, err)
	}
}

func TestTextToSpeech(t *testing.T) {
	testCases := []struct {
		name               string
		excludeAPIKey      bool
		queries            []elevenlabs.QueryFunc
		expQueryString     string
		testRequestBody    any
		expResponseBody    []byte
		expectedRespStatus int
	}{
		{
			name:          "No API key and no queries",
			excludeAPIKey: true,
			testRequestBody: elevenlabs.TextToSpeechRequest{
				ModelID: "model1",
				Text:    "Test text",
			},
			expResponseBody:    testRespBodies["TestTextToSpeech"],
			expectedRespStatus: http.StatusOK,
		},
		{
			name:           "With API key and latency optimizations query",
			excludeAPIKey:  false,
			queries:        []elevenlabs.QueryFunc{elevenlabs.LatencyOptimizations(2)},
			expQueryString: "optimize_streaming_latency=2",
			testRequestBody: elevenlabs.TextToSpeechRequest{
				ModelID: "model1",
				Text:    "Test text",
			},
			expResponseBody:    testRespBodies["TestTextToSpeech"],
			expectedRespStatus: http.StatusOK,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestAPIKey := mockAPIKey
			if tc.excludeAPIKey {
				requestAPIKey = ""
			}
			server := testServer(t, testServerConfig{
				KeyOptional:         tc.excludeAPIKey,
				expectedMethod:      http.MethodPost,
				expectedContentType: contentTypeJSON,
				expectedQueryStr:    tc.expQueryString,
				statusCode:          tc.expectedRespStatus,
				responseBody:        tc.expResponseBody,
			})
			defer server.Close()

			client := elevenlabs.NewMockClient(context.Background(), server.URL, requestAPIKey, mockTimeout)
			respBody, err := client.TextToSpeech("voiceID", tc.testRequestBody.(elevenlabs.TextToSpeechRequest), tc.queries...)

			if err != nil {
				t.Errorf("Expected no errors, got error: %q", err)
			}

			if string(respBody) != string(tc.expResponseBody) {
				t.Errorf("Expected response %q, got %q", string(tc.expResponseBody), string(respBody))
			}
		})
	}
}

func TestGetModels(t *testing.T) {
	respBody := testRespBodies["TestGetModels"]
	server := testServer(t, testServerConfig{
		expectedMethod:      http.MethodGet,
		expectedContentType: contentTypeJSON,
		expectedAccept:      "application/json",
		statusCode:          http.StatusOK,
		responseBody:        respBody,
	})
	defer server.Close()
	client := elevenlabs.NewMockClient(context.Background(), server.URL, mockAPIKey, mockTimeout)
	models, err := client.GetModels()
	if err != nil {
		t.Errorf("Expected no errors from `GetModels`, got \"%T\" error: %q", err, err)
	}
	if len(models) != 1 {
		t.Fatalf("Expected unmarshalled response to contain exactly one model, got %d", len(models))
	}
	var expModels []elevenlabs.Model
	if err := json.Unmarshal(respBody, &expModels); err != nil {
		t.Fatalf("Failed to unmarshal test respBody: %s", err)
	}
	if !reflect.DeepEqual(expModels, models) {
		t.Errorf("Unexpected Model in response: %+v", models[0])
	}
}

func TestGetVoices(t *testing.T) {
	respBody := testRespBodies["TestGetVoices"]
	server := testServer(t, testServerConfig{
		expectedMethod:      http.MethodGet,
		expectedContentType: contentTypeJSON,
		expectedAccept:      "application/json",
		statusCode:          http.StatusOK,
		responseBody:        respBody,
	})
	defer server.Close()
	client := elevenlabs.NewMockClient(context.Background(), server.URL, mockAPIKey, mockTimeout)
	voices, err := client.GetVoices()
	if err != nil {
		t.Errorf("Expected no errors from `GetVoices`, got \"%T\" error: %q", err, err)
	}
	if len(voices) != 1 {
		t.Fatalf("Expected unmarshalled response to contain exactly one model, got %d", len(voices))
	}
	var voicesResp elevenlabs.GetVoicesResponse
	if err := json.Unmarshal(respBody, &voicesResp); err != nil {
		t.Fatalf("Failed to unmarshal test respBody: %s", err)
	}
	if !reflect.DeepEqual(voicesResp.Voices, voices) {
		t.Errorf("Unexpected Voice in response: %+v", voices[0])
	}
}

func TestGetDefaultVoiceSettings(t *testing.T) {
	respBody := testRespBodies["TestGetDefaultVoiceSettings"]
	server := testServer(t, testServerConfig{
		expectedMethod:      http.MethodGet,
		expectedContentType: contentTypeJSON,
		expectedAccept:      "application/json",
		statusCode:          http.StatusOK,
		responseBody:        respBody,
	})
	defer server.Close()
	client := elevenlabs.NewMockClient(context.Background(), server.URL, mockAPIKey, mockTimeout)
	vSettings, err := client.GetDefaultVoiceSettings()
	if err != nil {
		t.Errorf("Expected no errors from `GetDefaultVoiceSettings`, got \"%T\" error: %q", err, err)
	}
	var expSettings elevenlabs.VoiceSettings
	if err := json.Unmarshal(respBody, &expSettings); err != nil {
		t.Fatalf("Failed to unmarshal test respBody: %s", err)
	}
	if !reflect.DeepEqual(expSettings, vSettings) {
		t.Errorf("Unexpected VoiceSettings in response: %+v", vSettings)
	}
}

func TestGetVoiceSettings(t *testing.T) {
	respBody := testRespBodies["TestGetVoiceSettings"]
	server := testServer(t, testServerConfig{
		expectedMethod:      http.MethodGet,
		expectedContentType: contentTypeJSON,
		expectedAccept:      "application/json",
		statusCode:          http.StatusOK,
		responseBody:        respBody,
	})
	defer server.Close()
	client := elevenlabs.NewMockClient(context.Background(), server.URL, mockAPIKey, mockTimeout)
	vSettings, err := client.GetVoiceSettings("TestVoiceID")
	if err != nil {
		t.Errorf("Expected no errors from `GetVoiceSettings`, got \"%T\" error: %q", err, err)
	}
	var expSettings elevenlabs.VoiceSettings
	if err := json.Unmarshal(respBody, &expSettings); err != nil {
		t.Fatalf("Failed to unmarshal test respBody: %s", err)
	}
	if !reflect.DeepEqual(expSettings, vSettings) {
		t.Errorf("Unexpected VoiceSettings in response: %+v", vSettings)
	}
}

func TestGetVoice(t *testing.T) {
	respBody := testRespBodies["TestGetVoice"]
	testCases := []struct {
		name           string
		queries        []elevenlabs.QueryFunc
		expQueryString string
	}{
		{
			name: "No queries",
		},
		{
			name:           "With settings query",
			queries:        []elevenlabs.QueryFunc{elevenlabs.WithSettings()},
			expQueryString: "with_settings=true",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := testServer(t, testServerConfig{
				expectedMethod:      http.MethodGet,
				expectedContentType: contentTypeJSON,
				expectedAccept:      "application/json",
				statusCode:          http.StatusOK,
				responseBody:        []byte(respBody),
			})
			defer server.Close()
			client := elevenlabs.NewMockClient(context.Background(), server.URL, mockAPIKey, mockTimeout)
			voice, err := client.GetVoice("TestVoiceID", tc.queries...)
			if err != nil {
				t.Errorf("Expected no errors from `GetVoice`, got \"%T\" error: %q", err, err)
			}
			var expVoice elevenlabs.Voice
			if err := json.Unmarshal([]byte(respBody), &expVoice); err != nil {
				t.Fatalf("Failed to unmarshal test respBody: %s", err)
			}
			if !reflect.DeepEqual(expVoice, voice) {
				t.Errorf("Unexpected Voice in response: %+v", voice)
			}
		})
	}
}

func TestDeleteVoice(t *testing.T) {
	server := testServer(t, testServerConfig{
		expectedMethod:      http.MethodDelete,
		expectedContentType: contentTypeJSON,
		expectedAccept:      "application/json",
		statusCode:          http.StatusOK,
	})
	defer server.Close()
	client := elevenlabs.NewMockClient(context.Background(), server.URL, mockAPIKey, mockTimeout)
	err := client.DeleteVoice("TestVoiceID")
	if err != nil {
		t.Errorf("Expected no errors from `DeleteVoice`, got \"%T\" error: %q", err, err)
	}
}

func TestEditVoiceSettings(t *testing.T) {
	server := testServer(t, testServerConfig{
		expectedMethod:      http.MethodPost,
		expectedContentType: contentTypeJSON,
		expectedAccept:      "application/json",
		statusCode:          http.StatusOK,
	})
	defer server.Close()
	client := elevenlabs.NewMockClient(context.Background(), server.URL, mockAPIKey, mockTimeout)
	err := client.EditVoiceSettings("voiceID", elevenlabs.VoiceSettings{Stability: 0.2, SimilarityBoost: 0.7})
	if err != nil {
		t.Errorf("Expected no errors, got error: %q", err)
	}
}

func TestAddVoice(t *testing.T) {
	testCases := []struct {
		name        string
		paths       []string
		expRespBody []byte
		expError    bool
	}{
		{
			name:        "with existing sample file",
			paths:       []string{"testdata/fake.mp3"},
			expRespBody: []byte(`{"voice_id":"TestVoiceId"}`),
			expError:    false,
		},
		{
			name:        "with non-existant sample file",
			paths:       []string{"testdata/not-there.mp3"},
			expRespBody: []byte("{}"),
			expError:    true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := elevenlabs.AddEditVoiceRequest{
				Name:        "NewTestVoiceName",
				FilePaths:   tc.paths,
				Description: "New voice description here",
				Labels:      map[string]string{"accent": "australian", "foo": "bar"},
			}
			server := testServer(t, testServerConfig{
				expectedMethod:      http.MethodPost,
				expectedContentType: contentMultipart,
				expectedAccept:      "application/json",
				statusCode:          http.StatusOK,
				responseBody:        tc.expRespBody,
			})
			defer server.Close()
			client := elevenlabs.NewMockClient(context.Background(), server.URL, mockAPIKey, mockTimeout)
			id, err := client.AddVoice(request)
			if err != nil {
				if !tc.expError {
					t.Errorf("Expected no errors, got error: %q", err)
				}
				return
			}
			if tc.expError {
				t.Error("Expected an error, got nil")
				return
			}

			if id != "TestVoiceId" {
				t.Errorf("Expected AddVoice to return voice ID %q, got %q", "TestVoiceId", id)
			}

		})
	}

}

func TestEditVoice(t *testing.T) {
	server := testServer(t, testServerConfig{
		expectedMethod:      http.MethodPost,
		expectedContentType: contentMultipart,
		expectedAccept:      "application/json",
		statusCode:          http.StatusOK,
	})
	defer server.Close()
	client := elevenlabs.NewMockClient(context.Background(), server.URL, mockAPIKey, mockTimeout)
	err := client.EditVoice("TestVoiceID", elevenlabs.AddEditVoiceRequest{Name: "TestVoice"})
	if err != nil {
		t.Errorf("Expected no errors, got error: %q", err)
	}
}

func TestDeleteSample(t *testing.T) {
	server := testServer(t, testServerConfig{
		expectedMethod:      http.MethodDelete,
		expectedContentType: contentTypeJSON,
		expectedAccept:      "application/json",
		statusCode:          http.StatusOK,
	})
	defer server.Close()
	client := elevenlabs.NewMockClient(context.Background(), server.URL, mockAPIKey, mockTimeout)
	err := client.DeleteSample("TestVoiceID", "TestSampleID")
	if err != nil {
		t.Errorf("Expected no errors from `DeleteSample`, got \"%T\" error: %q", err, err)
	}
}

func TestGetSampleAudio(t *testing.T) {
	expRespBody := testRespBodies["TestGetSampleAudio"]

	server := testServer(t, testServerConfig{
		expectedMethod:      http.MethodGet,
		expectedContentType: contentTypeJSON,
		expectedAccept:      "application/json",
		statusCode:          http.StatusOK,
		responseBody:        []byte(expRespBody),
	})
	defer server.Close()
	client := elevenlabs.NewMockClient(context.Background(), server.URL, mockAPIKey, mockTimeout)
	respBody, err := client.GetSampleAudio("TestVoiceID", "TestSampleID")
	if err != nil {
		t.Errorf("Expected no errors from `GetSampleAudio`, got \"%T\" error: %q", err, err)
	}
	if string(respBody) != string(expRespBody) {
		t.Errorf("Expected response %q, got %q", string(expRespBody), string(respBody))
	}
}
