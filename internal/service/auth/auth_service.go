package auth

import (
	"github.com/ougirez/diplom/internal/pkg/store"
)

type Service struct {
	store *store.Store
}

func NewService(store *store.Store) *Service {
	return &Service{store: store}
}

//func (svc *Service) SignupUser(ctx context.Context, request *domain.SignupUserRequest) (*domain.SignupUserResponse, error) {
//	if _, err := svc.store.GetUserByEmail(ctx, request.Email); !errors.Is(err, constants.ErrDBNotFound) {
//		if err == nil {
//			return nil, constants.ErrEmailAlreadyTaken
//		}
//		return nil, err
//	}
//
//	user := &domain.User{
//		UserName: request.UserName,
//		Email:    domain.NewNullString(request.Email),
//	}
//	if err := user.UserPassword.Init(request.Password); err != nil {
//		return nil, err
//	}
//
//	if err := svc.store.CreateUser(ctx, user); err != nil {
//		return nil, err
//	}
//
//	mainAccount, err := svc.CreateAccount(ctx, &domain.CreateAccountRequest{
//		UserID:   user.ID,
//		Currency: "RUB",
//	})
//	user.MainAccountNumber = mainAccount.Account.Number
//	err = svc.store.UpdateUser(ctx, user)
//	if err != nil {
//		return nil, err
//	}
//
//	authToken, err := utils.GenerateAuthToken(&utils.AuthTokenWrapper{UserID: user.ID})
//	if err != nil {
//		return nil, err
//	}
//
//	return &domain.SignupUserResponse{User: user, AuthToken: authToken}, nil
//}
//
//func (svc *Service) LoginUser(ctx context.Context, request *domain.LoginUserRequest) (*domain.LoginUserResponse, error) {
//	user, err := svc.store.GetUserByEmail(ctx, request.Email)
//	if err != nil {
//		return nil, err
//	}
//
//	if err := user.UserPassword.Validate(request.Password); err != nil {
//		return nil, err
//	}
//
//	logger.Debugf(ctx, "login: userID: [%v]", user.ID)
//
//	authToken, err := utils.GenerateAuthToken(&utils.AuthTokenWrapper{UserID: user.ID})
//	if err != nil {
//		return nil, err
//	}
//
//	return &domain.LoginUserResponse{User: user, AuthToken: authToken}, nil
//}
