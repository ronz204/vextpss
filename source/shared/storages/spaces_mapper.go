package storages

import "vextpss/source/secrets/core"

func toSpace(r SpaceRecord) core.Space {
	return core.Space{
		ID:        r.ID,
		Name:      r.Name,
		CreatedAt: r.CreatedAt,
	}
}
