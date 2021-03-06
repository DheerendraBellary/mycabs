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
	CabID  string `json:"cabid"`
	CityID string `json:"cityid,omitempty"`
}

//DeActivateCabRequest ...
type DeActivateCabRequest struct {
	ID string `json:"id"`
}

//ActivateCabRequest ...
type ActivateCabRequest struct {
	ID string `json:"id"`
}

//ChangeCityRequest ...
type ChangeCityRequest struct {
	CabID  string `json:"cabid"`
	CityID string `json:"cityid"`
}

//DemandCityResonse ...
type DemandCityResonse struct {
	CityID   string `json:"cityid"`
	CityName string `json:"cityname"`
}

//CabHistoryRequest ...
type CabHistoryRequest struct {
	CabID string `json:"cabid"`
}

//CabHistoryResonse ...
type CabHistoryResonse struct {
	History []string `json:"history"`
}

//////////////////////////////////////////////////////////////////////////////////
