package agent

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"time"

	"github.com/JoshPattman/jpf"
	"github.com/JoshPattman/jpf/parsers"
	"github.com/JoshPattman/jpf/pipelines"
)

//go:embed system.md
var defaultPrompt string

func New(model jpf.Model, logger *slog.Logger) *Agent {
	toolMap := make(map[string]Tool)
	enc := &msgEncoder{defaultPrompt}
	pars := parsers.SubstringJsonObject(parsers.NewJson[toolCallsMessage]())
	pipe := pipelines.NewOneShot(enc, pars, nil, model)

	return &Agent{
		logger:             logger,
		events:             make(chan Event),
		history:            make([]message, 0),
		pipeline:           pipe,
		collectionDuration: time.Second,
		tools:              toolMap,
		truncateAtLength:   50,
		truncateToLength:   30,
	}
}

type Agent struct {
	logger             *slog.Logger
	events             chan Event
	history            []message
	pipeline           jpf.Pipeline[[]message, toolCallsMessage]
	collectionDuration time.Duration
	tools              map[string]Tool
	truncateAtLength   int
	truncateToLength   int
}

func (a *Agent) AddTools(tools ...Tool) {
	for _, t := range tools {
		a.tools[t.Name()] = t
	}
	prompt := buildChangeToolsPrompt(slices.Collect(maps.Values(a.tools)))
	a.addEventsMessage(E(EventKind("available_tools_change"), prompt))
	a.logger.Info("changed tools", "active_tools", slices.Collect(maps.Keys(a.tools)))
}

func (a *Agent) SetPersonality(personality string) {
	prompt := fmt.Sprintf("You will act with the foillowing personality from now on: %s", personality)
	a.addEventsMessage(E(EventKind("personality_instruction"), prompt))
	a.logger.Info("personality has been set", "personality", personality)
}

func (a *Agent) Events() chan<- Event { return a.events }

func (a *Agent) Run() error {
	a.logger.Info("running event loop")
	var eventBuffer []Event
	var dispatch <-chan time.Time

	for {
		select {
		case event := <-a.events:
			eventBuffer = append(eventBuffer, event)
			if len(eventBuffer) == 1 {
				a.logger.Info("initial event occured, waiting to collect more")
				dispatch = time.After(a.collectionDuration)
			} else {
				a.logger.Info("additional event occured")
			}

		case <-dispatch:
			a.addEventsMessage(eventBuffer...)
			err := a.processUntilDone()
			if err != nil {
				return err
			}
			eventBuffer = nil
			dispatch = nil

		}
	}
}

func (a *Agent) addEventsMessage(events ...Event) {
	a.history = append(a.history, eventsMessage{events})
	a.logger.Info("events occured", "n", len(events))
}

func (a *Agent) clearIfTooLong() {
	if len(a.history) > a.truncateAtLength {
		a.history = a.history[len(a.history)-a.truncateToLength:]
		a.history = append(a.history, eventsMessage{[]Event{
			E(EventKind("conversation_clipping"), "The conversation has just been truncated (oldest messages were removed). Re-read your system prompt and ensure there is nothing you need to do."),
		}})
		a.logger.Info("truncated conversation")
	}
}

func (a *Agent) processUntilDone() error {
	a.logger.Info("processing events")
	for {
		result, _, err := a.pipeline.Call(context.Background(), a.history)
		if err != nil {
			a.logger.Error("failed to process events", "err", err)
			return err
		}
		a.history = append(a.history, result)

		if len(result.ToolCalls) == 0 {
			a.logger.Info("done processing events")
			return nil
		}

		responses := make([]string, 0, len(result.ToolCalls))

		for _, call := range result.ToolCalls {
			tool, ok := a.tools[call.ToolName]
			if !ok {
				responses = append(responses,
					fmt.Sprintf("Tool '%s' not found", call.ToolName),
				)
				continue
			}

			args := make(map[string]any)
			for _, arg := range call.Args {
				args[arg.ArgName] = arg.Value
			}

			out, err := tool.Call(args)
			if err != nil {
				responses = append(responses,
					fmt.Sprintf("Tool '%s' error: %v", call.ToolName, err),
				)
				continue
			}

			responses = append(responses, out)
		}

		a.history = append(a.history, topolResponseMessage{
			Responses: responses,
		})
	}
}
