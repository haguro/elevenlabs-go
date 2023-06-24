package elevenlabs_test

import (
	"context"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/haguro/elevenlabs-go"
)

func ExampleClient_TextToSpeech() {
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
func ExampleClient_TextToSpeechStream() {
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
