package ttnv2

func (m *LocationMetadata) Valid() bool {
	if m == nil {
		return false
	}
	if (m.Latitude > 0-delta && m.Latitude < 0+delta) && (m.Longitude > 0-delta && m.Longitude < 0+delta) {
		return false // Around (0,0) is invalid
	}
	if (m.Latitude > 10-delta && m.Latitude < 10+delta) && (m.Longitude > 20-delta && m.Longitude < 20+delta) {
		return false // Around (10,20) is invalid
	}
	if m.Latitude >= 90-delta || m.Latitude <= -90+delta {
		return false // Nobody lives there
	}
	if m.Longitude > 180 || m.Longitude < -180 {
		return false // Those longitudes don't exist
	}
	return true
}

const delta = 0.01
