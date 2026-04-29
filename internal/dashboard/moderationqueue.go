package dashboard

import (
    "fmt"
    "strings"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"

    "DungeonPlannerServer/internal/db/tables"
)

func NewModerationQueueTab(scenes []tables.Scene) fyne.CanvasObject {
    title := widget.NewLabelWithStyle("Moderation Queue", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
    queueSize := widget.NewLabel(fmt.Sprintf("Queue size: %d", len(scenes)))

    table := widget.NewTable(
        func() (int, int) {
            return len(scenes) + 1, 6
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
                scene := scenes[id.Row-1]
                switch id.Col {
                case 0:
                    label.SetText("Review")
                case 1:
                    label.SetText(scene.ID.String())
                case 2:
                    if scene.Name != nil {
                        label.SetText(*scene.Name)
                    } else {
                        label.SetText("—")
                    }
                case 3:
                    if scene.Author != nil {
                        label.SetText(*scene.Author)
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

    table.SetColumnWidth(0, 80)
    table.SetColumnWidth(1, 400)
    table.SetColumnWidth(2, 160)
    table.SetColumnWidth(3, 200)
    table.SetColumnWidth(4, 140)
    table.SetColumnWidth(5, 200)

    return container.NewBorder(
        container.NewVBox(title, queueSize),
        nil, nil, nil,
        table,
    )
}