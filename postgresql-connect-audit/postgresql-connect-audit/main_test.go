package main

import (
	"testing"
	"time"
)

func TestParseConnectionLog(t *testing.T) {
	testCases := []struct {
		name     string
		message  string
		expected *ConnectionInfo
	}{
		{
			name:    "有効な接続ログ（新形式）",
			message: "2025-08-31 06:31:53 UTC:10.0.128.64(35340):yuta@postgres:[2599]:LOG:  connection authorized: user=yuta database=postgres application_name=psql SSL enabled",
			expected: &ConnectionInfo{
				UserName:     "yuta",
				DatabaseName: "postgres",
				ClientIP:     "10.0.128.64",
			},
		},
		{
			name:     "無効なログメッセージ",
			message:  "2024-01-01 12:00:00 UTC: some other log message",
			expected: nil,
		},
		{
			name:    "フォールバックパターン",
			message: "2024-01-01 12:00:00 UTC:10.0.1.50(54321):admin@postgres:[54321]:LOG: connection authorized",
			expected: &ConnectionInfo{
				UserName:     "admin",
				DatabaseName: "postgres",
				ClientIP:     "10.0.1.50",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseConnectionLog(tc.message)
			
			if tc.expected == nil {
				if result != nil {
					t.Errorf("期待値: nil, 実際: %+v", result)
				}
				return
			}
			
			if result == nil {
				t.Errorf("期待値: %+v, 実際: nil", tc.expected)
				return
			}
			
			if result.UserName != tc.expected.UserName {
				t.Errorf("ユーザー名 - 期待値: %s, 実際: %s", tc.expected.UserName, result.UserName)
			}
			
			if result.DatabaseName != tc.expected.DatabaseName {
				t.Errorf("データベース名 - 期待値: %s, 実際: %s", tc.expected.DatabaseName, result.DatabaseName)
			}
			
			if result.ClientIP != tc.expected.ClientIP {
				t.Errorf("クライアントIP - 期待値: %s, 実際: %s", tc.expected.ClientIP, result.ClientIP)
			}
		})
	}
}

func TestConnectionInfoValidation(t *testing.T) {
	connInfo := ConnectionInfo{
		Timestamp:    time.Now(),
		UserName:     "testuser",
		DatabaseName: "testdb",
		ClientIP:     "192.168.1.100",
		LogMessage:   "test log message",
	}

	if connInfo.UserName == "" {
		t.Error("ユーザー名が空です")
	}
	
	if connInfo.DatabaseName == "" {
		t.Error("データベース名が空です")
	}
	
	if connInfo.ClientIP == "" {
		t.Error("クライアントIPが空です")
	}
}