package service

import "github.com/arnokay/arnobot-shared/service"

type Services struct {
	MessageService        *MessageService
	PlatformModuleService *service.PlatformModuleIn
	UserCommandService    *UserCommandService
	CmdManagerService     *CmdManagerService
	UserCmdManagerService *UserCmdManagerService
}
