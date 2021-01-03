/*
 * package mycabsapi defines the json apis for http mycabs service.
 */

package mycabsapi

//////////////////////////////////////////////////////////////////////////////////

//City ...
type City struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}

//Cab ...
type Cab struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	City  string `json:"city"`
	State string `json:"state,omitempty"`
}

//OnboardCityRequest ...
type OnboardCityRequest struct {
	Name string `json:"name"`
}

//OnboardCityResponse ...
type OnboardCityResponse struct {
	ID string `json:"id,omitempty"`
}

//RegisterCabRequest ...
type RegisterCabRequest struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	CityID string `json:"cityid"`
}

//RegisterCabResponse ...
type RegisterCabResponse struct {
	ID string `json:"id,omitempty"`
}

//BookingRequest ...
type BookingRequest struct {
	From    string `json:"from"`
	To      string `json:"to"`
	CabType string `json:"cabtype"`
}

//BookingResponse ...
type BookingResponse struct {
	CabID   string `json:"cabid,omitempty"`
	CabName string `json:"cabname,omitempty"`
}

//EndTripRequest ...
type EndTripRequest struct {
	CabID string `json:"cabid"`
}

//DeActivateCabRequest ...
type DeActivateCabRequest struct {
	ID string `json:"id"`
}

//ActivateCabRequest ...
type ActivateCabRequest struct {
	ID string `json:"id"`
}

//////////////////////////////////////////////////////////////////////////////////
