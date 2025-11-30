package commands

import (
	"testing"
)

func TestLinkCmd_ValidateArgs(t *testing.T) {
	cmd := linkCmd{}

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid flags",
			args: map[string]any{
				"child":  "svc1",
				"parent": "backend",
			},
			wantErr: false,
		},
		{
			name: "Valid positional",
			args: map[string]any{
				"args": []string{"svc1;backend"},
			},
			wantErr: false,
		},
		{
			name: "Valid mixed",
			args: map[string]any{
				"child":  "svc1",
				"parent": "backend",
				"args":   []string{"svc2;backend"},
			},
			wantErr: false,
		},
		{
			name: "Missing child flag",
			args: map[string]any{
				"parent": "backend",
			},
			wantErr: true,
		},
		{
			name: "Missing parent flag",
			args: map[string]any{
				"child": "svc1",
			},
			wantErr: true,
		},
		{
			name:    "No args",
			args:    map[string]any{},
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
