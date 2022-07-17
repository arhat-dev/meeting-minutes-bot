package telegram

import (
	"context"
	"fmt"
	"sync"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"go.uber.org/multierr"

	"arhat.dev/mbot/pkg/rt"
)

var _ rt.Conversation = (*conversationImpl)(nil)

type conversationImpl struct {
	bot *tgBot

	peer tg.InputPeerClass
}

type uploadResult struct {
	uploadedFile message.MultiMediaOption
	err          error
}

// Context implements rt.Conversation
func (c *conversationImpl) Context() context.Context {
	return c.bot.Context()
}

// SendMessage implements rt.Conversation
func (c *conversationImpl) SendMessage(ctx context.Context, opts rt.SendMessageOptions) (msgIDs []rt.MessageID, err error) {
	var (
		text  []styling.StyledTextOption
		media []message.MultiMediaOption
		resp  tg.UpdatesClass

		resultCh chan uploadResult
		wg       sync.WaitGroup

		cancel context.CancelFunc
	)

	sz := len(opts.MessageBody)
	builder := c.bot.sender.To(c.peer).Reply(int(opts.ReplyTo))

	if len(opts.Callbacks) != 0 {
		var markup tg.ReplyInlineMarkup

		markup.Rows = make([]tg.KeyboardButtonRow, len(opts.Callbacks))
		for i, cbs := range opts.Callbacks {
			markup.Rows[i].Buttons = make([]tg.KeyboardButtonClass, len(cbs))
			for j, cb := range cbs {
				switch {
				case !cb.OnClick.IsNil():
					// TODO: support callback
					markup.Rows[i].Buttons[j] = &tg.KeyboardButtonCallback{
						Text: cb.Text,
						Data: []byte{},
					}
				case !cb.URL.IsNil():
					markup.Rows[i].Buttons[j] = &tg.KeyboardButtonURL{
						Text: cb.Text,
						URL:  cb.URL.Get(),
					}
				default:
					markup.Rows[i].Buttons[j] = &tg.KeyboardButton{
						Text: cb.Text,
					}
				}
			}
		}

		builder = builder.Markup(&markup)
	}

	for i := 0; i < sz; i++ {
		sp := &opts.MessageBody[i]

		if !sp.IsMedia() {
			text = append(text, getStyleOptionForSpan(sp))
			continue
		}

		if sp.Size == 0 {
			continue
		}

		if resultCh == nil {
			resultCh = make(chan uploadResult, 1)
			ctx, cancel = context.WithCancel(c.bot.Context())
			defer cancel()

			wg.Add(1)
			go func() {
				wg.Wait()
				close(resultCh)
			}()
		} else {
			wg.Add(1)
		}

		go c.upload(ctx, &wg, sp, resultCh)
	}

	if resultCh != nil {
		for result := range resultCh {
			if result.err != nil {
				err = multierr.Append(err, result.err)
				// cancel to stop all other uploading tasks
				cancel()
				continue
			}

			if err == nil {
				media = append(media, result.uploadedFile)
			}
		}

		if err != nil {
			return
		}
	}

	// if there is media span, send media first
	switch len(media) {
	case 0: // no media spans
		resp, err = builder.StyledText(ctx, text...)
		text = nil
	case 1: // only one media span
		resp, err = builder.Media(ctx, media[0])
	default: // multiple media spans
		resp, err = builder.Album(ctx, media[0], media[1:]...)
	}
	if err != nil {
		return
	}

	_ = handleMessageSent(resp, &msgIDs)

	// send text
	if len(text) != 0 {
		var resp2 tg.UpdatesClass
		resp2, err = builder.StyledText(ctx, text...)
		if err == nil {
			_ = handleMessageSent(resp2, &msgIDs)
		}
	}

	return
}

func (c *conversationImpl) upload(ctx context.Context, wg *sync.WaitGroup, sp *rt.Span, resultCh chan<- uploadResult) {
	defer wg.Done()

	var (
		result       uploadResult
		uploadedFile tg.InputFileClass
	)

	uploadedFile, result.err = c.bot.uploader.Upload(ctx, uploader.NewUpload(sp.Filename, sp.Data, sp.Size))
	if result.err != nil {
		select {
		case <-ctx.Done():
		case resultCh <- result:
		}

		return
	}

	caption := translateTextSpans(sp.Caption)
	flag := sp.Flags
	switch {
	case flag.IsAudio():
		result.uploadedFile = message.Audio(uploadedFile, caption...).
			Title("").     // TODO
			Performer(""). // TODO
			Duration(sp.Duration)
	case flag.IsVoice():
		result.uploadedFile = message.Audio(uploadedFile, caption...).
			Duration(sp.Duration).
			Voice()
	case flag.IsVideo():
		result.uploadedFile = message.Video(uploadedFile, caption...).
			Duration(sp.Duration).
			SupportsStreaming()
	case flag.IsImage():
		// TODO
		// media = append(media,
		// 	message.Photo(uploadedFile, caption...),
		// )
	default:
		result.uploadedFile = message.File(uploadedFile, caption...).
			Filename(sp.Filename).
			MIME(sp.ContentType)
	}

	select {
	case <-ctx.Done():
	case resultCh <- result:
	}
}

