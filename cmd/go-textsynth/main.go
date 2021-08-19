package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/AlecAivazis/survey/v2"
	"github.com/inancgumus/screen"

	"github.com/rany2/go-textsynth/pkg/normalizenewlines"
	"github.com/rany2/go-textsynth/pkg/windowsnewlines"

	"golang.org/x/term"
)

// Create HTTP transports to share pool of connections
var tr = http.DefaultTransport.(*http.Transport).Clone()
var client = &http.Client{Transport: tr}

// PromptMaxSize sets the max prompt size limit to send to the API
const PromptMaxSize = 4095

// SeedLimit sets seed limit, anything over that limit causes the API to return an error
const SeedLimit = 2147483647

// keyExists is responsible for checking if server responded with json key
func keyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}

// validateModel checks if the model requested is available on Text Synth
func validateModel(model string) {
	allowedModels := map[string]bool{
		"gpt2_345M":  true,
		"gpt2_1558M": true,
		"gptj_6B":    true,
	}
	if !allowedModels[model] {
		log.Fatal("model must be either gpt2_345M, gpt2_1558M, or gptj_6B.")
	}
}

// Prompts the user after Text Synth was either interrupted or finished
func whatNow() string {
	response := struct {
		WhatNow string
	}{}
	prompt := []*survey.Question{
		{
			Name: "WhatNow",
			Prompt: &survey.Select{
				Message: "What now?",
				Options: []string{"Continue", "Retry", "Exit"},
			},
			Validate: survey.Required,
		},
	}
	err := survey.Ask(prompt, &response)
	if err != nil {
		log.Fatal(err)
	}
	return response.WhatNow
}

// promptCheck checks if the service would accept the prompt or not
func promptCheck(prompt string) {
	if len(prompt) > PromptMaxSize {
		log.Fatalf("The service doesn't accept prompt sizes greater than %d bytes. Current prompt size is %d bytes.", PromptMaxSize, len(prompt))
	}
}

// communicate connects to the Text Synth server to send the prompt and show it to the user
func communicate(model string, j map[string]interface{}, dontNormalizeNewline bool) string {
	if term.IsTerminal(syscall.Stdin) && term.IsTerminal(syscall.Stdout) {
		screen.Clear()
		screen.MoveTopLeft()
	}

	fmt.Printf("%s", j["prompt"].(string))

	request, err := json.Marshal(&j)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://bellard.org/textsynth/api/v1/engines/"+model+"/completions", bytes.NewBuffer(request))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", "https://github.com/rany2/go-textsynth")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalf("Service returned %d status code. Expected 200.", resp.StatusCode)
	}

	s := bufio.NewScanner(resp.Body)
	var newPrompt = j["prompt"].(string)

	finished := make(chan bool, 1)
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	go func() {
		for s.Scan() {
			var m map[string]interface{}
			err := json.Unmarshal(s.Bytes(), &m)
			if err == nil {
				if keyExists(m, "text") {
					if !dontNormalizeNewline && lineBreak == "\n" {
						fmt.Printf("%s", string(normalizenewlines.Run([]byte(m["text"].(string)))))
					} else if !dontNormalizeNewline && lineBreak == "\r\n" {
						fmt.Printf("%s", string(windowsnewlines.Run([]byte(m["text"].(string)))))
					} else {
						fmt.Printf("%s", m["text"].(string))
					}
					newPrompt += m["text"].(string)
				}
			}
		}
		finished <- true
	}()
	go func() {
		<-sigchan
		cancel()
	}()
	<-finished
	signal.Stop(sigchan) // stop listening on ctrl-c
	close(sigchan)       // close channel to end goroutine
	return newPrompt
}

func main() {
	model := flag.String("model", "gptj_6B", "Select a model (gpt2_345M, gpt2_1558M, or gptj_6B)")
	prompt := flag.String("prompt", "", "Prompt to send to Text Synth")
	promptfile := flag.String("promptfile", "", "Like prompt but read from file")
	temperature := flag.Float64("temperature", 1.0, "Divide the logits (=log(probability) of the tokens) by the temperature value (0.1 <= temperature <= 10)")
	topK := flag.Float64("top-k", 40, "Keep only the top-k tokens with the highest probability (1 <= top-k <= 1000)")
	topP := flag.Float64("top-p", 0.9, "Keep the top tokens having cumulative probability >= top-p (0 < top-p <= 1)")
	seed := flag.Uint("seed", 0, "Seed of the random number generator. Use 0 for a random seed.")
	dontNormalizeNewline := flag.Bool("dont-normalize-newline", false, "Do not convert Windows and Mac OS line endings to Unix")
	flag.Parse()

	// Check if the model requested exists
	validateModel(*model)

	if *promptfile != "" && *prompt != "" {
		log.Fatal("prompt and promptfile are mutually exclusive.")
	} else if *promptfile != "" {
		data, err := os.ReadFile(*promptfile)
		if err != nil {
			log.Fatal(err)
		}
		*prompt = string(data)
	} else if *prompt == "" {
		log.Fatal("prompt must be set via -prompt or -promptfile.")
	}

	if !*dontNormalizeNewline {
		*prompt = string(normalizenewlines.Run([]byte(*prompt)))
	}

	if *temperature < 0.1 || *temperature > 10.0 {
		log.Fatal("temperature must be between 0.1 and 10.")
	}

	if *topK < 1 || *topK > 1000 {
		log.Fatal("top_k must be between 1 and 1000.")
	}

	if *topP <= 0 || *topP > 1 {
		log.Fatal("invalid top_p value (0 < top-p <= 1).")
	}

	// No need to check if negative because flag.Uint deals with that
	if *seed > SeedLimit {
		log.Fatalf("seed cannot be greater than %d", SeedLimit)
	}

	j := make(map[string]interface{})
	j["temperature"] = *temperature
	j["top_k"] = *topK
	j["top_p"] = *topP
	j["seed"] = *seed
	j["stream"] = true

outer:
	for {
		promptCheck(*prompt)
		j["prompt"] = *prompt
		var newPrompt = communicate(*model, j, *dontNormalizeNewline)
		if term.IsTerminal(syscall.Stdin) && term.IsTerminal(syscall.Stdout) {
			fmt.Printf("%s", lineBreak)
			switch whatNow() {
			case "Continue":
				*prompt = newPrompt
			case "Retry":
				break
			default:
				break outer
			}
		} else {
			break outer
		}
	}
}
