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
	VehicleId
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

func (m MotoristId) getIdentifiers() spry.Identifiers {
	return spry.Identifiers{"License": m.License, "State": m.State}
}

func (m Motorist) GetIdentifierSet() spry.IdentifierSet {
	return spry.IdentifierSet{
		"Motorist": []spry.Identifiers{m.getIdentifiers()},
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
		VehicleId: vr.VehicleId,
		Type:      vr.Type,
		Make:      vr.Make,
		Model:     vr.Model,
		Color:     vr.Color,
	}
}

func (vr VehicleRegistered) updateVehicle(v *Vehicle) {
	v.Color = vr.Color
	v.Model = vr.Model
	v.Make = vr.Make
	v.Type = vr.Type
	v.VehicleId = vr.VehicleId
}

func (vr VehicleRegistered) GetIdentifierSet() spry.IdentifierSet {
	return spry.IdentifierSet{
		"Motorist": []spry.Identifiers{vr.getIdentifiers()},
		"Vehicle":  []spry.Identifiers{vr.GetIdentifiers()},
	}
}

func (vr VehicleRegistered) Apply(actor any) any {
	switch a := actor.(type) {
	case *Motorist:
		a.MotoristId = vr.MotoristId
		a.Vehicles = append(a.Vehicles, vr.toVehicle())
	case *Vehicle:
		vr.updateVehicle(a)
	}
	return actor
}

type RegisterVehicle struct {
	MotoristId
	VehicleId
	Type  string
	Make  string
	Model string
	Color string
}

func (rv RegisterVehicle) GetIdentifierSet() spry.IdentifierSet {
	return spry.IdentifierSet{
		"Motorist": []spry.Identifiers{rv.getIdentifiers()},
		"Vehicle":  []spry.Identifiers{rv.GetIdentifiers()},
	}
}

func (rv RegisterVehicle) Handle(actor any) ([]spry.Event, []error) {
	switch actor.(type) {
	case Motorist:
		return []spry.Event{
			VehicleRegistered{
				MotoristId: rv.MotoristId,
				VehicleId:  rv.VehicleId,
				EventMetadata: spry.EventMetadata{
					CreatedBy:  "Motorist",
					CreatedFor: "Vehicle",
				},
				Type:  rv.Type,
				Make:  rv.Make,
				Model: rv.Model,
				Color: rv.Color,
			},
		}, nil
	}
	return []spry.Event{}, nil
}
