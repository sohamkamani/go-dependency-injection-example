package service

import (
	"fmt"
	"github.com/sohamkamani/go-dependency-injection-example/database"
)

type Service struct {
	Store database.Store
}

func (s *Service) GetNumber(ID int) error {
	// Use the `Get` method of the dependency to retreive the value of the database entry
	result, err := s.Store.Get(ID)
	if err != nil {
		return err
	}
	// Perform some validation, and output an error if it is too high
	if result > 10 {
		return fmt.Errorf("result too high: %d", result)
	}
	// Return nil, if the result is valid
	return nil
}

func NewGetNumber(store database.Store) func(int) error {
	return func(ID int) error {
		// Use the `Get` method of the dependency to retreive the value of the database entry
		result, err := store.Get(ID)
		if err != nil {
			return err
		}
		// Perform some validation, and output an error if it is too high
		if result > 10 {
			return fmt.Errorf("result too high: %d", result)
		}
		// Return nil, if the result is valid
		return nil
	}
}