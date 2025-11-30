package commands

import (
	"testing"
)

func TestRegisterCmd_ValidateArgs(t *testing.T) {
	cmd := registerCmd{}

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid flags",
			args: map[string]any{
				"name": "repo1",
				"path": "./repo1",
			},
			wantErr: false,
		},
		{
			name: "Valid positional",
			args: map[string]any{
				"args": []string{"repo1;./repo1"},
			},
			wantErr: false,
		},
		{
			name: "Valid mixed (flags and positional)",
			args: map[string]any{
				"name": "repo1",
				"path": "./repo1",
				"args": []string{"repo2;./repo2"},
			},
			wantErr: false,
		},
		{
			name: "Missing name flag",
			args: map[string]any{
				"path": "./repo1",
			},
			wantErr: true,
		},
		{
			name: "Missing path flag",
			args: map[string]any{
				"name": "repo1",
			},
			wantErr: true,
		},
		{
			name:    "No args",
			args:    map[string]any{},
			wantErr: true,
		},
		{
			name: "Empty positional list",
			args: map[string]any{
				"args": []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cmd.ValidateArgs(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("ValidateArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