// translateTextSpans converts text spans to telegram styled text options skipping all media spans
func translateTextSpans(spans []rt.Span) (text []styling.StyledTextOption) {
	sz := len(spans)
	for i := 0; i < sz; i++ {
		if spans[i].IsMedia() {
			continue
		}

		text = append(text, getStyleOptionForSpan(&spans[i]))
	}

	return
}

func getStyleOptionForSpan(ent *rt.Span) styling.StyledTextOption {
	var styles []entity.Formatter

	if ent.IsURL() {
		if len(ent.URL) != 0 {
			styles = append(styles, entity.TextURL(ent.URL))
		} else {
			styles = append(styles, entity.URL())
		}
	}

	if ent.IsEmail() {
		styles = append(styles, entity.Email())
	}

	if ent.IsPhoneNumber() {
		styles = append(styles, entity.Phone())
	}

	if ent.IsMention() {
		if len(ent.Hint) != 0 {
			// TODO:
			styles = append(styles, entity.MentionName(&tg.InputUser{}))
		} else {
			styles = append(styles, entity.Mention())
		}
	}

	if ent.IsBlockquote() {
		styles = append(styles, entity.Blockquote())
	}

	if ent.IsBold() {
		styles = append(styles, entity.Bold())
	}

	if ent.IsItalic() {
		styles = append(styles, entity.Italic())
	}

	if ent.IsStrikethrough() {
		styles = append(styles, entity.Strike())
	}

	if ent.IsUnderline() {
		styles = append(styles, entity.Underline())
	}

	if ent.IsPre() {
		styles = append(styles, entity.Pre(ent.Hint))
	}

	if ent.IsCode() {
		styles = append(styles, entity.Code())
	}

	if len(styles) == 0 {
		return styling.Plain(ent.Text)
	}

	text := ent.Text
	return styling.Custom(func(b *entity.Builder) error {
		b.Format(text, styles...)
		return nil
	})
}

// handleMessageSent return first message id of the sent message
func handleMessageSent(updCls tg.UpdatesClass, msgIDs *[]rt.MessageID) (err error) {
	switch resp := updCls.(type) {
	case *tg.UpdatesTooLong:
		err = fmt.Errorf("too many updates")
	case *tg.UpdateShortMessage:
		*msgIDs = append(*msgIDs, rt.MessageID(resp.GetID()))
	case *tg.UpdateShortChatMessage:
		*msgIDs = append(*msgIDs, rt.MessageID(resp.GetID()))
	case *tg.UpdateShort:
		id, _ := extractMsgID(resp.GetUpdate())
		*msgIDs = append(*msgIDs, id)
	case *tg.UpdatesCombined:
		upds := resp.GetUpdates()
		for i := range upds {
			var (
				id rt.MessageID
				ok bool
			)
			id, ok = extractMsgID(upds[i])
			if ok {
				*msgIDs = append(*msgIDs, id)
			}
		}
	case *tg.Updates:
		upds := resp.GetUpdates()
		for i := range upds {
			var (
				id rt.MessageID
				ok bool
			)
			id, ok = extractMsgID(upds[i])
			if ok {
				*msgIDs = append(*msgIDs, id)
			}
		}
	case *tg.UpdateShortSentMessage:
		*msgIDs = append(*msgIDs, rt.MessageID(resp.GetID()))
	default:
		err = fmt.Errorf("unexpected response type %T", resp)
		return
	}

	return
}

func extractMsgID(upd tg.UpdateClass) (msgID rt.MessageID, hasMsgID bool) {
	switch u := upd.(type) {
	case *tg.UpdateNewMessage:
		return rt.MessageID(u.GetMessage().GetID()), true
	case *tg.UpdateMessageID:
		return rt.MessageID(u.GetID()), true
	default:
		return
	}
}
