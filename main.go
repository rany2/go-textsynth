package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"bufio"
	"log"
	"bytes"
	"flag"
	"os"
	"os/signal"
	promptui "github.com/manifoldco/promptui"
	tm "github.com/buger/goterm"
)

func keyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}

func whatNow() string {
	prompt := promptui.Select{
		Label: "What now?",
		Items: []string{"Continue", "Retry", "Exit"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		log.Fatal(err)
	}
	return result
}

func main() {
	model := flag.String("model", "gptj_6B", "Select a model (gpt2_345M, gpt2_1558M, or gptj_6B)")
	prompt := flag.String("prompt", "", "Prompt to send to Text Synth")
	temperature := flag.Float64("temperature", 1.0, "Divide the logits (=log(probability) of the tokens) by the temperature value (0.1 <= temperature <= 10)")
	top_k := flag.Float64("top_k", 40, "Keep only the top-k tokens with the highest probability (1 <= top-k <= 1000)")
	top_p := flag.Float64("top_p", 0.9, "Keep the top tokens having cumulative probability >= top-p (0 < top-p <= 1)")
	seed := flag.Uint("seed", 0, "Seed of the random number generator. Use 0 for a random seed.")
	flag.Parse()

	allowedModels := map[string]bool {
		"gpt2_345M": true,
		"gpt2_1558M": true,
		"gptj_6B": true,
	}
	if !allowedModels[*model] {
		log.Fatal("Model must be either gpt2_345M, gpt2_1558M, or gptj_6B.")
	}

	if *prompt == "" {
		log.Fatal("A prompt must be set.")
	}

	if *temperature < 0.1 || *temperature > 10.0 {
		log.Fatal("Temperature must be between 0.1 and 10.")
	}

	if *top_k < 1 || *top_k > 1000 {
		log.Fatal("Top-k must be between 1 and 1000.")
	}

	if *top_p < 0 || *top_p > 1 {
		log.Fatal("Top-p must be between 0 and 1.")
	}

	var j map[string]interface{}
	_ = json.Unmarshal([]byte("{}"), &j)
	j["temperature"] = *temperature
	j["top_k"] = *top_k
	j["top_p"] = *top_p
	j["seed"] = *seed
	j["stream"] = true

	outer:
		for {
			if len(*prompt) >= 4096 {
				log.Fatalf("The service doesn't accept prompt sizes greater than 4095 bytes. Current prompt size is %d bytes.", len(*prompt))
			}
			//fmt.Print ("\x1b[0;0H") // move cursor to top
			//fmt.Print ("\x1b[0J") // clear screen down
			tm.Clear() // clear screen using a library
			tm.MoveCursor(1,1)  // move to top
			tm.Flush() // send changes
			j["prompt"] = *prompt
			request, err := json.Marshal(&j)
			if err != nil {
				log.Fatal(err)
			}

			resp, err := http.Post("https://bellard.org/textsynth/api/v1/engines/"+ *model +"/completions",
				"application/json",
				bytes.NewBuffer(request))
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf ("%s", *prompt)
			s := bufio.NewScanner(resp.Body)
			var newPrompt = *prompt

			finished := make(chan bool, 1)
			go func(){
				sigchan := make (chan os.Signal, 1)
				signal.Notify(sigchan, os.Interrupt)
				for s.Scan() {
					select {
					case <-sigchan:
						finished <- true
						return
					default:
						var m map[string]interface{}
						err := json.Unmarshal(s.Bytes(), &m)
						if err == nil {
							if keyExists(m, "text") {
								fmt.Printf ("%s", m["text"].(string))
								newPrompt += m["text"].(string)
							}
						}
					}
				}
				finished <- true
			}()
			<-finished
			if err := s.Err(); err != nil {
				log.Fatal(err)
			}

			fmt.Println()
			switch whatNow() {
				case "Continue":
					*prompt = newPrompt
				case "Retry":
					break
				default:
					break outer
			}
		}
}
