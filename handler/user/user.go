package userHandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/a-h/templ"
	"github.com/fossyy/filekeeper/app"
	"github.com/fossyy/filekeeper/session"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/types/models"
	"github.com/fossyy/filekeeper/utils"
	"github.com/fossyy/filekeeper/view/client/user"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
	"net/http"
	"path/filepath"
	"strings"
)

var errorMessages = map[string]string{
	"password_not_match": "The passwords provided do not match. Please try again.",
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type ActionType string

const (
	UploadNewFile ActionType = "UploadNewFile"
	DeleteFile    ActionType = "DeleteFile"
	Ping          ActionType = "Ping"
)

type WebsocketAction struct {
	Action ActionType `json:"action"`
}

type ActionUploadNewFile struct {
	Action    string `json:"action"`
	Name      string `json:"name"`
	Size      uint64 `json:"size"`
	Chunk     uint64 `json:"chunk"`
	StartHash string `json:"startHash"`
	EndHash   string `json:"endHash"`
	RequestID string `json:"requestID"`
}

func GET(w http.ResponseWriter, r *http.Request) {
	userSession := r.Context().Value("user").(types.User)
	if r.Header.Get("upgrade") == "websocket" {
		upgrade, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
			return
		}
		handlerWS(upgrade, userSession)
	}

	sessions, err := session.GetSessions(userSession.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	allowance, err := app.Server.Database.GetAllowance(userSession.UserID)
	if err != nil {
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
	if err := r.URL.Query().Get("error"); err != "" {
		message, ok := errorMessages[err]
		if !ok {
			message = "Unknown error occurred. Please contact support at bagas@fossy.my.id for assistance."
		}
		if r.Header.Get("hx-request") == "true" {
			component = userView.MainContent(userSession, allowanceStats, sessions, types.Message{
				Code:    0,
				Message: message,
			})
		} else {
			component = userView.Main("Filekeeper - User Page", userSession, allowanceStats, sessions, types.Message{
				Code:    0,
				Message: message,
			})
		}
	} else {
		if r.Header.Get("hx-request") == "true" {
			component = userView.MainContent(userSession, allowanceStats, sessions, types.Message{
				Code:    1,
				Message: "",
			})
		} else {
			component = userView.Main("Filekeeper - User Page", userSession, allowanceStats, sessions, types.Message{
				Code:    1,
				Message: "",
			})
		}
	}
	err = component.Render(r.Context(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		app.Server.Logger.Error(err.Error())
		return
	}
}

func handlerWS(conn *websocket.Conn, userSession types.User) {
	defer conn.Close()
	var err error
	var message []byte

	for {
		_, message, err = conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				app.Server.Logger.Error("Unexpected connection closure:", err)
			} else {
				app.Server.Logger.Error("Connection closed:", err)
			}
			return
		}
		var action WebsocketAction
		err = json.Unmarshal(message, &action)
		if err != nil {
			app.Server.Logger.Error("Error unmarshalling WebsocketAction:", err)
			sendErrorResponse(conn, action.Action, "Internal Server Error")
			continue
		}

		switch action.Action {
		case UploadNewFile:
			var uploadNewFile ActionUploadNewFile
			err = json.Unmarshal(message, &uploadNewFile)
			if err != nil {
				app.Server.Logger.Error("Error unmarshalling ActionUploadNewFile:", err)
				sendErrorResponse(conn, action.Action, "Internal Server Error")
				continue
			}
			var file *models.File
			file, err = app.Server.Database.GetUserFile(uploadNewFile.Name, userSession.UserID.String())
			allowedFileTypes := []string{"jpg", "jpeg", "png", "gif", "bmp", "tiff", "pdf", "doc", "docx", "txt", "odt", "xls", "xlsx", "ppt", "pptx", "zip", "rar", "tar", "gz", "7z", "bz2", "exe", "bin", "sh", "bat", "cmd", "msi", "apk"}
			isAllowedFileType := func(fileType string) bool {
				for _, allowed := range allowedFileTypes {
					if fileType == allowed {
						return true
					}
				}
				return false
			}
			fileName := strings.Split(uploadNewFile.Name, ".")
			fileType := fileName[len(fileName)-1]
			if !isAllowedFileType(fileType) {
				fileType = "doc"
			}

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					newFile := models.File{
						ID:         uuid.New(),
						OwnerID:    userSession.UserID,
						Name:       uploadNewFile.Name,
						Size:       uploadNewFile.Size,
						StartHash:  uploadNewFile.StartHash,
						EndHash:    uploadNewFile.EndHash,
						Type:       fileType,
						TotalChunk: uploadNewFile.Chunk,
						Downloaded: 0,
					}
					err := app.Server.Database.CreateFile(&newFile)
					if err != nil {
						sendErrorResponse(conn, action.Action, "Error Creating File")
						continue
					}
					fileData := &types.FileWithDetail{
						ID:         newFile.ID,
						OwnerID:    newFile.OwnerID,
						Name:       newFile.Name,
						Size:       newFile.Size,
						Downloaded: newFile.Downloaded,
					}
					fileData.Chunk = make(map[string]bool)

					saveFolder := filepath.Join("uploads", userSession.UserID.String(), newFile.ID.String(), newFile.Name)

					pattern := fmt.Sprintf("%s/chunk_*", saveFolder)
					chunkFiles, err := filepath.Glob(pattern)
					if err != nil {
						app.Server.Logger.Error(err.Error())
						fileData.Done = false
					} else {
						for i := 0; i <= int(newFile.TotalChunk); i++ {
							fileData.Chunk[fmt.Sprintf("chunk_%d", i)] = false
						}

						for _, chunkFile := range chunkFiles {
							var chunkIndex int
							fmt.Sscanf(filepath.Base(chunkFile), "chunk_%d", &chunkIndex)

							fileData.Chunk[fmt.Sprintf("chunk_%d", chunkIndex)] = true
						}

						for i := 0; i <= int(newFile.TotalChunk); i++ {
							if !fileData.Chunk[fmt.Sprintf("chunk_%d", i)] {
								fileData.Done = false
								break
							}
						}
					}
					sendSuccessResponseWithID(conn, action.Action, fileData, uploadNewFile.RequestID)
					continue
				} else {
					sendErrorResponse(conn, action.Action, "Unknown error")
					continue
				}
			}
			if uploadNewFile.StartHash != file.StartHash || uploadNewFile.EndHash != file.EndHash {
				sendErrorResponse(conn, action.Action, "File Is Different")
				continue
			}
			fileData := &types.FileWithDetail{
				ID:         file.ID,
				OwnerID:    file.OwnerID,
				Name:       file.Name,
				Size:       file.Size,
				Downloaded: file.Downloaded,
				Chunk:      make(map[string]bool),
				Done:       true,
			}

			saveFolder := filepath.Join("uploads", userSession.UserID.String(), fileData.ID.String(), fileData.Name)
			pattern := fmt.Sprintf("%s/chunk_*", saveFolder)
			chunkFiles, err := filepath.Glob(pattern)

			if err != nil {
				app.Server.Logger.Error(err.Error())
				fileData.Done = false
			} else {
				for i := 0; i <= int(file.TotalChunk); i++ {
					fileData.Chunk[fmt.Sprintf("chunk_%d", i)] = false
				}

				for _, chunkFile := range chunkFiles {
					var chunkIndex int
					fmt.Sscanf(filepath.Base(chunkFile), "chunk_%d", &chunkIndex)

					fileData.Chunk[fmt.Sprintf("chunk_%d", chunkIndex)] = true
				}

				for i := 0; i <= int(file.TotalChunk); i++ {
					if !fileData.Chunk[fmt.Sprintf("chunk_%d", i)] {
						fileData.Done = false
						break
					}
				}
			}

			sendSuccessResponseWithID(conn, action.Action, fileData, uploadNewFile.RequestID)
			continue
		case Ping:
			sendSuccessResponse(conn, action.Action, map[string]string{"message": "received"})
			continue
		}
	}
}

