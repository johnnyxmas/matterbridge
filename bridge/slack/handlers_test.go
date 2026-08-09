package bslack

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/42wim/matterbridge/bridge"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A regular top-level message as delivered by the Events API / Socket
// Mode: no "thread_ts" key at all (that's only present on replies), no
// "message" key (that's only present on message_changed/message_deleted).
const regularTopLevelMessageJSON = `{
	"type": "message",
	"user": "U123HUMAN",
	"text": "hello",
	"ts": "1786248460.671289",
	"channel": "C0APH6R72HE",
	"channel_type": "channel",
	"event_ts": "1786248460.671289"
}`

func TestMessageEventUnmarshalSynthesizesThreadTimestamp(t *testing.T) {
	var ev slackevents.MessageEvent
	require.NoError(t, json.Unmarshal([]byte(regularTopLevelMessageJSON), &ev))

	require.NotNil(t, ev.Message, "Message should be synthesized for a regular top-level message")
	assert.Equal(t, ev.Message.Timestamp, ev.Message.ThreadTimestamp,
		"a top-level message is the root of its own thread; ThreadTimestamp must match "+
			"Timestamp or skipMessageEvent's unfurl check (see #266) treats every regular "+
			"message as an unfurl and drops it")
}

func TestSkipMessageEventDoesNotDropRegularMessages(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(ioutil.Discard)
	cfg := &bridge.Config{Bridge: &bridge.Bridge{Log: logrus.NewEntry(logger)}}
	b := newBridge(cfg)
	b.si = &slack.Bot{ID: "BBOTID", UserID: "UBOTUSER"}

	var ev slackevents.MessageEvent
	require.NoError(t, json.Unmarshal([]byte(regularTopLevelMessageJSON), &ev))

	assert.False(t, b.skipMessageEvent(&ev),
		"a plain message from a real human user must not be skipped")
}
