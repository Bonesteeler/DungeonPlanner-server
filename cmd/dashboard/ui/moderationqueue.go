package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"dungeonplannerserver-dashboard/internal/queries"
	"dungeonplannerserver/dto"
)

var (
    moderationScenes    []dto.SceneResponse
    moderationTable     *widget.Table
    moderationQueueSize *widget.Label
)

func NewModerationQueueTab(scenes []dto.SceneResponse) fyne.CanvasObject {
		queueContent := BuildModerationQueueContent(scenes)
		return queueContent
}

func BuildModerationQueueContent(scenes []dto.SceneResponse) *fyne.Container {
		fmt.Println("Building moderattion queue with", len(scenes), "scenes")
    moderationScenes = scenes

    title := widget.NewLabelWithStyle("Moderation Queue", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
    moderationQueueSize = widget.NewLabel(fmt.Sprintf("Queue size: %d", len(moderationScenes)))

    moderationTable = widget.NewTable(
        func() (int, int) {
            return len(moderationScenes) + 1, 6
        },
        func() fyne.CanvasObject {
            return widget.NewLabel("")
        },
        func(id widget.TableCellID, cell fyne.CanvasObject) {
            label := cell.(*widget.Label)
            if id.Row == 0 {
                headers := []string{"Action", "Scene ID", "Scene Name", "Author", "Unique Tile Count", "Unique Tile IDs"}
                label.TextStyle = fyne.TextStyle{Bold: true}
                label.SetText(headers[id.Col])
            } else {
                scene := moderationScenes[id.Row-1]
                switch id.Col {
                case 0:
                    label.SetText("Review")
                case 1:
                    label.SetText(scene.ID)
                case 2:
                    if scene.Name != "" {
                        label.SetText(scene.Name)
                    } else {
                        label.SetText("—")
                    }
                case 3:
                    if scene.Author != "" {
                        label.SetText(scene.Author)
                    } else {
                        label.SetText("—")
                    }
                case 4:
                    label.SetText(fmt.Sprintf("%d", len(scene.UniqueTileIDs)))
                case 5:
                    label.SetText(strings.Join(scene.UniqueTileIDs, ", "))
                }
            }
        },
    )

    moderationTable.SetColumnWidth(0, 80)
    moderationTable.SetColumnWidth(1, 400)
    moderationTable.SetColumnWidth(2, 160)
    moderationTable.SetColumnWidth(3, 200)
    moderationTable.SetColumnWidth(4, 140)
    moderationTable.SetColumnWidth(5, 200)

    return container.NewBorder(
        container.NewVBox(title, moderationQueueSize),
        nil, nil, nil,
        moderationTable,
    )
}

func RefreshModerationQueueTab() {
    if moderationTable == nil {
        return
    }
    moderationScenes = queries.GetAllScenes()
		BuildModerationQueueContent(moderationScenes)
    moderationTable.Refresh()
}