package usecase_test

import (
	"context"
	"testing"

	"github.com/benderr/gophermart/internal/domain/balance"
	"github.com/benderr/gophermart/internal/domain/balance/usecase"
	"github.com/benderr/gophermart/internal/domain/balance/usecase/mocks"
	mocklogger "github.com/benderr/gophermart/internal/logger/mock_logger"
	mocktransactor "github.com/benderr/gophermart/internal/transactor/mock_transactor"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWithdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBalanceRepo := mocks.NewMockBalanceRepo(ctrl)
	mockWithdrawsRepo := mocks.NewMockWithdrawsRepo(ctrl)
	mockTransactor := mocktransactor.New()
	mockLogger := mocklogger.New()
	balanceUsecase := usecase.New(mockBalanceRepo, mockWithdrawsRepo, mockTransactor, mockLogger)

	t.Run("Withdraw success", func(t *testing.T) {

		userid := "testuserid"
		ordernum := "ordernum"
		var withdraw float64 = 65
		mockBalanceRepo.EXPECT().GetBalanceByUser(gomock.Any(), gomock.Any(), userid).Return(&balance.Balance{
			Current:   100,
			Withdrawn: 20,
		}, nil)

		mockWithdrawsRepo.EXPECT().Create(gomock.Any(), gomock.Any(), userid, ordernum, withdraw).Return(nil)

		mockBalanceRepo.EXPECT().Withdraw(gomock.Any(), gomock.Any(), userid, withdraw).Return(nil)

		err := balanceUsecase.Withdraw(context.Background(), userid, ordernum, withdraw)

		assert.NoError(t, err, "error calling Withdraw")
	})

	t.Run("Withdraw error insufficient funds", func(t *testing.T) {

		userid := "testuserid"
		ordernum := "ordernum"
		var withdraw float64 = 120
		mockBalanceRepo.EXPECT().GetBalanceByUser(gomock.Any(), gomock.Any(), userid).Return(&balance.Balance{
			Current:   100,
			Withdrawn: 20,
		}, nil)

		err := balanceUsecase.Withdraw(context.Background(), userid, ordernum, withdraw)

		if assert.Error(t, err) {
			assert.Equal(t, balance.ErrInsufficientFunds, err)
		}
	})

}
