package models

import "time"

// Ветка обсуждения на форуме.
type Thread struct {

	// Идентификатор ветки обсуждения.
	Id int `json:"id,omitempty"`

	// Заголовок ветки обсуждения.
	Title string `json:"title"`

	// Пользователь, создавший данную тему.
	Author string `json:"author"`

	// Форум, в котором расположена данная ветка обсуждения.
	Forum string `json:"forum,omitempty"`

	// Описание ветки обсуждения.
	Message string `json:"message"`

	// Кол-во голосов непосредственно за данное сообщение форума.
	Votes int `json:"votes,omitempty"`

	// Человекопонятный URL (https://ru.wikipedia.org/wiki/%D0%A1%D0%B5%D0%BC%D0%B0%D0%BD%D1%82%D0%B8%D1%87%D0%B5%D1%81%D0%BA%D0%B8%D0%B9_URL). В данной структуре slug опционален и не может быть числом.
	Slug string `json:"slug,omitempty"`

	// Дата создания ветки на форуме.
	Created time.Time `json:"created,omitempty"`
}
