/*
 * forum
 *
 * Тестовое задание для реализации проекта \"Форумы\" на курсе по базам данных в Технопарке Mail.ru (https://park.mail.ru).
 *
 * API version: 0.1.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package post

import (
	models2 "go-server-server-generated/src/forum/models"
	"go-server-server-generated/src/thread/models"
	swagger "go-server-server-generated/src/user/models"
)

// Полная информация о сообщении, включая связанные объекты.
type PostFull struct {
	Post Post `json:"post,omitempty"`

	Author *swagger.User `json:"author,omitempty"`

	Thread *models.Thread `json:"thread,omitempty"`

	Forum *models2.Forum `json:"forum,omitempty"`
}
