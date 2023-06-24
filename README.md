# elevenlabs-go

![Go version](https://img.shields.io/badge/go-1.18-blue)
![License](https://img.shields.io/github/license/haguro/elevenlabs-go)
![Tests](https://github.com/haguro/elevenlabs-go/actions/workflows/tests.yml/badge.svg?branch=main&event=push)
[![codecov](https://codecov.io/gh/haguro/elevenlabs-go/branch/main/graph/badge.svg?token=UM33DSSTAG)](https://codecov.io/gh/haguro/elevenlabs-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/haguro/elevenlabs-go)](https://goreportcard.com/report/github.com/haguro/elevenlabs-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/haguro/elevenlabs-go.svg)](https://pkg.go.dev/github.com/haguro/elevenlabs-go)

This is a Go client library for the [ElevenLabs](https://elevenlabs.io/) voice cloning and speech synthesis platform. It provides a basic interface for Go programs to interact with the ElevenLabs [API](https://docs.elevenlabs.io/api-reference).

## Installation

```bash
go get github.com/haguro/elevenlabs-go
```

## Example Usage

Make sure to replace `"your-api-key"` in all examples with your actual API key and `"voiceID"` with the ID of the voice model you want to use. Refer to the official Elevenlabs [API documentation](https://docs.elevenlabs.io/api-reference/quick-start/introduction) for further details.

Full documentation is available [here](https://pkg.go.dev/github.com/haguro/elevenlabs-go).

### Using a New Client Instance

Using the `NewClient` function returns a new `Client` instance will allow to pass a parent context, your API key and a timeout duration.

```go
package main

import (
 "context"
 "log"
 "os"
 "time"

 "github.com/haguro/elevenlabs-go"
)

func main() {
 // Create a new client
 client := elevenlabs.NewClient(context.Background(), "your-api-key", 30*time.Second)

 // Create a TextToSpeechRequest
 ttsReq := elevenlabs.TextToSpeechRequest{
  Text:    "Hello, world! My name is Adam, nice to meet you!",
  ModelID: "eleven_monolingual_v1",
 }

 // Call the TextToSpeech method on the client, using the "Adam"'s voice ID.
 audio, err := client.TextToSpeech("pNInz6obpgDQGcFmaJgB", ttsReq)
 if err != nil {
  log.Fatal(err)
 }

 // Write the audio file bytes to disk
 if err := os.WriteFile("adam.mp3", audio, 0644); err != nil {
  log.Fatal(err)
 }

 log.Println("Successfully generated audio file")
}
```

### Using the Default Client and proxy functions

The library has a default client you can configure and use with proxy functions that wrap method calls to the default client. The default client has a default timeout set to 30 seconds and is configured with `context.Background()` as the the parent context. You will only need to set your API key at minimum when taking advantage of the default client. Here's the a version of the above example above using shorthand functions only.

```go
package main

import (
 "log"
 "os"
 "time"

 el "github.com/haguro/elevenlabs-go"
)

func main() {
 // Set your API key
 el.SetAPIKey("your-api-key")

 // Set a different timeout (optional)
 el.SetTimeout(15 * time.Second)

 // Call the TextToSpeech method on the client, using the "Adam"'s voice ID.
 audio, err := el.TextToSpeech("pNInz6obpgDQGcFmaJgB", el.TextToSpeechRequest{
   Text:    "Hello, world! My name is Adam, nice to meet you!",
   ModelID: "eleven_monolingual_v1",
  })
 if err != nil {
  log.Fatal(err)
 }

 // Write the audio file bytes to disk
 if err := os.WriteFile("adam.mp3", audio, 0644); err != nil {
  log.Fatal(err)
 }

 log.Println("Successfully generated audio file")
}
```

### Streaming

The Elevenlabs API allows streaming of audio "as it is being generated". In elevenlabs-go, you'll want to pass an `io.Writer` to the `TextToSpeechStream` method where the stream will be continuously copied to. _Note that you will need to set the client timeout to a high enough value to ensure that request does not time out mid-stream_.

```go
package main

import (
 "context"
 "log"
 "os/exec"
 "time"

 "github.com/haguro/elevenlabs-go"
)

func main() {
 message := `The concept of "flushing" typically applies to I/O buffers in many programming 
languages, which store data temporarily in memory before writing it to a more permanent location
like a file or a network connection. Flushing the buffer means writing all the buffered data
immediately, even if the buffer isn't full.`

 // Set your API key
 elevenlabs.SetAPIKey("your-api-key")

 // Set a large enough timeout to ensure the stream is not interrupted.
 elevenlabs.SetTimeout(1 * time.Minute)

 // We'll use mpv to play the audio from the stream piped to standard input
 cmd := exec.CommandContext(context.Background(), "mpv", "--no-cache", "--no-terminal", "--", "fd://0")

 // Get a pipe connected to the mpv's standard input
 pipe, err := cmd.StdinPipe()
 if err != nil {
  log.Fatal(err)
 }

 // Attempt to run the command in a separate process
 if err := cmd.Start(); err != nil {
  log.Fatal(err)
 }

 // Stream the audio to the pipe connected to mpv's standard input
 if err := elevenlabs.TextToSpeechStream(
  pipe,
  "pNInz6obpgDQGcFmaJgB",
  elevenlabs.TextToSpeechRequest{
   Text:    message,
   ModelID: "eleven_multilingual_v1",
  }); err != nil {
  log.Fatalf("Got %T error: %q\n", err, err)
 }

 // Close the pipe when all stream has been copied to the pipe
 if err := pipe.Close(); err != nil {
  log.Fatalf("Could not close pipe: %s", err)
 }
 log.Print("Streaming finished.")

 // Wait for mpv to exit. With the pipe closed, it will do that as
 // soon as it finishes playing
 if err := cmd.Wait(); err != nil {
  log.Fatal(err)
 }

 log.Print("All done.")
}
```

## Status and Future Plans

As of the time of writing (June 24, 2023), the library provides Go bindings for 100% of Elevenlabs's API methods. I do plan to add few more utility type functions should there be some need or enough request for them.

According to Elevenlabs, the API is still considered experimental and is subject to and likely to change.

## Contributing

Contributions are welcome! If you have any ideas, improvements, or bug fixes, please open an issue or submit a pull request.

## Looking for a Python library?

The Elevenlabs's official [Python library](https://github.com/elevenlabs/elevenlabs-python) is excellent and fellow Pythonistas are encourage to use it (and also to give Go, a [go](https://gobyexample.com/) 😉🩵)!

## Disclaimer

This is an independent project and is not affiliated with or endorsed by Elevenlabs. Elevenlabs and its trademarks are the property of their respective owners. The purpose of this project is to provide a client library to facilitate access to the public API provided Elevenlabs within Go programs. Any use of Elevenlabs's trademarks within this project is for identification purposes only and does not imply endorsement, sponsorship, or affiliation.

## License

This project is licensed under the [MIT License](LICENSE).

## Warranty

This code library is provided "as is" and without any warranties whatsoever. Use at your own risk. More details in the [LICENSE](LICENSE) file.
