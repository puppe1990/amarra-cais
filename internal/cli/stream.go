package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

type streamOpts struct {
	dryRun bool
	live   bool
}

func chatPageTpl(opts streamOpts) string {
	if opts.live {
		return tplChatPageLive
	}
	return tplChatPage
}

func scaffoldStreamChat(dir string, opts streamOpts) error {
	data := dataForHandler("chat")
	data.ModulePath = moduleFromDir(dir)

	migRel, _, err := nextMigrationFile(dir, "chat", opts.dryRun)
	if err != nil {
		return err
	}

	files := map[string]string{
		filepath.Join("internal/models", "conversation.go"):          tplConversationModel,
		filepath.Join("internal/models", "message.go"):               tplMessageModel,
		filepath.Join("internal/handlers", "chat.go"):                tplChatHandler,
		filepath.Join("internal/handlers", "chat_test.go"):           tplChatHandlerTest,
		filepath.Join("web/templates/pages", "conversations.html"):   tplConversationsPage,
		filepath.Join("web/templates/pages", "chat.html"):            chatPageTpl(opts),
		filepath.Join("web/templates/partials", "message.html"):      tplMessagePartial,
		filepath.Join("web/templates/partials", "chat_history.html"): tplChatHistoryPartial,
		migRel: tplChatMigration,
	}

	for path, tpl := range files {
		full := filepath.Join(dir, path)
		if _, err := os.Stat(full); err == nil {
			return fmt.Errorf("%s already exists", path)
		}
		if err := writeScaffoldTemplate(full, tpl, data, path, opts.dryRun); err != nil {
			return err
		}
	}

	if err := patchStoreForStreamChat(dir, data, opts.dryRun); err != nil {
		return err
	}
	if opts.live {
		livePath := filepath.Join(dir, "internal/handlers", "chat_live.go")
		if err := writeScaffoldTemplate(livePath, tplChatLive, data, "internal/handlers/chat_live.go", opts.dryRun); err != nil {
			return err
		}
		if err := patchLiveViewsRegister(dir, scaffoldData{Snake: "chat", Pascal: "Chat", ModulePath: data.ModulePath}, opts.dryRun); err != nil {
			return err
		}
	}
	if err := patchRoutesForStreamChat(dir, opts.dryRun); err != nil {
		return err
	}
	return patchLayoutNavForStreamChat(dir, opts.dryRun)
}
