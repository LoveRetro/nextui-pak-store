// Package i18n provides a tiny key→value lookup table backed by NextUI's
// shared lang files. It is intentionally allocation-light and safe to
// call concurrently after Init.
//
// Lang files live in /mnt/SDCARD/.system/res/lang/<code>.lang (the same
// directory the NextUI launcher reads). The active language code is
// read from /mnt/SDCARD/.userdata/shared/minuisettings.txt
// ("language=fr", "language=en", ...). Missing files, missing keys, or
// any I/O failure degrade silently — T(key) returns the key as-is.
package i18n

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	defaultLangDir      = "/mnt/SDCARD/.system/res/lang"
	defaultSettingsFile = "/mnt/SDCARD/.userdata/shared/minuisettings.txt"
	fallbackCode        = "en"
)

var (
	mu      sync.RWMutex
	table   = map[string]string{}
	active  = fallbackCode
	inited  bool
	langDir = defaultLangDir
)

// Init reads minuisettings.txt to discover the active language and
// loads the matching lang file (with the English file underneath so
// untranslated keys fall back to their English text). Safe to call
// multiple times.
func Init() {
	mu.Lock()
	defer mu.Unlock()

	code := readLangCode(defaultSettingsFile)
	load(code)
	active = code
	inited = true
}

// Reload re-reads minuisettings.txt and the lang files. Use when the
// user changes language without restarting the app.
func Reload() {
	Init()
}

// Active returns the currently active language code (e.g. "fr", "en").
func Active() string {
	mu.RLock()
	defer mu.RUnlock()
	return active
}

// T looks up key in the translation table. If the key is not found,
// the key itself is returned — this keeps the caller readable when no
// translation has been written yet.
func T(key string) string {
	mu.RLock()
	defer mu.RUnlock()
	if v, ok := table[key]; ok {
		return v
	}
	return key
}

// Tf is like T followed by fmt.Sprintf. Use for keys that contain
// %s/%d placeholders so callers do not need to import "fmt" just to
// format a translated string.
func Tf(key string, args ...any) string {
	return fmt.Sprintf(T(key), args...)
}

func load(code string) {
	table = map[string]string{}

	// English baseline so untranslated keys still render in English.
	parseInto(filepath.Join(langDir, "en.lang"), table)
	if code != "" && code != "en" {
		parseInto(filepath.Join(langDir, code+".lang"), table)
	}
}

func parseInto(path string, dst map[string]string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		// Decode common escapes so multi-line values can sit on a
		// single line in the .lang file ("Foo\\nBar" → "Foo\nBar").
		val = strings.NewReplacer(`\n`, "\n", `\t`, "\t", `\\`, `\`).Replace(val)
		dst[key] = val
	}
}

func readLangCode(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return fallbackCode
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if strings.HasPrefix(line, "language=") {
			code := strings.TrimSpace(strings.TrimPrefix(line, "language="))
			if code != "" {
				return code
			}
		}
	}
	return fallbackCode
}