func sendErrorResponse(conn *websocket.Conn, action ActionType, message string) {
	response := map[string]interface{}{
		"action":  action,
		"status":  "error",
		"message": message,
	}
	marshal, err := json.Marshal(response)
	if err != nil {
		app.Server.Logger.Error("Error marshalling error response:", err)
		return
	}
	err = conn.WriteMessage(websocket.TextMessage, marshal)
	if err != nil {
		app.Server.Logger.Error("Error writing error response:", err)
	}
}

func sendSuccessResponse(conn *websocket.Conn, action ActionType, response interface{}) {
	responseJSON := map[string]interface{}{
		"action":   action,
		"status":   "success",
		"response": response,
	}
	marshal, err := json.Marshal(responseJSON)
	if err != nil {
		app.Server.Logger.Error("Error marshalling success response:", err)
		return
	}
	err = conn.WriteMessage(websocket.TextMessage, marshal)
	if err != nil {
		app.Server.Logger.Error("Error writing success response:", err)
	}
}

func sendSuccessResponseWithID(conn *websocket.Conn, action ActionType, response interface{}, responseID string) {
	responseJSON := map[string]interface{}{
		"action":     action,
		"status":     "success",
		"response":   response,
		"responseID": responseID,
	}
	marshal, err := json.Marshal(responseJSON)
	if err != nil {
		app.Server.Logger.Error("Error marshalling success response:", err)
		return
	}
	err = conn.WriteMessage(websocket.TextMessage, marshal)
	if err != nil {
		app.Server.Logger.Error("Error writing success response:", err)
	}
}
