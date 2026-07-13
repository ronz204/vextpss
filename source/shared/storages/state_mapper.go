package storages

import "vextpss/source/secrets/core"

func toState(records []MetaRecord) core.State {
	m := make(map[string]string, len(records))
	for _, r := range records {
		m[r.Key] = r.Value
	}
	return core.StateFromMap(m)
}

func fromState(state core.State) []MetaRecord {
	m := state.ToMap()
	records := make([]MetaRecord, 0, len(m))
	for k, v := range m {
		records = append(records, MetaRecord{Key: k, Value: v})
	}
	return records
}
