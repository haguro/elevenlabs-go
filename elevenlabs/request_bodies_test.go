package elevenlabs_test

var testRespBodies = map[string][]byte{

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
}
