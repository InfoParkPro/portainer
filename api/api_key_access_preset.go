package portainer

// EffectiveAccessPreset returns the API key access preset currently applied to
// requests authenticated with this key.
func (apiKey APIKey) EffectiveAccessPreset(now int64) APIKeyAccessPreset {
	basePreset := apiKey.AccessPreset
	if basePreset == "" {
		basePreset = APIKeyAccessPresetManage
	}

	if basePreset == APIKeyAccessPresetDisabled {
		return basePreset
	}

	if apiKey.TemporaryAccessPreset == "" || apiKey.TemporaryAccessExpiresAt <= now {
		return basePreset
	}

	if apiKeyAccessPresetRank(apiKey.TemporaryAccessPreset) <= apiKeyAccessPresetRank(basePreset) {
		return basePreset
	}

	return apiKey.TemporaryAccessPreset
}

func apiKeyAccessPresetRank(preset APIKeyAccessPreset) int {
	switch preset {
	case APIKeyAccessPresetDisabled:
		return 0
	case APIKeyAccessPresetReadOnly:
		return 1
	case APIKeyAccessPresetPower:
		return 2
	case "", APIKeyAccessPresetManage:
		return 3
	default:
		return 0
	}
}
