package api

import (
	"runtime"
	"sync/atomic"
	"time"

	gethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/mackerelio/go-osstat/memory"
)

const (
	StatusUpdateInterval = 60 // seconds
)

type Stats struct {
	NumGoroutine     int    `json:"numGoroutine"`
	NumGC            uint32 `json:"numGC"`
	MemAllocMB       uint64 `json:"memAllocMB"`
	MemSysMB         uint64 `json:"memSysMB"`
	OsMemTotalMB     uint64 `json:"osMemTotalMB"`
	OsMemUsedMB      uint64 `json:"osMemUsedMB"`
	OsMemCachedMB    uint64 `json:"osMemCachedMB"`
	OsMemFreeMB      uint64 `json:"osMemFreeMB"`
	OsMemActiveMB    uint64 `json:"osMemActiveMB"`
	OsMemInactiveMB  uint64 `json:"osMemInactiveMB"`
	OsMemSwapTotalMB uint64 `json:"osMemSwapTotalMB"`
	OsMemSwapUsedMB  uint64 `json:"osMemSwapUsedMB"`
	OsMemSwapFreeMB  uint64 `json:"osMemSwapFreeMB"`
	NumEthCall       uint64 `json:"numEthCall"`
}

type DebugAPI interface {
	GetStats() Stats
	GetSeq(addr gethcmn.Address) hexutil.Uint64
}

func newDebugAPI(ethAPI *ethAPI) DebugAPI {
	return &debugAPI{ethAPI: ethAPI}
}

type debugAPI struct {
	ethAPI         *ethAPI
	lastUpdateTime int64
	stats          Stats
}

func (api *debugAPI) GetSeq(addr gethcmn.Address) hexutil.Uint64 {
	return hexutil.Uint64(api.ethAPI.backend.GetSeq(addr))
}

func (api *debugAPI) GetStats() Stats {
	now := time.Now().Unix()
	lastUpdateTime := atomic.LoadInt64(&api.lastUpdateTime)
	if now > lastUpdateTime+StatusUpdateInterval {
		if atomic.CompareAndSwapInt64(&api.lastUpdateTime, lastUpdateTime, now) {
			api.updateStats()
		}
	}

	return api.stats
}

func (api *debugAPI) updateStats() {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	api.stats.NumGoroutine = runtime.NumGoroutine()
	api.stats.NumGC = memStats.NumGC
	api.stats.MemAllocMB = toMB(memStats.Alloc)
	api.stats.MemSysMB = toMB(memStats.Sys)

	osMemStats, err := memory.Get()
	if err == nil {
		api.stats.OsMemTotalMB = toMB(osMemStats.Total)
		api.stats.OsMemUsedMB = toMB(osMemStats.Used)
		api.stats.OsMemCachedMB = toMB(osMemStats.Cached)
		api.stats.OsMemFreeMB = toMB(osMemStats.Free)
		api.stats.OsMemActiveMB = toMB(osMemStats.Active)
		api.stats.OsMemInactiveMB = toMB(osMemStats.Inactive)
		api.stats.OsMemSwapTotalMB = toMB(osMemStats.SwapTotal)
		api.stats.OsMemSwapUsedMB = toMB(osMemStats.SwapUsed)
		api.stats.OsMemSwapFreeMB = toMB(osMemStats.SwapFree)
	}

	api.stats.NumEthCall = atomic.LoadUint64(&api.ethAPI.numCall)
}

func toMB(n uint64) uint64 {
	return n / 1024 / 1024
}
