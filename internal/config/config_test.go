package config

import "testing"

func fakeGetenv(values map[string]string) func(string) string {
	return func(key string) string {
		return values[key]
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		values  map[string]string
		wantErr bool
		want    Config
	}{
		{
			name: "正常系: 必須項目が揃っていれば読み込める",
			values: map[string]string{
				"DISCORD_BOT_TOKEN": "token123",
				"DATABASE_URL":      "postgres://localhost/db",
			},
			wantErr: false,
			want: Config{
				DiscordBotToken: "token123",
				DatabaseURL:     "postgres://localhost/db",
				AdminPort:       "8080",
			},
		},
		{
			name: "正常系: 任意項目を指定するとデフォルト値を上書きする",
			values: map[string]string{
				"DISCORD_BOT_TOKEN": "token123",
				"DATABASE_URL":      "postgres://localhost/db",
				"ADMIN_PORT":        "9090",
			},
			wantErr: false,
			want: Config{
				DiscordBotToken: "token123",
				DatabaseURL:     "postgres://localhost/db",
				AdminPort:       "9090",
			},
		},
		{
			name: "異常系: DISCORD_BOT_TOKENが未設定だとエラーになる",
			values: map[string]string{
				"DATABASE_URL": "postgres://localhost/db",
			},
			wantErr: true,
		},
		{
			name: "異常系: DATABASE_URLが未設定だとエラーになる",
			values: map[string]string{
				"DISCORD_BOT_TOKEN": "token123",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(fakeGetenv(tt.values))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("Load() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
