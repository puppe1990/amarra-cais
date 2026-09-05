package cli

const tplConversationModel = `package models

import "time"

type Conversation struct {
	ID        int64
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
`

const tplMessageModel = `package models

import "time"

type Message struct {
	ID             int64
	ConversationID int64
	Role           string
	Content        string
	CreatedAt      time.Time
}
`

const tplChatHandler = `package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/puppe1990/amarra-cais/pkg/amarra/stream"
	"github.com/puppe1990/amarra-cais/pkg/amarra/view"
	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/chat"
	"github.com/puppe1990/amarra-cais/pkg/cais/httpx"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/meta"
	"github.com/puppe1990/amarra-cais/pkg/cais/validate"

	"{{.ModulePath}}/internal/models"
	"{{.ModulePath}}/internal/store"
)

type ChatHandler struct {
	views   *view.Renderer
	store   store.Store
	site    meta.Site
	catalog *i18n.Catalog
	cfg     cais.Config
}

func NewChatHandler(views *view.Renderer, s store.Store, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *ChatHandler {
	return &ChatHandler{views: views, store: s, site: site, catalog: catalog, cfg: cfg}
}

type conversationsPageData struct {
	meta.Site
	Conversations []models.Conversation
}

type chatPageData struct {
	meta.Site
	Conversation models.Conversation
	Messages     []models.Message
	StreamURL    string
	MessagesURL  string
}

func messageRole(role string) chat.Role {
	if role == "user" {
		return chat.RoleUser
	}
	return chat.RoleAssistant
}

func streamAssistantPreview(w http.ResponseWriter, text string) {
	if text == "" {
		return
	}
	step := len(text) / 6
	if step < 1 {
		step = 1
	}
	for end := step; end < len(text)+step; end += step {
		if end > len(text) {
			end = len(text)
		}
		_ = stream.WriteOp(w, stream.Op{Kind: "morph", Target: "chat-live", HTML: chat.LiveBubble(text[:end])})
		if end == len(text) {
			break
		}
	}
}

func (h *ChatHandler) List(w http.ResponseWriter, r *http.Request) {
	convs, err := h.store.ListConversations()
	if err != nil {
		http.Error(w, "could not load conversations", http.StatusInternalServerError)
		return
	}
	view.Write(w, r, h.views, view.Page{
		Layout: "app",
		Name:   "conversations",
		Data: conversationsPageData{
			Site:          meta.ForRequest(h.site, r),
			Conversations: convs,
		},
	}, h.cfg)
}

func (h *ChatHandler) Create(w http.ResponseWriter, r *http.Request) {
	id, err := h.store.InsertConversation("New chat")
	if err != nil {
		http.Error(w, "could not create conversation", http.StatusInternalServerError)
		return
	}
	httpx.SeeOther(w, r, fmt.Sprintf("/chat/%d", id))
}

func (h *ChatHandler) Show(w http.ResponseWriter, r *http.Request, id int64) {
	conv, err := h.store.FindConversationByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	msgs, err := h.store.ListMessages(id)
	if err != nil {
		http.Error(w, "could not load messages", http.StatusInternalServerError)
		return
	}
	msgs = chat.TrimForDisplay(msgs, 80)
	view.Write(w, r, h.views, view.Page{
		Layout: "app",
		Name:   "chat",
		Data: chatPageData{
			Site:         meta.ForRequest(h.site, r),
			Conversation: conv,
			Messages:     msgs,
			StreamURL:    fmt.Sprintf("/chat/%d/stream", id),
			MessagesURL:  fmt.Sprintf("/chat/%d/messages", id),
		},
	}, h.cfg)
}

func (h *ChatHandler) Stream(w http.ResponseWriter, r *http.Request, id int64) {
	stream.RelaySSE(w)
	lastID := int64(0)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			msgs, err := h.store.ListMessagesSince(id, lastID)
			if err != nil {
				continue
			}
			for _, m := range msgs {
				if m.Role == "assistant" {
					streamAssistantPreview(w, m.Content)
				}
				bubble := chat.MessageBubble
				if m.Role == "assistant" {
					bubble = chat.SafeMessageBubble
				}
				_ = stream.WriteOp(w, stream.Op{
					Kind:   "append",
					Target: "chat-history",
					HTML:   bubble(messageRole(m.Role), m.Content, m.CreatedAt),
				})
				lastID = m.ID
			}
		}
	}
}

func (h *ChatHandler) PostMessage(w http.ResponseWriter, r *http.Request, id int64) {
	content := r.FormValue("content")
	var errs validate.FieldErrors
	if content == "" {
		errs.Add("content", "Message is required")
	}
	if errs.Any() {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if _, err := h.store.FindConversationByID(id); err != nil {
		http.NotFound(w, r)
		return
	}
	if _, err := h.store.InsertMessage(models.Message{ConversationID: id, Role: "user", Content: content}); err != nil {
		http.Error(w, "could not save message", http.StatusInternalServerError)
		return
	}
	_ = h.store.UpdateConversationTitle(id, content)
	if _, err := h.store.InsertMessage(models.Message{ConversationID: id, Role: "assistant", Content: "Echo: " + content}); err != nil {
		http.Error(w, "could not save reply", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(chat.MessageBubble(chat.RoleUser, content, time.Now().UTC())))
}

func (h *ChatHandler) ListMessages(w http.ResponseWriter, r *http.Request, id int64) {
	msgs, err := h.store.ListMessages(id)
	if err != nil {
		http.Error(w, "could not load messages", http.StatusInternalServerError)
		return
	}
	msgs = chat.TrimForDisplay(msgs, 80)
	var buf strings.Builder
	for _, m := range msgs {
		if m.Role == "assistant" {
			buf.WriteString(chat.SafeMessageBubble(messageRole(m.Role), m.Content, m.CreatedAt))
		} else {
			buf.WriteString(chat.MessageBubble(messageRole(m.Role), m.Content, m.CreatedAt))
		}
	}
	_, _ = w.Write([]byte(buf.String()))
}
`

const tplChatHandlerTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/testutil"

	appi18n "{{.ModulePath}}/internal/i18n"
)

func TestChatHandler_List_Returns200(t *testing.T) {
	h := NewChatHandler(setupTestViews(t), setupTestStore(t), testSite(), appi18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/chat", nil)
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestChatHandler_Show_Returns200(t *testing.T) {
	s := setupTestStore(t)
	id, err := s.InsertConversation("Test")
	if err != nil {
		t.Fatal(err)
	}
	h := NewChatHandler(setupTestViews(t), s, testSite(), appi18n.DefaultCatalog(), cais.Config{})

	req := testutil.NewRequest(http.MethodGet, "/chat/1", testutil.PathValue("id", "1"))
	rr := httptest.NewRecorder()
	h.Show(rr, req, id)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "id=\"chat-history\"") {
		t.Error("missing chat-history")
	}
	if !strings.Contains(body, "data-amarra-stream") {
		t.Error("missing data-amarra-stream")
	}
}

func TestChatHandler_Show_NotFound_Returns404(t *testing.T) {
	h := NewChatHandler(setupTestViews(t), setupTestStore(t), testSite(), appi18n.DefaultCatalog(), cais.Config{})

	req := testutil.NewRequest(http.MethodGet, "/chat/999", testutil.PathValue("id", "999"))
	rr := httptest.NewRecorder()
	h.Show(rr, req, 999)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestChatHandler_PostMessage_ReturnsUserBubble(t *testing.T) {
	s := setupTestStore(t)
	id, err := s.InsertConversation("Test")
	if err != nil {
		t.Fatal(err)
	}
	h := NewChatHandler(setupTestViews(t), s, testSite(), appi18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodPost, "/chat/1/messages", strings.NewReader("content=hello"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.PostMessage(rr, req, id)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	testutil.AssertHTMLContains(t, rr.Body.String(), "hello", "cais-msg-user")
}
`

const tplConversationsPage = `{{"{{"}} define "content" {{"}}"}}
<section class="max-w-2xl mx-auto px-4 py-8">
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold text-slate-900">Conversations</h1>
    <.form action="/chat" method="post">
      {{"{{"}} csrfField .CSRFToken {{"}}"}}
      <.button type="submit">New chat</.button>
    </.form>
  </div>
  <ul class="space-y-2">
    {{"{{"}} range .Conversations {{"}}"}}
    <li>
      {{"{{"}} if .Title {{"}}"}}
      {{"{{"}} linkTo (printf "/chat/%d" .ID) .Title {{"}}"}}
      {{"{{"}} else {{"}}"}}
      {{"{{"}} linkTo (printf "/chat/%d" .ID) "Untitled" {{"}}"}}
      {{"{{"}} end {{"}}"}}
    </li>
    {{"{{"}} else {{"}}"}}
    <li class="text-sm text-slate-500">No conversations yet.</li>
    {{"{{"}} end {{"}}"}}
  </ul>
</section>
{{"{{"}} end {{"}}"}}
`

const tplChatPage = `{{"{{"}} define "content" {{"}}"}}
<section class="cais-chat-shell max-w-2xl mx-auto px-4 flex flex-col min-h-0 h-[calc(100dvh-8rem)] sm:h-[calc(100dvh-6rem)]">
  <div class="flex items-center gap-3 py-3 shrink-0 border-b border-slate-200">
    {{"{{"}} linkTo "/chat" "←" {{"}}"}}
    <h1 class="text-lg font-bold text-slate-900 truncate">{{"{{"}} if .Conversation.Title {{"}}"}}{{"{{"}} .Conversation.Title {{"}}"}}{{"{{"}} else {{"}}"}}Chat{{"{{"}} end {{"}}"}}</h1>
  </div>
  <div id="chat-messages" class="flex-1 overflow-y-auto py-3 min-h-0" data-amarra-stream="{{"{{"}} .StreamURL {{"}}"}}" data-amarra-target="chat-history">
    <div id="chat-history" class="flex flex-col gap-3">
      {{"{{"}}- range .Messages {{"}}"}}
      <div class="rounded-xl px-4 py-2 max-w-[85%] {{"{{"}}if eq .Role "assistant"{{"}}"}}bg-white border border-slate-200 self-start{{"{{"}}else{{"}}"}}bg-indigo-600 text-white self-end{{"{{"}}end{{"}}"}}">
        <p class="text-sm whitespace-pre-wrap">{{"{{"}} .Content {{"}}"}}</p>
      </div>
      {{"{{"}}- end {{"}}"}}
    </div>
    <div id="chat-live" class="flex flex-col gap-3"></div>
  </div>
  <.form action="{{"{{"}} printf "/chat/%d/messages" .Conversation.ID {{"}}"}}" method="post">
    {{"{{"}} csrfField .CSRFToken {{"}}"}}
    <textarea name="content" rows="2" class="flex-1 rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Message…" required></textarea>
    <.button type="submit">Send</.button>
  </.form>
</section>
{{"{{"}} end {{"}}"}}
`

const tplMessagePartial = `{{"{{"}} define "message" {{"}}"}}
<div class="rounded-xl px-4 py-2 max-w-[85%] {{"{{"}}if eq .Role "assistant"{{"}}"}}bg-white border border-slate-200 self-start{{"{{"}}else{{"}}"}}bg-indigo-600 text-white self-end{{"{{"}}end{{"}}"}}">
  <p class="text-sm whitespace-pre-wrap">{{"{{"}} .Content {{"}}"}}</p>
</div>
{{"{{"}} end {{"}}"}}
`

const tplChatHistoryPartial = `{{"{{"}} define "chat_history" {{"}}"}}
{{"{{"}} range .Messages {{"}}"}}
<div class="rounded-xl px-4 py-2 max-w-[85%] {{"{{"}}if eq .Role "assistant"{{"}}"}}bg-white border border-slate-200 self-start{{"{{"}}else{{"}}"}}bg-indigo-600 text-white self-end{{"{{"}}end{{"}}"}}">
  <p class="text-sm whitespace-pre-wrap">{{"{{"}} .Content {{"}}"}}</p>
</div>
{{"{{"}} end {{"}}"}}
{{"{{"}} end {{"}}"}}
`

const tplChatMigration = `-- up
CREATE TABLE conversations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE messages (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  conversation_id INTEGER NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
  content TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);

-- down
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversations;
`
