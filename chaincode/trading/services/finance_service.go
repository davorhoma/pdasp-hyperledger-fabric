package services

import "chaincode/trading/models"

func DepositToEntity(entity interface{}, amount float64) error {
	switch e := entity.(type) {
	case *models.User:
		return DepositToUser(e, amount)
	case *models.Merchant:
		return DepositToMerchant(e, amount)
	default:
		return ErrInvalidInput
	}
}
