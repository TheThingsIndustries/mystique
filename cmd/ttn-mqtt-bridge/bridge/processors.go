package bridge

import (
	"strings"

	"github.com/TheThingsIndustries/mystique/pkg/ttnv2"
)

// ProcessUplink processes the uplink for the given gateway.
// Note that the gateway may be nil if it could not be refreshed.
func ProcessUplink(gateway *Gateway, uplink *ttnv2.UplinkMessage) {
	if lorawan := uplink.ProtocolMetadata.GetLoRaWAN(); lorawan != nil {
		lorawan.FCnt = 0
	}
}

// ProcessStatus processes the status for the given gateway.
// Note that the gateway may be nil if it could not be refreshed.
func ProcessStatus(gateway *Gateway, status *ttnv2.StatusMessage) {
	if !status.Location.Valid() {
		status.Location = nil
	}
	if gateway != nil {
		if gateway.AntennaLocation != nil {
			if status.Location == nil {
				status.Location = &ttnv2.LocationMetadata{
					Latitude:  float32(gateway.AntennaLocation.Latitude),
					Longitude: float32(gateway.AntennaLocation.Longitude),
					Source:    ttnv2.LocationMetadata_REGISTRY,
				}
			}
			if status.Location.Altitude == 0 {
				status.Location.Altitude = int32(gateway.AntennaLocation.Altitude)
			}
		}
		if status.FrequencyPlan == "" && gateway.FrequencyPlan != "" {
			status.FrequencyPlan = gateway.FrequencyPlan
		}
		if status.Platform == "" {
			platform := []string{}
			if brand := gateway.Attributes["brand"]; brand != "" {
				platform = append(platform, brand)
			}
			if model := gateway.Attributes["model"]; model != "" {
				platform = append(platform, model)
			}
			status.Platform = strings.Join(platform, " ")
		}
		if status.Description == "" {
			status.Description = gateway.Attributes["description"]
		}
	}
	status.Bridge = "Mystique MQTT Bridge"
}
