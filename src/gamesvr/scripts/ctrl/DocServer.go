package ctrl

import (
	"gamekit/src/base/control"
	"gamekit/src/base/logger"
	"encoding/json"
	"os"
	"sort"
)

func buildDocList() []*control.CtrlHandleData {
	handlers := control.GetAllHandlers()
	items := make([]*control.CtrlHandleData, 0, len(handlers))
	for _, h := range handlers {
		items = append(items, h)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Cmd < items[j].Cmd
	})
	return items
}

func InitDocCache() {
	items := buildDocList()
	jsonData, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		logger.Log.Error("InitDocCache marshal error: %v", err)
		return
	}

	outputFile := "ctrl_commands.json"
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		logger.Log.Error("InitDocCache write file error: %v", err)
		return
	}
	logger.Log.Info("InitDocCache write ctrl commands to %s, count: %d", outputFile, len(items))
}