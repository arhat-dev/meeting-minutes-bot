package generator

import (
	"arhat.dev/meeting-minutes-bot/pkg/message"
)

func CreateFuncMap() FuncMap {
	return map[string]interface{}{
		"findMessage": func(messages []message.Interface, id string) message.Interface {
			// TODO: use binary search?
			for _, m := range messages {
				if m.ID() == id {
					return m
				}
			}

			return nil
		},
	}
}
