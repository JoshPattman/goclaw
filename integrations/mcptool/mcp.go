package mcptool

import (
	"context"
	"errors"
	"fmt"
	"goclaw/agent"
	"os/exec"
	"slices"
	"strings"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

func CreateCommand(command []string) (*client.Client, error) {
	cmd := exec.Command(command[0], command[1:]...)
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

func CreateClient(addr string, customHeaders map[string]string) (*client.Client, error) {
	httpTransport, err := transport.NewStreamableHTTP(
		addr,
		transport.WithHTTPHeaders(customHeaders),
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

func initClient(c *client.Client) error {
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "MCP-Agent",
		Version: "1.0.0",
	}
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}

	_, err := c.Initialize(context.Background(), initRequest)
	if err != nil {
		return err
	}
	return nil
}

func CreateToolsFromMCP(client *client.Client) ([]agent.Tool, error) {
	ctx := context.Background()
	result, err := client.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, err
	}
	tools := make([]agent.Tool, len(result.Tools))
	for i, mcpTool := range result.Tools {
		agentTool, err := createTool(client, mcpTool)
		if err != nil {
			return nil, err
		}
		tools[i] = agentTool
	}
	return tools, nil
}

func createTool(client *client.Client, tool mcp.Tool) (agent.Tool, error) {
	return &mcpTool{client, tool}, nil
}

type mcpTool struct {
	client *client.Client
	tool   mcp.Tool
}

func (m *mcpTool) Call(args map[string]any) (string, error) {
	res, err := m.client.CallTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      m.tool.Name,
			Arguments: args,
		},
	})
	if err != nil {
		return "", err
	}
	contents := make([]string, 0)
	for _, c := range res.Content {
		if c, ok := c.(mcp.TextContent); ok {
			contents = append(contents, c.Text)
		}
	}
	if len(contents) == 0 {
		return "", errors.New("tool returned no content")
	}
	return strings.Join(contents, "\n\n\n"), nil
}

func (m *mcpTool) Def() agent.ToolDef {
	desc := []string{m.tool.Description}
	for pname, prop := range m.tool.InputSchema.Properties {
		var required string
		if slices.Contains(m.tool.InputSchema.Required, pname) {
			required = " [required]"
		}
		propMap, ok := prop.(map[string]any)
		if !ok {
			propMap = map[string]any{}
		}
		propType, ok := propMap["type"].(string)
		if !ok {
			propType = "Type not documented, please infer"
		}
		propDesc, ok := propMap["description"].(string)
		if !ok {
			propDesc = "Description not documented, please infer"
		}
		desc = append(
			desc,
			fmt.Sprintf("Param%s `%s` (%s): %s", required, pname, propType, propDesc),
		)
	}
	return agent.ToolDef{
		Name: m.tool.Name,
		Desc: strings.Join(desc, "\n"),
	}
}
