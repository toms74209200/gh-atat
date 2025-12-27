package config

import (
	"maps"
	"reflect"
	"testing"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		wantErr     bool
		wantConfig  map[ConfigKey]any
		checkKey    ConfigKey
		keyExists   bool
		expectedVal any
	}{
		{
			name:        "object with key",
			input:       []byte(`{"repositories": ["owner/repo1", "another/repo2"]}`),
			wantErr:     false,
			checkKey:    Repositories,
			keyExists:   true,
			expectedVal: []any{"owner/repo1", "another/repo2"},
		},
		{
			name:        "empty object key",
			input:       []byte(`{"repositories": []}`),
			wantErr:     false,
			checkKey:    Repositories,
			keyExists:   true,
			expectedVal: []any{},
		},
		{
			name:       "empty input",
			input:      []byte(``),
			wantErr:    false,
			wantConfig: map[ConfigKey]any{},
		},
		{
			name:       "whitespace input",
			input:      []byte(`   `),
			wantErr:    false,
			wantConfig: map[ConfigKey]any{},
		},
		{
			name:    "invalid JSON",
			input:   []byte(`["owner/repo1"`),
			wantErr: true,
		},
		{
			name:    "invalid syntax",
			input:   []byte(`{invalid json}`),
			wantErr: true,
		},
		{
			name:      "unknown key skipped",
			input:     []byte(`{"unknown": "value"}`),
			wantErr:   false,
			checkKey:  Repositories,
			keyExists: false,
		},
		{
			name:    "valid JSON array",
			input:   []byte(`["value1", "value2"]`),
			wantErr: true,
		},
		{
			name:    "string primitive",
			input:   []byte(`"string value"`),
			wantErr: true,
		},
		{
			name:    "number primitive",
			input:   []byte(`123`),
			wantErr: true,
		},
		{
			name:    "boolean primitive",
			input:   []byte(`true`),
			wantErr: true,
		},
		{
			name:    "null primitive",
			input:   []byte(`null`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseConfig(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			// Check if specific config is expected
			if tt.wantConfig != nil {
				if !reflect.DeepEqual(config, tt.wantConfig) {
					t.Errorf("ParseConfig() = %v, want %v", config, tt.wantConfig)
				}
				return
			}

			// Check key existence
			if tt.checkKey != "" {
				_, exists := config[tt.checkKey]
				if exists != tt.keyExists {
					t.Errorf("key %v exists = %v, want %v", tt.checkKey, exists, tt.keyExists)
				}

				// Check expected value if key should exist
				if tt.keyExists && tt.expectedVal != nil {
					if !reflect.DeepEqual(config[tt.checkKey], tt.expectedVal) {
						t.Errorf("config[%v] = %v, want %v", tt.checkKey, config[tt.checkKey], tt.expectedVal)
					}
				}
			}
		})
	}
}

func TestUpdateConfig(t *testing.T) {
	tests := []struct {
		name               string
		baseConfig         map[ConfigKey]any
		updates            map[ConfigKey]any
		expectedResult     map[ConfigKey]any
		checkBaseUnchanged bool
	}{
		{
			name:       "add new key to empty base",
			baseConfig: make(map[ConfigKey]any),
			updates: map[ConfigKey]any{
				Repositories: []any{"owner/repo1"},
			},
			expectedResult: map[ConfigKey]any{
				Repositories: []any{"owner/repo1"},
			},
			checkBaseUnchanged: true,
		},
		{
			name: "overwrite existing key",
			baseConfig: map[ConfigKey]any{
				Repositories: []any{"owner/repo1"},
			},
			updates: map[ConfigKey]any{
				Repositories: []any{"owner/repo2"},
			},
			expectedResult: map[ConfigKey]any{
				Repositories: []any{"owner/repo2"},
			},
			checkBaseUnchanged: true,
		},
		{
			name: "overwrite with different type",
			baseConfig: map[ConfigKey]any{
				Repositories: []any{"owner/repo1"},
			},
			updates: map[ConfigKey]any{
				Repositories: "owner/repo2",
			},
			expectedResult: map[ConfigKey]any{
				Repositories: "owner/repo2",
			},
			checkBaseUnchanged: true,
		},
		{
			name: "empty updates",
			baseConfig: map[ConfigKey]any{
				Repositories: []any{"owner/repo1"},
			},
			updates: make(map[ConfigKey]any),
			expectedResult: map[ConfigKey]any{
				Repositories: []any{"owner/repo1"},
			},
			checkBaseUnchanged: true,
		},
		{
			name:       "empty base with updates",
			baseConfig: make(map[ConfigKey]any),
			updates: map[ConfigKey]any{
				Repositories: []any{"owner/repo1"},
			},
			expectedResult: map[ConfigKey]any{
				Repositories: []any{"owner/repo1"},
			},
			checkBaseUnchanged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clone base config to verify it's not modified
			originalBase := maps.Clone(tt.baseConfig)

			resultConfig := UpdateConfig(tt.baseConfig, tt.updates)

			// Check result
			if !reflect.DeepEqual(resultConfig, tt.expectedResult) {
				t.Errorf("UpdateConfig() = %v, want %v", resultConfig, tt.expectedResult)
			}

			// Check that base config is not modified
			if tt.checkBaseUnchanged && !reflect.DeepEqual(tt.baseConfig, originalBase) {
				t.Errorf("baseConfig was modified: got %v, want %v", tt.baseConfig, originalBase)
			}
		})
	}
}
