package phonebook

type Contact struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Surname    string   `json:"surname"`
	SecondName string   `json:"second_name"`
	Numbers    []string `json:"numbers"`
}
