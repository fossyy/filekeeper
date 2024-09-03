package admin

import (
	"encoding/json"
	adminIndex "github.com/fossyy/filekeeper/view/admin/index"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"log"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type SystemStats struct {
	TotalMemoryGB     float64 `json:"total_memory_gb"`
	MemoryUsedGB      float64 `json:"memory_used_gb"`
	CpuUsagePercent   float64 `json:"cpu_usage_percent"`
	UploadSpeedMbps   float64 `json:"upload_speed_mbps"`
	DownloadSpeedMbps float64 `json:"download_speed_mbps"`
}

func SetupRoutes() *http.ServeMux {
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//users, err := app.Admin.Database.GetAllUsers()
		//if err != nil {
		//	http.Error(w, "Unable to retrieve users", http.StatusInternalServerError)
		//	return
		//}
		//w.Header().Set("Content-Type", "application/json")
		//if err := json.NewEncoder(w).Encode(users); err != nil {
		//	http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		//	return
		//}
		adminIndex.Main().Render(r.Context(), w)
		return
	})

	handler.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		for {
			handlerWS(conn)
		}
	})

	fileServer := http.FileServer(http.Dir("./public"))
	handler.Handle("/public/", http.StripPrefix("/public", fileServer))
	return handler
}

func handlerWS(conn *websocket.Conn) {
	prevCounters, _ := net.IOCounters(false)
	for {
		vMem, _ := mem.VirtualMemory()

		totalMemoryGB := float64(vMem.Total) / (1024 * 1024 * 1024)
		memoryUsedGB := float64(vMem.Used) / (1024 * 1024 * 1024)

		cpuPercent, _ := cpu.Percent(time.Second, false)

		currentCounters, _ := net.IOCounters(false)

		uploadBytes := currentCounters[0].BytesSent - prevCounters[0].BytesSent
		downloadBytes := currentCounters[0].BytesRecv - prevCounters[0].BytesRecv
		uploadSpeedMbps := float64(uploadBytes) * 8 / (1024 * 1024) / 2
		downloadSpeedMbps := float64(downloadBytes) * 8 / (1024 * 1024) / 2

		prevCounters = currentCounters

		stats := SystemStats{
			TotalMemoryGB:     totalMemoryGB,
			MemoryUsedGB:      memoryUsedGB,
			CpuUsagePercent:   cpuPercent[0],
			UploadSpeedMbps:   uploadSpeedMbps,
			DownloadSpeedMbps: downloadSpeedMbps,
		}

		statsJson, _ := json.Marshal(stats)

		err := conn.WriteMessage(websocket.TextMessage, statsJson)
		if err != nil {
			conn.Close()
			return
		}

		time.Sleep(2 * time.Second)
	}

}
