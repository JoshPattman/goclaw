package mcptool

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/JoshPattman/cg"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func New(name string, factory ClientFactory) cg.Plugin {
	return &plugin{name, factory}
}

type plugin struct {
	name    string
	factory ClientFactory
}

func (p *plugin) Load() ([]cg.Tool, <-chan cg.Event, func(), error) {
	c, err := p.factory.CreateClient()
	if err != nil {
		return nil, nil, nil, err
	}
	err = initClient(c)
	if err != nil {
		return nil, nil, nil, err
	}

	ctx := context.Background()
	result, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, nil, nil, err
	}
	tools := make([]cg.Tool, len(result.Tools))
	for i, mcpTool := range result.Tools {
		agentTool, err := createTool(c, mcpTool)
		if err != nil {
			return nil, nil, nil, err
		}
		tools[i] = agentTool
	}
	return tools, nil, nil, nil
}

func (p *plugin) Name() string {
	return "mcp@" + p.name
}
func (p *plugin) Internal() bool { return false }

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

func createTool(client *client.Client, tool mcp.Tool) (cg.Tool, error) {
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

func (m *mcpTool) Def() cg.ToolDef {
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
	return cg.ToolDef{
		Name: m.tool.Name,
		Desc: strings.Join(desc, "\n"),
	}
}
