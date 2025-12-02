package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	proModel      = flag.String("pro", "llama3.2", "PRO model")
	conModel      = flag.String("con", "mistral", "CON model")
	judgeModel    = flag.String("judge", "phi3", "Judge model")
	proEndpoint   = flag.String("pro-endpoint", "http://127.0.0.1:11434", "PRO endpoint")
	conEndpoint   = flag.String("con-endpoint", "http://127.0.0.1:11434", "CON endpoint")
	judgeEndpoint = flag.String("judge-endpoint", "http://127.0.0.1:11434", "Judge endpoint")
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

func queryModel(model, prompt, endpoint string) (string, error) {
	reqBody, _ := json.Marshal(OllamaReq{Model: model, Prompt: prompt, Stream: false})
	resp, err := http.Post(endpoint+"/api/generate", "application/json", strings.NewReader(string(reqBody)))
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
	flag.Parse()

	fmt.Println("EchoArena v2.1 – Local LLM Debate Arena (separate endpoints for PRO/CON/Judge)")
	fmt.Print("\nEnter debate topic: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	topic := scanner.Text()

	fmt.Printf("\n⚔️ Topic: %s\n", topic)
	fmt.Printf("🤖 PRO: %s @ %s\n   CON: %s @ %s\n   Judge: %s @ %s\n\n",
	*proModel, *proEndpoint, *conModel, *conEndpoint, *judgeModel, *judgeEndpoint)

	transcript := []string{}

	for round := 1; round <= 3; round++ {
		proResp, _ := queryModel(*proModel, fmt.Sprintf("You are PRO. Strongly argue FOR: %s (round %d). <150 words", topic, round), *proEndpoint)
		transcript = append(transcript, fmt.Sprintf("Round %d – PRO (%s):\n%s", round, *proModel, proResp))
		fmt.Printf("Round %d – PRO:\n%s\n\n", round, proResp)

		conResp, _ := queryModel(*conModel, fmt.Sprintf("You are CON. Rebut PRO. Topic: %s (round %d). <150 words", topic, round), *conEndpoint)
		transcript = append(transcript, fmt.Sprintf("Round %d – CON (%s):\n%s", round, *conModel, conResp))
		fmt.Printf("Round %d – CON:\n%s\n\n", round, conResp)

		time.Sleep(2 * time.Second)
	}

	judgePrompt := "Judge this debate. Score PRO and CON 1–10 on logic, evidence, persuasion. Declare clear winner.\n\n" + strings.Join(transcript, "\n\n")
	verdict, _ := queryModel(*judgeModel, judgePrompt, *judgeEndpoint)
	fmt.Printf("🏛️ JUDGE VERDICT:\n%s\n", verdict)

	os.WriteFile("debate.md", []byte("# EchoArena Debate\n\n"+strings.Join(transcript, "\n\n")+"\n\nJUDGE: "+verdict), 0644)
	fmt.Println("\n📄 Full transcript saved to debate.md")
}
