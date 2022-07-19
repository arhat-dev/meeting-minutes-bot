package main

// bot platforms
import (
	_ "arhat.dev/mbot/pkg/bot/discord"
	_ "arhat.dev/mbot/pkg/bot/github"
	_ "arhat.dev/mbot/pkg/bot/gitlab"
	_ "arhat.dev/mbot/pkg/bot/gitter"
	_ "arhat.dev/mbot/pkg/bot/irc"
	_ "arhat.dev/mbot/pkg/bot/line"
	_ "arhat.dev/mbot/pkg/bot/matrix"
	_ "arhat.dev/mbot/pkg/bot/mattermost"
	_ "arhat.dev/mbot/pkg/bot/reddit"
	_ "arhat.dev/mbot/pkg/bot/slack"
	_ "arhat.dev/mbot/pkg/bot/telegram"
	_ "arhat.dev/mbot/pkg/bot/vk"
	_ "arhat.dev/mbot/pkg/bot/whatsapp"
)
