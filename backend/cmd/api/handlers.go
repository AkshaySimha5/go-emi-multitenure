package main

import (
	"encoding/json"
	"math"
	"net/http"
	"sync"
)

type jsonResponse struct {
	Emi36 float64 `json:"emi36"`
	Emi60 float64 `json:"emi60"`
	Emi72 float64 `json:"emi72"`
}
type LoanDetails struct {
	LoanAmount     float64 `json:"loanamount"`
	RateOfInterest float64 `json:"rateofinterest"`
}
type EMIDetails struct {
	Tenure int
	EMI    float64
}

func calculateEMI(wg *sync.WaitGroup, LoanAmount float64, RateOfInterest float64, tenure int, emiChannel chan EMIDetails) {
	defer wg.Done()

	rate := RateOfInterest / 12 / 100
	n := float64(tenure * 12)
	emi := LoanAmount * rate * math.Pow(1+rate, n) / (math.Pow(1+rate, n) - 1)

	emiChannel <- EMIDetails{Tenure: tenure, EMI: emi}
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	var input LoanDetails
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}
	emiChannel := make(chan EMIDetails, 3)
	var wg sync.WaitGroup

	wg.Add(3)
	go calculateEMI(&wg, input.LoanAmount, input.RateOfInterest, 3, emiChannel)
	go calculateEMI(&wg, input.LoanAmount, input.RateOfInterest, 5, emiChannel)
	go calculateEMI(&wg, input.LoanAmount, input.RateOfInterest, 6, emiChannel)
	go func() {
		wg.Wait()
		close(emiChannel)
	}()

	emis := make(map[int]float64)
	for emi := range emiChannel {
		emis[emi.Tenure] = math.Round(emi.EMI)
	}
	payload := jsonResponse{
		Emi36: emis[3],
		Emi60: emis[5],
		Emi72: emis[6],
	}
	out, _ := json.MarshalIndent(payload, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(out)
}
