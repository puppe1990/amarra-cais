package cli

const tplConversationsPage = `{{"{{"}} define "content" {{"}}"}}
<section class="max-w-2xl mx-auto px-4 py-8">
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold text-slate-900">Conversations</h1>
    <.form action="/chat" method="post">
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
      <div class="rounded-xl px-4 py-2 max-w-[85%] {{"{{"}}if eq .Role "assistant"{{"}}"}}bg-white border border-slate-200 self-start{{"{{"}}else{{"}}"}}bg-copper text-ink self-end{{"{{"}}end{{"}}"}}">
        <p class="text-sm whitespace-pre-wrap">{{"{{"}} .Content {{"}}"}}</p>
      </div>
      {{"{{"}}- end {{"}}"}}
    </div>
    <div id="chat-live" class="flex flex-col gap-3"></div>
  </div>
  <.form action="{{"{{"}} printf "/chat/%d/messages" .Conversation.ID {{"}}"}}" method="post">
    <textarea name="content" rows="2" class="flex-1 rounded-xl border border-slate-200 px-3 py-2 text-sm" placeholder="Message…" required></textarea>
    <.button type="submit">Send</.button>
  </.form>
</section>
{{"{{"}} end {{"}}"}}
`

const tplMessagePartial = `{{"{{"}} define "message" {{"}}"}}
<div class="rounded-xl px-4 py-2 max-w-[85%] {{"{{"}}if eq .Role "assistant"{{"}}"}}bg-white border border-slate-200 self-start{{"{{"}}else{{"}}"}}bg-copper text-ink self-end{{"{{"}}end{{"}}"}}">
  <p class="text-sm whitespace-pre-wrap">{{"{{"}} .Content {{"}}"}}</p>
</div>
{{"{{"}} end {{"}}"}}
`

const tplChatHistoryPartial = `{{"{{"}} define "chat_history" {{"}}"}}
{{"{{"}} range .Messages {{"}}"}}
<div class="rounded-xl px-4 py-2 max-w-[85%] {{"{{"}}if eq .Role "assistant"{{"}}"}}bg-white border border-slate-200 self-start{{"{{"}}else{{"}}"}}bg-copper text-ink self-end{{"{{"}}end{{"}}"}}">
  <p class="text-sm whitespace-pre-wrap">{{"{{"}} .Content {{"}}"}}</p>
</div>
{{"{{"}} end {{"}}"}}
{{"{{"}} end {{"}}"}}
`
