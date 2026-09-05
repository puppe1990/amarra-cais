package cli

// Simple chat history + Stream SSE target (data-amarra-stream, no hx-ext).
const tplPartialChatSSE = `{{"{{"}}- define "chat_sse" -{{"}}"}}
<div
  id="chat-messages"
  class="flex flex-col gap-3 min-h-[12rem]"
  data-amarra-stream="{{"{{"}} .StreamURL {{"}}"}}"
  data-amarra-target="chat-history"
>
  <div id="chat-history" class="flex flex-col gap-3 min-h-[12rem]">
    {{"{{"}}- range .Messages {{"}}"}}
    <div class="rounded-xl px-4 py-2 max-w-[85%] {{"{{"}}if eq .Role "assistant"{{"}}"}}bg-white border border-slate-200 self-start{{"{{"}}else{{"}}"}}bg-indigo-600 text-white self-end{{"{{"}}end{{"}}"}}">
      <p class="text-sm whitespace-pre-wrap">{{"{{"}} .Content {{"}}"}}</p>
    </div>
    {{"{{"}}- end {{"}}"}}
  </div>
  <div id="chat-live" class="flex flex-col gap-3"></div>
</div>
<div
  id="chat-thinking"
  class="hidden rounded-xl px-4 py-2 max-w-[85%] bg-slate-100 border border-slate-200 self-start text-sm text-slate-500"
  role="status"
  aria-live="polite"
>
  Agente pensando…
</div>
{{"{{"}}- end -{{"}}"}}`

// Agent-mode chat slots for token-level Stream SSE.
const tplPartialChatSSEAgent = `{{"{{"}}- define "chat_sse_agent" -{{"}}"}}
<div class="cais-chat-messages-wrap relative flex-1 min-h-0 flex flex-col">
  <div id="chat-messages" class="cais-chat-messages flex-1 overflow-y-auto overflow-x-hidden max-w-full overscroll-contain py-3 min-h-0" data-amarra-stream="{{"{{"}} .StreamURL {{"}}"}}" data-amarra-target="chat-history">
    <div class="flex flex-col gap-3 min-h-full justify-end">
      <div id="chat-history" class="flex flex-col gap-3">
        {{"{{"}}- range .Messages {{"}}"}}
        <div class="rounded-xl px-4 py-2 max-w-[85%] {{"{{"}}if eq .Role "assistant"{{"}}"}}bg-white border border-slate-200 self-start{{"{{"}}else{{"}}"}}bg-indigo-600 text-white self-end{{"{{"}}end{{"}}"}}">
          <p class="text-sm whitespace-pre-wrap">{{"{{"}} .Content {{"}}"}}</p>
        </div>
        {{"{{"}}- end {{"}}"}}
      </div>
      <div id="chat-stream" class="flex flex-col gap-3"></div>
      <div id="chat-live" class="flex flex-col gap-3"></div>
      <div
        id="chat-thinking"
        class="hidden cais-thinking flex items-center gap-2.5 max-w-[85%] rounded-2xl rounded-bl-sm bg-slate-100 border border-slate-200 px-4 py-3 text-sm text-slate-600 shadow-xs self-start"
        role="status"
        aria-live="polite"
        aria-hidden="true"
      >
        <span class="cais-thinking-dots shrink-0" aria-hidden="true"><span></span><span></span><span></span></span>
        <span id="chat-thinking-label">Thinking...</span>
      </div>
    </div>
  </div>
  <button type="button" id="chat-scroll-down" class="cais-chat-scroll-down hidden" aria-hidden="true" aria-label="Jump to latest messages">
    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M19 9l-7 7-7-7" />
    </svg>
  </button>
</div>
{{"{{"}}- end -{{"}}"}}`
