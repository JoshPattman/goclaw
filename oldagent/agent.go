package oldagent

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"
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
	a.addEventsMessage(toolChangeEvent{
		slices.Collect(maps.Values(a.tools)),
	})
	a.logger.Info("changed tools", "active_tools", slices.Collect(maps.Keys(a.tools)))
}

func (a *Agent) SetPersonality(personality string) {
	a.addEventsMessage(personalityInstruction{personality})
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

func (a *Agent) clearIfTooLong() {
	if len(a.history) > a.truncateAtLength {
		a.history = a.history[len(a.history)-a.truncateToLength:]
		a.addEventsMessage(conversationClippedEvent{})
		a.logger.Info("truncated conversation")
	}
}

const doneToolName = "end_iteration"

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
			a.logger.Info("agent called no tools so we need to remind it this is not how it stops processing")
			a.addHistory(needToEndMessage{})
			continue
		}
		if len(result.ToolCalls) == 1 && result.ToolCalls[0].ToolName == doneToolName {
			a.logger.Info("agent called end iteration tool so we can stop")
			return nil
		}

		toolNames := make([]string, len(result.ToolCalls))
		for i, t := range result.ToolCalls {
			toolNames[i] = t.ToolName
		}
		a.logger.Info("agent called tools", "tool_names", strings.Join(toolNames, ";"))

		responses := make([]string, 0, len(result.ToolCalls))

		for _, call := range result.ToolCalls {
			if call.ToolName == doneToolName {
				responses = append(responses, "You can only call the end iteration tool by itself - you cannot call other tools at the same time as it")
			}
			tool, ok := a.tools[call.ToolName]
			if !ok {
				responses = append(responses,
					fmt.Sprintf("Tool '%s' not found", call.ToolName),
				)
				a.logger.Warn("a tool was not found", "tool", call.ToolName)
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
				a.logger.Warn("a tool errored", "tool", call.ToolName, "err", err)
				continue
			}

			responses = append(responses, out)
		}

		a.history = append(a.history, toolResponseMessage{
			Responses: responses,
		})
	}
}

func (a *Agent) addEventsMessage(events ...Event) {
	a.addHistory(eventsMessage{events})
	eventNames := make([]string, len(events))
	for i, e := range events {
		eventNames[i] = string(e.EventKind())
	}
	a.logger.Info("events occured", "n", len(events), "names", strings.Join(eventNames, ";"))
}

func (a *Agent) addHistory(messages ...message) {
	a.history = append(a.history, messages...)
}
