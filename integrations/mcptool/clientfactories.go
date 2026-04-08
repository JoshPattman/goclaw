package mcptool

import (
	"os/exec"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
)

type ClientFactory interface {
	CreateClient() (*client.Client, error)
}

func ClientFromCommand(command []string) ClientFactory {
	return commandClientFactory{command}
}

func ClientFromHTTP(addr string, customHeaders map[string]string) ClientFactory {
	return httpClientFactory{addr, customHeaders}
}

type commandClientFactory struct {
	command []string
}

func (f commandClientFactory) CreateClient() (*client.Client, error) {
	cmd := exec.Command(f.command[0], f.command[1:]...)
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	commandTransport := transport.NewIO(out, in, nil)

	c := client.NewClient(commandTransport)

	err = initClient(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

type httpClientFactory struct {
	addr          string
	customHeaders map[string]string
}

func (f httpClientFactory) CreateClient() (*client.Client, error) {
	httpTransport, err := transport.NewStreamableHTTP(
		f.addr,
		transport.WithHTTPHeaders(f.customHeaders),
	)
	if err != nil {
		return nil, err
	}
	c := client.NewClient(httpTransport)

	err = initClient(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
