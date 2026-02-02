package main

type Address struct {
	Street  string
	City    string
	Country string
}

type User struct {
	Name      string
	Age       int
	Addresses []Address
	Metadata  map[string]interface{}
}

func createUser(name string, age int) User {
	return User{
		Name: name,
		Age:  age,
		Addresses: []Address{
			{Street: "123 Main St", City: "NYC", Country: "USA"},
		},
		Metadata: map[string]interface{}{
			"role":   "admin",
			"active": true,
		},
	}
}

func updateUser(u *User, newName string) {
	u.Name = newName
	u.Metadata["updated"] = true
}
