package cleaning

type List map[string]Record

func (List) TableName() string {
	return "cleaning"
}

func (e List) Permissions(p ...string) string {
	return ""
}

func (e List) Owner(o ...uint) uint {
	return 0
}

func (e List) Users(u ...uint) []uint {
	return []uint{}
}

func (e List) Groups(g ...uint) []uint {
	return []uint{}
}

func (List) IDValue(...uint) uint {
	return 0
}

func (e List) UniqueCode(uc ...string) string {
	return ""
}

func NewList() List {
	return List{}
}

func (List) Complete() error {
	return nil
}

func (List) Hasher() error {
	return nil
}

func (List) Prepare() error {
	return nil
}
