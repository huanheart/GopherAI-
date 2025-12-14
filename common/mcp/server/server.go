package server

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// test
// WeatherAPIClient 是一个模拟的天气API客户端
type WeatherAPIClient struct {
	baseURL string
}

// WeatherResponse 表示天气API的响应
type WeatherResponse struct {
	Location    string  `json:"location"`
	Temperature float64 `json:"temperature"`
	Condition   string  `json:"condition"`
	Humidity    int     `json:"humidity"`
	WindSpeed   float64 `json:"windSpeed"`
}

// NewWeatherAPIClient 创建一个新的天气API客户端
func NewWeatherAPIClient(baseURL string) *WeatherAPIClient {
	return &WeatherAPIClient{
		baseURL: baseURL,
	}
}

// GetWeather 查询指定城市的天气
func (c *WeatherAPIClient) GetWeather(ctx context.Context, city string) (*WeatherResponse, error) {
	// 模拟响应数据
	responses := map[string]*WeatherResponse{
		"北京": {
			Location:    "北京",
			Temperature: 15.5,
			Condition:   "多云",
			Humidity:    65,
			WindSpeed:   12.3,
		},
		"上海": {
			Location:    "上海",
			Temperature: 18.2,
			Condition:   "晴天",
			Humidity:    70,
			WindSpeed:   8.5,
		},
		"广州": {
			Location:    "广州",
			Temperature: 22.1,
			Condition:   "阴天",
			Humidity:    80,
			WindSpeed:   6.8,
		},
		"深圳": {
			Location:    "深圳",
			Temperature: 21.8,
			Condition:   "晴天",
			Humidity:    75,
			WindSpeed:   7.2,
		},
		"杭州": {
			Location:    "杭州",
			Temperature: 16.3,
			Condition:   "小雨",
			Humidity:    85,
			WindSpeed:   5.4,
		},
	}

	if resp, ok := responses[city]; ok {
		return resp, nil
	}

	// 如果城市不在模拟数据中，返回默认值
	return &WeatherResponse{
		Location:    city,
		Temperature: 15.0,
		Condition:   "未知",
		Humidity:    50,
		WindSpeed:   10.0,
	}, nil
}

// NewMCPServer 创建一个新的MCP服务器实例
func NewMCPServer() *server.MCPServer {
	// 创建天气API客户端
	weatherClient := NewWeatherAPIClient("https://api.weather.com")

	mcpServer := server.NewMCPServer(
		"weather-query-server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	// 添加天气查询工具
	mcpServer.AddTool(
		mcp.NewTool("get_weather",
			mcp.WithDescription("获取指定城市的天气信息"),
			mcp.WithString("city",
				mcp.Description("要查询天气的城市名称"),
				mcp.Required(),
			),
		), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			arguments := request.GetArguments()
			city, ok := arguments["city"].(string)
			if !ok {
				return nil, fmt.Errorf("invalid city argument")
			}

			// 查询天气
			weather, err := weatherClient.GetWeather(ctx, city)
			if err != nil {
				return nil, fmt.Errorf("failed to get weather: %w", err)
			}

			// 格式化天气信息
			weatherInfo := fmt.Sprintf("城市: %s\n温度: %.1f°C\n天气状况: %s\n湿度: %d%%\n风速: %.1f km/h",
				weather.Location,
				weather.Temperature,
				weather.Condition,
				weather.Humidity,
				weather.WindSpeed)

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: weatherInfo,
					},
				},
			}, nil
		})

	return mcpServer
}

// StartServer 启动MCP服务器
// transportType: "stdio" 或 "http"
// httpAddr: 当transportType为"http"时，HTTP服务器监听的地址（例如":8080"）
func StartServer(transportType string, httpAddr string) error {
	mcpServer := NewMCPServer()

	if transportType == "http" {
		httpServer := server.NewStreamableHTTPServer(mcpServer)
		log.Printf("HTTP server listening on %s/mcp", httpAddr)
		if err := httpServer.Start(httpAddr); err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	} else {
		if err := server.ServeStdio(mcpServer); err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	return nil
}
