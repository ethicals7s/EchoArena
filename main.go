package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type OllamaReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResp struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func queryOllama(model, prompt string) (string, error) {
	reqBody, _ := json.Marshal(OllamaReq{Model: model, Prompt: prompt, Stream: false})
	resp, err := http.Post("http://127.0.0.1:11434/api/generate", "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var r OllamaResp
	json.Unmarshal(body, &r)
	return strings.TrimSpace(r.Response), nil
}

func main() {
	fmt.Println("EchoArena v1 – Local LLM Debate Arena")
	fmt.Print("\nEnter debate topic: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	topic := scanner.Text()

	pro := "llama3.2"
	con := "mistral"
	judge := "phi3"

	fmt.Printf("\nTopic: %s\n", topic)
	fmt.Printf("PRO: %s    |    CON: %s    |    Judge: %s\n\n", pro, con, judge)

	transcript := []string{}

	for round := 1; round <= 3; round++ {
		proResp, _ := queryOllama(pro, fmt.Sprintf("You are PRO. Strongly argue FOR: %s (round %d). <150 words", topic, round))
		transcript = append(transcript, fmt.Sprintf("Round %d – PRO (%s):\n%s", round, pro, proResp))
		fmt.Printf("Round %d – PRO:\n%s\n\n", round, proResp)

		conResp, _ := queryOllama(con, fmt.Sprintf("You are CON. Rebut PRO. Topic: %s (round %d). <150 words", topic, round))
		transcript = append(transcript, fmt.Sprintf("Round %d – CON (%s):\n%s", round, con, conResp))
		fmt.Printf("Round %d – CON:\n%s\n\n", round, conResp)

		time.Sleep(2 * time.Second)
	}

	judgePrompt := "Judge this debate. Score PRO and CON 1–10 on logic, evidence, persuasion. Declare clear winner.\n\n" + strings.Join(transcript, "\n\n")
	verdict, _ := queryOllama(judge, judgePrompt)
	fmt.Printf("JUDGE VERDICT:\n%s\n", verdict)

	os.WriteFile("debate.md", []byte("# EchoArena Debate\n\n"+strings.Join(transcript, "\n\n")+"\n\nJUDGE: "+verdict), 0644)
	fmt.Println("\nFull transcript saved to debate.md")
}
