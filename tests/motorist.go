package tests

import (
	"github.com/arobson/spry"
	"github.com/arobson/spry/core"
)

type VehicleId struct {
	VIN string
}

func (v VehicleId) GetIdentifiers() spry.Identifiers {
	return spry.Identifiers{"VIN": v.VIN}
}

type Vehicle struct {
	VIN   string
	Type  string
	Make  string
	Model string
	Color string
}

func (v Vehicle) GetIdentifiers() spry.Identifiers {
	return spry.Identifiers{"VIN": v.VIN}
}

type Motorist struct {
	MotoristId
	Name     string
	Vehicles []Vehicle
}

type MotoristId struct {
	License string
	State   string
}

func (m MotoristId) GetIdentifiers() spry.Identifiers {
	return spry.Identifiers{"License": m.License, "State": m.State}
}

func (m Motorist) GetIdentifierSet() spry.IdentifierSet {
	return spry.IdentifierSet{
		"Vehicle": core.Mapper(
			m.Vehicles,
			func(v Vehicle) spry.Identifiers { return v.GetIdentifiers() },
		),
	}
}

type VehicleRegistered struct {
	spry.EventMetadata
	MotoristId
	VehicleId
	Type  string
	Make  string
	Model string
	Color string
}

func (vr VehicleRegistered) toVehicle() Vehicle {
	return Vehicle{
		VIN:   vr.VIN,
		Type:  vr.Type,
		Make:  vr.Make,
		Model: vr.Model,
		Color: vr.Color,
	}
}

func (vr VehicleRegistered) Apply(actor any) any {
	switch a := actor.(type) {
	case Motorist:
		a.Vehicles = append(a.Vehicles, vr.toVehicle())
	}
	return actor
}

type RegisterVehicle struct {
	MotoristId
	VehicleId
	Type         string
	Make         string
	Model        string
	Color        string
	OwnerLicense string
	OwnerState   string
}

func (rv RegisterVehicle) GetIdentifierSet() spry.IdentifierSet {
	return spry.IdentifierSet{
		"Player":  []spry.Identifiers{{"License": rv.License, "State": rv.State}},
		"Vehicle": []spry.Identifiers{{"VIN": rv.VIN}},
	}
}

func (rv RegisterVehicle) Handle(actor any) ([]spry.Event, error) {
	switch actor.(type) {
	case *Motorist:
		return []spry.Event{
			VehicleRegistered{
				EventMetadata: spry.EventMetadata{
					CreatedBy:  "Motorist",
					CreatedFor: "Vehicle",
				},
			},
		}, nil
	}
	return []spry.Event{}, nil
}
