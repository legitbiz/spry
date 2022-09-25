package tests

import (
	"github.com/arobson/spry"
	"github.com/arobson/spry/core"
)

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
	License  string
	State    string
	Name     string
	Vehicles []Vehicle
}

func (m Motorist) GetIdentifiers() spry.Identifiers {
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
	VIN   string
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
	VIN          string
	Type         string
	Make         string
	Model        string
	Color        string
	OwnerLicense string
	OwnerState   string
}

func (rv RegisterVehicle) Handle(actor any) ([]spry.Event, error) {
	switch a := actor.(type) {
	case *Motorist:
		v := Vehicle{
			VIN:   rv.VIN,
			Type:  rv.Type,
			Make:  rv.Make,
			Model: rv.Model,
			Color: rv.Color,
		}
		a.Vehicles = append(a.Vehicles, v)
	}
	return []spry.Event{}, nil
}
