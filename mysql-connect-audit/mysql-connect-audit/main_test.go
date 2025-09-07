package main

import (
	"testing"
)

func TestParseConnectionLog(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected *ConnectionInfo
	}{
		{
			name:    "MySQL Connect with database specified",
			message: "2025-09-07T06:41:11.701820Z	  252 Connect	test28@10.0.139.222 on appdb using TCP/IP",
			expected: &ConnectionInfo{
				UserName:     "test28",
				DatabaseName: "appdb",
				ClientIP:     "10.0.139.222",
			},
		},
		{
			name:    "MySQL Connect without database specified",
			message: "2025-09-07T06:40:32.160890Z	  249 Connect	test28@10.0.139.222 on  using TCP/IP",
			expected: &ConnectionInfo{
				UserName:     "test28",
				DatabaseName: "指定なし",
				ClientIP:     "10.0.139.222",
			},
		},
		{
			name:    "MySQL Connect with different user and IP",
			message: "2024-01-01T12:00:00.123456Z	    1 Connect	myuser@192.168.1.100 on mydb using TCP/IP",
			expected: &ConnectionInfo{
				UserName:     "myuser",
				DatabaseName: "mydb",
				ClientIP:     "192.168.1.100",
			},
		},
		{
			name:    "MySQL Connect with empty database name",
			message: "2024-01-01T12:00:00.123456Z	    1 Connect	admin@172.16.0.1 on  using TCP/IP",
			expected: &ConnectionInfo{
				UserName:     "admin",
				DatabaseName: "指定なし",
				ClientIP:     "172.16.0.1",
			},
		},
		{
			name:     "System user should be ignored",
			message:  "2024-01-01T12:00:00.123456Z	    1 Connect	rdsadmin@localhost on mysql using TCP/IP",
			expected: nil,
		},
		{
			name:     "Non-connect message should be ignored",
			message:  "2024-01-01T12:00:00.123456Z	    1 Query	SELECT * FROM users",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseConnectionLog(tt.message)
			
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, but got %+v", result)
				}
				return
			}
			
			if result == nil {
				t.Errorf("Expected %+v, but got nil", tt.expected)
				return
			}
			
			if result.UserName != tt.expected.UserName {
				t.Errorf("UserName: expected %s, got %s", tt.expected.UserName, result.UserName)
			}
			
			if result.DatabaseName != tt.expected.DatabaseName {
				t.Errorf("DatabaseName: expected %s, got %s", tt.expected.DatabaseName, result.DatabaseName)
			}
			
			if result.ClientIP != tt.expected.ClientIP {
				t.Errorf("ClientIP: expected %s, got %s", tt.expected.ClientIP, result.ClientIP)
			}
		})
	}
}

func TestParseActualMySQLLogs(t *testing.T) {
	// 実際のMySQLログ形式でのテスト
	actualLogs := []struct {
		name     string
		message  string
		expected *ConnectionInfo
	}{
		{
			name:    "Actual MySQL log with database",
			message: "2025-09-07T06:41:11.701820Z	  252 Connect	test28@10.0.139.222 on appdb using TCP/IP",
			expected: &ConnectionInfo{
				UserName:     "test28",
				DatabaseName: "appdb",
				ClientIP:     "10.0.139.222",
			},
		},
		{
			name:    "Actual MySQL log without database",
			message: "2025-09-07T06:40:32.160890Z	  249 Connect	test28@10.0.139.222 on  using TCP/IP",
			expected: &ConnectionInfo{
				UserName:     "test28",
				DatabaseName: "指定なし",
				ClientIP:     "10.0.139.222",
			},
		},
	}

	for _, tt := range actualLogs {
		t.Run(tt.name, func(t *testing.T) {
			result := parseConnectionLog(tt.message)
			
			if result == nil {
				t.Errorf("Expected %+v, but got nil", tt.expected)
				return
			}
			
			if result.UserName != tt.expected.UserName {
				t.Errorf("UserName: expected %s, got %s", tt.expected.UserName, result.UserName)
			}
			
			if result.DatabaseName != tt.expected.DatabaseName {
				t.Errorf("DatabaseName: expected %s, got %s", tt.expected.DatabaseName, result.DatabaseName)
			}
			
			if result.ClientIP != tt.expected.ClientIP {
				t.Errorf("ClientIP: expected %s, got %s", tt.expected.ClientIP, result.ClientIP)
			}

			// タイムスタンプが正しく解析されているかチェック
			expectedTime := "2025-09-07 06:41:11"
			if tt.name == "Actual MySQL log without database" {
				expectedTime = "2025-09-07 06:40:32"
			}
			actualTime := result.Timestamp.Format("2006-01-02 15:04:05")
			if actualTime != expectedTime {
				t.Errorf("Timestamp: expected %s, got %s", expectedTime, actualTime)
			}
		})
	}
}