package core

import (
	"strings"
	"testing"
)

func anonymousTestToken() string        { return anonymousTokenConstructor() }
func anonymousTokenConstructor() string { return AutoKey("button") }

func TestAutoKeyUsesComponentPrefix(t *testing.T) {
	var keys Keys
	keys.BeginFrame()
	for _, component := range []string{"button", "input", "checkbox"} {
		token := AutoKey(component)
		if !strings.HasPrefix(token, "\x00"+component+":") {
			t.Fatalf("token = %q, want component prefix %q", token, component)
		}
		resolved := keys.FullKey("tooltip:" + token + "-input")
		wantPrefix := "~rtooltip:%00" + component + ":"
		if !strings.HasPrefix(resolved, wantPrefix) || !strings.HasSuffix(resolved, ":0-input") {
			t.Fatalf("resolved key = %q, want component prefix and independent layout index", resolved)
		}
	}
}

func TestAutoKeysResolveAtLayoutPerContext(t *testing.T) {
	fixed := anonymousTestToken()
	var first, second Keys
	first.BeginFrame()
	fixedKey := first.Claim(KindClickable, fixed)
	dynamicKey := first.Claim(KindClickable, anonymousTestToken())
	if fixedKey == dynamicKey {
		t.Fatal("prebuilt and rebuilt widgets share an identity")
	}
	second.BeginFrame()
	second.Claim(KindClickable, anonymousTestToken())
	if got := first.FullKey(fixed); got != fixedKey {
		t.Fatalf("another context changed identity: %q, want %q", got, fixedKey)
	}
	first.BeginFrame()
	if got := first.Claim(KindClickable, fixed); got != fixedKey {
		t.Fatalf("prebuilt identity changed: %q, want %q", got, fixedKey)
	}
	if got := first.Claim(KindClickable, anonymousTestToken()); got != dynamicKey {
		t.Fatalf("rebuilt identity changed: %q, want %q", got, dynamicKey)
	}
}

func TestAutoKeySuffixSharesOwnerIdentity(t *testing.T) {
	var keys Keys
	keys.BeginFrame()
	token := anonymousTestToken()
	owner := keys.FullKey(token)
	child := keys.FullKey(token + "-input")
	keys.BeginFrame()
	next := anonymousTestToken()
	if got := keys.FullKey(next + "-input"); got != child {
		t.Fatalf("child identity changed: %q, want %q", got, child)
	}
	if got := keys.FullKey(next); got != owner {
		t.Fatalf("owner identity changed: %q, want %q", got, owner)
	}
}

func TestAutoKeyPreservesExplicitKeys(t *testing.T) {
	var keys Keys
	keys.BeginFrame()
	const explicit = "\x00button:page.go:42\x00123\x00"
	if got, want := keys.FullKey(explicit), keys.fullKey(explicit); got != want {
		t.Fatalf("explicit key was interpreted as a token: %q, want %q", got, want)
	}
}

func TestAutoKeyResolutionUsesEncodedScope(t *testing.T) {
	for _, path := range [][]string{{"a/b"}, {"a\x00b"}} {
		var keys Keys
		keys.BeginFrame()
		keys.Push("a")
		keys.Push("b")
		keys.FullKey(anonymousTestToken())
		keys.FullKey(anonymousTestToken())
		restore := keys.UseScope(path)
		first := keys.FullKey(anonymousTestToken())
		restore()
		keys.BeginFrame()
		restore = keys.UseScope(path)
		if got := keys.FullKey(anonymousTestToken()); got != first {
			t.Fatalf("identity depends on another scope for path %q: %q, want %q", path, got, first)
		}
		restore()
	}
}

func TestAnonymousKeysCountDistinctScopesIndependently(t *testing.T) {
	var keys Keys
	keys.BeginFrame()
	keys.Push("a")
	keys.Push("b")
	keys.Claim(KindClickable, "")
	keys.Claim(KindClickable, "")
	restore := keys.UseScope([]string{"a/b"})
	first := keys.Claim(KindClickable, "")
	restore()
	keys.BeginFrame()
	restore = keys.UseScope([]string{"a/b"})
	defer restore()
	if got := keys.Claim(KindClickable, ""); got != first {
		t.Fatalf("another path changed the anonymous position: %q, want %q", got, first)
	}
}

func TestMeasurementRestoresAutoKeyOrder(t *testing.T) {
	ctx := new(Context)
	ctx.keys.BeginFrame()
	fixed := anonymousTestToken()
	first := ctx.keys.FullKey(fixed)
	restore := PushMeasurement(ctx)
	ctx.keys.FullKey(anonymousTestToken())
	restore()
	second := ctx.keys.FullKey(anonymousTestToken())
	ctx.keys.BeginFrame()
	if got := ctx.keys.FullKey(fixed); got != first {
		t.Fatalf("measurement changed first identity: %q, want %q", got, first)
	}
	if got := ctx.keys.FullKey(anonymousTestToken()); got != second {
		t.Fatalf("measurement consumed a position: %q, want %q", got, second)
	}
}
