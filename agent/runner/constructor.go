package runner

import (
	"goclaw/agent"
	"goclaw/agent/files"
	"goclaw/agent/runner/messages"
	"goclaw/agent/runner/runnertools"
	"log/slog"
	"math"
	"os"
	"sync"
	"time"

	"github.com/JoshPattman/jpf"
	"github.com/JoshPattman/jpf/pipelines"
)

func New(model jpf.Model, memoryLoc string, fs files.FileSystem, opts ...runnerOpt) agent.Agent {
	nothingLoggHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(math.MaxInt),
	})
	setup := agentRunnerSetup{
		logger:             slog.New(nothingLoggHandler),
		prompt:             getDefaultPrompt(),
		collectionDuration: time.Second,
	}
	for _, o := range opts {
		o(&setup)
	}
	enc := buildEncoder(setup.prompt)
	//pars := parsers.SubstringJsonObject(parsers.NewJson[messages.ToolCallsMessage]())
	pars := &firstJsonObjectParser[messages.ToolCallsMessage]{}
	pipe := pipelines.NewOneShot(enc, pars, nil, model)

	events := make(chan agent.Event)

	runner := &agentRunner{
		setup.logger,
		events,
		make([]messages.Message, 0),
		enc,
		pipe,
		setup.collectionDuration,
		nil,
		&sync.Mutex{},
		memoryLoc,
		fs,
		&sync.Mutex{},
		make(chan struct{}),
		5,
		5,
		16000,
	}
	runner.AddPlugin(runnertools.Plugin(events, fs))
	return runner
}

type runnerOpt func(*agentRunnerSetup)

type agentRunnerSetup struct {
	logger             *slog.Logger
	collectionDuration time.Duration
	prompt             agent.JsonObject
}

func WithLogger(logger *slog.Logger) runnerOpt {
	return func(ar *agentRunnerSetup) {
		ar.logger = logger
	}
}

func WithCollectDuration(dur time.Duration) runnerOpt {
	return func(ar *agentRunnerSetup) {
		ar.collectionDuration = dur
	}
}

func WithPrompt(prompt agent.JsonObject) runnerOpt {
	return func(ar *agentRunnerSetup) {
		ar.prompt = prompt
	}
}
