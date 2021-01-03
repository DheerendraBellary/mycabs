package mycabsservice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mycabs/mycabsapi"
	"net/http"
)

//OnboardCityHandler ...
func OnboardCityHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("OnboardCityHandler: Received OnboardCity Request")
	switch method := r.Method; method {
	case http.MethodPost:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			errMsg := fmt.Sprintf("OnboardCityHandler: Request Read Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			return
		}
		req := &mycabsapi.OnboardCityRequest{}
		err = json.Unmarshal(body, req)
		if err != nil {
			errMsg := fmt.Sprintf("OnboardCityHandler: Request Processing Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			return
		}

		err = validateOnboardCityReq(req)
		if err != nil {
			errMsg := fmt.Sprintf("OnboardCityHandler: Request Validation Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			return
		}

		cityID, err := OnboardCity(req)
		if err != nil {
			errMsg := fmt.Sprintf("OnboardCityHandler: OnboardCity Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusInternalServerError, errMsg)
			return
		}

		onboardResp := mycabsapi.OnboardCityResponse{ID: cityID}
		resp, err := json.Marshal(onboardResp)
		if err != nil {
			errMsg := fmt.Sprintf("OnboardCityHandler: Response Building Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusInternalServerError, errMsg)
			return
		}

		fmt.Printf("Citi Onboarded... ID: %v\n", cityID)
		writeResponse(w, resp)

	default:
		errMsg := fmt.Sprintf("OnboardCityHandler: Invalide Request Method. %v\n", method)
		fmt.Printf(errMsg)
		writeErrorResponse(w, http.StatusBadRequest, errMsg)
		return
	}
}

//RegisterCabHandler ...
func RegisterCabHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("RegisterCabHandler: Received RegisterCab Request")
	switch method := r.Method; method {
	case http.MethodPost:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			errMsg := fmt.Sprintf("RegisterCabHandler: Request Read Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			return
		}
		req := &mycabsapi.RegisterCabRequest{}
		err = json.Unmarshal(body, req)
		if err != nil {
			errMsg := fmt.Sprintf("RegisterCabHandler: Request Processing Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			return
		}

		err = validateRegisterCabReq(req)
		if err != nil {
			errMsg := fmt.Sprintf("RegisterCabHandler: Request Validation Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			return
		}

		cabID, err := RegisterCab(req)
		if err != nil {
			errMsg := fmt.Sprintf("RegisterCabHandler: RegisterCab Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusInternalServerError, errMsg)
			return
		}

		onboardResp := mycabsapi.OnboardCityResponse{ID: cabID}
		resp, err := json.Marshal(onboardResp)
		if err != nil {
			errMsg := fmt.Sprintf("RegisterCabHandler: Response Building Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusInternalServerError, errMsg)
			return
		}

		fmt.Printf("Cab Registered.... ID: %v\n", cabID)
		writeResponse(w, resp)

	default:
		errMsg := fmt.Sprintf("RegisterCabHandler: Invalide Request Method. %v\n", method)
		fmt.Printf(errMsg)
		writeErrorResponse(w, http.StatusBadRequest, errMsg)
		return
	}
}

//BookCabHandler ...
func BookCabHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("BookCabHandler: Received BookCab Request")
	switch method := r.Method; method {
	case http.MethodPost:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			errMsg := fmt.Sprintf("BookCabHandler: Request Read Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			return
		}
		req := &mycabsapi.BookingRequest{}
		err = json.Unmarshal(body, req)
		if err != nil {
			errMsg := fmt.Sprintf("BookCabHandler: Request Processing Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			return
		}

		err = validateBookingReq(req)
		if err != nil {
			errMsg := fmt.Sprintf("BookCabHandler: Request Validation Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			return
		}

		cab, err := BookCab(req)
		if err != nil {
			errMsg := fmt.Sprintf("BookCabHandler: RegisterCab Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusInternalServerError, errMsg)
			return
		}

		if cab == nil {
			fmt.Printf("BookCabHandler: No cabs were found\n")
			writeResponse(w, []byte{})
			return
		}

		bookingResp := mycabsapi.BookingResponse{
			CabID:   cab.ID,
			CabName: cab.Name,
		}
		resp, err := json.Marshal(bookingResp)
		if err != nil {
			errMsg := fmt.Sprintf("BookCabHandler: Response Building Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusInternalServerError, errMsg)
			return
		}

		fmt.Printf("Cab Booked... ID: %v\n", bookingResp)
		writeResponse(w, resp)

	default:
		errMsg := fmt.Sprintf("BookCabHandler: Invalide Request Method. %v\n", method)
		fmt.Printf(errMsg)
		writeErrorResponse(w, http.StatusBadRequest, errMsg)
		return
	}
}

//EndTripHandler ...
func EndTripHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("EndTripHandler: Received EndTrip Request")
	switch method := r.Method; method {
	case http.MethodPost:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			errMsg := fmt.Sprintf("EndTripHandler: Request Read Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			return
		}
		req := &mycabsapi.EndTripRequest{}
		err = json.Unmarshal(body, req)
		if err != nil {
			errMsg := fmt.Sprintf("EndTripHandler: Request Processing Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			return
		}

		err = validateEndTripReq(req)
		if err != nil {
			errMsg := fmt.Sprintf("EndTripHandler: Request Validation Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			return
		}

		err = EndTrip(req)
		if err != nil {
			errMsg := fmt.Sprintf("EndTripHandler: EndTrip Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusInternalServerError, errMsg)
			return
		}

		fmt.Printf("Trip Ended.... ID: %v\n", req.CabID)
		writeResponse(w, []byte{})

	default:
		errMsg := fmt.Sprintf("EndTripHandler: Invalide Request Method. %v\n", method)
		fmt.Printf(errMsg)
		writeErrorResponse(w, http.StatusBadRequest, errMsg)
		return
	}
}

//DeActivateCabHandler ...
func DeActivateCabHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DeActivateCabHandler: Received DeActivateCab Request")
	switch method := r.Method; method {
	case http.MethodPost:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			errMsg := fmt.Sprintf("DeActivateCabHandler: Request Read Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			return
		}
		req := &mycabsapi.DeActivateCabRequest{}
		err = json.Unmarshal(body, req)
		if err != nil {
			errMsg := fmt.Sprintf("DeActivateCabHandler: Request Processing Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			return
		}

		err = validateDeActivateCabReq(req)
		if err != nil {
			errMsg := fmt.Sprintf("DeActivateCabHandler: Request Validation Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			return
		}

		err = DeActivateCab(req)
		if err != nil {
			errMsg := fmt.Sprintf("DeActivateCabHandler: DeActivateCab Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusInternalServerError, errMsg)
			return
		}

		fmt.Printf("Decativated.... ID: %v\n", req.ID)
		writeResponse(w, []byte{})

	default:
		errMsg := fmt.Sprintf("DeActivateCabHandler: Invalide Request Method. %v\n", method)
		fmt.Printf(errMsg)
		writeErrorResponse(w, http.StatusBadRequest, errMsg)
		return
	}
}

//ActivateCabHandler ...
func ActivateCabHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ActivateCabHandler: Received ActivateCab Request")
	switch method := r.Method; method {
	case http.MethodPost:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			errMsg := fmt.Sprintf("ActivateCabHandler: Request Read Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			return
		}
		req := &mycabsapi.ActivateCabRequest{}
		err = json.Unmarshal(body, req)
		if err != nil {
			errMsg := fmt.Sprintf("ActivateCabHandler: Request Processing Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			return
		}

		err = validateActivateCabReq(req)
		if err != nil {
			errMsg := fmt.Sprintf("ActivateCabHandler: Request Validation Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusBadRequest, errMsg)
			return
		}

		err = ActivateCab(req)
		if err != nil {
			errMsg := fmt.Sprintf("ActivateCabHandler: DeActivateCab Failed. Err: %v\n", err)
			fmt.Printf(errMsg)
			writeErrorResponse(w, http.StatusInternalServerError, errMsg)
			return
		}

		fmt.Printf("Activated .... ID: %v\n", req.ID)
		writeResponse(w, []byte{})

	default:
		errMsg := fmt.Sprintf("ActivateCabHandler: Invalide Request Method. %v\n", method)
		fmt.Printf(errMsg)
		writeErrorResponse(w, http.StatusBadRequest, errMsg)
		return
	}
}
