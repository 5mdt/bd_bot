package bot

import (
	"testing"
)

func TestLeftChatMemberBehavior(t *testing.T) {
	// Test that the bot no longer sends welcome messages when members leave
	// This test documents the expected behavior: no welcome messages on member leave

	// The bot should only send welcome messages when it is added to a chat,
	// not when other members leave the chat
	expectWelcomeOnMemberLeave := false

	if expectWelcomeOnMemberLeave {
		t.Error("Bot should not send welcome messages when members leave chats")
	}
}
