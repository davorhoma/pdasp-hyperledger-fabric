package services

import "chaincode/trading/models"

func CreateUser(id, firstName, lastName, email string) (*models.User, error) {
	if id == "" || firstName == "" || lastName == "" || email == "" {
		return nil, ErrInvalidInput
	}

	return &models.User{
		DocType:   models.DocTypeUser,
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Invoices:  []string{},
		Balance:   0,
	}, nil
}

func CreateMultipleUsers(usersData []struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
}) ([]*models.User, error) {
	users := make([]*models.User, 0, len(usersData))
	for _, u := range usersData {
		user, err := CreateUser(u.ID, u.FirstName, u.LastName, u.Email)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func DepositToUser(u *models.User, amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	u.Balance += amount
	return nil
}

func WithdrawFromUser(u *models.User, amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	if u.Balance < amount {
		return ErrInsufficientFunds
	}

	u.Balance -= amount
	return nil
}
