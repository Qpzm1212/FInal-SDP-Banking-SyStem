package main

import (
	"ESHKEREE/internal/service"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
)

func httpCodeFor(err error) int {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrAccountNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrAccountExists):
		return http.StatusConflict
	case errors.Is(err, service.ErrInsufficientFunds), errors.Is(err, service.ErrCreditLimitExceeded):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

func writeErr(c echo.Context, err error) error {
	return c.JSON(httpCodeFor(err), map[string]string{"error": err.Error()})
}

func parseFloatParam(raw, field string) (float64, error) {
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %q must be a number", service.ErrInvalidInput, field)
	}
	return val, nil
}

// Handles starting the binary either from repo root or cmd/api.
func resolveWebRoot() string {
	for _, dir := range []string{"web", "../../web", "../web"} {
		if info, err := os.Stat(dir + "/index.html"); err == nil && !info.IsDir() {
			return dir
		}
	}
	return "web"
}

func main() {
	e := echo.New()
	bank := service.NewBankFacade()

	webRoot := resolveWebRoot()
	e.Logger.Infof("serving static files from: %s", webRoot)
	e.Static("/", webRoot)

	e.GET("/api/accounts", func(c echo.Context) error {
		return c.JSON(http.StatusOK, bank.ListAccounts())
	})

	e.POST("/api/account", func(c echo.Context) error {
		acc, err := bank.OpenAccount(c.FormValue("type"), c.FormValue("id"))
		if err != nil {
			return writeErr(c, err)
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":      "Account opened",
			"account_id":   acc.ID,
			"account_type": acc.Type,
			"balance":      acc.Balance,
		})
	})

	e.GET("/api/transactions", func(c echo.Context) error {
		return c.JSON(http.StatusOK, bank.ListTransactions(100))
	})

	e.POST("/api/transaction", func(c echo.Context) error {
		amount, err := parseFloatParam(c.FormValue("amount"), "amount")
		if err != nil {
			return writeErr(c, err)
		}
		res, err := bank.Transfer(c.FormValue("account_id"), amount)
		if err != nil {
			return writeErr(c, err)
		}
		status := "Transfer sent"
		if res.Suspicious {
			status = "Transfer sent, but flagged as SUSPICIOUS. The security team has been notified."
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":      status,
			"suspicious":  res.Suspicious,
			"reason":      res.Reason,
			"balance":     res.Balance,
			"transaction": res.Record,
		})
	})

	e.POST("/api/loan", func(c echo.Context) error {
		amount, err := parseFloatParam(c.FormValue("amount"), "amount")
		if err != nil {
			return writeErr(c, err)
		}
		loan, err := bank.ApplyForLoan(c.FormValue("borrower"), amount)
		if err != nil {
			return writeErr(c, err)
		}
		return c.JSON(http.StatusOK, loan)
	})

	e.GET("/api/interest", func(c echo.Context) error {
		balance, err := parseFloatParam(c.QueryParam("balance"), "balance")
		if err != nil {
			return writeErr(c, err)
		}
		interest, err := bank.CalculateYield(c.QueryParam("type"), balance)
		if err != nil {
			return writeErr(c, err)
		}
		return c.JSON(http.StatusOK, map[string]float64{"interest": interest})
	})

	e.Logger.Fatal(e.Start(":8080"))
}
