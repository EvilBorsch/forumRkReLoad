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
	"time"
)

// Сообщение внутри ветки обсуждения на форуме.
type Post struct {

	// Идентификатор данного сообщения.
	Id int `json:"id,omitempty"`

	// Идентификатор родительского сообщения (0 - корневое сообщение обсуждения).
	Parent *int `json:"parent,omitempty"`

	// Автор, написавший данное сообщение.
	Author string `json:"author"`

	// Собственно сообщение форума.
	Message string `json:"message"`

	// Истина, если данное сообщение было изменено.
	IsEdited bool `json:"isEdited,omitempty"`

	// Идентификатор форума (slug) данного сообещния.
	Forum string `json:"forum,omitempty"`

	// Идентификатор ветви (id) обсуждения данного сообещния.
	Thread int `json:"thread,omitempty"`

	// Дата создания сообщения на форуме.
	Created time.Time `json:"created,omitempty"`

	Parents []uint8
}