package fileHandler

import (
	"fmt"
	"github.com/a-h/templ"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
	fileView "github.com/fossyy/filekeeper/view/client/file"
	"net/http"
	"strconv"
)

func GET(w http.ResponseWriter, r *http.Request) {
	userSession := r.Context().Value("user").(types.User)
	files, err := app.Server.Database.GetFiles(userSession.UserID.String())
	if err != nil {
		app.Server.Logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var filesData []types.FileData
	for i := 0; i < len(files); i++ {
		filesData = append(filesData, types.FileData{
			ID:         files[i].ID.String(),
			Name:       files[i].Name,
			Size:       utils.ConvertFileSize(files[i].Size),
			IsPrivate:  files[i].IsPrivate,
			Type:       files[i].Type,
			Downloaded: strconv.FormatUint(files[i].Downloaded, 10),
		})
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
		component = fileView.MainContent(filesData, userSession, allowanceStats)
	} else {
		component = fileView.Main("File Dashboard", filesData, userSession, allowanceStats)
	}
	err = component.Render(r.Context(), w)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
