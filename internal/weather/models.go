package weather

type WeatherResponse struct {
	Location Location `json:"location"`
	Current  Current  `json:"current"`
}

type Location struct {
	Name           string  `json:"name"`
	Region         string  `json:"region"`
	Country        string  `json:"country"`
	Lat            float64 `json:"lat"`
	Lon            float64 `json:"lon"`
	TzID           string  `json:"tz_id"`
	LocaltimeEpoch int64   `json:"localtime_epoch"`
	Localtime      string  `json:"localtime"`
}

type Current struct {
	LastUpdatedEpoch int64       `json:"last_updated_epoch"`
	LastUpdated      string      `json:"last_updated"`
	TempC            float64     `json:"temp_c"`
	TempF            float64     `json:"temp_f"`
	IsDay            int         `json:"is_day"`
	Condition        Condition   `json:"condition"`
	WindMph          float64     `json:"wind_mph"`
	WindKph          float64     `json:"wind_kph"`
	WindDegree       float64     `json:"wind_degree"`
	WindDir          string      `json:"wind_dir"`
	PressureMb       float64     `json:"pressure_mb"`
	PressureIn       float64     `json:"pressure_in"`
	PrecipMm         float64     `json:"precip_mm"`
	PrecipIn         float64     `json:"precip_in"`
	Humidity         int         `json:"humidity"`
	Cloud            int         `json:"cloud"`
	FeelslikeC       float64     `json:"feelslike_c"`
	FeelslikeF       float64     `json:"feelslike_f"`
	VisKm            float64     `json:"vis_km"`
	VisMiles         float64     `json:"vis_miles"`
	UV               int         `json:"uv"`
	GustMph          float64     `json:"gust_mph"`
	GustKph          float64     `json:"gust_kph"`
	AirQuality       *AirQuality `json:"air_quality,omitempty"`
}

type Condition struct {
	Text string `json:"text"`
	Icon string `json:"icon"`
	Code int    `json:"code"`
}

type AirQuality struct {
	CO           float64 `json:"co"`
	NO2          float64 `json:"no2"`
	O3           float64 `json:"o3"`
	SO2          float64 `json:"so2"`
	PM25         float64 `json:"pm2_5"`
	PM10         float64 `json:"pm10"`
	UsEpaIndex   int     `json:"us-epa-index"`
	GbDefraIndex int     `json:"gb-defra-index"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
