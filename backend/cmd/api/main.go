package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
)

type EmiRequest struct {
	LoanAmount   float64 `json:"loanAmount"`
	InterestRate float64 `json:"interestRate"`
}

type EmiResponse struct {
	EMI float64 `json:"emi"`
}

func calculateEMI(principal float64, annualInterestRate float64, numberOfInstallments int) float64 {
	monthlyInterestRate := annualInterestRate / 12.0 / 100.0
	emi := (principal * monthlyInterestRate * math.Pow(1+monthlyInterestRate, float64(numberOfInstallments))) / (math.Pow(1+monthlyInterestRate, float64(numberOfInstallments)) - 1)
	return emi
}

func CalculateEMIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request EmiRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	fmt.Println(request.InterestRate, request.LoanAmount)

	var wg sync.WaitGroup
	wg.Add(3)

	emiResults := make(map[int]float64)

	go func() {
		defer wg.Done()
		emi := calculateEMI(request.LoanAmount, request.InterestRate, 36)
		emiResults[36] = emi
	}()

	go func() {
		defer wg.Done()
		emi := calculateEMI(request.LoanAmount, request.InterestRate, 60)
		emiResults[60] = emi
	}()

	go func() {
		defer wg.Done()
		emi := calculateEMI(request.LoanAmount, request.InterestRate, 72)
		emiResults[72] = emi
	}()

	wg.Wait()

	response := map[string]float64{
		"emi36": emiResults[36],
		"emi60": emiResults[60],
		"emi72": emiResults[72],
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/", CalculateEMIHandler)
	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
