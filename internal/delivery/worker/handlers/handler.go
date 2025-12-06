package handlers

import "github.com/AsaHero/e-wallet/internal/usecase/notifications"

type Handler struct {
	NotificationUsecase *notifications.Module
}
