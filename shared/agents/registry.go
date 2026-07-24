package agents

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed registry.json
var registryJSON []byte

// OnErrorStop fails the whole pipeline when the agent errors.
const OnErrorStop = "stop"

// OnErrorSkip continues the pipeline when types stay compatible.
const OnErrorSkip = "skip"

// Agent is one orchestrated microservice entry from the shared registry.
type Agent struct {
	Key     string   `json:"key"`
	Label   string   `json:"label"`
	LabelID string   `json:"labelId"`
	Desc    string   `json:"desc"`
	Inputs  []string `json:"inputs"`
	Outputs []string `json:"outputs"`
	EnvURL  string   `json:"envUrl"`
	OnError string   `json:"onError"`
}

type fileShape struct {
	Agents []Agent `json:"agents"`
}

var (
	loadOnce sync.Once
	all      []Agent
	byKey    map[string]Agent
	loadErr  error
)

func load() {
	loadOnce.Do(func() {
		var shape fileShape
		if err := json.Unmarshal(registryJSON, &shape); err != nil {
			loadErr = fmt.Errorf("parse agent registry: %w", err)
			byKey = map[string]Agent{}
			return
		}
		all = shape.Agents
		byKey = make(map[string]Agent, len(all))
		for _, a := range all {
			key := strings.ToLower(strings.TrimSpace(a.Key))
			a.Key = key
			if a.OnError != OnErrorSkip {
				a.OnError = OnErrorStop
			}
			byKey[key] = a
		}
	})
}

// All returns every registered agent in registry order.
func All() []Agent {
	load()
	out := make([]Agent, len(all))
	copy(out, all)
	return out
}

// Get returns an agent by key.
func Get(key string) (Agent, bool) {
	load()
	a, ok := byKey[strings.ToLower(strings.TrimSpace(key))]
	return a, ok
}

// MustLoad panics if the embedded registry is invalid (startup guard).
func MustLoad() {
	load()
	if loadErr != nil {
		panic(loadErr)
	}
}

// OutputType returns the primary output type for an agent (default text).
func OutputType(key string) string {
	a, ok := Get(key)
	if !ok || len(a.Outputs) == 0 {
		return "text"
	}
	return a.Outputs[0]
}

// InputCompatible reports whether outputType is accepted by nextAgent.
func InputCompatible(outputType, nextAgent string) bool {
	a, ok := Get(nextAgent)
	if !ok {
		return false
	}
	for _, in := range a.Inputs {
		if in == outputType {
			return true
		}
	}
	return false
}

// ErrorPolicy returns stop|skip for an agent.
func ErrorPolicy(key string) string {
	a, ok := Get(key)
	if !ok {
		return OnErrorStop
	}
	return a.OnError
}

// LabelID returns the casual Indonesian display name used in chat progress.
func LabelID(key string) string {
	a, ok := Get(key)
	if !ok {
		return strings.ReplaceAll(key, "_", " ")
	}
	if a.LabelID != "" {
		return a.LabelID
	}
	return a.Label
}

// Label returns the English/product label.
func Label(key string) string {
	a, ok := Get(key)
	if !ok {
		return strings.ReplaceAll(key, "_", " ")
	}
	return a.Label
}

// EnvURLKeys returns map[agentKey]envVarName for config loading.
func EnvURLKeys() map[string]string {
	load()
	out := make(map[string]string, len(all))
	for _, a := range all {
		out[a.Key] = a.EnvURL
	}
	return out
}

// RouterSystemPrompt builds the LLM routing catalog from the registry.
func RouterSystemPrompt() string {
	load()
	var b strings.Builder
	b.WriteString(`You are the routing engine for the Joki Tugas Multi-Agent Orchestrator.
Your task is to analyze the user's request and plan a sequential pipeline of agents to fulfill it.

Available agents are:
`)
	for i, a := range all {
		in := strings.Join(a.Inputs, "|")
		out := strings.Join(a.Outputs, "|")
		fmt.Fprintf(&b, "%d. %s (input: %s -> output: %s) - %s\n", i+1, a.Key, in, out, a.Desc)
	}
	b.WriteString(`
Rules:
- Respond ONLY with a valid JSON object containing a "pipeline" key which is an array of strings representing the agent names in sequential order of execution.
- Example JSON response:
{
  "pipeline": ["web_scraper", "summarizer", "ppt_generator"]
}
- Do not include markdown formatting or extra text outside the JSON object.
- The pipeline should be a realistic, type-compatible sequence.`)
	return b.String()
}
