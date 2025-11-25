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

type ForecastResponse struct {
	Location Location `json:"location"`
	Current  Current  `json:"current"`
	Forecast Forecast `json:"forecast"`
	Alerts   *Alerts  `json:"alerts,omitempty"`
}

type Forecast struct {
	ForecastDay []ForecastDay `json:"forecastday"`
}

type ForecastDay struct {
	Date      string `json:"date"`
	DateEpoch int64  `json:"date_epoch"`
	Day       Day    `json:"day"`
	Astro     Astro  `json:"astro"`
	Hour      []Hour `json:"hour"`
}

type Day struct {
	MaxTempC          float64   `json:"maxtemp_c"`
	MaxTempF          float64   `json:"maxtemp_f"`
	MinTempC          float64   `json:"mintemp_c"`
	MinTempF          float64   `json:"mintemp_f"`
	AvgTempC          float64   `json:"avgtemp_c"`
	AvgTempF          float64   `json:"avgtemp_f"`
	MaxWindMph        float64   `json:"maxwind_mph"`
	MaxWindKph        float64   `json:"maxwind_kph"`
	TotalPrecipMm     float64   `json:"totalprecip_mm"`
	TotalPrecipIn     float64   `json:"totalprecip_in"`
	AvgVisKm          float64   `json:"avgvis_km"`
	AvgVisMiles       float64   `json:"avgvis_miles"`
	AvgHumidity       int       `json:"avghumidity"`
	DailyWillItRain   int       `json:"daily_will_it_rain"`
	DailyChanceOfRain int       `json:"daily_chance_of_rain"`
	DailyWillItSnow   int       `json:"daily_will_it_snow"`
	DailyChanceOfSnow int       `json:"daily_chance_of_snow"`
	Condition         Condition `json:"condition"`
	UV                int       `json:"uv"`
}

type Astro struct {
	Sunrise          string `json:"sunrise"`
	Sunset           string `json:"sunset"`
	Moonrise         string `json:"moonrise"`
	Moonset          string `json:"moonset"`
	MoonPhase        string `json:"moon_phase"`
	MoonIllumination string `json:"moon_illumination"`
}

type Hour struct {
	TimeEpoch    int64     `json:"time_epoch"`
	Time         string    `json:"time"`
	TempC        float64   `json:"temp_c"`
	TempF        float64   `json:"temp_f"`
	IsDay        int       `json:"is_day"`
	Condition    Condition `json:"condition"`
	WindMph      float64   `json:"wind_mph"`
	WindKph      float64   `json:"wind_kph"`
	WindDegree   float64   `json:"wind_degree"`
	WindDir      string    `json:"wind_dir"`
	PressureMb   float64   `json:"pressure_mb"`
	PressureIn   float64   `json:"pressure_in"`
	PrecipMm     float64   `json:"precip_mm"`
	PrecipIn     float64   `json:"precip_in"`
	Humidity     int       `json:"humidity"`
	Cloud        int       `json:"cloud"`
	FeelslikeC   float64   `json:"feelslike_c"`
	FeelslikeF   float64   `json:"feelslike_f"`
	WindchillC   float64   `json:"windchill_c"`
	WindchillF   float64   `json:"windchill_f"`
	HeatindexC   float64   `json:"heatindex_c"`
	HeatindexF   float64   `json:"heatindex_f"`
	DewpointC    float64   `json:"dewpoint_c"`
	DewpointF    float64   `json:"dewpoint_f"`
	WillItRain   int       `json:"will_it_rain"`
	ChanceOfRain int       `json:"chance_of_rain"`
	WillItSnow   int       `json:"will_it_snow"`
	ChanceOfSnow int       `json:"chance_of_snow"`
	VisKm        float64   `json:"vis_km"`
	VisMiles     float64   `json:"vis_miles"`
	GustMph      float64   `json:"gust_mph"`
	GustKph      float64   `json:"gust_kph"`
	UV           int       `json:"uv"`
}

type Alerts struct {
	Alert []Alert `json:"alert"`
}

type Alert struct {
	Headline    string `json:"headline"`
	Msgtype     string `json:"msgtype"`
	Severity    string `json:"severity"`
	Urgency     string `json:"urgency"`
	Areas       string `json:"areas"`
	Category    string `json:"category"`
	Certainty   string `json:"certainty"`
	Event       string `json:"event"`
	Note        string `json:"note"`
	Effective   string `json:"effective"`
	Expires     string `json:"expires"`
	Desc        string `json:"desc"`
	Instruction string `json:"instruction"`
}
