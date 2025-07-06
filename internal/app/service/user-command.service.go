package service

import (
	"context"
	"encoding/json"
	

	"github.com/arnokay/arnobot-shared/apperror"
	"github.com/arnokay/arnobot-shared/applog"
	"github.com/arnokay/arnobot-shared/data"
	"github.com/arnokay/arnobot-shared/db"
	"github.com/arnokay/arnobot-shared/storage"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
)

type UserCommandService struct {
	cache             jetstream.KeyValue
	store             storage.Storager
	cmdManagerService *CmdManagerService

	logger applog.Logger
}

func NewUserCommandService(
	cache jetstream.KeyValue,
	store storage.Storager,
	commandManager *CmdManagerService,
) *UserCommandService {
	logger := applog.NewServiceLogger("user-command-service")

	return &UserCommandService{
		cache:             cache,
		store:             store,
		cmdManagerService: commandManager,

		logger: logger,
	}
}

func getCommandKVKey(userID uuid.UUID, name string) string {
	return "ucs." + userID.String() + "." + name
}

func (s *UserCommandService) GetOne(ctx context.Context, arg data.UserCommandGetOne) (data.UserCommand, error) {
	if val, err := s.cache.Get(ctx, getCommandKVKey(arg.UserID, arg.Name)); err == nil {
		var userCommand data.UserCommand
		json.Unmarshal(val.Value(), &userCommand)
		return userCommand, nil
	} else {
		s.logger.DebugContext(ctx, "missing cache for get user command, making db call", "err", err)
	}

	fromDB, err := s.store.Query(ctx).CoreUserCommandGetOne(ctx, db.CoreUserCommandGetOneParams{
		UserID: arg.UserID,
		Name:   arg.Name,
	})
	if err != nil {
		return data.UserCommand{}, s.store.HandleErr(ctx, err)
	}

	userCommand := data.NewUserCommandFromDB(fromDB)
	b, _ := json.Marshal(userCommand)

	_, err = s.cache.Put(ctx, getCommandKVKey(userCommand.UserID, userCommand.Name), b)
	if err != nil {
		s.logger.WarnContext(ctx, "cannot cache put user command", "err", err)
	}

	return userCommand, nil
}

func (s *UserCommandService) GetByUserID(ctx context.Context, userID uuid.UUID) ([]data.UserCommand, error) {
	fromDBs, err := s.store.Query(ctx).CoreUserCommandGetByUserID(ctx, userID)
	if err != nil {
		return nil, s.store.HandleErr(ctx, err)
	}

	var userCommands []data.UserCommand

	for _, fromDB := range fromDBs {
		userCommands = append(userCommands, data.NewUserCommandFromDB(fromDB))
	}

	return userCommands, nil
}

func (s *UserCommandService) Create(ctx context.Context, arg data.UserCommandCreate) (data.UserCommand, error) {
	if s.cmdManagerService.IsCommand(arg.Name) {
		return data.UserCommand{}, apperror.New(apperror.CodeInvalidInput, "default command has this name", nil)
	}

	fromDB, err := s.store.Query(ctx).CoreUserCommandCreate(ctx, db.CoreUserCommandCreateParams{
		UserID: arg.UserID,
		Name:   arg.Name,
		Text:   arg.Text,
		Reply:  arg.Reply,
	})
	if err != nil {
		return data.UserCommand{}, s.store.HandleErr(ctx, err)
	}

	userCommand := data.NewUserCommandFromDB(fromDB)
	b, _ := json.Marshal(userCommand)

	_, err = s.cache.Create(ctx, getCommandKVKey(userCommand.UserID, userCommand.Name), b)
	if err != nil {
		s.logger.WarnContext(ctx, "cannot cache create user command", "err", err)
	}

	return userCommand, nil
}

func (s *UserCommandService) Update(ctx context.Context, arg data.UserCommandUpdate) (data.UserCommand, error) {
	if arg.NewName != nil && s.cmdManagerService.IsCommand(*arg.NewName) {
		return data.UserCommand{}, apperror.New(apperror.CodeInvalidInput, "default command has this name", nil)
	}

	fromDB, err := s.store.Query(ctx).CoreUserCommandUpdate(ctx, db.CoreUserCommandUpdateParams{
		UserID:  arg.UserID,
		Name:    arg.Name,
		NewName: arg.NewName,
		Text:    arg.Text,
		Reply:   arg.Reply,
	})
	if err != nil {
		return data.UserCommand{}, s.store.HandleErr(ctx, err)
	}

	userCommand := data.NewUserCommandFromDB(fromDB)
	b, _ := json.Marshal(userCommand)

	_, err = s.cache.Put(ctx, getCommandKVKey(userCommand.UserID, userCommand.Name), b)
	if err != nil {
		s.logger.WarnContext(ctx, "cannot cache updated user command", "err", err)
	}

	return userCommand, nil
}

func (s *UserCommandService) Delete(ctx context.Context, arg data.UserCommandDelete) (data.UserCommand, error) {
	fromDB, err := s.store.Query(ctx).CoreUserCommandDelete(ctx, db.CoreUserCommandDeleteParams{
		UserID: arg.UserID,
		Name:   arg.Name,
	})
	if err != nil {
		return data.UserCommand{}, s.store.HandleErr(ctx, err)
	}

	userCommand := data.NewUserCommandFromDB(fromDB)

	err = s.cache.Purge(ctx, getCommandKVKey(userCommand.UserID, userCommand.Name))
	if err != nil {
		s.logger.WarnContext(ctx, "cannot purge cache user command", "err", err)
	}

	return userCommand, nil
}
