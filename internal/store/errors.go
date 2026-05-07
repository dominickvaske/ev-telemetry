package store

import "errors"

// ErrVehicleNotFound is returned when a vehicle is not found in the fleetstore.
var ErrVehicleNotFound = errors.New("vehicle not found")
