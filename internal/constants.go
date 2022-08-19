package internal

import (
	"fmt"
	"time"

	"github.com/4wings/cli/types"
)

const DATA_FOLDER = "data"
const DATASETS_FILE = "datasets.json"
const TEMP_FILES_FILE = "files.json"
const STATUS_IMPORTING = "importing"
const STATUS_COMPLETED = "completed"
const STATUS_ERROR = "error"

var DATASETS_PATH = fmt.Sprintf("./%s/%s", DATA_FOLDER, DATASETS_FILE)
var TEMP_FILES_PATH = fmt.Sprintf("./%s/%s", DATA_FOLDER, TEMP_FILES_FILE)
var GEE_DATASETS = []types.Dataset{{
	ID:          "public-global-water-temperature:v20220801",
	Name:        "Sea surface temperature",
	Description: "Sea surface temperature is the water temperature at the ocean's surface. The Hybrid Coordinate Ocean Model (HYCOM) is a data-assimilative hybrid isopycnal-sigma-pressure (generalized) coordinate ocean model. The subset of HYCOM data hosted in EE contains the variables salinity, temperature, velocity, and elevation. They have been interpolated to a uniform 0.08 degree lat/long grid between 80.48°S and 80.48°N. The salinity, temperature, and velocity variables have been interpolated to 40 standard z-levels. Source: HYCOM",
	Source:      "GEE",
	Type:        "4wings",
	StartDate:   time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
	EndDate:     time.Date(2022, 6, 30, 0, 0, 0, 0, time.UTC),
	Unit:        "ºC",
	Configuration: types.Configuration{
		Images:    []string{"", "HYCOM/sea_temp_salinity", "projects/world-fishing-827/assets/sea_temperature_month", ""},
		Band:      "water_temp_0",
		Min:       -32768,
		Max:       32763,
		Scale:     0.001,
		Offset:    20,
		Intervals: []string{"day", "month"},
	},
}, {
	ID:          "public-global-water-salinity:v20220801",
	Name:        "Sea surface salinity",
	Description: "Sea surface salinity is a key parameter to estimate the influence of oceans on climate. Along with temperature, salinity is a key factor that determines the density of ocean water and thus determines the convection and re-emergence of water masses. The thermohaline circulation crosses all the oceans in surface and at depth, driven by temperature and salinity. A global “conveyor belt” is a simple model of the large-scale thermohaline circulation. Deep-water forms in the North Atlantic, sinks, moves south, circulates around Antarctica, and finally enters the Indian, Pacific, and Atlantic basins. Currents bring cold water masses from north to south and vice versa. This thermohaline circulation greatly influences the formation of sea ice at the world’s poles, and carries ocean food sources and sea life around the planet, as well as affects rainfall patterns, wind patterns, hurricanes and monsoons. Source: EU Copernicus Marine Service Information.",
	Source:      "GEE",
	Type:        "4wings",
	StartDate:   time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
	EndDate:     time.Date(2022, 6, 30, 0, 0, 0, 0, time.UTC),
	Unit:        "ºC",
	Configuration: types.Configuration{
		Images:    []string{"", "HYCOM/sea_temp_salinity", "projects/world-fishing-827/assets/sea_salinity_month", ""},
		Band:      "salinity_0",
		Min:       -32768,
		Max:       32763,
		Scale:     0.001,
		Offset:    20,
		Intervals: []string{"day", "month"},
	},
}, {
	ID:          "public-global-chlorophyl:v20220801",
	Name:        "Chlorophyll-a concentration",
	Description: "Chlorophyll-a is the light-harvesting pigment found in all photosynthetic plants. Its concentration in the ocean is used as an index of phytoplankton biomass and, as such, is a key input to primary productivity models. The moderate resolution imaging spectroradiometer (MODIS) instrument aboard NASA's Terra and Aqua satellites measures ocean color every day, from which global chlorophyll-a concentrations are derived. Ocean phytoplankton chemically fix carbon through photosynthesis, taking in dissolved carbon dioxide and producing oxygen. Through this process, marine plants capture about an equal amount of carbon as does photosynthesis by land vegetation. Changes in the amount of phytoplankton indicate the change in productivity of the oceans and provide a key ocean link for global climate change monitoring. Scientists use chlorophyll in modeling Earth's biogeochemical cycles such as the carbon cycle or the nitrogen cycle. Additionally, on short time scales, chlorophyll can be used to trace oceanographic currents, jets, and plumes. The 1 kilometer resolution and nearly daily global coverage of the MODIS data thus allows scientists to observe mesoscale oceanographic features in coastal and estuarine environments, which are of increasing importance in marine science studies. Source: NASA Earth Observations.",
	Source:      "GEE",
	Type:        "4wings",
	StartDate:   time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
	EndDate:     time.Date(2022, 1, 31, 0, 0, 0, 0, time.UTC),
	Unit:        "mg/m^3",
	Configuration: types.Configuration{
		Images:    []string{"", "NASA/OCEANDATA/MODIS-Aqua/L3SMI", "projects/world-fishing-827/assets/sea_chlorophyl_month", ""},
		Band:      "chlor_a",
		Min:       0,
		Max:       99.99,
		Scale:     1,
		Offset:    0,
		Intervals: []string{"day", "month"},
	},
}, {
	ID:          "public-global-climate-projections-tasmin:v20220801",
	Name:        "Near surface air temperature",
	Description: "Daily mean of the daily-minimum near-surface air temperature",
	Source:      "GEE",
	Type:        "4wings",
	StartDate:   time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
	EndDate:     time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC),
	Unit:        "K",
	Configuration: types.Configuration{
		Images:    []string{"", "", "projects/world-fishing-827/assets/climate_projections_tasmin_month", ""},
		Band:      "tasmin",
		Min:       165.31,
		Max:       318.89,
		Scale:     1,
		Offset:    0,
		Intervals: []string{"month"},
	},
}, {
	ID:          "public-global-climate-projections-pr:v20220801",
	Name:        "Near surface precipitation",
	Description: "Daily mean of precipitation at surface; includes both liquid and solid phases from all types of clouds (both large-scale and convective)",
	Source:      "GEE",
	Type:        "4wings",
	StartDate:   time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
	EndDate:     time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC),
	Unit:        "kg/(m^2*s)",
	Configuration: types.Configuration{
		Images:          []string{"", "projects/world-fishing-827/assets/climate_projections_day", "projects/world-fishing-827/assets/climate_projections_pr_month", ""},
		Band:            "pr",
		Min:             0,
		Max:             0.0016,
		Scale:           1,
		Offset:          0,
		Intervals:       []string{"day", "month"},
		ValueMultiplier: 10000,
	},
}, {
	ID:          "public-global-terra-atmosphere",
	Name:        "Aerosol optical thickness",
	Description: "Aerosol optical thickness at 0.55 microns for both ocean (best) and land (corrected): mean of daily mean",
	Source:      "GEE",
	Type:        "4wings",
	StartDate:   time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
	EndDate:     time.Date(2030, 12, 31, 0, 0, 0, 0, time.UTC),
	Unit:        "kg/(m^2*s)",
	Configuration: types.Configuration{
		Images:    []string{"", "", "MODIS/061/MOD08_M3", ""},
		Band:      "Aerosol_Optical_Depth_Land_Ocean_Mean_Mean",
		Min:       -100,
		Max:       5000,
		Scale:     0.001,
		Offset:    0,
		Intervals: []string{"month"},
	},
}}
var GFW_DATASETS = []types.Dataset{{
	ID:          "public-global-fishing-effort:v20201001",
	Name:        "Apparent Fishing Effort",
	Description: "Global Fishing Watch uses data about a vessel’s identity, type, location, speed, direction and more that is broadcast using the Automatic Identification System (AIS) and collected via satellites and terrestrial receivers. AIS was developed for safety/collision-avoidance. Global Fishing Watch analyzes AIS data collected from vessels that our research has identified as known or possible commercial fishing vessels, and applies a fishing presence algorithm to determine “apparent fishing activity” based on changes in vessel speed and direction. The algorithm classifies each AIS broadcast data point for these vessels as either apparently fishing or not fishing and shows the former on the Global Fishing Watch fishing activity heat map. AIS data as broadcast may vary in completeness, accuracy and quality. Also, data collection by satellite or terrestrial receivers may introduce errors through missing or inaccurate data. Global Fishing Watch’s fishing presence algorithm is a best effort mathematically to identify “apparent fishing activity.” As a result, it is possible that some fishing activity is not identified as such by Global Fishing Watch; conversely, Global Fishing Watch may show apparent fishing activity where fishing is not actually taking place. For these reasons, Global Fishing Watch qualifies designations of vessel fishing activity, including synonyms of the term “fishing activity,” such as “fishing” or “fishing effort,” as “apparent,” rather than certain. Any/all Global Fishing Watch information about “apparent fishing activity” should be considered an estimate and must be relied upon solely at your own risk. Global Fishing Watch is taking steps to make sure fishing activity designations are as accurate as possible. Global Fishing Watch fishing presence algorithms are developed and tested using actual fishing event data collected by observers, combined with expert analysis of vessel movement data resulting in the manual classification of thousands of known fishing events. Global Fishing Watch also collaborates extensively with academic researchers through our research program to share fishing activity classification data and automated classification techniques.",
	Source:      "GFW",
	Type:        "4wings",
	StartDate:   time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC),
	Unit:        "hours",
	Configuration: types.Configuration{
		Intervals: []string{"hour", "day", "10days"},
	},
}}

var TILE_QUERY_PARAMS_V1 = []string{`filters\[\d+\]`, `datasets\[\d+\]`, `vessel-groups\[\d+\]`, "date-range", "proxy", "format", "temporal-aggregation", "interval", "comparison-range", "style"}
