// Code generated by "cmd/codegen/main.go". run `go generate` when adding new `Client` methods; DO NOT EDIT.

package elevenlabs

func TextToSpeech(voiceID string, ttsReq TextToSpeechRequest, queries ...QueryFunc) ([]byte, error) {
	return getDefaultClient().TextToSpeech(voiceID, ttsReq, queries...)
}

func GetModels() ([]Model, error) {
	return getDefaultClient().GetModels()
}

func GetVoices() ([]Voice, error) {
	return getDefaultClient().GetVoices()
}

func GetDefaultVoiceSettings() (VoiceSettings, error) {
	return getDefaultClient().GetDefaultVoiceSettings()
}

func GetVoiceSettings(voiceId string) (VoiceSettings, error) {
	return getDefaultClient().GetVoiceSettings(voiceId)
}

func GetVoice(voiceId string, queries ...QueryFunc) (Voice, error) {
	return getDefaultClient().GetVoice(voiceId, queries...)
}

func DeleteVoice(voiceId string) error {
	return getDefaultClient().DeleteVoice(voiceId)
}

func EditVoiceSettings(voiceId string, settings VoiceSettings) error {
	return getDefaultClient().EditVoiceSettings(voiceId, settings)
}

func AddVoice(voiceReq AddEditVoiceRequest) (string, error) {
	return getDefaultClient().AddVoice(voiceReq)
}

func EditVoice(voiceId string, voiceReq AddEditVoiceRequest) error {
	return getDefaultClient().EditVoice(voiceId, voiceReq)
}

func DeleteSample(voiceId, sampleId string) error {
	return getDefaultClient().DeleteSample(voiceId, sampleId)
}

func GetSampleAudio(voiceId, sampleId string) ([]byte, error) {
	return getDefaultClient().GetSampleAudio(voiceId, sampleId)
}

func GetHistory(queries ...QueryFunc) (GetHistoryResponse, NextHistoryPageFunc, error) {
	return getDefaultClient().GetHistory(queries...)
}

func GetHistoryItem(itemId string) (HistoryItem, error) {
	return getDefaultClient().GetHistoryItem(itemId)
}

func DeleteHistoryItem(itemId string) error {
	return getDefaultClient().DeleteHistoryItem(itemId)
}

func GetHistoryItemAudio(itemId string) ([]byte, error) {
	return getDefaultClient().GetHistoryItemAudio(itemId)
}

func DownloadHistoryAudio(dlReq DownloadHistoryRequest) ([]byte, error) {
	return getDefaultClient().DownloadHistoryAudio(dlReq)
}
