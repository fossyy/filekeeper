package fileHandler

import (
	"fmt"
	"github.com/a-h/templ"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	fileView "github.com/fossyy/filekeeper/view/client/file"
	"net/http"
	"path/filepath"
	"strconv"
)

func GET(w http.ResponseWriter, r *http.Request) {
	userSession := r.Context().Value("user").(types.User)
	query := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")
	var fileStatus types.FileStatus
	if status == "private" {
		fileStatus = types.Private
	} else if status == "public" {
		fileStatus = types.Public
	} else {
		fileStatus = types.All
	}
	files, err := app.Server.Database.GetFiles(userSession.UserID.String(), query, fileStatus)
	if err != nil {
		app.Server.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var filesData []types.FileData

	for _, file := range files {
		saveFolder := filepath.Join("uploads", userSession.UserID.String(), file.ID.String())

		pattern := fmt.Sprintf("%s/chunk_*", saveFolder)
		chunkFiles, err := filepath.Glob(pattern)

		missingChunk := err != nil || len(chunkFiles) != int(file.TotalChunk)

		filesData = append(filesData, types.FileData{
			ID:         file.ID.String(),
			Name:       file.Name,
			Size:       utils.ConvertFileSize(file.Size),
			IsPrivate:  file.IsPrivate,
			Type:       file.Type,
			Done:       !missingChunk,
			Downloaded: strconv.FormatUint(file.Downloaded, 10),
		})
	}

	if r.Header.Get("hx-request") == "true" {
		component := fileView.FileTable(filesData)
		err := component.Render(r.Context(), w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
	allowance, err := app.Server.Database.GetAllowance(userSession.UserID)
	if err != nil {
		app.Server.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	usage, err := app.Server.Service.GetUserStorageUsage(userSession.UserID.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	allowanceStats := &types.Allowance{
		AllowanceByte:        utils.ConvertFileSize(allowance.AllowanceByte),
		AllowanceUsedByte:    utils.ConvertFileSize(usage),
		AllowanceUsedPercent: fmt.Sprintf("%.2f", float64(usage)/float64(allowance.AllowanceByte)*100),
	}

	var component templ.Component
	if r.Header.Get("hx-request") == "true" {
		component = fileView.MainContent("Filekeeper - File Dashboard", filesData, userSession, allowanceStats)
	} else {
		component = fileView.Main("Filekeeper - File Dashboard", filesData, userSession, allowanceStats)
	}

	err = component.Render(r.Context(), w)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
