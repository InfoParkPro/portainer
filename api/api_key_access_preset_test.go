package portainer

import "testing"

func TestAPIKeyEffectiveAccessPreset(t *testing.T) {
	now := int64(100)

	tests := []struct {
		name   string
		apiKey APIKey
		want   APIKeyAccessPreset
	}{
		{
			name:   "legacy empty preset is manage",
			apiKey: APIKey{},
			want:   APIKeyAccessPresetManage,
		},
		{
			name: "active temporary elevation is applied",
			apiKey: APIKey{
				AccessPreset:             APIKeyAccessPresetPower,
				TemporaryAccessPreset:    APIKeyAccessPresetManage,
				TemporaryAccessExpiresAt: now + 1,
			},
			want: APIKeyAccessPresetManage,
		},
		{
			name: "expired temporary elevation is ignored",
			apiKey: APIKey{
				AccessPreset:             APIKeyAccessPresetPower,
				TemporaryAccessPreset:    APIKeyAccessPresetManage,
				TemporaryAccessExpiresAt: now,
			},
			want: APIKeyAccessPresetPower,
		},
		{
			name: "temporary lower preset is ignored",
			apiKey: APIKey{
				AccessPreset:             APIKeyAccessPresetPower,
				TemporaryAccessPreset:    APIKeyAccessPresetReadOnly,
				TemporaryAccessExpiresAt: now + 1,
			},
			want: APIKeyAccessPresetPower,
		},
		{
			name: "disabled base cannot be elevated",
			apiKey: APIKey{
				AccessPreset:             APIKeyAccessPresetDisabled,
				TemporaryAccessPreset:    APIKeyAccessPresetManage,
				TemporaryAccessExpiresAt: now + 1,
			},
			want: APIKeyAccessPresetDisabled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.apiKey.EffectiveAccessPreset(now); got != tt.want {
				t.Fatalf("EffectiveAccessPreset() = %q, want %q", got, tt.want)
			}
		})
	}
}
