package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/domain"
	"github.com/lgustavopalmieri/healing-specialist/internal/modules/specialist/features/create/application/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestCreateSpecialistCommand_Execute_Robust - Versão robusta dos testes
func TestCreateSpecialistCommand_Execute_Robust(t *testing.T) {
	input := CreateSpecialistDTO{
		Name:          "Dr. João Silva",
		Email:         "joao.silva@email.com",
		Phone:         "+5511999999999",
		Specialty:     "Cardiologia",
		LicenseNumber: "CRM123456",
		Description:   "Cardiologista experiente",
		Keywords:      []string{"cardiologia", "coração"},
		AgreedToShare: true,
	}

	t.Run("SUCCESS - All validations pass", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSpecialistCreateRepositoryInterface(ctrl)
		mockGateway := mocks.NewMockSpecialistCreateExternalGatewayInterface(ctrl)
		mockEventPublisher := mocks.NewMockEventDispatcher(ctrl)
		mockTracer := mocks.NewMockTracer(ctrl)
		mockLogger := mocks.NewMockLogger(ctrl)
		mockSpan := mocks.NewMockSpan(ctrl)
		mockApiSpan := mocks.NewMockSpan(ctrl)

		command := NewCreateSpecialistCommand(mockRepo, mockGateway, mockEventPublisher, mockTracer, mockLogger)

		// ORDEM ARQUITETURAL GARANTIDA
		gomock.InOrder(
			mockTracer.EXPECT().Start(gomock.Any(), CreateSpecialistSpanName).Return(context.Background(), mockSpan),
			mockRepo.EXPECT().ValidateUniqueness(gomock.Any(), gomock.Any(), input.Email, input.LicenseNumber).Return(nil),
			mockTracer.EXPECT().Start(gomock.Any(), "ValidateLicenseExternal").Return(context.Background(), mockApiSpan),
			mockGateway.EXPECT().ValidateLicenseNumber(gomock.Any(), input.LicenseNumber).Return(true, nil),
			mockApiSpan.EXPECT().End(),
			mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, specialist *domain.Specialist) (*domain.Specialist, error) {
				return specialist, nil
			}),
			mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Return(nil),
			mockLogger.EXPECT().Info(gomock.Any(), SpecialistCreatedSuccessMessage, gomock.Any(), gomock.Any()),
			mockSpan.EXPECT().End(),
		)

		result, err := command.Execute(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, input.Name, result.Name)
	})

	t.Run("INVALID_LICENSE - Gateway returns false", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSpecialistCreateRepositoryInterface(ctrl)
		mockGateway := mocks.NewMockSpecialistCreateExternalGatewayInterface(ctrl)
		mockEventPublisher := mocks.NewMockEventDispatcher(ctrl)
		mockTracer := mocks.NewMockTracer(ctrl)
		mockLogger := mocks.NewMockLogger(ctrl)
		mockSpan := mocks.NewMockSpan(ctrl)
		mockApiSpan := mocks.NewMockSpan(ctrl)

		command := NewCreateSpecialistCommand(mockRepo, mockGateway, mockEventPublisher, mockTracer, mockLogger)

		gomock.InOrder(
			mockTracer.EXPECT().Start(gomock.Any(), CreateSpecialistSpanName).Return(context.Background(), mockSpan),
			mockRepo.EXPECT().ValidateUniqueness(gomock.Any(), gomock.Any(), input.Email, input.LicenseNumber).Return(nil),
			mockTracer.EXPECT().Start(gomock.Any(), "ValidateLicenseExternal").Return(context.Background(), mockApiSpan),
			mockGateway.EXPECT().ValidateLicenseNumber(gomock.Any(), input.LicenseNumber).Return(false, nil),
			mockLogger.EXPECT().Warn(gomock.Any(), InvalidLicenseNumberMessage, gomock.Any()),
			mockApiSpan.EXPECT().End(),
			mockSpan.EXPECT().End(),
		)

		// GARANTIA: Operações que NÃO devem ocorrer
		mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
		mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)

		result, err := command.Execute(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, domain.ErrInvalidLicense, err)
	})

	t.Run("GATEWAY_ERROR - External service fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSpecialistCreateRepositoryInterface(ctrl)
		mockGateway := mocks.NewMockSpecialistCreateExternalGatewayInterface(ctrl)
		mockEventPublisher := mocks.NewMockEventDispatcher(ctrl)
		mockTracer := mocks.NewMockTracer(ctrl)
		mockLogger := mocks.NewMockLogger(ctrl)
		mockSpan := mocks.NewMockSpan(ctrl)
		mockApiSpan := mocks.NewMockSpan(ctrl)

		command := NewCreateSpecialistCommand(mockRepo, mockGateway, mockEventPublisher, mockTracer, mockLogger)
		gatewayError := errors.New("external service error")

		gomock.InOrder(
			mockTracer.EXPECT().Start(gomock.Any(), CreateSpecialistSpanName).Return(context.Background(), mockSpan),
			mockRepo.EXPECT().ValidateUniqueness(gomock.Any(), gomock.Any(), input.Email, input.LicenseNumber).Return(nil),
			mockTracer.EXPECT().Start(gomock.Any(), "ValidateLicenseExternal").Return(context.Background(), mockApiSpan),
			mockGateway.EXPECT().ValidateLicenseNumber(gomock.Any(), input.LicenseNumber).Return(false, gatewayError),
			mockApiSpan.EXPECT().RecordError(gatewayError),
			mockLogger.EXPECT().Error(gomock.Any(), ErrLicenseValidationMessage, gomock.Any(), gomock.Any()),
			mockApiSpan.EXPECT().End(),
			mockSpan.EXPECT().End(),
		)

		// GARANTIA: Operações que NÃO devem ocorrer
		mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
		mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)

		result, err := command.Execute(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrLicenseValidation, err)
	})

	t.Run("TIMEOUT_GUARANTEED - External validation times out", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSpecialistCreateRepositoryInterface(ctrl)
		mockGateway := mocks.NewMockSpecialistCreateExternalGatewayInterface(ctrl)
		mockEventPublisher := mocks.NewMockEventDispatcher(ctrl)
		mockTracer := mocks.NewMockTracer(ctrl)
		mockLogger := mocks.NewMockLogger(ctrl)
		mockSpan := mocks.NewMockSpan(ctrl)
		mockApiSpan := mocks.NewMockSpan(ctrl)

		command := NewCreateSpecialistCommand(mockRepo, mockGateway, mockEventPublisher, mockTracer, mockLogger)

		// Expectativas básicas
		mockTracer.EXPECT().Start(gomock.Any(), CreateSpecialistSpanName).Return(context.Background(), mockSpan)
		mockRepo.EXPECT().ValidateUniqueness(gomock.Any(), gomock.Any(), input.Email, input.LicenseNumber).Return(nil)
		mockTracer.EXPECT().Start(gomock.Any(), "ValidateLicenseExternal").Return(context.Background(), mockApiSpan)
		mockApiSpan.EXPECT().End()
		mockSpan.EXPECT().End()

		// GARANTIA DE TIMEOUT: Gateway bloqueia até context cancelar
		gatewayExecuted := make(chan struct{})
		mockGateway.EXPECT().ValidateLicenseNumber(gomock.Any(), input.LicenseNumber).DoAndReturn(
			func(ctx context.Context, licenseNumber string) (bool, error) {
				close(gatewayExecuted) // Sinaliza execução
				<-ctx.Done()           // Aguarda timeout
				return false, ctx.Err()
			})

		// GARANTIA: Operações que NÃO devem ocorrer
		mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
		mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)

		// Execute com medição de tempo
		start := time.Now()
		result, err := command.Execute(context.Background(), input)
		duration := time.Since(start)

		// VALIDAÇÕES RIGOROSAS:

		// 1. Gateway DEVE ter executado
		select {
		case <-gatewayExecuted:
			// OK
		default:
			t.Fatal("Gateway nunca foi executado - falso positivo!")
		}

		// 2. Tempo próximo ao timeout (800ms ± margem)
		expectedTimeout := 800 * time.Millisecond
		if duration < expectedTimeout-100*time.Millisecond || duration > expectedTimeout+300*time.Millisecond {
			t.Fatalf("Duração inesperada: %v (esperado ~%v)", duration, expectedTimeout)
		}

		// 3. Erro específico de timeout
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrExternalValidationTimeout, err)
	})

	t.Run("UNIQUENESS_FAILURE - Early return before gateway", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockSpecialistCreateRepositoryInterface(ctrl)
		mockGateway := mocks.NewMockSpecialistCreateExternalGatewayInterface(ctrl)
		mockEventPublisher := mocks.NewMockEventDispatcher(ctrl)
		mockTracer := mocks.NewMockTracer(ctrl)
		mockLogger := mocks.NewMockLogger(ctrl)
		mockSpan := mocks.NewMockSpan(ctrl)

		command := NewCreateSpecialistCommand(mockRepo, mockGateway, mockEventPublisher, mockTracer, mockLogger)
		uniquenessError := errors.New("email already exists")

		gomock.InOrder(
			mockTracer.EXPECT().Start(gomock.Any(), CreateSpecialistSpanName).Return(context.Background(), mockSpan),
			mockRepo.EXPECT().ValidateUniqueness(gomock.Any(), gomock.Any(), input.Email, input.LicenseNumber).Return(uniquenessError),
			mockSpan.EXPECT().RecordError(uniquenessError),
			mockLogger.EXPECT().Error(gomock.Any(), ErrUniquenessValidationMessage, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()),
			mockSpan.EXPECT().End(),
		)

		// GARANTIA: Gateway NÃO deve ser chamado
		mockGateway.EXPECT().ValidateLicenseNumber(gomock.Any(), gomock.Any()).Times(0)
		mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
		mockEventPublisher.EXPECT().Dispatch(gomock.Any(), gomock.Any()).Times(0)

		result, err := command.Execute(context.Background(), input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, uniquenessError, err)
	})
}
