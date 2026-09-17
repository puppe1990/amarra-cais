package cli

const tplChatPageLive = `{{"{{"}} define "content" {{"}}"}}
<section amarra-live="chat" data-amarra-topic="{{"{{"}} printf "chat:%d" .Conversation.ID {{"}}"}}" class="cais-chat-shell max-w-2xl mx-auto px-4 flex flex-col min-h-0 h-[calc(100dvh-8rem)]">
  <div class="flex items-center gap-3 py-3 shrink-0 border-b border-foam/10">
    {{"{{"}} linkTo "/chat" "←" {{"}}"}}
    <h1 class="font-serif text-lg text-foam truncate">{{"{{"}} if .Conversation.Title {{"}}"}}{{"{{"}} .Conversation.Title {{"}}"}}{{"{{"}} else {{"}}"}}Chat{{"{{"}} end {{"}}"}}</h1>
  </div>
  <div id="chat-history" class="flex-1 overflow-y-auto py-3 min-h-0 flex flex-col gap-3">
    {{"{{"}}- range .Messages {{"}}"}}
    <div class="rounded-xl px-4 py-2 max-w-[85%] {{"{{"}}if eq .Role "assistant"{{"}}"}}border border-foam/20 text-foam self-start{{"{{"}}else{{"}}"}}bg-copper text-ink self-end{{"{{"}}end{{"}}"}}">
      <p class="text-sm whitespace-pre-wrap">{{"{{"}} .Content {{"}}"}}</p>
    </div>
    {{"{{"}}- end {{"}}"}}
  </div>
  <form amarra-submit="send" class="flex gap-2 py-3">
    <textarea name="content" rows="2" class="flex-1 border border-copper/40 bg-ink text-foam px-3 py-2 text-sm" placeholder="Message…" required></textarea>
    <button type="submit" class="bg-copper px-5 py-2 font-mono text-[11px] uppercase tracking-[0.22em] text-ink">Send</button>
  </form>
</section>
{{"{{"}} end {{"}}"}}
`

const tplChatLive = `package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/puppe1990/amarra-cais/pkg/amarra/live"

	"{{.ModulePath}}/internal/models"
	"{{.ModulePath}}/internal/store"
)

type ChatLive struct {
	store  store.Store
	convID int64
	html   string
}

func NewChatLive(s store.Store) *ChatLive {
	return &ChatLive{store: s}
}

func (v *ChatLive) Mount(_ context.Context, sock live.Socket) error {
	topic := sock.Topic()
	idStr := strings.TrimPrefix(topic, "chat:")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("chat topic: %w", err)
	}
	v.convID = id
	return v.reload()
}

func (v *ChatLive) Handle(_ context.Context, ev live.Event) error {
	if ev.Name != "send" {
		return fmt.Errorf("unknown event %s", ev.Name)
	}
	var payload struct {
		Content string ` + "`json:\"content\"`" + `
	}
	_ = json.Unmarshal(ev.Payload, &payload)
	content := strings.TrimSpace(payload.Content)
	if content == "" {
		return nil
	}
	if _, err := v.store.InsertMessage(models.Message{ConversationID: v.convID, Role: "user", Content: content}); err != nil {
		return err
	}
	if _, err := v.store.InsertMessage(models.Message{ConversationID: v.convID, Role: "assistant", Content: "ok: " + content}); err != nil {
		return err
	}
	return v.reload()
}

func (v *ChatLive) Render() live.Rendered {
	return live.Rendered{Target: "chat-history", HTML: v.html}
}

func (v *ChatLive) reload() error {
	msgs, err := v.store.ListMessages(v.convID)
	if err != nil {
		return err
	}
	var b strings.Builder
	for _, m := range msgs {
		if m.Role == "assistant" {
			b.WriteString(` + "`" + `<div class="rounded-xl px-4 py-2 max-w-[85%] border border-foam/20 text-foam self-start"><p class="text-sm whitespace-pre-wrap">` + "`" + `)
		} else {
			b.WriteString(` + "`" + `<div class="rounded-xl px-4 py-2 max-w-[85%] bg-copper text-ink self-end"><p class="text-sm whitespace-pre-wrap">` + "`" + `)
		}
		b.WriteString(html.EscapeString(m.Content))
		b.WriteString("</p></div>")
	}
	v.html = b.String()
	return nil
}
`
