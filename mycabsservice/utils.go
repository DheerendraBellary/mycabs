package mycabsservice

import (
	"errors"
	"mycabs/mycabsapi"
	"net/http"
)

/////////////////-------------Some Utility Functions-------------////////////

func writeResponse(w http.ResponseWriter, jsonResp []byte) {
	w.Header().Add("content-type", "application/json")
	w.Header().Add("charset", "utf-8")
	w.Write(jsonResp)
}

func writeErrorResponse(w http.ResponseWriter, httpStatus int, errMsg string) {
	w.Header().Add("errormsg", errMsg)
	w.WriteHeader(httpStatus)

}

//validateOnboardCityReq ...
func validateOnboardCityReq(req *mycabsapi.OnboardCityRequest) error {
	if req.Name == "" {
		return errors.New("validateOnboardCityReq: Name Cannot be Empty")
	}
	return nil
}

//validateRegisterCabReq ...
func validateRegisterCabReq(req *mycabsapi.RegisterCabRequest) error {
	if req.Name == "" || req.Type == "" || req.CityID == "" {
		return errors.New("validateRegisterCabReq: Name/Type/CityID Cannot be Empty")
	}

	//Improvements TODO: Validate for proper cityId and type

	return nil
}

//validateRegisterCabReq ...
func validateBookingReq(req *mycabsapi.BookingRequest) error {
	if req.From == "" || req.To == "" || req.CabType == "" {
		return errors.New("validateBookingReq: From/To/Type cannot be Empty")
	}

	//Improvements TODO: Validate for proper cityId in From and To fileds
	//Improvements TODO: Validate for proper Type

	return nil
}

//validateEndTripReq ...
func validateEndTripReq(req *mycabsapi.EndTripRequest) error {
	if req.CabID == "" {
		return errors.New("validateEndTripReq: CabID be Empty")
	}
	return nil
}

//validateEndTripReq ...
func validateDeActivateCabReq(req *mycabsapi.DeActivateCabRequest) error {
	if req.ID == "" {
		return errors.New("validateDeActivateCabReq: ID be Empty")
	}
	return nil
}

//validateEndTripReq ...
func validateActivateCabReq(req *mycabsapi.ActivateCabRequest) error {
	if req.ID == "" {
		return errors.New("validateActivateCabReq: ID be Empty")
	}
	return nil
}
