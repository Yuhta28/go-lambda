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

// parseConnectionLog ログメッセージから接続情報を解析
func parseConnectionLog(message string) *ConnectionInfo {
	// PostgreSQLの接続ログパターンを解析
	// 実際の形式: "2025-08-31 06:31:53 UTC:10.0.128.64(35340):yuta@postgres:[2599]:LOG:  connection authorized: user=yuta database=postgres application_name=psql SSL enabled"
	
	connectionPattern := regexp.MustCompile(`connection\s+(received|authorized)`)
	if !connectionPattern.MatchString(message) {
		return nil
	}

	// システムユーザーを除外
	systemUsers := []string{"rdsadmin", "rdshm"}
	for _, sysUser := range systemUsers {
		if strings.Contains(message, fmt.Sprintf(`identity="%s"`, sysUser)) ||
		   strings.Contains(message, fmt.Sprintf(`user=%s`, sysUser)) ||
		   strings.Contains(message, fmt.Sprintf(`%s@`, sysUser)) {
			return nil
		}
	}

	// 新しいログ形式に対応した正規表現
	// パターン: "YYYY-MM-DD HH:MM:SS UTC:IP(PORT):USER@DATABASE:[PID]:LOG:  connection authorized: user=USER database=DATABASE ..."
	logPattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})\s+UTC:([^:]+):\s*([^@]+)@([^:]+):\[(\d+)\]:LOG:\s+connection\s+(authorized|received):\s+user=(\w+)\s+database=(\w+)`)
	matches := logPattern.FindStringSubmatch(message)
	
	if len(matches) < 9 {
		log.Printf("ログメッセージの解析に失敗しました: %s", message)
		// フォールバック: 簡単なパターンで再試行
		simplePattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})\s+UTC:([^:]+):\s*([^@]+)@([^:]+):`)
		simpleMatches := simplePattern.FindStringSubmatch(message)
		if len(simpleMatches) >= 5 {
			timestamp, err := time.Parse("2006-01-02 15:04:05", simpleMatches[1])
			if err != nil {
				timestamp = time.Now()
			} else {
				// UTCとして解析されたタイムスタンプをJSTに変換
				jst, _ := time.LoadLocation("Asia/Tokyo")
				timestamp = timestamp.In(jst)
			}
			
			// IPアドレスから(PORT)部分を除去
			clientIP := regexp.MustCompile(`^([^(]+)`).FindStringSubmatch(simpleMatches[2])
			ip := "不明"
			if len(clientIP) > 1 {
				ip = clientIP[1]
			}
			
			return &ConnectionInfo{
				Timestamp:    timestamp,
				UserName:     strings.TrimSpace(simpleMatches[3]),
				DatabaseName: strings.TrimSpace(simpleMatches[4]),
				ClientIP:     ip,
				LogMessage:   message,
			}
		}
		return nil
	}

	timestamp, err := time.Parse("2006-01-02 15:04:05", matches[1])
	if err != nil {
		log.Printf("タイムスタンプの解析に失敗しました: %v", err)
		timestamp = time.Now()
	} else {
		// UTCとして解析されたタイムスタンプをJSTに変換
		jst, _ := time.LoadLocation("Asia/Tokyo")
		timestamp = timestamp.In(jst)
	}

	// IPアドレスから(PORT)部分を除去
	clientIPRaw := matches[2]
	clientIP := regexp.MustCompile(`^([^(]+)`).FindStringSubmatch(clientIPRaw)
	ip := "不明"
	if len(clientIP) > 1 {
		ip = clientIP[1]
	}

	return &ConnectionInfo{
		Timestamp:    timestamp,
		UserName:     strings.TrimSpace(matches[7]), // user=XXX から抽出
		DatabaseName: strings.TrimSpace(matches[8]), // database=XXX から抽出
		ClientIP:     ip,
		LogMessage:   message,
	}
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
		Username:  "Aurora DB Monitor",
		IconEmoji: ":shark:",
		Text:      "🔗 Aurora PostgreSQLへの新しい接続が検出されました",
		Attachments: []Attachment{
			{
				Color: "good",
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