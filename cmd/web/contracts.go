package main

import "abahjoseph.com/snippetbox/pkg/models"

type SnippetsInterface interface {
	Insert(string, string, string) (int, error)
	Get(int) (*models.Snippet, error)
	Latest() ([]*models.Snippet, error)
}

type UsersInterface interface {
	Insert(string, string, string) error
	Authenticate(string, string) (int, error)
	Get(int) (*models.User, error)
}
