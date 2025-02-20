package rcon

type TPSStat struct {
	ID, Name string
	Ticktime float64
	TPS      float64
}

type EntityCount struct {
	Name  string
	Count int
}

type DynmapRenderStat struct {
	Dim                          string
	Processed, Rendered, Updated int
}

type DynmapChunkloadingStat struct {
	State    string
	Count    int
	Duration float64
}

type TickStats struct {
	Target        float64
	Average       float64
	P50, P95, P99 float64
}
