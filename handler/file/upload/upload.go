package uploadHandler

import (
	"fmt"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/types"
	"io"
	"net/http"
	"strconv"
)

func POST(w http.ResponseWriter, r *http.Request) {
	userSession, ok := r.Context().Value("user").(types.User)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	fileID := r.PathValue("id")
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		app.Server.Logger.Error(err.Error())
		return
	}

	file, err := app.Server.Cache.GetFile(r.Context(), fileID)
	if err != nil {
		app.Server.Logger.Error("error getting upload info: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if userSession.UserID != file.OwnerID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	index, err := strconv.Atoi(r.FormValue("index"))
	if err != nil || index < 0 || uint64(index) >= file.TotalChunk {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fileByte, _, err := r.FormFile("chunk")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		app.Server.Logger.Error("error getting upload chunk: " + err.Error())
		return
	}
	defer fileByte.Close()

	buffer, err := io.ReadAll(fileByte)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error("error copying byte to file dst: " + err.Error())
		return
	}
	if len(buffer) > 2*1024*1024 || (index < int(file.TotalChunk)-1 && len(buffer) != 2*1024*1024) {
		http.Error(w, "Invalid chunk size", http.StatusBadRequest)
		return
	}

	err = app.Server.Storage.Add(r.Context(), fmt.Sprintf("%s/%s/chunk_%d", file.OwnerID.String(), file.ID.String(), index), buffer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error("error copying byte to file dst: " + err.Error())
		return
	}
	err = app.Server.Cache.UpdateFileChunk(r.Context(), file.ID, file.OwnerID, strconv.Itoa(index), file.TotalChunk)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
	return
}
