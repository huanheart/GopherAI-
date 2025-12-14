package client

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// MCPClient 是MCP客户端的封装
// 它提供了一个类对象接口来与MCP服务器交互
type MCPClient struct {
	c *client.Client
}

// NewMCPClient 创建一个新的MCP客户端实例
// transportType: "stdio" 或 "http"
// stdioCmd: 当transportType为"stdio"时，要执行的命令
// httpURL: 当transportType为"http"时，HTTP传输的URL
func NewMCPClient(transportType string, stdioCmd string, httpURL string) (*MCPClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var c *client.Client

	if transportType == "stdio" {
		fmt.Println("正在初始化stdio客户端...")

		// 解析命令和参数
		args := parseCommand(stdioCmd)
		if len(args) == 0 {
			return nil, fmt.Errorf("无效的stdio命令")
		}

		// 创建命令和stdio传输
		command := args[0]
		cmdArgs := args[1:]

		// 创建带有详细日志的stdio传输
		stdioTransport := transport.NewStdio(command, nil, cmdArgs...)

		// 使用传输创建客户端
		c = client.NewClient(stdioTransport)

		// 启动客户端
		if err := c.Start(ctx); err != nil {
			return nil, fmt.Errorf("启动客户端失败: %w", err)
		}

		// 为stderr设置日志（如果可用）
		if stderr, ok := client.GetStderr(c); ok {
			go func() {
				buf := make([]byte, 4096)
				for {
					n, err := stderr.Read(buf)
					if err != nil {
						if err != io.EOF {
							log.Printf("读取stderr错误: %v", err)
						}
						return
					}
					if n > 0 {
						fmt.Fprintf(io.Discard, "[Server] %s", buf[:n])
					}
				}
			}()
		}
	} else if transportType == "http" {
		fmt.Println("正在初始化HTTP客户端...")
		// 创建HTTP传输
		httpTransport, err := transport.NewStreamableHTTP(httpURL)
		if err != nil {
			return nil, fmt.Errorf("创建HTTP传输失败: %w", err)
		}

		// 使用传输创建客户端
		c = client.NewClient(httpTransport)
	} else {
		return nil, fmt.Errorf("无效的传输类型: %s", transportType)
	}

	return &MCPClient{c: c}, nil
}

// Initialize 初始化客户端
func (m *MCPClient) Initialize(ctx context.Context) (*mcp.InitializeResult, error) {
	// 设置通知处理程序
	m.c.OnNotification(func(notification mcp.JSONRPCNotification) {
		fmt.Printf("收到通知: %s\n", notification.Method)
	})

	// 初始化客户端
	fmt.Println("正在初始化客户端...")
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "MCP-Go Weather Client",
		Version: "1.0.0",
	}
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}

	serverInfo, err := m.c.Initialize(ctx, initRequest)
	if err != nil {
		return nil, fmt.Errorf("初始化失败: %w", err)
	}

	// 显示服务器信息
	fmt.Printf("连接到服务器: %s (版本 %s)\n",
		serverInfo.ServerInfo.Name,
		serverInfo.ServerInfo.Version)

	return serverInfo, nil
}

// Ping 执行健康检查
func (m *MCPClient) Ping(ctx context.Context) error {
	fmt.Println("正在执行健康检查...")
	if err := m.c.Ping(ctx); err != nil {
		return fmt.Errorf("健康检查失败: %w", err)
	}
	fmt.Println("服务器正常运行并响应")
	return nil
}

// CallTool 调用MCP工具
func (m *MCPClient) CallTool(ctx context.Context, toolName string, args map[string]any) (*mcp.CallToolResult, error) {
	callToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		},
	}

	result, err := m.c.CallTool(ctx, callToolRequest)
	if err != nil {
		return nil, fmt.Errorf("调用工具失败: %w", err)
	}

	return result, nil
}

// CallWeatherTool 调用get_weather工具
func (m *MCPClient) CallWeatherTool(ctx context.Context, city string) (*mcp.CallToolResult, error) {
	fmt.Printf("正在查询城市 %s 的天气...\n", city)

	callToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_weather",
			Arguments: map[string]any{
				"city": city,
			},
		},
	}

	result, err := m.c.CallTool(ctx, callToolRequest)
	if err != nil {
		return nil, fmt.Errorf("调用工具失败: %w", err)
	}

	return result, nil
}

// GetToolResultText 获取工具结果中的文本内容
func (m *MCPClient) GetToolResultText(result *mcp.CallToolResult) string {
	var text string
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			text += textContent.Text + "\n"
		}
	}
	return text
}

// Close 关闭客户端连接
func (m *MCPClient) Close() {
	if m.c != nil {
		m.c.Close()
	}
}

// parseCommand 将命令字符串分割为命令和参数
func parseCommand(cmd string) []string {
	// 这是一个简单的实现，不处理引号或转义
	var result []string
	var current string
	var inQuote bool
	var quoteChar rune

	for _, r := range cmd {
		switch {
		case r == ' ' && !inQuote:
			if current != "" {
				result = append(result, current)
				current = ""
			}
		case (r == '"' || r == '\''):
			if inQuote && r == quoteChar {
				inQuote = false
				quoteChar = 0
			} else if !inQuote {
				inQuote = true
				quoteChar = r
			} else {
				current += string(r)
			}
		default:
			current += string(r)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}
