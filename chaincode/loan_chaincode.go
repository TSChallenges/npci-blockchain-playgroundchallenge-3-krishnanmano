package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type LoanContract struct {
	contractapi.Contract
}

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

var loanNotFoundErr = fmt.Errorf("Loan not found")

// TODO: Implement ApplyForLoan
func (c *LoanContract) ApplyForLoan(ctx contractapi.TransactionContextInterface, loanID, applicantName string, loanAmount float64, termMonths int, interestRate float64) error {
	existingLoan, err := c.getLoan(ctx, loanID)
	if err != nil && !errors.Is(err, loanNotFoundErr) {
		return err
	}

	if existingLoan != nil {
		return fmt.Errorf("Loan already exists")
	}

	if loanAmount < 0 ||
		loanID == "" ||
		applicantName == "" ||
		termMonths < 0 ||
		interestRate < 0 {
		return fmt.Errorf("invalid Loan details")
	}

	status := "Pending"
	loan := Loan{
		LoanID:        loanID,
		ApplicantName: applicantName,
		LoanAmount:    loanAmount,
		TermMonths:    termMonths,
		InterestRate:  interestRate,
		Outstanding:   loanAmount,
		Status:        status,
	}

	loanJSON, err := json.Marshal(loan)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(loanID, loanJSON)
	if err != nil {
		return fmt.Errorf("Failed to store data: %v", err)
	}

	return nil
}

// TODO: Implement ApproveLoan
func (c *LoanContract) ApproveLoan(ctx contractapi.TransactionContextInterface, loanID string, status string) error {
	loan, err := c.getLoan(ctx, loanID)
	if err != nil {
		return err
	}

	if loan.Status == "Rejected" {
		return fmt.Errorf("Loan status is Rejected")
	}

	if loan.Status != "Pending" {
		return fmt.Errorf("Loan status is not in pending state")
	}

	loan.Status = status
	loanJSON, err := json.Marshal(loan)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(loanID, loanJSON)
	if err != nil {
		return fmt.Errorf("Failed to store data: %v", err)
	}

	return nil
}

// TODO: Implement MakeRepayment
func (c *LoanContract) MakeRepayment(ctx contractapi.TransactionContextInterface, loanID string, repaymentAmount float64) error {
	if repaymentAmount < 0 {
		return fmt.Errorf("repayment amount must be greater than zero")
	}

	loanBytes, err := ctx.GetStub().GetState(loanID)
	if err != nil {
		return fmt.Errorf("erro while getting Loan details")
	}

	if loanBytes == nil {
		return fmt.Errorf("error while querying loan details for repayment")
	}

	var loan Loan
	err = json.Unmarshal(loanBytes, &loan)
	if err != nil {
		return fmt.Errorf("error while unmarshalling")
	}

	switch loan.Status {
	case "Closed":
		return fmt.Errorf("loan with ID %s is already closed", loanID)
	case "Rejected":
		return fmt.Errorf("loan with ID %s is already rejected", loanID)
	}

	// Validate repayment amount
	if repaymentAmount <= 0 {
		return fmt.Errorf("repayment amount must be greater than zero")
	}

	if repaymentAmount > loan.Outstanding {
		return fmt.Errorf("repayment amount exceeds outstanding balance")
	}

	loan.Outstanding -= repaymentAmount
	loan.Repayments = append(loan.Repayments, repaymentAmount)

	// If the loan is fully repaid, update status
	if loan.Outstanding == 0 {
		loan.Status = "Closed"
	}

	// Serialize the updated loan data
	updatedLoanJSON, err := json.Marshal(loan)
	if err != nil {
		return fmt.Errorf("failed to marshal updated loan data: %v", err)
	}

	// Store updated loan back in the ledger
	err = ctx.GetStub().PutState(loanID, updatedLoanJSON)
	if err != nil {
		return fmt.Errorf("failed to update loan repayment: %v", err)
	}
	return nil
}

func (c *LoanContract) getLoan(ctx contractapi.TransactionContextInterface, loanID string) (*Loan, error) {
	existingLoan, err := ctx.GetStub().GetState(loanID)
	if err != nil {
		return nil, err
	}

	if existingLoan == nil {
		return nil, loanNotFoundErr
	}

	var loan Loan
	err = json.Unmarshal(existingLoan, &loan)
	if err != nil {
		return nil, err
	}

	return &loan, nil
}

func (c *LoanContract) CheckLoanBalance(ctx contractapi.TransactionContextInterface, loanID string) (*Loan, error) {
	return c.getLoan(ctx, loanID)
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(LoanContract))
	if err != nil {
		fmt.Printf("Error creating loan chaincode: %s", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting loan chaincode: %s", err)
	}
}
