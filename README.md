# elevenlabs-go

![Go version](https://img.shields.io/badge/go-1.18-blue)
![License](https://img.shields.io/github/license/haguro/elevenlabs-go)
![Tests](https://github.com/haguro/elevenlabs-go/actions/workflows/tests.yml/badge.svg?branch=main&event=push)
[![codecov](https://codecov.io/gh/haguro/elevenlabs-go/branch/main/graph/badge.svg?token=UM33DSSTAG)](https://codecov.io/gh/haguro/elevenlabs-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/haguro/elevenlabs-go)](https://goreportcard.com/report/github.com/haguro/elevenlabs-go)

This is a Go client library for the ElevenLabs text-to-speech platform. It provides a simple interface to interact with the ElevenLabs API. *Please note: This library is very much a work in progress and far from ready or complete.*

## Installation

```bash
go get github.com/haguro/elevenlabs-go
```

## Example Usage

### Using the Default Client (Shorthand)

The library has a default client you can configure and use as a shorthand. The default client has the timeout set to 30 seconds by default and is configured with `context.Background()` as the the parent context.

```go
package main

import (
 "fmt"
 "github.com/haguro/elevenlabs-go/elevenlabs"
)

func main() {
 // Set your API key
 elevenlabs.SetAPIKey("your-api-key")

 // Set a different timeout (optional)
 elevenlabs.SetTimeout(5 * time.Second)

 // Call the TextToSpeech function
 audio, err := elevenlabs.TextToSpeech("voiceID",
  elevenlabs.TextToSpeechRequest{
   Text: "Hello, world!",
 })
 if err != nil {
  fmt.Println(err)
  os.Exit(1)
 }

 if err := os.WriteFile("out.mp3", audio, 0644); err != nil {
  fmt.Println(err)
  os.Exit(1)
 }

 fmt.Println("Successfully generated audio file out.mp3")
}
```

### Using a New Client Instance

Using the `NewClient` method to instantiate a new `Client` instance will allow to pass your own context.

```go
package main

import (
 "context"
 "time"
 "github.com/haguro/elevenlabs-go/elevenlabs"
)

func main() {
 // Create a new client
 client := elevenlabs.NewClient(context.Background(), "your-api-key", 30*time.Second)

 // Create a TextToSpeechRequest
 ttsReq := elevenlabs.TextToSpeechRequest{
  Text: "Hello, world!",
 }

 // Call the TextToSpeech method on the client
 audio, err := client.TextToSpeech("voiceID", ttsReq)
 if err != nil {
  fmt.Println(err)
  os.Exit(1)
 }

 if err := os.WriteFile("out.mp3", audio, 0644); err != nil {
  fmt.Println(err)
  os.Exit(1)
 }

 fmt.Println("Successfully generated audio file out.mp3")
}
```

In both examples, replace `"your-api-key"` with your actual API key and `"voiceID"` with the ID of the voice model you want to use. The `TextToSpeechRequest` struct should be filled with the text you want to convert to speech. Refer to the official Elevenlabs [API documentation](https://docs.elevenlabs.io/api-reference/quick-start/introduction) for further details.

## Contributing

Contributions are welcome! If you have any ideas, improvements, or bug fixes, please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).
