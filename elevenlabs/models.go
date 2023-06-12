package elevenlabs

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type Language struct {
	LanguageId string `json:"language_id"`
	Name       string `json:"name"`
}

type Model struct {
	CanBeFineTuned       bool       `json:"can_be_finetuned"`
	CanDoTextToSpeech    bool       `json:"can_do_text_to_speech"`
	CanDoVoiceConversion bool       `json:"can_do_voice_conversion"`
	Description          string     `json:"description"`
	Languages            []Language `json:"languages"`
	ModelId              string     `json:"model_id"`
	Name                 string     `json:"name"`
	TokenCostFactor      float32    `json:"token_cost_factor"`
}

type TextToSpeechRequest struct {
	Text          string         `json:"text"`
	ModelID       string         `json:"model_id,omitempty"`
	VoiceSettings *VoiceSettings `json:"voice_settings,omitempty"`
}

type GetVoicesResponse struct {
	Voices []Voice `json:"voices"`
}

type AddVoiceResponse struct {
	VoiceId string `json:"voice_id"`
}

type Voice struct {
	AvailableForTiers []string          `json:"available_for_tiers"`
	Category          string            `json:"category"`
	Description       string            `json:"description"`
	FineTuning        FineTuning        `json:"fine_tuning"`
	Labels            map[string]string `json:"labels"`
	Name              string            `json:"name"`
	PreviewUrl        string            `json:"preview_url"`
	Samples           []VoiceSample     `json:"samples"`
	Settings          VoiceSettings     `json:"settings,omitempty"`
	Sharing           VoiceSharing      `json:"sharing"`
	VoiceId           string            `json:"voice_id"`
}

type VoiceSettings struct {
	SimilarityBoost float32 `json:"similarity_boost"`
	Stability       float32 `json:"stability"`
}

type VoiceSharing struct {
	ClonedByCount       int    `json:"cloned_by_count"`
	HistoryItemSampleId string `json:"history_item_sample_id"`
	LikedByCount        int    `json:"liked_by_count"`
	OriginalVoiceId     string `json:"original_voice_id"`
	PublicOwnerId       string `json:"public_owner_id"`
	Status              string `json:"status"`
}

type VoiceSample struct {
	FileName  string `json:"file_name"`
	Hash      string `json:"hash"`
	MimeType  string `json:"mime_type"`
	SampleId  string `json:"sample_id"`
	SizeBytes int    `json:"size_bytes"`
}

type FineTuning struct {
	FineTuningRequested       bool                  `json:"fine_tuning_requested"`
	FineTuningState           FineTuningState       `json:"finetuning_state"`
	IsAllowedToFineTune       bool                  `json:"is_allowed_to_fine_tune"`
	Language                  string                `json:"language"`
	ModelId                   string                `json:"model_id"`
	SliceIds                  []string              `json:"slice_ids"`
	VerificationAttempts      []VerificationAttempt `json:"verification_attempts"`
	VerificationAttemptsCount int                   `json:"verification_attempts_count"`
	VerificationFailures      []string              `json:"verification_failures"`
}

type FineTuningState string

type VerificationAttempt struct {
	Accepted            bool      `json:"accepted"`
	DateUnix            int       `json:"date_unix"`
	LevenshteinDistance float32   `json:"levenshtein_distance"`
	Recording           Recording `json:"recording"`
	Similarity          float32   `json:"similarity"`
	Text                string    `json:"text"`
}

type Recording struct {
	MimeType       string `json:"mime_type"`
	RecordingId    string `json:"recording_id"`
	SizeBytes      int    `json:"size_bytes"`
	Transcription  string `json:"transcription"`
	UploadDateUnix int    `json:"upload_date_unix"`
}

type AddEditVoiceRequest struct {
	Name        string
	FilePaths   []string
	Description string
	Labels      map[string]string
}

func (r *AddEditVoiceRequest) buildRequestBody() (*bytes.Buffer, string, error) {
	var b bytes.Buffer

	w := multipart.NewWriter(&b)

	if err := w.WriteField("name", r.Name); err != nil {
		return nil, "", err
	}
	if r.Description != "" {
		if err := w.WriteField("description", r.Description); err != nil {
			return nil, "", err
		}
	}
	if len(r.Labels) > 0 {
		labelsJson, err := json.Marshal(r.Labels)
		if err != nil {
			return nil, "", err
		}
		if err := w.WriteField("labels", string(labelsJson)); err != nil {
			return nil, "", err
		}
	}

	for _, file := range r.FilePaths {
		f, err := os.Open(file)
		if err != nil {
			return nil, "", err
		}
		defer f.Close()

		fw, err := w.CreateFormFile("files", filepath.Base(file))
		if err != nil {
			return nil, "", err
		}
		if _, err = io.Copy(fw, f); err != nil {
			return nil, "", err
		}
	}

	err := w.Close()
	if err != nil {
		return nil, "", err

	}

	return &b, w.FormDataContentType(), nil
}
