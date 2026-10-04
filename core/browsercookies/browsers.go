package browsercookies

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type browserFamily int

const (
	familyFirefox browserFamily = iota
	familyChromium
)

// browserSpec describes where a browser keeps its profiles on each OS.
type browserSpec struct {
	name   string
	family browserFamily
	// roots maps GOOS to profile roots. A leading "~/" is the home directory,
	// "%APPDATA%" and "%LOCALAPPDATA%" the Windows application data folders.
	roots map[string][]string
	// keyringName is the "<name> Safe Storage" entry of Chromium browsers.
	keyringName string
}

var browserSpecs = []browserSpec{
	{name: "Firefox", family: familyFirefox, roots: map[string][]string{
		"linux":   {"~/.mozilla/firefox", "~/snap/firefox/common/.mozilla/firefox", "~/.var/app/org.mozilla.firefox/.mozilla/firefox"},
		"darwin":  {"~/Library/Application Support/Firefox/Profiles"},
		"windows": {`%APPDATA%\Mozilla\Firefox\Profiles`},
	}},
	{name: "LibreWolf", family: familyFirefox, roots: map[string][]string{
		"linux":   {"~/.librewolf", "~/.var/app/io.gitlab.librewolf-community/.librewolf"},
		"darwin":  {"~/Library/Application Support/LibreWolf/Profiles"},
		"windows": {`%APPDATA%\LibreWolf\Profiles`},
	}},
	{name: "Zen", family: familyFirefox, roots: map[string][]string{
		"linux":   {"~/.zen", "~/.var/app/app.zen_browser.zen/.zen"},
		"darwin":  {"~/Library/Application Support/zen/Profiles"},
		"windows": {`%APPDATA%\zen\Profiles`},
	}},
	{name: "Floorp", family: familyFirefox, roots: map[string][]string{
		"linux":   {"~/.floorp"},
		"darwin":  {"~/Library/Application Support/Floorp/Profiles"},
		"windows": {`%APPDATA%\Floorp\Profiles`},
	}},
	{name: "Waterfox", family: familyFirefox, roots: map[string][]string{
		"linux":   {"~/.waterfox"},
		"darwin":  {"~/Library/Application Support/Waterfox/Profiles"},
		"windows": {`%APPDATA%\Waterfox\Profiles`},
	}},
	{name: "Chrome", family: familyChromium, keyringName: "Chrome", roots: map[string][]string{
		"linux":   {"~/.config/google-chrome"},
		"darwin":  {"~/Library/Application Support/Google/Chrome"},
		"windows": {`%LOCALAPPDATA%\Google\Chrome\User Data`},
	}},
	{name: "Chromium", family: familyChromium, keyringName: "Chromium", roots: map[string][]string{
		"linux":   {"~/.config/chromium"},
		"darwin":  {"~/Library/Application Support/Chromium"},
		"windows": {`%LOCALAPPDATA%\Chromium\User Data`},
	}},
	{name: "Brave", family: familyChromium, keyringName: "Brave", roots: map[string][]string{
		"linux":   {"~/.config/BraveSoftware/Brave-Browser"},
		"darwin":  {"~/Library/Application Support/BraveSoftware/Brave-Browser"},
		"windows": {`%LOCALAPPDATA%\BraveSoftware\Brave-Browser\User Data`},
	}},
	{name: "Edge", family: familyChromium, keyringName: "Chromium", roots: map[string][]string{
		"linux":   {"~/.config/microsoft-edge"},
		"darwin":  {"~/Library/Application Support/Microsoft Edge"},
		"windows": {`%LOCALAPPDATA%\Microsoft\Edge\User Data`},
	}},
	{name: "Vivaldi", family: familyChromium, keyringName: "Chrome", roots: map[string][]string{
		"linux":   {"~/.config/vivaldi"},
		"darwin":  {"~/Library/Application Support/Vivaldi"},
		"windows": {`%LOCALAPPDATA%\Vivaldi\User Data`},
	}},
}

// macKeyringName is the keychain account of Chromium browsers that differs
// from their Linux keyring name.
var macKeyringNames = map[string]string{"Edge": "Microsoft Edge", "Vivaldi": "Vivaldi"}

// profile is one browser profile with a cookie database.
type profile struct {
	spec         browserSpec
	name         string
	databasePath string
}

// environment is what discovery reads: the OS, the user's folders and the key sources.
type environment struct {
	goos            string
	homeDir         string
	appDataDir      string
	localAppDataDir string
	tempDir         string
	password        keyringPassword
}

func (env environment) expandRoot(root string) string {
	switch {
	case strings.HasPrefix(root, "~/"):
		return filepath.Join(env.homeDir, filepath.FromSlash(strings.TrimPrefix(root, "~/")))
	case strings.HasPrefix(root, `%APPDATA%\`):
		return filepath.Join(env.appDataDir, filepath.FromSlash(strings.ReplaceAll(strings.TrimPrefix(root, `%APPDATA%\`), `\`, "/")))
	case strings.HasPrefix(root, `%LOCALAPPDATA%\`):
		return filepath.Join(env.localAppDataDir, filepath.FromSlash(strings.ReplaceAll(strings.TrimPrefix(root, `%LOCALAPPDATA%\`), `\`, "/")))
	default:
		return root
	}
}

// profiles lists the profiles of every known browser that have a cookie database.
func (env environment) profiles() []profile {
	found := []profile{}
	for _, spec := range browserSpecs {
		for _, root := range spec.roots[env.goos] {
			found = append(found, env.profilesIn(spec, env.expandRoot(root))...)
		}
	}
	return found
}

func (env environment) profilesIn(spec browserSpec, root string) []profile {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	found := []profile{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		databasePath := cookieDatabasePath(spec.family, filepath.Join(root, entry.Name()))
		if databasePath != "" {
			found = append(found, profile{spec: spec, name: entry.Name(), databasePath: databasePath})
		}
	}
	sort.Slice(found, func(left, right int) bool { return found[left].name < found[right].name })
	return found
}

// cookieDatabasePath returns the cookie database of a profile directory, or "" when it has none.
func cookieDatabasePath(family browserFamily, profileDir string) string {
	candidates := []string{filepath.Join(profileDir, "cookies.sqlite")}
	if family == familyChromium {
		candidates = []string{filepath.Join(profileDir, "Network", "Cookies"), filepath.Join(profileDir, "Cookies")}
	}
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
}
