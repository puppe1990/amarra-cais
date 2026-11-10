package cli

import (
	"os"
	"path/filepath"
	"strings"
)

func destroyStreamChat(dir string, dryRun, force bool) error {
	data := dataForHandler("chat")
	files := []string{
		filepath.Join("internal/models", "conversation.go"),
		filepath.Join("internal/models", "message.go"),
		filepath.Join("internal/handlers", "chat.go"),
		filepath.Join("internal/handlers", "chat_test.go"),
		filepath.Join("internal/handlers", "chat_live.go"),
		filepath.Join("web/templates/pages", "conversations.html"),
		filepath.Join("web/templates/pages", "chat.html"),
		filepath.Join("web/templates/partials", "message.html"),
		filepath.Join("web/templates/partials", "chat_history.html"),
	}
	migrationsDir := filepath.Join(dir, "internal/store/migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.Contains(e.Name(), "_chat.sql") {
			files = append(files, filepath.Join("internal/store/migrations", e.Name()))
		}
	}
	if err := removeGeneratedFiles(dir, files, dryRun, force); err != nil {
		return err
	}
	if err := unpatchRoutesForHandler(dir, data, dryRun); err != nil {
		return err
	}
	if err := unpatchLive(dir, data, dryRun); err != nil {
		return err
	}
	if err := unpatchStoreForStreamChat(dir, dryRun); err != nil {
		return err
	}
	if err := unpatchLayoutNavForStreamChat(dir, dryRun); err != nil {
		return err
	}
	return ensureMigrationPlaceholder(dir, dryRun)
}

func unpatchStoreForStreamChat(dir string, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	names := nameSet(
		"ListConversations",
		"FindConversationByID",
		"InsertConversation",
		"UpdateConversationTitle",
		"ListMessages",
		"ListMessagesSince",
		"InsertMessage",
	)
	content, err := removeStoreMethodsNamed(string(body), names)
	if err != nil {
		return err
	}
	content = dropUnusedImport(content, "time")
	return updateScaffoldFile(path, []byte(content), "internal/store/store.go", dryRun)
}

func unpatchLayoutNavForStreamChat(dir string, dryRun bool) error {
	path, _ := layoutNavFile(dir)
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	link := `          <a href="/chat" class="px-3 py-2 font-mono text-[10px] uppercase tracking-[0.22em] text-foam/50 hover:text-foam flex-shrink-0">Chat</a>
`
	content := strings.Replace(string(body), link, "", 1)
	rel := strings.TrimPrefix(path, dir+string(os.PathSeparator))
	return updateScaffoldFile(path, []byte(content), rel, dryRun)
}
