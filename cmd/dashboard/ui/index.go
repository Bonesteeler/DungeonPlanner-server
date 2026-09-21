package ui

import (
    "fmt"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"

    "dungeonplannerserver/dto"
    "dungeonplannerserver-dashboard/internal/queries"
)

var indexContent *fyne.Container

func NewIndexTab(scenes []dto.SceneResponse) fyne.CanvasObject {
    indexContent = buildIndexContent(scenes)
    return indexContent
}

func buildIndexContent(scenes []dto.SceneResponse) *fyne.Container {
		fmt.Println("Building index content with", len(scenes), "scenes")
    title := widget.NewLabelWithStyle("Scene Statistics", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

    total := len(scenes)

    statusCounts := make(map[dto.ModerationStatus]int)
    uniqueTileCount := 0
    for _, s := range scenes {
        statusCounts[0]++
        uniqueTileCount += len(s.UniqueTileIDs)
    }

    pending := statusCounts[dto.ModerationStatusPending]
    approved := statusCounts[dto.ModerationStatusApproved]
    rejected := statusCounts[dto.ModerationStatusRejected]

    avgTiles := 0.0
    if total > 0 {
        avgTiles = float64(uniqueTileCount) / float64(total)
    }

    stats := container.NewGridWithColumns(2,
        widget.NewLabelWithStyle("Total Scenes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
        widget.NewLabel(fmt.Sprintf("%d", total)),

        widget.NewLabelWithStyle("Pending Review", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
        widget.NewLabel(fmt.Sprintf("%d", pending)),

        widget.NewLabelWithStyle("Approved", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
        widget.NewLabel(fmt.Sprintf("%d", approved)),

        widget.NewLabelWithStyle("Rejected", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
        widget.NewLabel(fmt.Sprintf("%d", rejected)),

        widget.NewLabelWithStyle("Avg. Unique Tiles per Scene", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
        widget.NewLabel(fmt.Sprintf("%.1f", avgTiles)),
    )

    card := widget.NewCard("", "", stats)

    return container.NewBorder(
        title,
        nil, nil, nil,
        container.NewCenter(card),
    )
}

func RefreshIndexTab() {
	scenes := queries.GetAllScenes()
	if indexContent == nil {
			return
	}
	refreshed := buildIndexContent(scenes)
	indexContent.Objects = refreshed.Objects
	indexContent.Refresh()
}