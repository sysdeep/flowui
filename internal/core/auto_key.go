package core

import (
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
)

var autoKeySequence atomic.Uint64

// autoTokenMarker reserves an internal format distinct from ordinary keys that
// happen to contain NULs and numbers. Explicit application keys must not use it.
const autoTokenMarker = "\x00@"

type autoTokenScope struct {
	scope string
	token string
}

// AutoKey returns a construction token, resolved per Context at layout time.
// Identity depends on component, call site, scope and first-resolution order.
// Dynamic content uses explicit keys or stable layout.Key scopes instead.
func AutoKey(component string) string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file, line = "unknown", 0
	}
	// Constructors only mint a token. Each Context resolves its site and layout
	// order independently, including widgets built before the frame starts.
	site := filepath.ToSlash(file) + ":" + strconv.Itoa(line)
	return "\x00" + component + ":" + site + autoTokenMarker +
		strconv.FormatUint(autoKeySequence.Add(1), 10) + "\x00"
}

// nextAutoKey is shared by constructor tokens and the empty-key fallback.
// Encoding the scope before counting keeps distinct path segments independent.
func (k *Keys) nextAutoKey(site string) string {
	if k.autoCounts == nil {
		k.autoCounts = make(map[string]int)
	}
	counter := k.fullKey(site)
	index := k.autoCounts[counter]
	k.autoCounts[counter] = index + 1
	return site + ":" + strconv.Itoa(index)
}

type autoKeyParts struct {
	prefix string
	site   string
	token  string
	suffix string
}

func parseAutoKey(key string) (autoKeyParts, bool) {
	start := strings.IndexByte(key, 0)
	if start < 0 {
		return autoKeyParts{}, false
	}
	header, sequence, ok := strings.Cut(key[start:], autoTokenMarker)
	if !ok || strings.IndexByte(header, ':') <= 1 || strings.ContainsRune(header[1:], 0) {
		return autoKeyParts{}, false
	}
	id, suffix, ok := strings.Cut(sequence, "\x00")
	if !ok {
		return autoKeyParts{}, false
	}
	if _, err := strconv.ParseUint(id, 10, 64); err != nil {
		return autoKeyParts{}, false
	}
	end := len(key) - len(suffix)
	return autoKeyParts{prefix: key[:start], site: header, token: key[start:end], suffix: suffix}, true
}

func (k *Keys) resolveAutoKey(key string) string {
	parts, ok := parseAutoKey(key)
	if !ok {
		return key
	}
	if k.autoTokens == nil {
		k.autoTokens = make(map[autoTokenScope]string)
	}
	identity := autoTokenScope{scope: k.fullKey(""), token: parts.token}
	resolved, ok := k.autoTokens[identity]
	if !ok {
		resolved = k.nextAutoKey(parts.site)
		k.autoTokens[identity] = resolved
	}
	return parts.prefix + resolved + parts.suffix
}
