package uploadHandler

import (
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"io"
	"net/http"
	"strconv"
)

func POST(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	file, err := app.Server.Service.GetFile(fileID)
	if err != nil {
		app.Server.Logger.Error("error getting upload info: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rawIndex := r.FormValue("index")
	index, err := strconv.Atoi(rawIndex)
	if err != nil {
		return
	}

	fileByte, _, err := r.FormFile("chunk")
	if err != nil {
		app.Server.Logger.Error("error getting upload info: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer fileByte.Close()

	buffer, err := io.ReadAll(fileByte)
	if err != nil {
		app.Server.Logger.Error("error copying byte to file dst: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = app.Server.Storage.Add(r.Context(), fmt.Sprintf("%s/%s/chunk_%d", file.OwnerID.String(), file.ID.String(), index), buffer)
	if err != nil {
		app.Server.Logger.Error("error copying byte to file dst: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	return
}
