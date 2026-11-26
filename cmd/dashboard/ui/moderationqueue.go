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
	selectedSceneLabel  *widget.Label
	moderationError     *widget.Label
	acceptSceneButton   *widget.Button
	rejectSceneButton   *widget.Button
	selectedSceneID     string
)

func NewModerationQueueTab(scenes []dto.SceneResponse) fyne.CanvasObject {
	return BuildModerationQueueContent(scenes)
}

func BuildModerationQueueContent(scenes []dto.SceneResponse) *fyne.Container {
	moderationScenes = scenes
	selectedSceneID = ""

	title := widget.NewLabelWithStyle("Moderation Queue", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	moderationQueueSize = widget.NewLabel(fmt.Sprintf("Queue size: %d", len(moderationScenes)))
	selectedSceneLabel = widget.NewLabel("No scene selected")
	moderationError = widget.NewLabel("")
	moderationError.Wrapping = fyne.TextWrapWord

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
				return
			}

			label.TextStyle = fyne.TextStyle{}
			scene := moderationScenes[id.Row-1]
			switch id.Col {
			case 0:
				label.SetText("Select")
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
		},
	)

	moderationTable.SetColumnWidth(0, 80)
	moderationTable.SetColumnWidth(1, 400)
	moderationTable.SetColumnWidth(2, 160)
	moderationTable.SetColumnWidth(3, 200)
	moderationTable.SetColumnWidth(4, 140)
	moderationTable.SetColumnWidth(5, 200)

	acceptSceneButton = widget.NewButton("Accept", func() {
		moderateSelectedScene(dto.ModerationStatusApproved)
	})
	rejectSceneButton = widget.NewButton("Reject", func() {
		moderateSelectedScene(dto.ModerationStatusRejected)
	})
	acceptSceneButton.Disable()
	rejectSceneButton.Disable()

	moderationTable.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 || id.Row > len(moderationScenes) {
			clearModerationSelection()
			return
		}
		scene := moderationScenes[id.Row-1]
		selectedSceneID = scene.ID
		name := scene.Name
		if name == "" {
			name = "Unnamed scene"
		}
		selectedSceneLabel.SetText(fmt.Sprintf("Selected: %s (%s)", name, scene.ID))
		moderationError.SetText("")
		acceptSceneButton.Enable()
		rejectSceneButton.Enable()
	}

	reviewControls := container.NewVBox(
		selectedSceneLabel,
		container.NewHBox(acceptSceneButton, rejectSceneButton),
		moderationError,
	)
	return container.NewBorder(
		container.NewVBox(title, moderationQueueSize),
		reviewControls,
		nil,
		nil,
		moderationTable,
	)
}

func moderateSelectedScene(status dto.ModerationStatus) {
	if selectedSceneID == "" {
		return
	}
	if err := queries.UpdateSceneModerationStatus(selectedSceneID, status); err != nil {
		moderationError.SetText(err.Error())
		return
	}
	RefreshModerationQueueTab()
}

func clearModerationSelection() {
	selectedSceneID = ""
	if selectedSceneLabel != nil {
		selectedSceneLabel.SetText("No scene selected")
	}
	if acceptSceneButton != nil {
		acceptSceneButton.Disable()
	}
	if rejectSceneButton != nil {
		rejectSceneButton.Disable()
	}
	if moderationTable != nil {
		moderationTable.UnselectAll()
	}
}

func RefreshModerationQueueTab() {
	if moderationTable == nil {
		return
	}
	scenes, err := queries.GetPendingScenes()
	if err != nil {
		moderationError.SetText(fmt.Sprintf("Failed to load moderation queue: %v", err))
		return
	}
	moderationScenes = scenes
	moderationQueueSize.SetText(fmt.Sprintf("Queue size: %d", len(moderationScenes)))
	moderationError.SetText("")
	clearModerationSelection()
	moderationTable.Refresh()
}
