package cli

import (
	"reflect"
	"testing"
)

// TestScaffoldTemplates_allRequiredPresent guards against accidental removal
// when editing tpl_scaffold_*.go files.
func TestScaffoldTemplates_allRequiredPresent(t *testing.T) {
	v := reflect.ValueOf(struct {
		TplGoMod                  string
		TplEmptyCSS               string
		TplAir                    string
		TplEnvExample             string
		TplWebEmbed               string
		TplMain                   string
		TplMainBlank              string
		TplConsole                string
		TplApp                    string
		TplAppBlank               string
		TplRoutes                 string
		TplRoutesMinimal          string
		TplRoutesBlank            string
		TplHomeHandler            string
		TplHomeTest               string
		TplDashboardHandler       string
		TplDashboardTest          string
		TplHelpersTest            string
		TplGenericHandler         string
		TplGenericHandlerTest     string
		TplStore                  string
		TplStoreMinimal           string
		TplStoreTest              string
		TplStoreTestMinimal       string
		TplMigrations             string
		TplMigrationInit          string
		TplLayout                 string
		TplLayoutWelcome          string
		TplLayoutMinimal          string
		TplLayoutBlank            string
		TplCaisLogo               string
		TplPageHome               string
		TplPageDashboard          string
		TplPartialChatSSE         string
		TplPartialChatSSEAgent    string
		TplGenericPage            string
		TplInputCSS               string
		TplTailwind               string
		TplPackageJSON            string
		TplMakefile               string
		TplCIWorkflow             string
		TplPreCommitConfig        string
		TplGolangci               string
		TplPrettierrc             string
		TplPrettierignore         string
		TplGitignore              string
		TplREADME                 string
		TplREADMEBlank            string
		TplAgentsMD               string
		TplI18nCatalog            string
		TplI18nEn                 string
		TplI18nPt                 string
		TplI18nTest               string
		TplSeeds                  string
		TplSeedsMinimal           string
		TplUserModel              string
		TplMigrationAuth          string
		TplAuthHandler            string
		TplStorePasswordReset     string
		TplStorePasswordResetTest string
		TplAuthTest               string
		TplPageLogin              string
		TplPageSignup             string
		TplPageForgotPassword     string
		TplPageResetPassword      string
	}{
		TplGoMod:                  tplGoMod,
		TplEmptyCSS:               tplEmptyCSS,
		TplAir:                    tplAir,
		TplEnvExample:             tplEnvExample,
		TplWebEmbed:               tplWebEmbed,
		TplMain:                   tplMain,
		TplMainBlank:              tplMainBlank,
		TplConsole:                tplConsole,
		TplApp:                    tplApp,
		TplAppBlank:               tplAppBlank,
		TplRoutes:                 tplRoutes,
		TplRoutesMinimal:          tplRoutesMinimal,
		TplRoutesBlank:            tplRoutesBlank,
		TplHomeHandler:            tplHomeHandler,
		TplHomeTest:               tplHomeTest,
		TplDashboardHandler:       tplDashboardHandler,
		TplDashboardTest:          tplDashboardTest,
		TplHelpersTest:            tplHelpersTest,
		TplGenericHandler:         tplGenericHandler,
		TplGenericHandlerTest:     tplGenericHandlerTest,
		TplStore:                  tplStore,
		TplStoreMinimal:           tplStoreMinimal,
		TplStoreTest:              tplStoreTest,
		TplStoreTestMinimal:       tplStoreTestMinimal,
		TplMigrations:             tplMigrations,
		TplMigrationInit:          tplMigrationInit,
		TplLayout:                 tplLayout,
		TplLayoutWelcome:          tplLayoutWelcome,
		TplLayoutMinimal:          tplLayoutMinimal,
		TplLayoutBlank:            tplLayoutBlank,
		TplCaisLogo:               tplCaisLogo,
		TplPageHome:               tplPageHome,
		TplPageDashboard:          tplPageDashboard,
		TplPartialChatSSE:         tplPartialChatSSE,
		TplPartialChatSSEAgent:    tplPartialChatSSEAgent,
		TplGenericPage:            tplGenericPage,
		TplInputCSS:               tplInputCSS,
		TplTailwind:               tplTailwind,
		TplPackageJSON:            tplPackageJSON,
		TplMakefile:               tplMakefile,
		TplCIWorkflow:             tplCIWorkflow,
		TplPreCommitConfig:        tplPreCommitConfig,
		TplGolangci:               tplGolangci,
		TplPrettierrc:             tplPrettierrc,
		TplPrettierignore:         tplPrettierignore,
		TplGitignore:              tplGitignore,
		TplREADME:                 tplREADME,
		TplREADMEBlank:            tplREADMEBlank,
		TplAgentsMD:               tplAgentsMD,
		TplI18nCatalog:            tplI18nCatalog,
		TplI18nEn:                 tplI18nEn,
		TplI18nPt:                 tplI18nPt,
		TplI18nTest:               tplI18nTest,
		TplSeeds:                  tplSeeds,
		TplSeedsMinimal:           tplSeedsMinimal,
		TplUserModel:              tplUserModel,
		TplMigrationAuth:          tplMigrationAuth,
		TplAuthHandler:            tplAuthHandler,
		TplStorePasswordReset:     tplStorePasswordReset,
		TplStorePasswordResetTest: tplStorePasswordResetTest,
		TplAuthTest:               tplAuthTest,
		TplPageLogin:              tplPageLogin,
		TplPageSignup:             tplPageSignup,
		TplPageForgotPassword:     tplPageForgotPassword,
		TplPageResetPassword:      tplPageResetPassword,
	})

	for i := 0; i < v.NumField(); i++ {
		s := v.Field(i).String()
		if s == "" {
			t.Errorf("template %s is empty", v.Type().Field(i).Name)
		}
	}
}
