package birc

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeNick(t *testing.T) {
	assert.Equal(t, "alice", sanitizeNick("alice"))
	assert.Equal(t, "J0hnny-Xm4s", sanitizeNick("J0hnny Xm4s"))
	assert.Equal(t, "a-b-c", sanitizeNick("a:b c"))
}

func TestRelayMsgNick(t *testing.T) {
	// basic case: sanitized name plus "/<protocol>" suffix
	assert.Equal(t, "alice/discord", relayMsgNick("alice", "discord", 32))

	// spaces/punctuation in the display name get sanitized before the
	// suffix is appended
	assert.Equal(t, "J0hnny-Xm4s/discord", relayMsgNick("J0hnny Xm4s", "discord", 32))

	// no protocol: just the sanitized nick, no trailing separator
	assert.Equal(t, "alice", relayMsgNick("alice", "", 32))

	// NICKLEN <= 0 (unknown): falls back to a conservative default rather
	// than truncating everything away or leaving it unbounded
	got := relayMsgNick("a-fairly-long-display-name-here", "discord", 0)
	assert.LessOrEqual(t, len(got), 30)
	assert.Contains(t, got, "/discord")

	// truncation must preserve the full suffix - a suffix that gets cut
	// off reproduces the exact bug this function exists to prevent
	got = relayMsgNick("a-fairly-long-display-name-here", "discord", 15)
	assert.Equal(t, "a-fairl/discord", got)
	assert.Equal(t, 15, len(got))

	// truncation must not split a multi-byte UTF-8 rune
	got = relayMsgNick("Приветствиемир", "discord", 15)
	assert.True(t, utf8.ValidString(got))
	assert.Regexp(t, `/discord$`, got)

	// suffix longer than the nick budget: degrade to just the suffix
	// rather than producing an empty or negative-length base
	got = relayMsgNick("bob", "discord", 5)
	assert.Regexp(t, `/discord$`, got)
}
