package commands

import (
	"testing"
)

func TestStageCmd_ValidateArgs(t *testing.T) {
	cmd := StageCommand{}

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid args",
			args: map[string]any{
				"args": []string{"file1.txt"},
			},
			wantErr: false,
		},
		{
			name: "Valid dot",
			args: map[string]any{
				"args": []string{"."},
			},
			wantErr: false,
		},
		{
			name:    "No args",
			args:    map[string]any{},
			wantErr: true,
		},
		{
			name: "Empty args list",
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
