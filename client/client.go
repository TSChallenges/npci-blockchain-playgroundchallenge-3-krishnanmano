package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/hash"
)

type Loan struct {
	LoanID        string    `json:"loanID"`
	ApplicantName string    `json:"applicantName"`
	LoanAmount    float64   `json:"loanAmount"`
	TermMonths    int       `json:"termMonths"`
	InterestRate  float64   `json:"interestRate"`
	Outstanding   float64   `json:"outstanding"`
	Status        string    `json:"status"`
	Repayments    []float64 `json:"repayments"`
}

const (
	channelName   = "mychannel"
	chaincodeName = "loan"
)

var now = time.Now()

func main() {
	clientConnection := newGrpcConnection()
	defer clientConnection.Close()

	id := newIdentity()
	sign := newSign()

	gateway, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithHash(hash.SHA256),
		client.WithClientConnection(clientConnection),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer gateway.Close()

	network := gateway.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)

	loanName := "loan2"

	// TODO: Call ApplyForLoan
	_, err = contract.SubmitTransaction("ApplyForLoan", loanName, "John Doe", "5000", "12", "5.5")
	if err != nil {
		log.Fatalf("Failed to apply for loan: %v", err)
	}
	fmt.Println("Loan successfully applied.")

	_, err = contract.SubmitTransaction("ApproveLoan", loanName, "Approved")
	if err != nil {
		log.Fatalf("Failed to approve loan: %v", err)
	}
	fmt.Println("Loan status updated to Approved.")

	_, err = contract.SubmitTransaction("MakeRepayment", loanName, "1000")
	if err != nil {
		log.Fatalf("Failed to process repayment: %v", err)
	}
	fmt.Println("Repayment recorded. Outstanding balance updated.")

	// TODO: Call CheckLoanBalance
	result, err := contract.EvaluateTransaction("CheckLoanBalance", loanName)
	if err != nil {
		log.Fatalf("Failed to get loan balance: %v", err)
	}

	var loan Loan
	err = json.Unmarshal([]byte(result), &loan)
	if err != nil {
		log.Fatalf("Failed to unmarshal result: %v", err)
	}

	fmt.Printf("Loan details: %v", loan.Outstanding)
}
