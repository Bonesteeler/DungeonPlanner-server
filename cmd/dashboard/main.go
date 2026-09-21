package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"dungeonplannerserver-dashboard/ui"
	"dungeonplannerserver-dashboard/internal/queries"
)

func main() {
    a := app.New()
    w := a.NewWindow("Dungeon Planner Dashboard")

		scenes := queries.GetAllScenes()

		tabs := container.NewAppTabs(
			container.NewTabItem("Login", ui.NewLoginUI()),
			container.NewTabItem("Main", ui.NewIndexTab(scenes)),
			container.NewTabItem("Moderation Queue", ui.NewModerationQueueTab(scenes)),
		)

		tabs.SetTabLocation(container.TabLocationTop)

    w.SetContent(tabs)
    w.Resize(fyne.NewSize(1200, 500))
    w.ShowAndRun()
}