package session

import "github.com/oligoden/chassis/storage/gosql"

type UserRecords []gosql.User

func (UserRecords) TableName() string {
	return "users"
}

func (e UserRecords) Permissions(p ...string) string {
	return ""
}

func (e UserRecords) Owner(o ...uint) uint {
	return 0
}

func (e UserRecords) Users(u ...uint) []uint {
	return []uint{}
}

func (e UserRecords) Groups(g ...uint) []uint {
	return []uint{}
}

func (UserRecords) IDValue(...uint) uint {
	return 0
}

func (e UserRecords) UniqueCode(uc ...string) string {
	return ""
}
