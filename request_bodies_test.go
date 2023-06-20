package elevenlabs_test

var testRespBodies = map[string][]byte{
	"TestAPIErrorOnBadRequestAndUnauthorized": []byte(`{
  "detail": {
    "status": "needs_authorization",
    "message": "Neither authorization header nor xi-api-key received, please provide one."
  }
}`),
	"TestValidationErrorOnUnprocessableEntity": []byte(`{
  "detail": [
    {
      "loc": [
        "string",
        0
      ],
      "msg": "string",
      "type": "string"
    }
  ]
}`),
	"TestGetModels": []byte(`[
	{
		"model_id": "TestModelID",
		"name": "TestModelName",
		"can_be_finetuned": true,
		"can_do_text_to_speech": true,
		"can_do_voice_conversion": true,
		"token_cost_factor": 0,
		"description": "TestModelDescription",
		"languages": [
			{
				"language_id": "LangIDEnglish",
				"name": "English"
			}
		]
	}
]`),

	"TestGetVoices": []byte(`{
  "voices": [
    {
      "voice_id": "string",
      "name": "string",
      "samples": [
        {
          "sample_id": "string",
          "file_name": "string",
          "mime_type": "string",
          "size_bytes": 0,
          "hash": "string"
        }
      ],
      "category": "string",
      "fine_tuning": {
        "model_id": "string",
        "language": "string",
        "is_allowed_to_fine_tune": true,
        "fine_tuning_requested": true,
        "finetuning_state": "not_started",
        "verification_attempts": [
          {
            "text": "string",
            "date_unix": 0,
            "accepted": true,
            "similarity": 0,
            "levenshtein_distance": 0,
            "recording": {
              "recording_id": "string",
              "mime_type": "string",
              "size_bytes": 0,
              "upload_date_unix": 0,
              "transcription": "string"
            }
          }
        ],
        "verification_failures": [
          "string"
        ],
        "verification_attempts_count": 0,
        "slice_ids": [
          "string"
        ]
      },
      "labels": {
        "additionalProp1": "string",
        "additionalProp2": "string",
        "additionalProp3": "string"
      },
      "description": "string",
      "preview_url": "string",
      "available_for_tiers": [
        "string"
      ],
      "settings": {
        "stability": 0,
        "similarity_boost": 0
      },
      "sharing": {
        "status": "string",
        "history_item_sample_id": "string",
        "original_voice_id": "string",
        "public_owner_id": "string",
        "liked_by_count": 0,
        "cloned_by_count": 0
      }
    }
  ]
}`),

	"TestGetDefaultVoiceSettings": []byte(`{
  "stability": 0.1,
  "similarity_boost": 0.2
}`),

	"TestGetVoiceSettings": []byte(`{
  "stability": 0.7,
  "similarity_boost": 0.9
}`),

	"TestGetVoice": []byte(`{
  "voice_id": "string",
  "name": "string",
  "samples": [
    {
      "sample_id": "string",
      "file_name": "string",
      "mime_type": "string",
      "size_bytes": 0,
      "hash": "string"
    }
  ],
  "category": "string",
  "fine_tuning": {
    "model_id": "string",
    "language": "string",
    "is_allowed_to_fine_tune": true,
    "fine_tuning_requested": true,
    "finetuning_state": "not_started",
    "verification_attempts": [
      {
        "text": "string",
        "date_unix": 0,
        "accepted": true,
        "similarity": 0,
        "levenshtein_distance": 0,
        "recording": {
          "recording_id": "string",
          "mime_type": "string",
          "size_bytes": 0,
          "upload_date_unix": 0,
          "transcription": "string"
        }
      }
    ],
    "verification_failures": [
      "string"
    ],
    "verification_attempts_count": 0,
    "slice_ids": [
      "string"
    ]
  },
  "labels": {
    "additionalProp1": "string",
    "additionalProp2": "string",
    "additionalProp3": "string"
  },
  "description": "string",
  "preview_url": "string",
  "available_for_tiers": [
    "string"
  ],
  "settings": {
    "stability": 0.3,
    "similarity_boost": 0.7
  },
  "sharing": {
    "status": "string",
    "history_item_sample_id": "string",
    "original_voice_id": "string",
    "public_owner_id": "string",
    "liked_by_count": 0,
    "cloned_by_count": 0
  }
}`),
	"TestGetSampleAudio": []byte("testaudiobytes"),
	"TestTextToSpeech":   []byte("testaudiobytes"),
	"TestGetHistory-NoMore": []byte(`{
  "history":[],
  "last_history_item_id": "",
  "has_more": false
}`),
	"TestGetHistory-HasMore": []byte(`{
  "history":[],
  "last_history_item_id": "fake-history-id",
  "has_more": true
}`),
	"TestGetHistoryItem": []byte(`{
  "history_item_id": "TestHistoryItemID",
  "request_id": "string",
  "voice_id": "string",
  "voice_name": "string",
  "text": "string",
  "date_unix": 0,
  "character_count_change_from": 0,
  "character_count_change_to": 0,
  "content_type": "string",
  "state": "created",
  "settings": {},
  "feedback": {
    "thumbs_up": true,
    "feedback": "string",
    "emotions": true,
    "inaccurate_clone": true,
    "glitches": true,
    "audio_quality": true,
    "other": true,
    "review_status": "not_reviewed"
  }
}`),
	"TestDownloadHistoryAudio": []byte("testhistoryitemaudiobytes"),
	"TestGetSubscription": []byte(`
{
  "tier": "string",
  "character_count": 0,
  "character_limit": 0,
  "can_extend_character_limit": true,
  "allowed_to_extend_character_limit": true,
  "next_character_count_reset_unix": 0,
  "voice_limit": 0,
  "professional_voice_limit": 0,
  "can_extend_voice_limit": true,
  "can_use_instant_voice_cloning": true,
  "can_use_professional_voice_cloning": true,
  "currency": "usd",
  "status": "trialing",
  "next_invoice": {
    "amount_due_cents": 0,
    "next_payment_attempt_unix": 0
  },
  "has_open_invoices": true
}`),
	"TestGetUser": []byte(`{
  "subscription": {
    "tier": "string",
    "character_count": 0,
    "character_limit": 0,
    "can_extend_character_limit": true,
    "allowed_to_extend_character_limit": true,
    "next_character_count_reset_unix": 0,
    "voice_limit": 0,
    "professional_voice_limit": 0,
    "can_extend_voice_limit": true,
    "can_use_instant_voice_cloning": true,
    "can_use_professional_voice_cloning": true,
    "currency": "usd",
    "status": "trialing"
  },
  "is_new_user": true,
  "xi_api_key": "string",
  "can_use_delayed_payment_methods": true
}`),
}
