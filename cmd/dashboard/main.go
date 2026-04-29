package main

import (
    "strings"
		"os"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"

		"github.com/labstack/echo/v4"

		"DungeonPlannerServer/internal/dashboard"
		"DungeonPlannerServer/internal/db"
		"DungeonPlannerServer/internal/db/tables"
)

func main() {
		e := echo.New()
    a := app.New()
    w := a.NewWindow("Dungeon Planner Dashboard")

 		secretBytes, err := os.ReadFile("secrets/password.txt")
		if err != nil {
			e.Logger.Fatal("Error reading secret file:", err)
		}
		dbPassword := strings.TrimSpace(string(secretBytes))
 
		dbconnection, err := db.Connect(dbPassword)
		if err != nil {
			e.Logger.Fatal("Database connection failed:", err)
		}

		scenes, err := tables.GetAllScenes(dbconnection)
		if err != nil {
			e.Logger.Fatal("Failed to get scenes:", err)
		}

		tabs := container.NewAppTabs(
			container.NewTabItem("Main", dashboard.NewIndexTab(scenes)),
			container.NewTabItem("Moderation Queue", dashboard.NewModerationQueueTab(scenes)),
		)

		tabs.SetTabLocation(container.TabLocationTop)

    w.SetContent(tabs)
    w.Resize(fyne.NewSize(1200, 500))
    w.ShowAndRun()
}