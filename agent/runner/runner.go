package runner

import (
	"context"
	"fmt"
	"goclaw/agent"
	"goclaw/agent/files"
	"goclaw/agent/runner/messages"
	"goclaw/agent/runner/runnertools"
	"log/slog"
	"strings"
	"time"

	"github.com/JoshPattman/jpf"

	_ "embed"
)

type agentRunner struct {
	logger             *slog.Logger
	events             chan agent.Event
	history            []messages.Message
	pipeline           jpf.Pipeline[encoderInput, messages.ToolCallsMessage]
	collectionDuration time.Duration
	tools              []agent.Tool
	memoryLoc          string
	fs                 files.FileSystem
}

func (a *agentRunner) AddTools(tools ...agent.Tool) {
	for _, t := range tools {
		if a.toolExists(t.Def().Name) {
			continue
		}
		a.tools = append(a.tools, t)
	}
	a.logger.Info("changed tools", "active_tools", a.toolNames())
}

func (a *agentRunner) Events() chan<- agent.Event { return a.events }

func (a *agentRunner) Run() error {
	a.logger.Info("running event loop")
	var eventBuffer []agent.Event
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

func (a *agentRunner) processUntilDone() error {
	a.logger.Info("processing events")
	for {
		inputData := encoderInput{
			a.history,
			a.toolDefs(),
			time.Now(),
			a.memoryLoc,
			a.workingMemory(),
		}
		result, _, err := a.pipeline.Call(context.Background(), inputData)
		if err != nil {
			a.logger.Error("failed to process events", "err", err)
			return err
		}
		a.addHistory(result)

		if len(result.ToolCalls) == 0 {
			a.logger.Info("agent called no tools so we need to remind it this is not how it stops processing")
			a.addHistory(messages.NeedToExplicitlyStopMessage())
			continue
		}
		if len(result.ToolCalls) == 1 && result.ToolCalls[0].ToolName == runnertools.DoneToolName() {
			a.logger.Info("agent called end iteration tool so we can stop")
			return nil
		}
		a.logToolCalls(result.ToolCalls)

		responses := make([]string, 0, len(result.ToolCalls))

		for _, call := range result.ToolCalls {
			tool := a.lookupTool(call.ToolName)
			if tool == nil {
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
		a.addHistory(messages.ToolResponseMessage(responses))
	}
}

func (a *agentRunner) logToolCalls(toolCalls []messages.ToolCall) {
	toolNames := make([]string, len(toolCalls))
	for i, t := range toolCalls {
		toolNames[i] = t.ToolName
	}
	a.logger.Info("agent called tools", "tool_names", strings.Join(toolNames, ";"))
}

func (a *agentRunner) workingMemory() string {
	bs, err := a.fs.Read(a.memoryLoc)
	if err != nil {
		return fmt.Sprintf("There was an error loading your working memory: %s", err.Error())
	}
	return string(bs)
}

func (a *agentRunner) addEventsMessage(events ...agent.Event) {
	a.addHistory(messages.EventsMessage(events...))
	eventNames := make([]string, len(events))
	for i, e := range events {
		eventNames[i] = e.Kind()
	}
	a.logger.Info("events occured", "n", len(events), "names", strings.Join(eventNames, ";"))
}

func (a *agentRunner) addHistory(messages ...messages.Message) {
	a.history = append(a.history, messages...)
}

func (a *agentRunner) lookupTool(name string) agent.Tool {
	for _, t := range a.tools {
		if t.Def().Name == name {
			return t
		}
	}
	return nil
}

func (a *agentRunner) toolExists(name string) bool {
	return a.lookupTool(name) != nil
}

func (a *agentRunner) toolNames() []string {
	names := make([]string, len(a.tools))
	for i, t := range a.tools {
		names[i] = t.Def().Name
	}
	return names
}

func (a *agentRunner) toolDefs() []agent.ToolDef {
	defs := make([]agent.ToolDef, len(a.tools))
	for i, t := range a.tools {
		defs[i] = t.Def()
	}
	return defs
}
