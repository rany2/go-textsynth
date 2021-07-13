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
	"io"
	promptui "github.com/manifoldco/promptui"
	tm "github.com/buger/goterm"
)

// Create HTTP transports to share pool of connections while disabling compression
var tr = &http.Transport{}
var client = &http.Client{Transport: tr}

// Set prompt size limit and file chunk size
const PROMPT_MAX_SIZE = 4095;
const CHUNK_SIZE = 65536;

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
	promptfile := flag.String("promptfile", "", "Like prompt but read from file")
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
		log.Fatal("model must be either gpt2_345M, gpt2_1558M, or gptj_6B.")
	}

	if *promptfile != "" && *prompt != "" {
		log.Fatal("prompt and promptfile are mutually exclusive.")
	} else if *promptfile != "" {
		f, err := os.Open(*promptfile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		reader := bufio.NewReader(f)
		part := make([]byte, CHUNK_SIZE)
		for {
			if len(*prompt) > PROMPT_MAX_SIZE {
				log.Fatalf ("While reading file exceeded prompt limit of %d bytes, before aborting it was %d bytes.", PROMPT_MAX_SIZE, len(*prompt))
			} else {
				if count, err := reader.Read(part); err != nil {
					if err == io.EOF {
						break
					} else {
						log.Fatal (err)
					}
				} else {
					*prompt += string(part[:count])
				}
			}
		}
	} else if *prompt == "" {
		log.Fatal("prompt must be set via -prompt or -promptfile.")
	}

	if *temperature < 0.1 || *temperature > 10.0 {
		log.Fatal("temperature must be between 0.1 and 10.")
	}

	if *top_k < 1 || *top_k > 1000 {
		log.Fatal("top_k must be between 1 and 1000.")
	}

	if *top_p <= 0 || *top_p > 1 {
		log.Fatal("invalid top_p value (0 < top-p <= 1).")
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
			if len(*prompt) > PROMPT_MAX_SIZE {
				log.Fatalf("The service doesn't accept prompt sizes greater than %d bytes. Current prompt size is %d bytes.", PROMPT_MAX_SIZE, len(*prompt))
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

			req, err := http.NewRequest("POST", "https://bellard.org/textsynth/api/v1/engines/"+ *model +"/completions", bytes.NewBuffer(request))
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

			fmt.Printf ("%s", *prompt)
			s := bufio.NewScanner(resp.Body)
			var newPrompt = *prompt

			finished := make(chan bool, 1)
			sigchan := make (chan os.Signal, 1)
			signal.Notify(sigchan, os.Interrupt)
			go func(){
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
