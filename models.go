package elevenlabs

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	CanBeFineTuned                     bool       `json:"can_be_finetuned"`
	CanDoTextToSpeech                  bool       `json:"can_do_text_to_speech"`
	CanDoVoiceConversion               bool       `json:"can_do_voice_conversion"`
	CanUseSpeakerBoost                 bool       `json:"can_use_speaker_boost"`
	CanUseStyle                        bool       `json:"can_use_style"`
	Description                        string     `json:"description"`
	Languages                          []Language `json:"languages"`
	MaxCharactersRequestFreeUser       int        `json:"max_characters_request_free_user"`
	MaxCharactersRequestSubscribedUser int        `json:"max_characters_request_subscribed_user"`
	ModelId                            string     `json:"model_id"`
	Name                               string     `json:"name"`
	RequiresAlphaAccess                bool       `json:"requires_alpha_access"`
	ServesProVoices                    bool       `json:"serves_pro_voices"`
	TokenCostFactor                    float32    `json:"token_cost_factor"`
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
	AvailableForTiers       []string          `json:"available_for_tiers"`
	Category                string            `json:"category"`
	Description             string            `json:"description"`
	FineTuning              FineTuning        `json:"fine_tuning"`
	HighQualityBaseModelIds []string          `json:"high_quality_base_model_ids"`
	Labels                  map[string]string `json:"labels"`
	Name                    string            `json:"name"`
	PreviewUrl              string            `json:"preview_url"`
	Samples                 []VoiceSample     `json:"samples"`
	Settings                VoiceSettings     `json:"settings,omitempty"`
	Sharing                 VoiceSharing      `json:"sharing"`
	VoiceId                 string            `json:"voice_id"`
}

type VoiceSettings struct {
	SimilarityBoost float32 `json:"similarity_boost"`
	Stability       float32 `json:"stability"`
	Style           float32 `json:"style,omitempty"`
	SpeakerBoost    bool    `json:"use_speaker_boost,omitempty"`
}

type VoiceSharing struct {
	ClonedByCount          int               `json:"cloned_by_count"`
	DateUnix               int               `json:"date_unix"`
	Description            string            `json:"description"`
	DisableAtUnix          bool              `json:"disable_at_unix"`
	EnabledInLibrary       bool              `json:"enabled_in_library"`
	FinancialRewardEnabled bool              `json:"financial_reward_enabled"`
	FreeUsersAllowed       bool              `json:"free_users_allowed"`
	HistoryItemSampleId    string            `json:"history_item_sample_id"`
	Labels                 map[string]string `json:"labels"`
	LikedByCount           int               `json:"liked_by_count"`
	LiveModerationEnabled  bool              `json:"live_moderation_enabled"`
	Name                   string            `json:"name"`
	NoticePeriod           int               `json:"notice_period"`
	OriginalVoiceId        string            `json:"original_voice_id"`
	PublicOwnerId          string            `json:"public_owner_id"`
	Rate                   float32           `json:"rate"`
	ReviewMessage          string            `json:"review_message"`
	ReviewStatus           string            `json:"review_status"`
	Status                 string            `json:"status"`
	VoiceMixingAllowed     bool              `json:"voice_mixing_allowed"`
	WhitelistedEmails      []string          `json:"whitelisted_emails"`
}

type VoiceSample struct {
	FileName  string `json:"file_name"`
	Hash      string `json:"hash"`
	MimeType  string `json:"mime_type"`
	SampleId  string `json:"sample_id"`
	SizeBytes int    `json:"size_bytes"`
}

type FineTuning struct {
	FineTuningRequested         bool                  `json:"fine_tuning_requested"`
	FineTuningState             string                `json:"finetuning_state"`
	IsAllowedToFineTune         bool                  `json:"is_allowed_to_fine_tune"`
	Language                    string                `json:"language"`
	ManualVerification          ManualVerification    `json:"manual_verification"`
	ManualVerificationRequested bool                  `json:"manual_verification_requested"`
	SliceIds                    []string              `json:"slice_ids"`
	VerificationAttempts        []VerificationAttempt `json:"verification_attempts"`
	VerificationAttemptsCount   int                   `json:"verification_attempts_count"`
	VerificationFailures        []string              `json:"verification_failures"`
}

type ManualVerification struct {
	ExtraText       string `json:"extra_text"`
	Files           []File `json:"files"`
	RequestTimeUnix int    `json:"request_time_unix"`
}

type File struct {
	FileId         string `json:"file_id"`
	FileName       string `json:"file_name"`
	MimeType       string `json:"mime_type"`
	SizeBytes      int    `json:"size_bytes"`
	UploadDateUnix int    `json:"upload_date_unix"`
}

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

