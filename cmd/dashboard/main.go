package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"dungeonplannerserver-dashboard/internal/queries"
	"dungeonplannerserver-dashboard/ui"
)

const LoginTab = "Login"
const MainTab = "Main"
const ModerationQueueTab = "Moderation Queue"

func main() {
    a := app.New()
    w := a.NewWindow("Dungeon Planner Dashboard")

		scenes := queries.GetAllScenes()

		tabs := container.NewAppTabs(
			container.NewTabItem(LoginTab, ui.NewLoginUI()),
			container.NewTabItem(MainTab, ui.NewIndexTab(scenes)),
			container.NewTabItem(ModerationQueueTab, ui.NewModerationQueueTab(scenes)),
		)

		tabs.SetTabLocation(container.TabLocationTop)

		tabs.OnSelected = func(t *container.TabItem) {
			if t.Text == MainTab {
				ui.RefreshIndexTab()
			}
			if t.Text == ModerationQueueTab {
				ui.RefreshModerationQueueTab()
			}
		}

    w.SetContent(tabs)
    w.Resize(fyne.NewSize(1200, 500))
    w.ShowAndRun()
}