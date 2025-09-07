package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// SlackMessage Slackメッセージの構造体
type SlackMessage struct {
	Text        string       `json:"text"`
	Username    string       `json:"username,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment Slackメッセージの添付ファイル
type Attachment struct {
	Color     string  `json:"color"`
	Title     string  `json:"title"`
	Text      string  `json:"text"`
	Fields    []Field `json:"fields"`
	Timestamp int64   `json:"ts"`
}

// Field Slackメッセージのフィールド
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// ConnectionInfo データベース接続情報
type ConnectionInfo struct {
	Timestamp   time.Time
	UserName    string
	DatabaseName string
	ClientIP    string
	LogMessage  string
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event events.CloudwatchLogsEvent) error {
	log.Printf("CloudWatch Logsイベントを受信しました: %+v", event)

	// CloudWatch Logsデータをデコード
	data, err := event.AWSLogs.Parse()
	if err != nil {
		log.Printf("CloudWatch Logsデータの解析に失敗しました: %v", err)
		return err
	}

	// 各ログイベントを処理
	for _, logEvent := range data.LogEvents {
		connectionInfo := parseConnectionLog(logEvent.Message)
		if connectionInfo != nil {
			err := sendSlackNotification(*connectionInfo)
			if err != nil {
				log.Printf("Slack通知の送信に失敗しました: %v", err)
				continue
			}
			log.Printf("Slack通知を送信しました: %s", connectionInfo.UserName)
		}
	}

	return nil
}

// parseConnectionLog MySQLログメッセージから接続情報を解析
func parseConnectionLog(message string) *ConnectionInfo {
	// MySQLの接続ログパターンを解析
	// 実際の形式:
	// DB指定あり: "2025-09-07T06:41:11.701820Z	  252 Connect	test28@10.0.139.222 on appdb using TCP/IP"
	// DB指定なし: "2025-09-07T06:40:32.160890Z	  249 Connect	test28@10.0.139.222 on  using TCP/IP"
	
	connectPattern := regexp.MustCompile(`Connect`)
	if !connectPattern.MatchString(message) {
		return nil
	}

	// システムユーザーを除外
	systemUsers := []string{"rdsadmin", "mysql.session", "mysql.sys", "root"}
	for _, sysUser := range systemUsers {
		if strings.Contains(message, fmt.Sprintf(`%s@`, sysUser)) {
			return nil
		}
	}

	// MySQLのGeneral Logパターンに対応した正規表現
	// パターン1: DB指定あり "YYYY-MM-DDTHH:MM:SS.ffffffZ	ID Connect	USER@IP on DATABASE using TCP/IP"
	withDbPattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)\s+\d+\s+Connect\s+([^@]+)@([^\s]+)\s+on\s+(\w+)\s+using`)
	matches := withDbPattern.FindStringSubmatch(message)
	
	if len(matches) >= 5 {
		timestamp, err := time.Parse("2006-01-02T15:04:05.000000Z", matches[1])
		if err != nil {
			log.Printf("タイムスタンプの解析に失敗しました: %v", err)
			timestamp = time.Now()
		}

		return &ConnectionInfo{
			Timestamp:    timestamp,
			UserName:     strings.TrimSpace(matches[2]),
			DatabaseName: strings.TrimSpace(matches[4]),
			ClientIP:     strings.TrimSpace(matches[3]),
			LogMessage:   message,
		}
	}

	// パターン2: DB指定なし "YYYY-MM-DDTHH:MM:SS.ffffffZ	ID Connect	USER@IP on  using TCP/IP"
	withoutDbPattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)\s+\d+\s+Connect\s+([^@]+)@([^\s]+)\s+on\s+\s+using`)
	noDbMatches := withoutDbPattern.FindStringSubmatch(message)
	
	if len(noDbMatches) >= 4 {
		timestamp, err := time.Parse("2006-01-02T15:04:05.000000Z", noDbMatches[1])
		if err != nil {
			log.Printf("タイムスタンプの解析に失敗しました: %v", err)
			timestamp = time.Now()
		}

		return &ConnectionInfo{
			Timestamp:    timestamp,
			UserName:     strings.TrimSpace(noDbMatches[2]),
			DatabaseName: "指定なし",
			ClientIP:     strings.TrimSpace(noDbMatches[3]),
			LogMessage:   message,
		}
	}

	// パターン3: より柔軟なパターン（"on"の後に何もない、または空白のみ）
	flexiblePattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)\s+\d+\s+Connect\s+([^@]+)@([^\s]+)\s+on\s+([^\s]*)\s*using`)
	flexMatches := flexiblePattern.FindStringSubmatch(message)
	
	if len(flexMatches) >= 5 {
		timestamp, err := time.Parse("2006-01-02T15:04:05.000000Z", flexMatches[1])
		if err != nil {
			log.Printf("タイムスタンプの解析に失敗しました: %v", err)
			timestamp = time.Now()
		}

		dbName := strings.TrimSpace(flexMatches[4])
		if dbName == "" {
			dbName = "指定なし"
		}

		return &ConnectionInfo{
			Timestamp:    timestamp,
			UserName:     strings.TrimSpace(flexMatches[2]),
			DatabaseName: dbName,
			ClientIP:     strings.TrimSpace(flexMatches[3]),
			LogMessage:   message,
		}
	}

	log.Printf("MySQLログメッセージの解析に失敗しました: %s", message)
	return nil
}

// sendSlackNotification Slackに通知を送信
func sendSlackNotification(connInfo ConnectionInfo) error {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("SLACK_WEBHOOK_URL環境変数が設定されていません")
	}

	clusterID := os.Getenv("AURORA_CLUSTER_ID")
	if clusterID == "" {
		clusterID = "不明"
	}

	// Slackメッセージを構築
	message := SlackMessage{
		Username:  "Aurora MySQL Monitor",
		IconEmoji: ":dolphin:",
		Text:      "🔗 Aurora MySQLへの新しい接続が検出されました",
		Attachments: []Attachment{
			{
				Color: "#FF6B35", // MySQLのオレンジ色
				Title: "データベース接続情報",
				Fields: []Field{
					{
						Title: "ユーザー名",
						Value: connInfo.UserName,
						Short: true,
					},
					{
						Title: "データベース名",
						Value: connInfo.DatabaseName,
						Short: true,
					},
					{
						Title: "クライアントIP",
						Value: connInfo.ClientIP,
						Short: true,
					},
					{
						Title: "クラスターID",
						Value: clusterID,
						Short: true,
					},
					{
						Title: "接続時刻",
						Value: connInfo.Timestamp.Format("2006-01-02 15:04:05 JST"),
						Short: false,
					},
				},
				Timestamp: connInfo.Timestamp.Unix(),
			},
		},
	}

	// JSONにエンコード
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("JSONエンコードに失敗しました: %v", err)
	}

	// HTTPリクエストを送信
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("HTTP POSTリクエストに失敗しました: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack APIがエラーを返しました: %d", resp.StatusCode)
	}

	return nil
}