type DownloadHistoryRequest struct {
	HistoryItemIds []string `json:"history_item_ids"`
}

type GetHistoryResponse struct {
	History           []HistoryItem `json:"history"`
	LastHistoryItemId string        `json:"last_history_item_id"`
	HasMore           bool          `json:"has_more"`
}

type HistoryItem struct {
	CharacterCountChangeFrom int           `json:"character_count_change_from"`
	CharacterCountChangeTo   int           `json:"character_count_change_to"`
	ContentType              string        `json:"content_type"`
	DateUnix                 int           `json:"date_unix"`
	Feedback                 Feedback      `json:"feedback"`
	HistoryItemId            string        `json:"history_item_id"`
	ModelId                  string        `json:"model_id"`
	RequestId                string        `json:"request_id"`
	Settings                 VoiceSettings `json:"settings"`
	ShareLinkId              string        `json:"share_link_id"`
	State                    string        `json:"state"`
	Text                     string        `json:"text"`
	VoiceCategory            string        `json:"voice_category"`
	VoiceId                  string        `json:"voice_id"`
	VoiceName                string        `json:"voice_name"`
}

type Feedback struct {
	AudioQuality    bool    `json:"audio_quality"`
	Emotions        bool    `json:"emotions"`
	Feedback        string  `json:"feedback"`
	Glitches        bool    `json:"glitches"`
	InaccurateClone bool    `json:"inaccurate_clone"`
	Other           bool    `json:"other"`
	ReviewStatus    *string `json:"review_status,omitempty"`
	ThumbsUp        bool    `json:"thumbs_up"`
}

type Subscription struct {
	AllowedToExtendCharacterLimit  bool    `json:"allowed_to_extend_character_limit"`
	CanExtendCharacterLimit        bool    `json:"can_extend_character_limit"`
	CanExtendVoiceLimit            bool    `json:"can_extend_voice_limit"`
	CanUseInstantVoiceCloning      bool    `json:"can_use_instant_voice_cloning"`
	CanUseProfessionalVoiceCloning bool    `json:"can_use_professional_voice_cloning"`
	CharacterCount                 int     `json:"character_count"`
	CharacterLimit                 int     `json:"character_limit"`
	Currency                       string  `json:"currency"`
	NextCharacterCountResetUnix    int     `json:"next_character_count_reset_unix"`
	VoiceLimit                     int     `json:"voice_limit"`
	ProfessionalVoiceLimit         int     `json:"professional_voice_limit"`
	Status                         string  `json:"status"`
	Tier                           string  `json:"tier"`
	MaxVoiceAddEdits               int     `json:"max_voice_add_edits"`
	VoiceAddEditCounter            int     `json:"voice_add_edit_counter"`
	HasOpenInvoices                bool    `json:"has_open_invoices"`
	NextInvoice                    Invoice `json:"next_invoice"`
	withInvoicingDetails           bool
}

type Invoice struct {
	AmountDueCents         int `json:"amount_due_cents"`
	NextPaymentAttemptUnix int `json:"next_payment_attempt_unix"`
}

type User struct {
	Subscription                Subscription `json:"subscription"`
	FirstName                   string       `json:"first_name,omitempty"`
	IsNewUser                   bool         `json:"is_new_user"`
	IsOnboardingComplete        bool         `json:"is_onboarding_complete"`
	XiApiKey                    string       `json:"xi_api_key"`
	CanUseDelayedPaymentMethods bool         `json:"can_use_delayed_payment_methods"`
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
	buildFailed := func(err error) (*bytes.Buffer, string, error) {
		return nil, "", fmt.Errorf("failed to build request body: %w", err)
	}

	if err := w.WriteField("name", r.Name); err != nil {
		return buildFailed(err)
	}
	if r.Description != "" {
		if err := w.WriteField("description", r.Description); err != nil {
			return buildFailed(err)
		}
	}
	if len(r.Labels) > 0 {
		labelsJson, err := json.Marshal(r.Labels)
		if err != nil {
			return buildFailed(err)
		}
		if err := w.WriteField("labels", string(labelsJson)); err != nil {
			return buildFailed(err)
		}
	}

	for _, file := range r.FilePaths {
		f, err := os.Open(file)
		if err != nil {
			return buildFailed(err)
		}
		defer f.Close()

		fw, err := w.CreateFormFile("files", filepath.Base(file))
		if err != nil {
			return buildFailed(err)
		}
		if _, err = io.Copy(fw, f); err != nil {
			return buildFailed(err)
		}
	}

	err := w.Close()
	if err != nil {
		return buildFailed(err)
	}

	return &b, w.FormDataContentType(), nil
}
